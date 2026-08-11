# AIWeLink CI 与镜像发布加固设计

日期：2026-08-11

## 1. 背景与目标

当前 GitHub Actions 可以执行基础测试和镜像构建，但存在以下风险：

- `main` 与 `aiwelink-dev` 同时写入同一个语义版本镜像标签；
- 分支镜像在 CI 完成前发布；
- 部署检查存在但不属于必需检查；
- 前端只执行 213 个测试文件中的 6 个；
- feature push 与 pull request 对同一提交重复运行完整 CI；
- 推广注册只验证静态 Compose 内容，不验证 Sub2 与 Traffic 的调用契约；
- CI、镜像和 Release 使用不同 Node.js 主版本；
- 缺少代码扫描、凭据扫描和容器产物扫描。

本设计建立一个明确的质量门禁，使可部署镜像只能来自已通过全部必需检查的受信任分支提交，同时保持 `aiwelink-dev -> main` 的人工 PR 流程。任何工作流都不自动合并 `main`。

## 2. 触发与并发模型

`CI` 只接受以下事件：

- `pull_request`，目标分支为 `aiwelink-dev` 或 `main`；
- `push`，分支为 `aiwelink-dev` 或 `main`；
- `workflow_dispatch`，用于人工诊断。

feature 分支 push 不再单独运行 CI；创建或更新 PR 后由 `pull_request` 事件运行。并发组按工作流与 PR 编号或分支 ref 计算，新提交取消同一 PR 或同一受信任分支的旧运行，避免旧结果和旧镜像覆盖新结果。

定时安全扫描和真实灰度探测使用独立并发组，不与 PR CI 相互取消。

## 3. 统一质量门禁

CI 包含以下相互独立、可并行执行的工作：

1. `shell`：部署 shell 与 Apple Container 脚本测试；
2. `compose`：AIWeLink Compose 与 env 模板渲染检查；
3. `unit-tests`：Go unit build tag 测试；
4. `integration-tests`：Go integration build tag 测试；
5. `golangci-lint`：固定版本的 Go lint；
6. `frontend-checks`：ESLint、类型检查和生产构建；
7. `frontend-tests`：完整 `vitest run`，不维护手写关键测试白名单；
8. `backend-security`：固定版本 `govulncheck`；
9. `frontend-security`：生产依赖审计与有期限的例外校验；
10. `growth-contract`：Sub2 到 Fake Traffic 的推广注册绑定契约测试；
11. `secret-scan`：仓库历史/差异凭据扫描；
12. `filesystem-scan`：源代码与依赖文件系统漏洞扫描。

聚合 job `ci-gate` 使用 `if: always()` 检查上述所有 `needs` 结果，任何 required job 不是 `success` 都使门禁失败。分支保护最终只依赖稳定的 `ci-gate` 上下文，避免以后新增矩阵任务时忘记同步保护规则。

所有 Node.js job 与 Dockerfile 统一使用 Node.js 24。Go 版本继续由 `backend/go.mod` 的 `go 1.26.5` 驱动，不在 YAML 重复硬编码版本字符串。

## 4. 推广注册契约测试

PR CI 不访问生产 Traffic，也不依赖公网可用性。`growth-contract` 在测试进程中启动本地 Fake Traffic HTTP 服务，并让真实 Sub2 growth client/handler 对其发起请求。测试至少覆盖：

- `/r/{code}` 接收合法推广码并写入父域来源 Cookie；
- 注册请求携带来源 Cookie 后，Sub2 向配置的 Traffic bind endpoint 发送请求；
- service credential、site id、幂等键和必要请求字段符合 Traffic 契约；
- Traffic 成功响应时绑定完成；
- 连接超时、读取超时和非 2xx 响应不会阻塞前端注册响应；
- 临时失败进入加密 outbox，并可重试成功；
- 无来源 Cookie、非法推广码和关闭开关时不发送绑定请求。

如果现有单元测试已经覆盖某个场景，新增测试优先复用真实生产代码与现有测试基础设施，不复制实现逻辑。

## 5. 真实灰度探测

独立的定时与手动工作流探测 `https://api.aiwelink.cc` 和 Traffic 公网入口。该探测必须无副作用，不创建真实用户、不写注册绑定记录，也不输出 Cookie 或凭据。

允许验证：

- `/health` 可达；
- 专用测试推广码的 `/r/{code}` 返回预期状态、Location 和安全 Cookie 属性；
- Traffic 公共健康/契约端点可达；
- 响应延迟在明确的超时范围内。

在 Traffic 提供正式 dry-run bind endpoint 之前，线上探测不得伪造注册绑定。完整绑定语义由 PR 中的本地契约测试保证。灰度探测失败只告警，不自动回滚、不自动修改环境。

## 6. 镜像标签与发布门禁

分支镜像发布成为 `CI` 中依赖 `ci-gate` 的 job，并且只在受信任分支的 `push` 事件运行。PR、定时任务和手动诊断默认不获得私有仓库凭据。

标签所有权如下：

| 来源 | 可写标签 |
| --- | --- |
| `aiwelink-dev` push | `dev`, `dev-<12位SHA>` |
| `main` push | `latest`, `main-<12位SHA>` |
| `v*` Release | `<语义版本>`, Release 对应的不可变摘要 |

分支工作流不得写 `0.1.170-1` 一类语义版本标签。Release 工作流是语义版本标签的唯一写入者，并在发布前确认：

- tag 格式和仓库版本文件一致；
- tag commit 属于 `main` 历史；
- tag commit 存在成功的 `ci-gate` check run；
- 多架构镜像构建完成并通过漏洞扫描后才写入语义版本标签。

镜像部署应优先固定 `dev-<SHA>`、`main-<SHA>` 或 digest；`dev` 与 `latest` 仅作为人工发现和便捷跟踪标签。

## 7. 安全与供应链

- GitHub Actions 使用固定主版本，并为可执行工具固定具体版本；后续可由 Dependabot 更新；
- 增加 CodeQL 的 Go 与 JavaScript/TypeScript 分析；
- 启用 GitHub Secret Scanning（仓库权限支持时），同时在 PR 使用独立凭据扫描作为门禁；
- 镜像推送前执行高危/严重漏洞扫描；
- 保留 `pnpm audit` 例外机制，但清理过期且不再命中的例外，并把占位 owner 改为真实责任归属；
- 发布 job 只授予 `contents: read`，registry credential 仅注入发布步骤；Release 的 `contents: write` 与 `packages: write` 保持在最小需要范围。

## 8. 分支保护迁移

为避免把尚不存在的 check 设为 required 后造成死锁，迁移按以下顺序执行：

1. 功能分支创建到 `aiwelink-dev` 的 PR；
2. 确认 PR 上出现且通过 `ci-gate`；
3. 将 `aiwelink-dev` 必需检查切换为 `ci-gate`；
4. 合并到 `aiwelink-dev` 后，由后续 `aiwelink-dev -> main` PR 验证同一门禁；
5. 在用户明确允许 main PR 流程时，再将 `main` 必需检查切换为 `ci-gate`。

本任务不自动创建、批准或合并到 `main` 的 PR。当前用户的 review bypass 权限保持不变，但不得绕过质量门禁发布镜像。

## 9. 测试与验收

实现过程采用测试先行：先新增能证明旧工作流错误的静态工作流契约测试，再修改 YAML 和脚本使其通过。至少验证：

- feature push 不触发重复 CI；
- PR 和受信任分支 push 会触发 CI；
- `ci-gate` 包含全部 required jobs；
- publish 依赖 `ci-gate`，且 PR 不发布；
- dev、main、Release 标签集合互不冲突；
- Node.js 版本在 CI、Security、Release 与 Dockerfile 中一致；
- 完整前端测试命令不再使用固定六文件白名单；
- Fake Traffic 契约测试覆盖成功、超时、失败入队和重试；
- 灰度探测没有注册或绑定写操作；
- workflow YAML 可解析，shell 脚本通过语法检查；
- Go 单元/集成测试、前端 lint/typecheck/test/build 全部通过；
- Docker daemon 可用时执行本地镜像构建与扫描；不可用时由 PR Actions 提供最终容器验证证据。

## 10. 非目标

- 不自动合并 `aiwelink-dev` 或 `main`；
- 不修改生产部署 env 的真实秘密值；
- 不为 Traffic 增加有副作用的测试后门；
- 不在 CI 创建或删除真实用户；
- 不改变推广、注册或返佣的业务规则；
- 不把本次 CI 加固扩展为通用部署平台重构。
