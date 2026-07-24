# ---- Build Stage ----
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# ---- Runtime Stage ----
FROM alpine:latest

WORKDIR /app

# 安装 CA 证书（HTTPS 请求需要）
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
