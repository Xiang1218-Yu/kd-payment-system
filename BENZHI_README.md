# kd-payment-system

项目用途：为社区快递柜运营商（如丰巢）设计的格口动态定价与负载调度系统。解决不同时段、不同区域格口利用率不均的问题：根据历史取件数据与实时库存，动态调整格口"临时租赁价格"（如晚高峰加价），并引导快递员将包裹投递至利用率较低的邻近柜机，以平衡负载、提升整体周转率。项目源代码、依赖描述和评测专用 Docker 文件共同构成自包含任务；不依赖本机预编译二进制。

## 标准构建、运行和测试命令

后端位于 `backend/`，并通过 `go:embed` 将前端产物嵌入二进制。首次本地编译前先构建前端并复制产物：

```bash
cd frontend
npm install
npm run build
cd ..
mkdir -p backend/internal/handler/dist
cp -a frontend/dist/. backend/internal/handler/dist/

cd backend
go build ./...
go run ./cmd/server
go test ./...
```

## 前端标准命令

```bash
cd frontend
npm install
npm run build
```
## 评测容器

评测专用 Dockerfile 为 `benzhi.Dockerfile`，构建脚本为 `build_benzhi_docker.sh`。

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh my-go-task linux/arm64
./build_benzhi_docker.sh my-go-task linux/amd64
docker run -it my-go-task:latest
```
