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

Traffic 在用户访问 `https://aiwelink.cc/r/{code}` 后设置域为 `.aiwelink.cc`、名称为 `awl_growth_sid` 的 Cookie。端到端测试时，API 首页和后续注册页面都必须位于 `.aiwelink.cc` 下；直接访问服务器 IP 或 `IP:8080` 时，浏览器不会发送这个域 Cookie，因而无法完成绑定。

API 域同时提供后端推广入口 `https://api.aiwelink.cc/r/{code}`。Sub2API 只校验推广码格式并立即返回 `302` 到 `https://aiwelink.cc/r/{code}?entry=api`；Traffic 记录归因、设置父域 Cookie 后，再直接 `302` 到 `https://api.aiwelink.cc/`。API 首页重新加载后复用现有 `HomepageIntro` 入场动画；该链路不加载 `aiwelink.cc` 品牌主页，不查询 Sub2API 数据库，也不从 Sub2API 同步请求 Traffic。

Traffic 的 `8300` 端口继续只监听 loopback 或私网，不需要对公网开放，也不需要在每个分布式 Sub2API 节点反代 `8300`。浏览器访问的是现有公网 HTTPS 域名 `aiwelink.cc/r/*`。

推广注册只处理邮件注册接口 `POST /api/v1/auth/register`。`/internal/growth/logins` 和 `GROWTH_LOGIN_*` 属于旧的登录事件集成，不能启用推广注册绑定。

## 准备 `.env`

在服务器的 `deploy` 目录执行：

```bash
cp .env.aiwelink-dev.example .env
chmod 600 .env
```

编辑 `.env`，替换所有 `replace_` 开头的值。不要把真实 `.env` 提交到 Git。

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
GROWTH_REGISTRATION_REFERRAL_BASE_URL=https://aiwelink.cc/r
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
  GROWTH_REGISTRATION_REFERRAL_BASE_URL \
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
2. 访问一个真实 API 域推广链接：`https://api.aiwelink.cc/r/{code}`。
3. 在网络面板确认第一跳为 `https://aiwelink.cc/r/{code}?entry=api`，第二跳为 `https://api.aiwelink.cc/`；确认 API 首页播放现有入场动画，且链路未请求 `aiwelink.cc` 品牌主页资源。
4. 确认跳转后的浏览器会话中存在域为 `.aiwelink.cc` 的 `awl_growth_sid`。
5. 使用一个从未注册过的新邮箱完成邮件注册，不使用 OAuth、Passkey 或其他登录入口。
6. 在浏览器网络面板检查 `POST /api/v1/auth/register` 的请求头，确认 Cookie 中包含 `awl_growth_sid`。
7. 在 Traffic 中确认该新 Sub2API 用户已经绑定到对应推广来源。

发布时必须先部署支持 `entry=api` 和 `API_HOMEPAGE_URL` 的 Traffic 版本，再部署包含 `/r/{code}` 的 Sub2API 版本。顺序反过来会让第一跳已带 `entry=api`，但旧 Traffic 仍回到错误目标。

只通过 `http://服务器IP:8080` 打开 API 首页或注册页面不算有效验收，因为 `.aiwelink.cc` Cookie 不会发送给 IP 地址。

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
