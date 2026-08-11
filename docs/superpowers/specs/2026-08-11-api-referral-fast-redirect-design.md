# API 域推广入口快速跳转设计

> 日期：2026-08-11
> 状态：已确认设计，待实施
> 影响仓库：`AIwelink/sub2api-aiwelink-dev`、`AIwelink/traffic-analysis`

## 1. 背景

AIWeLink 的推广入口当前由 Traffic Analysis 处理：

```text
GET https://aiwelink.cc/r/{code}
```

Traffic 校验推广码、记录访问、更新 30 天末次邀请并设置父域 Cookie：

```text
awl_growth_sid; Domain=.aiwelink.cc; Path=/; Secure; HttpOnly; SameSite=Lax
```

Sub2API 已在邮件注册接口 `POST /api/v1/auth/register` 捕获该 Cookie，并通过共享 PostgreSQL outbox 异步调用 Traffic 完成不可变注册绑定。分布式 Sub2 节点共享数据库和 outbox，因此任意节点都能处理注册。

缺少的能力是 API 域本身没有后端推广入口：

```text
GET https://api.aiwelink.cc/r/{code}
```

同时，API 域入口不能等待 Vue UI 才开始归因，不能公开或反代 Traffic 的 `8300` 端口，也不能加载 `aiwelink.cc` 品牌主页。归因完成后应回到 API 域首页，并复用首页现有入场动画。

## 2. 已确认决策

1. Sub2API 后端新增精确的 `GET /r/{code}` 路由，不由前端处理。
2. Sub2API 不查询推广数据库、不签发 Growth Cookie，也不向 Traffic 发起服务器端 HTTP 请求。
3. 合法推广码由 Sub2API 立即 `302` 到 Traffic 的公开 HTTPS 域名，并附加固定入口枚举 `entry=api`。
4. Traffic 仍是推广码、访问记录、末次邀请和 Growth Cookie 的唯一权威实现。
5. Traffic 识别 `entry=api` 后直接跳转 `https://api.aiwelink.cc/`，不加载 `aiwelink.cc` 品牌主页，也不直接进入注册页。
6. API 首页在新的文档加载中自然播放现有 `HomepageIntro` 入场动画；不新增推广专用动画、query 参数或前端持久状态。
7. 普通 `https://aiwelink.cc/r/{code}` 保持现有主页跳转行为。
8. 不接受任意 `next`、`redirect`、origin 或 URL 参数，避免开放重定向。
9. Sub2API 与 Traffic 分别使用独立功能分支和 PR；Sub2 PR 只投 `aiwelink-dev`，不修改或合并 Sub2 `main`。

## 3. 请求链路

```text
浏览器
  -> GET https://api.aiwelink.cc/r/7km4q2xd
  -> Sub2API 后端校验 code 形状
  <- 302 Location: https://aiwelink.cc/r/7km4q2xd?entry=api
  -> Traffic GET /r/7km4q2xd?entry=api
  -> Traffic 校验 code、记录访问、更新末次邀请
  <- 302 Location: https://api.aiwelink.cc/
     Set-Cookie: awl_growth_sid=...; Domain=.aiwelink.cc; ...
  -> GET https://api.aiwelink.cc/（自动携带父域 Cookie）
  -> API 首页复用现有 HomepageIntro 入场动画
  -> POST https://api.aiwelink.cc/api/v1/auth/register
  -> 任意 Sub2 节点捕获 Cookie 并写共享 outbox
  -> Worker 调用 Traffic 注册绑定接口
```

整个流程只有后端重定向和 Traffic 写入，不等待前端脚本，也不请求主页 HTML、JavaScript、CSS 或图片资源。

## 4. Sub2API 设计

### 4.1 配置

`GrowthRegistrationConfig` 增加公开推广基址：

```dotenv
GROWTH_REGISTRATION_REFERRAL_BASE_URL=https://aiwelink.cc/r
```

启用推广注册时，该值必须满足：

- 绝对 HTTPS URL；
- 不含用户名、密码、query 或 fragment；
- path 规范化为单一 `/r`；
- 不从请求头、query 或前端环境变量覆盖。

配置只表示浏览器可访问的 Traffic 公共入口，不能复用内部注册绑定 endpoint。

### 4.2 路由

新增：

```text
GET /r/:code
```

处理规则：

- 推广码转为小写并按 Traffic 相同规则校验：`^[a-hj-km-np-z2-9]{8}$`；
- 合法值拼为 `{REFERRAL_BASE_URL}/{code}?entry=api`；
- 丢弃原请求携带的所有 query 和 fragment，不做透传；
- 返回 `302 Found`、`Cache-Control: no-store`；
- 不创建 HTTP client、数据库连接或后台任务；
- 非 GET 方法不匹配该路由；
- 格式非法或功能未启用时，安全地 `302` 到本域相对路径 `/`，不访问 Traffic。

相对 fallback 保证每个区域 API 域仍留在自己的首页，不需要新增第二个公开 URL 配置。

### 4.3 路由边界

该后端路由必须在 SPA fallback 之前注册，避免 `/r/{code}` 被嵌入前端吞掉。它不属于 `/api/v1`，不需要登录、CORS 或服务凭据，但继续经过现有恢复、安全响应头和请求日志中间件。

## 5. Traffic Analysis 设计

### 5.1 配置

Settings 增加固定 API 首页目标：

```dotenv
API_HOMEPAGE_URL=https://api.aiwelink.cc/
```

生产校验要求：

- HTTPS；
- host 精确为 `api.aiwelink.cc`；
- 默认端口；
- path 精确为 `/`；
- 不含用户名、密码、query 或 fragment。

该配置不是用户输入，不允许由请求覆盖。

### 5.2 入口枚举

现有路由扩展为：

```text
GET /r/{code}?entry=api
```

`entry` 只影响最终安全目标，不进入推广归因、来源字段、数据库或指标标签：

- 缺少 `entry`：使用现有 `PUBLIC_HOMEPAGE_URL`；
- `entry=api`：使用固定 `API_HOMEPAGE_URL`；
- 其他值：按缺少 `entry` 处理，不能解释成 URL。

公开用户即使手工添加 `entry=api`，也只能在两个受控 AIWeLink 页面之间选择，不能形成开放重定向。

### 5.3 统一目标选择

Traffic 在进入数据库逻辑前计算受控目标，并在所有结果中保持一致：

- `attribution_updated`；
- `invalid_code`；
- `excluded`；
- `database_error`。

当 `entry=api` 时，上述结果都直接到 API 首页。新的 API 文档加载会复用现有首页入场动画，Traffic 不负责动画状态。只有成功的计数访问会写来源并设置/刷新 `awl_growth_sid`；无效码、排除请求或数据库错误不伪造 Cookie，继续沿用现有归因规则。

## 6. 安全与性能

- `8300` 保持监听 `127.0.0.1` 或私网，不新增公网端口；浏览器请求的是现有 `https://aiwelink.cc/r/*`。
- Sub2 路由只执行字符串规范化和响应写入，不受 Traffic、数据库或网络延迟影响。
- Traffic 公共入口继续使用已有 code 校验、限速、bot 过滤、`Cache-Control: no-store` 和数据库事务。
- URL 通过结构化解析和固定枚举构造，不使用字符串接受任意目的地址。
- 日志不记录 Growth Cookie、服务凭据或完整请求正文；推广码也不新增为指标标签。
- 两次 `302` 是不反代、不复制 Traffic 逻辑且保证浏览器收到父域 Cookie 的最小必要网络成本。

## 7. 错误处理

| 场景 | 行为 |
| --- | --- |
| Sub2 收到合法 code | 立即 `302` 到 Traffic API 模式 |
| Sub2 收到非法 code | `302` 到本域 `/`，不访问 Traffic |
| Sub2 推广功能关闭 | `302` 到本域 `/` |
| Traffic 收到合法 code | 记录来源、设置 Cookie、`302` 到 API 首页 |
| Traffic 收到无效/停用/过期 code | 不设置新 Cookie，`302` 到 API 首页 |
| Traffic 判定 bot/排除 | 保留现有审计规则，`302` 到 API 首页 |
| Traffic 数据库故障 | 不伪造来源，`302` 到 API 首页并记录稳定错误结果 |
| Traffic 域名完全不可达 | 浏览器显示网络错误；Sub2 不进行同步探测或等待 |

## 8. 测试

### 8.1 Sub2API

- 合法小写和大写 code 规范化后返回精确 `302 Location`；
- 非法字符、长度、额外 path segment 不进入 Traffic；
- 原始 query 不能覆盖或追加到目标；
- 功能关闭时落到 `/`；
- 响应包含 `Cache-Control: no-store`；
- 路由在嵌入前端存在时仍优先于 SPA fallback；
- 处理器没有 HTTP client 或数据库依赖。

### 8.2 Traffic Analysis

- 无 `entry` 的现有路径继续跳主页；
- `entry=api` 在成功、无效码、排除和数据库错误时均跳 API 首页；
- 未知 `entry` 不能成为任意 Location；
- API 模式成功响应仍设置正确的 `.aiwelink.cc` Cookie；
- Cookie 属性、访问记录和末次邀请规则不变；
- Settings 拒绝不安全 API 首页 URL。

### 8.3 端到端

使用不自动跟随重定向的客户端逐跳验证：

1. API 域第一跳不访问前端资源；
2. Traffic 第二跳设置 Cookie 且 Location 为 API 首页；
3. API 首页在新文档加载后播放现有 `HomepageIntro`，无需推广专用状态；
4. 浏览器后续注册请求携带 `awl_growth_sid`；
5. 新邮箱注册产生 outbox 并在 Traffic 锁定对应推广来源；
6. 整条链路不请求 `aiwelink.cc/` 主页文档。

## 9. 发布顺序

1. 先合并并部署 Traffic PR，使 `entry=api` 成为向后兼容能力；
2. 验证原有主站 `/r/{code}` 行为未改变；
3. 再合并 Sub2 PR 到 `aiwelink-dev` 并等待镜像 Action；
4. 部署 Sub2 灰度实例，验证 API 域逐跳重定向和 Cookie；
5. 灰度通过后由运维更新正式 Sub2 镜像；
6. 任一应用回滚都不需要数据库回滚，本功能不新增 migration。

Traffic PR 和 Sub2 PR 均只创建、测试和交付，不由本任务自动合并。
