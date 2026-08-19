# 前端构建阶段：使用官方 Node 镜像，保证 node/npm 版本与 lockfile 兼容。
FROM node:20-bookworm-slim AS frontend-build
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# 评测镜像保留 Go 与 Node 两套工具链，方便模型在容器内继续操作源码。
FROM golang:1.26
WORKDIR /app

# node:20 官方镜像将 node/npm 安装在 /usr/local；复制它以保留前端工具链，
# 同时复用上一个阶段已经构建完成的静态资源。
COPY --from=frontend-build /usr/local /usr/local
COPY . .

# 后端是独立 Go module，且通过 go:embed 引用构建后的前端资源。将 dist 放到
# embed 指令期望的位置后，再从 backend 目录执行 Go 编译。
COPY --from=frontend-build /app/frontend/dist ./backend/internal/handler/dist
RUN node --version && npm --version \
    && cd backend && go build ./...

# 容器启动后进入 shell，方便模型操作。
CMD ["bash"]
