# syntax=docker/dockerfile:1

# ---- 构建阶段 ----
# qtcloud-crowd provider 无 CGO 依赖（纯 Go 标准库，无外部模块），静态链接产物，
# 可直接运行于 alpine（无需 gcc/musl-dev 工具链）。对齐 qtcloud-execute provider 模板。
FROM golang:1.26-alpine AS build

WORKDIR /src

# 依赖代理：默认国内可达的 goproxy.cn（proxy.golang.org 在部分网络环境不可达）
# 覆盖：docker build --build-arg GOPROXY=https://proxy.golang.org,direct .
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

# 仅 go.mod（无 go.sum——纯标准库无第三方依赖）；先拷贝模块文件，利用构建缓存
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/provider-server ./cmd/server

# ---- 运行阶段 ----
# 静态二进制，与 alpine 兼容（musl/libc 均可直跑）
FROM alpine:3.20

RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/provider-server /usr/local/bin/provider-server

USER app
EXPOSE 8080

# 监听地址由 QTCLOUD_CROWD_ADDR 控制（默认 :8080，见 cmd/server/main.go）
ENTRYPOINT ["provider-server"]
