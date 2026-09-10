# ---- 阶段 1：构建 ----
# modernc.org/sqlite 与 aws-sdk-go-v2 都是纯 Go，无需 CGO，用 alpine 即可产出静态二进制
FROM golang:1.27-alpine AS builder

WORKDIR /src

# 先单独拷贝依赖清单并下载，利用 Docker 层缓存（改源码不会触发重新拉依赖）
COPY go.mod go.sum ./
RUN go mod download

# 再拷贝全部源码
COPY . .

# 编译 blog 服务。-s -w 去掉调试符号减小体积；CGO 关闭保证静态链接
# （modernc sqlite / aws-sdk 均为纯 Go，无需 cgo）
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/blog ./blog

# 编译迁移工具，容器启动时先用它建表
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./tools/migrate

# ---- 阶段 2：运行 ----
FROM alpine:3.20

WORKDIR /app/blog

# 只需二进制、配置和迁移脚本
COPY --from=builder /out/blog ./blog
COPY --from=builder /out/migrate ./migrate
COPY blog/etc ./etc
COPY migrations ./migrations

# RustFS / S3 走 HTTPS：从 builder 阶段（golang:alpine 自带 CA 证书）拷贝过来，
# 避免 apk add 联网。Go 的 TLS 默认读 /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

# 数据与日志目录（运行时用 volume 挂载，不要写进镜像）
RUN mkdir -p /app/data /app/logs

EXPOSE 8812

# 配置文件是相对路径 etc/blog-api.yaml，二进制在 /app/blog 下运行，
# 因此 SqliteDSN 的 ../data -> /app/data，日志 ../logs -> /app/logs。
# 启动前先跑迁移（幂等），保证空 volume 也能自动建表。
CMD ["sh", "-c", "./migrate && exec ./blog -f etc/blog-api.yaml"]
