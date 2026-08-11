# AIWeLink 推广注册绑定部署

本文用于 AIWeLink 的 Sub2API 开发/灰度实例。该实例使用 `aiwelink-dev` 镜像、外部 PostgreSQL 和 Redis，并把邮件注册结果发送给 Traffic。

## 部署关系

| 项目 | 开发/灰度 | 正式环境 |
| --- | --- | --- |
| 宿主机端口 | `8080` | `8081` |
| 容器监听端口 | `8080` | `8080` |
| Compose/容器身份 | `sub2api-aiwelink-dev` / `sub2api-8080` | 使用独立身份，例如 `sub2api-8081` |
| PostgreSQL、Redis | 与正式环境共用 | 正式服务 |
| Traffic 绑定接口 | `https://aiwelink.cc/internal/growth/registrations/bind` | 相同 |

Traffic 在用户访问 `https://aiwelink.cc/r/{code}` 后设置域为 `.aiwelink.cc`、名称为 `awl_growth_sid` 的 Cookie。端到端测试时，注册页面也必须位于 `.aiwelink.cc` 下；直接访问服务器 IP 或 `IP:8080` 时，浏览器不会发送这个域 Cookie，因而无法完成绑定。

推广注册只处理邮件注册接口 `POST /api/v1/auth/register`。`/internal/growth/logins` 和 `GROWTH_LOGIN_*` 属于旧的登录事件集成，不能启用推广注册绑定。

## 准备 `.env`

在服务器的 `deploy` 目录执行：

```bash
cp .env.aiwelink-dev.example .env
chmod 600 .env
```

编辑 `.env`，替换所有 `replace_` 开头的值。不要把真实 `.env` 提交到 Git。

## `.env` 逐项说明

下面的模板按“Compose 运行参数、PostgreSQL、Redis、应用密钥、推广注册”分组。除特别标注外，变量名必须保持不变；值不要带示例中的反引号，也不要把密码或密钥写进 Compose 文件。

### Compose 运行参数

| 变量 | 示例/默认值 | 说明 |
| --- | --- | --- |
| `COMPOSE_PROJECT_NAME` | `sub2api-aiwelink-dev` | Docker Compose 项目名称，用于生成项目级资源标识。开发/灰度与正式环境必须不同，否则会发生容器、网络或数据卷命名冲突。 |
| `CONTAINER_NAME` | `sub2api-8080` | Sub2API 容器的固定名称，供 `docker logs`、`docker exec` 和健康检查使用。正式环境应改为独立名称，例如 `sub2api-8081`。 |
| `SUB2API_IMAGE` | `docker.aiwelink.cc/sub2api-aiwelink-dev:dev-<12位合并SHA>` | 要拉取的镜像。必须填 GitHub Action 发布的不可变 `dev-<SHA>` 标签，不要长期使用可变的 `dev` 或 `latest`，这样回滚时才能准确定位版本。 |
| `SUB2API_ENV_FILE` | `.env` | Compose 传给容器的环境文件路径，相对于 Compose 文件所在目录解析。服务器部署通常保持 `.env`，修改文件名时必须同时修改启动命令和该变量。 |
| `BIND_HOST` | `0.0.0.0` | 宿主机绑定地址。`0.0.0.0` 表示所有 IPv4 网卡；若只允许同机反向代理访问，可设为 `127.0.0.1`。这不是容器监听地址。 |
| `SERVER_PORT` | `8080` | 宿主机发布端口。开发/灰度使用 `8080`，正式使用 `8081`；Compose 会把它映射到容器的固定 `8080`，因此不要把容器端口改成 `8081`。 |
| `SERVER_MODE` | `release` | 应用运行模式。灰度和正式使用 `release`；`debug` 会增加调试输出，只用于临时排障。 |
| `RUN_MODE` | `standard` | 产品功能模式。`standard` 启用完整 SaaS、计费和余额校验；`simple` 是内部模式，会隐藏 SaaS 功能并跳过计费校验。 |
| `TZ` | `Asia/Shanghai` | IANA 时区，影响日志时间、定时任务和后台显示。灰度与正式应保持一致，避免排查时出现时间偏差。 |
| `SETUP_MIGRATION_TIMEOUT_SECONDS` | `600` | 启动时数据库迁移的最长等待秒数。共用数据库或迁移量较大时应留出足够时间；设为 `0` 使用程序内置默认值 60 秒。 |
| `ONEPANEL_NETWORK_NAME` | `1panel-network` | 已存在的外部 Docker 网络名称，反向代理和 Sub2API 必须加入同一网络。执行 `compose up` 前确认该网络已创建；它不是数据库或 Redis 的地址。 |

### PostgreSQL 连接

| 变量 | 示例/默认值 | 说明 |
| --- | --- | --- |
| `DATABASE_HOST` | `replace_with_postgresql_host` | 容器内可访问的 PostgreSQL 主机名或 IP。外部数据库部署不能填写 `localhost`，除非数据库确实运行在同一容器内。 |
| `DATABASE_PORT` | `5432` | PostgreSQL 服务对 Sub2API 暴露的 TCP 端口。 |
| `DATABASE_DBNAME` | `replace_with_database_name` | 已存在的数据库名称，灰度与正式共用同一套用户、推广 outbox 和迁移记录时必须填写同一个值。 |
| `DATABASE_USER` | `replace_with_database_user` | Sub2API 使用的 PostgreSQL 角色。该角色需要执行启动迁移，并对业务表拥有读写权限。 |
| `DATABASE_PASSWORD` | `replace_with_database_password` | `DATABASE_USER` 的密码，只保存于服务器 `.env`。不要添加引号；修改时要同时在 PostgreSQL 端轮换密码。 |
| `DATABASE_SSLMODE` | `disable` | PostgreSQL TLS 模式，可用 `disable`、`require`、`verify-ca`、`verify-full`。生产应按数据库服务商要求配置；`disable` 仅适用于已由私网或隧道保护的连接。 |
| `DATABASE_MAX_OPEN_CONNS` | `200` | 当前 Sub2API 进程允许打开的最大连接数（活跃和空闲总和）。所有灰度/正式副本的该值之和必须低于 PostgreSQL `max_connections` 并留出余量。 |
| `DATABASE_MAX_IDLE_CONNS` | `20` | 当前进程保留的最大空闲连接数，应小于或等于 `DATABASE_MAX_OPEN_CONNS`。过大浪费数据库连接槽位，过小会增加频繁建连。 |
| `DATABASE_CONN_MAX_LIFETIME_MINUTES` | `30` | 单个连接的最长生命周期，单位分钟。用于定期回收经过 NAT/LB 的陈旧连接；`0` 表示不限制，生产一般不建议设为 `0`。 |
| `DATABASE_CONN_MAX_IDLE_TIME_MINUTES` | `5` | 空闲连接的最长保留时间，单位分钟。超时后释放闲置连接；`0` 表示不限制。 |

### Redis 连接

| 变量 | 示例/默认值 | 说明 |
| --- | --- | --- |
| `REDIS_HOST` | `replace_with_redis_host` | 容器内可访问的 Redis 主机名或 IP。它必须能从 Sub2API 所在的外部网络访问。 |
| `REDIS_PORT` | `6379` | Redis 服务对 Sub2API 暴露的 TCP 端口。 |
| `REDIS_USERNAME` | 空值 | Redis ACL 用户名。未启用 ACL 或使用默认用户时留空。 |
| `REDIS_PASSWORD` | `replace_with_redis_password` | Redis 密码/ACL 密钥。只有明确关闭认证的内部开发 Redis 才能留空，服务器正式或灰度环境应配置认证。 |
| `REDIS_DB` | `0` | Redis 逻辑数据库编号。使用默认 Redis 配置时通常为 `0`；共用 Redis 的 Sub2API 实例必须保持一致，避免读写不同逻辑库。 |
| `REDIS_POOL_SIZE` | `1024` | 当前进程的 Redis 连接池上限。应结合 Redis `maxclients` 和副本数量调节，过大会耗尽 Redis 客户端连接。 |
| `REDIS_MIN_IDLE_CONNS` | `10` | 连接池预热保留的最小空闲连接数，可降低首请求延迟，但会持续占用 Redis 客户端槽位。 |
| `REDIS_ENABLE_TLS` | `false` | 是否使用 TLS 连接 Redis。只有外部 Redis 提供 TLS 端点时设为 `true`，并确认服务端证书链已在运行环境中受信。 |

### 持久化应用密钥

| 变量 | 示例/默认值 | 说明 |
| --- | --- | --- |
| `JWT_SECRET` | `openssl rand -hex 32` | JWT 签名密钥，至少需要 32 字节；推荐用命令生成 64 位十六进制字符串。必须在容器重启和版本更新间保持不变；更换会使现有登录会话全部失效。 |
| `TOTP_ENCRYPTION_KEY` | `openssl rand -hex 32` | 加密数据库中 2FA/TOTP 密钥的 AES 密钥，必须恰好是 64 位十六进制（解码后 32 字节）并持久化。更换后已有 TOTP 配置无法解密，用户将无法用 2FA 登录。 |

### 推广注册绑定

| 变量 | 示例/默认值 | 说明 |
| --- | --- | --- |
| `GROWTH_REGISTRATION_ENABLED` | `true` | 推广注册绑定总开关。只有 Traffic、共享数据库和域 Cookie 都准备好后才设为 `true`；关闭时不会创建或处理 outbox 任务。 |
| `GROWTH_REGISTRATION_ENDPOINT` | `https://aiwelink.cc/internal/growth/registrations/bind` | Sub2API 调用 Traffic 的完整 HTTPS 地址，必须包含 `/internal/growth/registrations/bind` 路径，并能从容器内解析和访问。 |
| `GROWTH_REGISTRATION_SITE_ID` | `aiwelink` | Traffic 站点标识，必须与 Traffic 的 `aiwelink` 配置键和站点 ID 完全一致。不要填写展示名称或域名。 |
| `GROWTH_REGISTRATION_SERVICE_CREDENTIAL` | Traffic 的 `SITE_SERVICE_CREDENTIALS_JSON.aiwelink` | Traffic 服务端凭据，必须逐字复制对应值，不得包含 JSON 引号、空格或换行。它是鉴权 token，不是加密密钥；不能与 outbox key 混用。 |
| `GROWTH_REGISTRATION_OUTBOX_ENCRYPTION_KEY` | `openssl rand -hex 32` | Sub2API outbox 的 AES 密钥，必须恰好 64 位且只含 `0-9a-fA-F`。使用同一 PostgreSQL 的所有启用实例必须填同一个值；不能填写普通 64 位字母数字 token，也不能复用 Traffic 凭据。 |
| `GROWTH_REGISTRATION_COOKIE_NAME` | `awl_growth_sid` | 浏览器推广会话 Cookie 名称，必须与 Traffic 设置的名称一致。Cookie 还必须覆盖 `.aiwelink.cc`；用服务器 IP 打开注册页不会发送该 Cookie。 |
| `GROWTH_REGISTRATION_CONNECT_TIMEOUT_SECONDS` | `2` | Sub2API 建立 Traffic TCP/TLS 连接的超时秒数。值太大会让故障上游拖慢 worker，值太小则可能在网络抖动时误判失败。 |
| `GROWTH_REGISTRATION_READ_TIMEOUT_SECONDS` | `5` | 等待 Traffic 返回绑定结果的超时秒数。超时属于可重试的上游失败，不限制用户注册接口本身的处理时间。 |

开发/灰度环境保留以下身份和端口：

```dotenv
COMPOSE_PROJECT_NAME=sub2api-aiwelink-dev
CONTAINER_NAME=sub2api-8080
BIND_HOST=0.0.0.0
SERVER_PORT=8080
```

`SUB2API_IMAGE` 应固定为 `aiwelink-dev` 合并后由 Action 生成的不可变镜像，而不是长期依赖可变的 `dev` 标签：

```dotenv
SUB2API_IMAGE=docker.aiwelink.cc/sub2api-aiwelink-dev:dev-<12位合并SHA>
```

如果用同一份 Compose 配置正式实例，至少需要使用独立身份和 `8081` 宿主机端口，避免容器、网络项目和数据卷命名冲突：

```dotenv
COMPOSE_PROJECT_NAME=sub2api-8081
CONTAINER_NAME=sub2api-8081
SERVER_PORT=8081
```

容器内部的 `SERVER_PORT` 由 Compose 固定为 `8080`，不能改成宿主机端口。

## 密钥规则

分别生成 JWT、TOTP 和推广 outbox 密钥，每条命令只用于一个配置项：

```bash
openssl rand -hex 32
openssl rand -hex 32
openssl rand -hex 32
```

每条输出都是 64 位十六进制字符串。分别写入：

```dotenv
JWT_SECRET=<第一条输出>
TOTP_ENCRYPTION_KEY=<第二条输出>
GROWTH_REGISTRATION_OUTBOX_ENCRYPTION_KEY=<第三条输出>
```

增长注册配置必须满足：

- `GROWTH_REGISTRATION_SERVICE_CREDENTIAL` 与 Traffic 的 `SITE_SERVICE_CREDENTIALS_JSON.aiwelink` 完全一致。
- `GROWTH_REGISTRATION_OUTBOX_ENCRYPTION_KEY` 是 Sub2API 专用密钥，不能复用服务凭据。
- 共用同一个 PostgreSQL 数据库且启用了增长注册的所有 Sub2API 实例，必须使用同一个 outbox 密钥。worker 会共同领取同一张 outbox 表中的任务；密钥不同会产生 `decrypt_failed` 死信。
- `JWT_SECRET` 和 `TOTP_ENCRYPTION_KEY` 也应保持固定，不能在每次容器启动时重新生成。

最终增长注册部分应为：

```dotenv
GROWTH_REGISTRATION_ENABLED=true
GROWTH_REGISTRATION_ENDPOINT=https://aiwelink.cc/internal/growth/registrations/bind
GROWTH_REGISTRATION_SITE_ID=aiwelink
GROWTH_REGISTRATION_SERVICE_CREDENTIAL=<Traffic中aiwelink对应的凭据>
GROWTH_REGISTRATION_OUTBOX_ENCRYPTION_KEY=<共用的64位十六进制outbox密钥>
GROWTH_REGISTRATION_COOKIE_NAME=awl_growth_sid
GROWTH_REGISTRATION_CONNECT_TIMEOUT_SECONDS=2
GROWTH_REGISTRATION_READ_TIMEOUT_SECONDS=5
```

## 校验并启动

以下命令均在服务器的 `deploy` 目录执行：

```bash
docker compose -f docker-compose.aiwelink-dev.yml config --quiet
docker compose -f docker-compose.aiwelink-dev.yml pull
docker compose -f docker-compose.aiwelink-dev.yml up -d
docker compose -f docker-compose.aiwelink-dev.yml ps
```

查看启动和迁移日志：

```bash
docker compose -f docker-compose.aiwelink-dev.yml logs --tail=200 sub2api
```

检查健康状态：

```bash
curl -fsS http://127.0.0.1:8080/health
docker inspect --format '{{json .State.Health}}' sub2api-8080
```

开发/灰度实例应该显示健康，并且宿主机只发布 `8080 -> 8080`。正式实例对应 `8081 -> 8080`。

## 检查容器配置

下面的命令会打印公开配置，并且只显示凭据和密钥的长度，不泄露其内容：

```bash
docker exec sub2api-8080 sh -lc '
for key in \
  GROWTH_REGISTRATION_ENABLED \
  GROWTH_REGISTRATION_ENDPOINT \
  GROWTH_REGISTRATION_SITE_ID \
  GROWTH_REGISTRATION_COOKIE_NAME \
  GROWTH_REGISTRATION_CONNECT_TIMEOUT_SECONDS \
  GROWTH_REGISTRATION_READ_TIMEOUT_SECONDS
do
  value=$(printenv "$key" || true)
  printf "%s=%s\n" "$key" "$value"
done
for key in \
  GROWTH_REGISTRATION_SERVICE_CREDENTIAL \
  GROWTH_REGISTRATION_OUTBOX_ENCRYPTION_KEY
do
  value=$(printenv "$key" || true)
  if [ -n "$value" ]; then
    printf "%s=<已设置，长度=%s>\n" "$key" "${#value}"
  else
    printf "%s=<未设置>\n" "$key"
  fi
done
'
```

从容器验证 Traffic HTTPS 路由和鉴权。无凭据请求的预期状态是 `401`：

```bash
docker exec sub2api-8080 sh -lc \
  "wget -S -O /dev/null -T 10 --header='Content-Type: application/json' --post-data='{}' https://aiwelink.cc/internal/growth/registrations/bind 2>&1 || true"
```

这里出现 `401` 表示域名、TLS、反代路由和鉴权边界都正常。`404` 表示反代路径错误，`502/504` 表示反代到 Traffic 的上游不可用。

## 端到端验收

1. 清理测试浏览器中现有的 `awl_growth_sid` Cookie。
2. 访问一个真实推广链接：`https://aiwelink.cc/r/{code}`。
3. 确认跳转后的浏览器会话中存在域为 `.aiwelink.cc` 的 `awl_growth_sid`。
4. 使用同一个浏览器会话，打开 `.aiwelink.cc` 下的开发/灰度注册域名。
5. 使用一个从未注册过的新邮箱完成邮件注册，不使用 OAuth、Passkey 或其他登录入口。
6. 在浏览器网络面板检查 `POST /api/v1/auth/register` 的请求头，确认 Cookie 中包含 `awl_growth_sid`。
7. 在 Traffic 中确认该新 Sub2API 用户已经绑定到对应推广来源。

只通过 `http://服务器IP:8080` 打开注册页面不算有效验收，因为 `.aiwelink.cc` Cookie 不会发送给 IP 地址。

## Outbox 排障

成功投递的 outbox 行会被删除，所以表中没有该行可能表示已经成功。检查当前积压和死信：

```sql
SELECT
  outbox_id,
  source_registration_id,
  site_id,
  external_user_id,
  attempt_count,
  last_http_status,
  last_error_code,
  last_request_id,
  available_at,
  dead_lettered_at,
  created_at
FROM growth_registration_outbox
ORDER BY outbox_id DESC
LIMIT 50;
```

常见结果：

| 状态 | 含义 | 处理 |
| --- | --- | --- |
| 没有新增行 | 请求没有携带 Cookie、功能未启用，或者已经快速投递成功 | 先检查容器八个变量和注册请求 Cookie，再检查 Traffic 绑定结果 |
| `401` / `403` | Traffic 服务凭据不匹配 | 对齐 `SITE_SERVICE_CREDENTIALS_JSON.aiwelink` |
| `404` | endpoint 或反代路径错误 | 必须使用本文的 `/internal/growth/registrations/bind` |
| `422` | `site_id`、会话或请求数据不被 Traffic 接受 | 核对 Traffic 的 `aiwelink` 站点配置和 Cookie 来源 |
| `decrypt_failed` | 领取任务的实例使用了不同 outbox 密钥 | 统一所有共库实例的 outbox 密钥 |
| `503` 且错误码为 `temporarily_unavailable` 或 `source_adapter_unavailable` | Traffic 暂时不可用 | Sub2API 自动退避重试 |
| 网络、DNS 或 TLS 错误 | 请求尚未到达 Traffic | Sub2API 自动重试；检查容器网络和 DNS |

`401`、`403`、`404`、`422` 和 `decrypt_failed` 会进入死信，且加密会话会被清除；修复配置后应使用新邮箱重新执行一次完整推广注册验收，旧死信不会自动恢复。

## 更新和回滚

合并到 `aiwelink-dev` 后，等待 `Publish AIWeLink Branch Image` Action 成功，再把 `.env` 中的镜像改为 Action 生成的不可变标签：

```dotenv
SUB2API_IMAGE=docker.aiwelink.cc/sub2api-aiwelink-dev:dev-<新的12位合并SHA>
```

更新：

```bash
docker compose -f docker-compose.aiwelink-dev.yml pull
docker compose -f docker-compose.aiwelink-dev.yml up -d
docker compose -f docker-compose.aiwelink-dev.yml ps
```

回滚时只把 `SUB2API_IMAGE` 改回上一个已经验证的不可变标签，然后再次执行 `pull` 和 `up -d`。不要删除或重建共用数据库；涉及数据库迁移的版本回滚应先确认迁移兼容性和备份。
