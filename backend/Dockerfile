# syntax=docker/dockerfile:1
#
# Multi-stage build for the locker pricing & scheduling system.
#   Stage 1 (frontend): build the Vite+React SPA. By default the frontend is
#     built inside the container from source (requires network access to npm).
#     If you are behind a restricted network or simply want to reuse a locally
#     built dist, pass --build-arg BUILD_FRONTEND=false and the build will
#     COPY the already-built frontend/dist from the build context instead.
#   Stage 2 (backend):  compile the Go binary, embedding the frontend dist.
#   Stage 3 (runtime):   minimal alpine image holding only the binary.
#
# Build context is the repository root so both frontend/ and backend/ are
# reachable. Build with:
#   docker build -f backend/Dockerfile -t kd-payment-system .
# or, no-network / reuse local build:
#   npm run build --prefix frontend
#   docker build -f backend/Dockerfile --build-arg BUILD_FRONTEND=false -t kd-payment-system .

ARG BUILD_FRONTEND=true

# ---------- Stage 1a: frontend (build from source, needs network) ----------
FROM node:20-alpine AS frontend-true
ARG BUILD_FRONTEND
WORKDIR /app/frontend
# Install deps and build from source. Cached by package.json unless it changes.
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ---------- Stage 1b: frontend (prebuilt, no network) ----------
FROM alpine:3.20 AS frontend-false
# When BUILD_FRONTEND=false, just copy the locally-built dist from the build
# context. Run `npm run build` in /frontend first so dist/ exists.
COPY frontend/dist /app/frontend/dist

# ---------- Stage 1 selector ----------
# Resolves to frontend-true (build in-container) or frontend-false (reuse
# local dist) depending on the BUILD_FRONTEND arg.
FROM frontend-${BUILD_FRONTEND} AS frontend-assets

# ---------- Stage 2: backend ----------
FROM golang:1.26-alpine AS backend
WORKDIR /app/backend
COPY backend/go.mod ./
COPY backend/ ./
# Embed the frontend dist (built or prebuilt) where the go:embed directive
# expects it.
COPY --from=frontend-assets /app/frontend/dist ./internal/handler/dist
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /kd-server ./cmd/server

# ---------- Stage 3: runtime ----------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 10001 app
WORKDIR /app
COPY --from=backend /kd-server /app/kd-server
USER app
EXPOSE 3600
ENV ADDR=:3600
ENTRYPOINT ["/app/kd-server"]
