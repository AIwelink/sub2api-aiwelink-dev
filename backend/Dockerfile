FROM golang:1.26.5-alpine

WORKDIR /app

RUN addgroup -g 1000 sub2api && \
    adduser -u 1000 -G sub2api -s /bin/sh -D sub2api

# 安装必要的工具
RUN apk add --no-cache git

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN VERSION_VALUE="$(./scripts/resolve-version.sh)" && \
    UPSTREAM_VERSION_VALUE="$(tr -d '\r\n' < ./cmd/server/UPSTREAM_VERSION)" && \
    go build -ldflags="-s -w -X main.Version=${VERSION_VALUE} -X main.UpstreamVersion=${UPSTREAM_VERSION_VALUE}" -o main ./cmd/server/

RUN mkdir -p /app/data && chown sub2api:sub2api /app/data

# 暴露端口
EXPOSE 8080

USER sub2api

# 运行应用
CMD ["./main"]
