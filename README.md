# 智能快递柜格口动态定价与调度系统

为社区快递柜运营商（如丰巢）设计的格口动态定价与负载调度系统。解决不同时段、不同区域格口利用率不均的问题：根据历史取件数据与实时库存，动态调整格口"临时租赁价格"（如晚高峰加价），并引导快递员将包裹投递至利用率较低的邻近柜机，以平衡负载、提升整体周转率。

包含完整的 **Go 后端** + **React(JSX) 前端**，代码遵循 **单一职责原则** 分层，**Docker 一键启动**。

---

## 核心能力

### 1. 动态定价

报价 = **基准价 × 时段系数 × 利用率系数 × 稀缺系数**

| 因子 | 取值 | 说明 |
|---|---|---|
| 基准价 | S ¥2 / M ¥3 / L ¥5 | 按格口尺寸 |
| 时段系数 | 0.6 ~ 1.4 | 晚高峰(17-21点) ×1.4、早高峰 ×1.1、深夜 ×0.6 促销 |
| 利用率系数 | 0.8 ~ 1.5 | <50% 空闲促销 ×0.8；>90% 爆满加价 ×1.5 |
| 稀缺系数 | 1.0 / 1.25 | 剩余格口 ≤3 时叠加 ×1.25 |

每次报价都返回完整的**构成明细**，前端逐步展示"基准价 → ×时段 → ×利用率 → ×稀缺 → 最终价"，让定价过程透明可解释。

### 2. 负载调度

快递员请求投递到某柜机时：

- 目标柜机有空闲格口且利用率 < 85% → **直接投递**。
- 否则 → 在**同区域**内按 Haversine 距离找邻近柜机，按 `(利用率升序, 距离升序)` 排序，推荐利用率最低的邻近柜机，并返回 Top-3 替代方案（含距离、利用率、报价）。
- 不会跨区域调度。

### 3. 时间模拟

右上角"模拟时钟"可推进时间，观察晚高峰加价系数实时生效——演示动态定价随时间的变化。

---

## 架构（单一职责分层）

```
依赖方向:  handler → service → {pricing, scheduler, store} → model
                                   ↑
                                 clock（时间抽象，可注入）
```

| 包 | 职责 | 关键文件 |
|---|---|---|
| `model` | 纯领域数据结构，无逻辑 | `locker.go` `parcel.go` `pricing.go` |
| `clock` | 时钟抽象（真实/可注入），供定价按时段计算 | `clock.go` |
| `store` | 内存仓储 + 种子数据 + 并发安全(RWMutex) | `store.go` `mutations.go` `seed.go` |
| `pricing` | 动态定价算法（纯函数，可单测） | `engine.go` `policy.go` |
| `scheduler` | 邻近柜机负载均衡（Haversine + 排序） | `scheduler.go` |
| `service` | 用例编排，组合 store/pricing/scheduler | `pricing_svc.go` `dropoff_svc.go` `pickup_svc.go` `stats_svc.go` `sim_svc.go` |
| `handler` | HTTP 协议适配（仅解码/调用/编码） | `router.go` `*_handler.go` |
| `cmd/server` | 入口，组装依赖 | `main.go` |

**SRP 落点**：`handler` 不触碰 `store`/`pricing`；`pricing` 不依赖 `http`；`scheduler` 依赖 `CabinetProvider` 接口而非 `store` 包，依赖箭头单向内聚。

前端同样遵循 SRP：`api/client.js` 只管网络、`pages/*` 各管一个用例、`components/*` 是可复用视觉原语。

---

## API 一览

| Method | Path | 作用 |
|---|---|---|
| GET | `/api/regions` | 区域列表与聚合利用率 |
| GET | `/api/cabinets?region=` | 柜机列表（含各尺寸可用数） |
| GET | `/api/cabinets/:id` | 柜机详情（全格口 + 各尺寸统计） |
| GET | `/api/pricing/:cabinetId?size=S|M|L` | 实时报价 + 构成明细 |
| POST | `/api/dropoff` | 投递：`{cabinetId, size}` → 调度结果 + 报价 + 占用格口 |
| POST | `/api/pickup` | 取件：`{lockerId}` → 释放格口、写历史 |
| GET | `/api/stats/dashboard` | 看板：区域排行/最拥挤最空闲柜机/时段取件分布/平均停留 |
| GET | `/api/sim/state` | 读取当前模拟时钟 |
| POST | `/api/sim/tick` | 推进时钟：`{"duration":"1h"}` |
| POST | `/api/sim/reset` | 重置时钟到真实时间 |

非 `/api/*` 路径回退到内嵌的 SPA `index.html`，支持前端路由刷新。

---

## 快速开始

### 方式一：Docker（推荐）

```bash
docker compose up --build
```

浏览器打开 <http://localhost:8080>。

也可单独 build：

```bash
docker build -f backend/Dockerfile -t kd-payment-system .
docker run -p 8080:8080 kd-payment-system
```

> Docker 多阶段构建：`node:20` 编译前端 → `golang:1.26` 编译后端并用 `embed.FS` 内嵌静态资源 → `alpine` 只装二进制。最终镜像仅约 26MB，单一可执行文件，零外部依赖。
>
> **内网/受限网络**：容器内 `npm ci` 需访问 npm 仓库。若网络不通，可先用本地构建前端再复用：
> ```bash
> npm --prefix frontend run build
> BUILD_FRONTEND=false docker compose up --build
> ```

### 方式二：本地开发

后端：

```bash
cd backend
go run ./cmd/server        # :8080
```

前端热更新（开发态，代理 `/api` 到后端）：

```bash
cd frontend
npm install
npm run dev                # :5173，浏览器开 http://localhost:5173
```

> 生产构建时前端需先 `npm run build` 并将 `frontend/dist/*` 复制到 `backend/internal/handler/dist/`，再 `go build`，这样 `embed.FS` 才能内嵌。

---

## 演示流程

1. **总览看板**：看到 3 个区域利用率差异——中央商务区最高（红色）、居民区B最低（绿色）。
2. **动态定价**：选一台爆满柜机，看报价明细（晚高峰 ×1.4、爆满 ×1.5、稀缺 ×1.25 叠加）。点右上角"→晚高峰"推进时钟，报价系数从 1.0 跳到 1.4。
3. **投递调度**：选爆满柜机发起投递，系统自动重定向到利用率更低的邻近柜机，显示距离与 Top-3 替代方案，并真实占用一个格口。
4. **柜机状态**：查看格口网格，点击已占用格口模拟取件，释放格口并写入历史记录。

---

## 测试

```bash
cd backend
go test ./internal/pricing/... ./internal/scheduler/...
```

- `pricing`：验证晚高峰/早高峰/深夜系数、利用率分档、稀缺叠加、明细链终点等于最终价。
- `scheduler`：验证满载柜机被重定向到利用率最低的邻近柜机、有空闲时直接确认、未知柜机报错、不跨区域调度、Haversine 距离。

---

## 技术栈

- **后端**：Go 1.26，标准库 `net/http`（Go 1.22+ 路由方法语法）、`embed`、`sync`。无第三方依赖。
- **前端**：React 18 + React Router 6 + Vite 5（JSX）。
- **数据**：内存存储 + 启动种子数据（3 区域 / 15 柜机 / ~450 格口 / 7 天约 3600 条取件历史，带晚高峰分布）。

---

## 设计说明

- **内存存储**：零外部依赖，重启重置。适合演示与原型；如需持久化可替换 `store` 实现而不动上层。
- **时钟抽象**：定价引擎不直接读 `time.Now()`，而是接收时间参数；`SimService` 通过 `clock.Manual` 推进时间，让"晚高峰加价"在演示中可见。
- **可解释定价**：返回 `Breakdown` 步骤链，前端逐步渲染，体现"为什么这个价"。
- **调度不跨区**：邻近推荐限制在同区域，符合快递员投递半径的现实约束。
