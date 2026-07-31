# AIWeLink Sub2API 测试环境 Docker 部署设计

**状态：** 已确认
**日期：** 2026-08-01

## 1. 目标

在服务器上直接拉取已验证的 Sub2API 开发镜像，连接现有远程测试 PostgreSQL 和 Redis，并把登录事件投递到实际 Traffic Analysis 接口。部署不在服务器构建源码，也不启动本地数据库或 Redis。

本配置只用于测试环境，不得连接生产数据库。

## 2. 交付文件

- `deploy/docker-compose.aiwelink-test.yml`：可提交的服务器 Compose 文件；
- `deploy/.env.aiwelink-test.example`：可提交的脱敏变量模板；
- `deploy/.env.aiwelink-test`：包含真实测试凭据的服务器配置，必须被 Git 忽略并设置为仅部署用户可读。

## 3. 容器设计

- 默认镜像：`docker.aiwelink.cc/sub2api-aiwelink-dev:dev-0c4864e109`；
- 容器名：`sub2api-aiwelink-test`；
- 主机端口：`0.0.0.0:8081`；
- 容器端口：`8080`；
- 重启策略：`unless-stopped`；
- 数据目录：命名卷挂载到 `/app/data`；
- 网络：加入已有外部网络 `1panel-network`；
- 文件句柄限制：soft/hard 均为 `100000`；
- Compose 只拉镜像，不包含 `build`、PostgreSQL 或 Redis service。

由于端口明确绑定到 `0.0.0.0`，服务器防火墙和 1Panel/OpenResty 必须限制不需要的公网访问。

## 4. 环境变量契约

真实 env 从现有 `deploy/.env.develop` 复制以下测试环境值，不在文档或 Git 中记录原值：

- PostgreSQL：`DATABASE_HOST`、`DATABASE_PORT`、`DATABASE_DBNAME`、`DATABASE_USER`、`DATABASE_PASSWORD`、`DATABASE_SSLMODE`；
- Redis：`REDIS_HOST`、`REDIS_PORT`、`REDIS_PASSWORD`、`REDIS_DB`；
- 应用密钥：`JWT_SECRET`、`TOTP_ENCRYPTION_KEY`；
- Growth：`GROWTH_LOGIN_ENABLED`、`GROWTH_LOGIN_ENDPOINT`、`GROWTH_SITE_ID`、`GROWTH_SERVICE_CREDENTIAL`、`GROWTH_OUTBOX_ENCRYPTION_KEY`、连接和读取超时；
- 运行参数：`BIND_HOST=0.0.0.0`、`SERVER_PORT=8081`、`SERVER_MODE=release`、`RUN_MODE=standard`、`TZ=Asia/Shanghai`。

Growth 登录地址使用 `https://aiwelink.cc/internal/growth/logins`。服务凭据、数据库密码、Redis 密码、JWT/TOTP 密钥和 Outbox 加密密钥不得输出到部署日志。

## 5. 初始化和健康检查

- `AUTO_SETUP=true`；
- `SETUP_MIGRATION_TIMEOUT_SECONDS=600`；
- 健康检查必须访问容器内部 `http://localhost:8080/health`，不能访问主机端口 `8081`；
- `start_period=240s`，为远程数据库初始化和后台服务加载预留时间；
- `interval=10s`、`timeout=5s`、`retries=12`。

本地验证中该镜像首次启动约需 3 分钟，启动宽限不能沿用旧配置的 30 秒。

## 6. 部署和回滚

部署顺序：

1. 将 Compose 和真实 env 放入同一 `deploy` 目录；
2. 对真实 env 执行 `chmod 600`；
3. 登录 `docker.aiwelink.cc`；
4. 执行 `docker compose --env-file .env.aiwelink-test -f docker-compose.aiwelink-test.yml pull`；
5. 执行 `docker compose --env-file .env.aiwelink-test -f docker-compose.aiwelink-test.yml up -d`；
6. 等待容器变为 `healthy`，再通过 `0.0.0.0:8081` 对外接流量。

回滚时只修改 `SUB2API_IMAGE` 为上一不可变标签或 digest，再执行 `pull` 和 `up -d`。不得通过覆盖测试数据库或删除 migration 回滚。

## 7. 验收标准

- Compose 展开和语法校验成功；
- 服务器拉取的远端镜像 digest 与发布记录一致；
- 容器无重启并进入 `healthy`；
- `GET /health` 返回 `200`；
- 应用可连接远程测试 PostgreSQL 和 Redis；
- Growth Worker 使用实际 HTTPS 接口，且测试凭据不出现在 Git、终端回显或日志中。
