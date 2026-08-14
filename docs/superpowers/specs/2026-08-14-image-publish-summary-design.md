# AIWeLink 镜像发布总结设计

日期：2026-08-14

## 背景与目标

`aiwelink-dev` 和 `main` 的分支镜像已经由 `.github/workflows/backend-ci.yml` 中的 `publish-image` job 发布，但工作流目前只在步骤日志中显示推送过程。使用者需要进入日志并手动定位镜像标签和 digest，且发布失败时缺少统一的排查入口。

本次改动在 `publish-image` job 末尾增加 GitHub Actions Job Summary。无论构建、扫描、认证或推送成功还是失败，总结都会生成，并明确展示目标镜像和当前执行结果。

## 总结内容

总结使用 Markdown 表格展示以下信息：

- 发布状态：`success`、`failure` 或 `skipped`；
- 失败阶段：checkout、Buildx 初始化、版本读取、元数据生成、构建、扫描、凭据校验、登录、推送或未知阶段；
- 分支和完整 commit SHA；
- 应用版本与上游版本；
- 可变标签：`dev` 或 `latest`；
- 不可变标签：`dev-<12位SHA>` 或 `main-<12位SHA>`；
- 成功推送后的 registry manifest digest；
- 使用不可变标签和 digest 的 `docker pull` 命令。

分支标签规则保持不变：

| 来源 | 可变标签 | 不可变标签 |
| --- | --- | --- |
| `aiwelink-dev` push | `docker.aiwelink.cc/sub2api-aiwelink-dev:dev` | `docker.aiwelink.cc/sub2api-aiwelink-dev:dev-<12位SHA>` |
| `main` push | `docker.aiwelink.cc/sub2api-aiwelink-dev:latest` | `docker.aiwelink.cc/sub2api-aiwelink-dev:main-<12位SHA>` |

应用版本和上游版本仅用于说明镜像内的软件版本，不新增语义版本镜像标签。语义版本标签仍由 Release workflow 独占。

## 工作流实现

保留现有 `publish-image` job 和发布门禁。为 checkout、Buildx 初始化、构建、扫描、凭据校验、登录和推送步骤增加稳定的 step id，最后增加一个 `if: always()` 的总结步骤。

总结步骤从 GitHub 上下文计算目标标签，因此即使前置步骤未执行，也能显示预期镜像。版本信息优先读取 `Read application version` 的输出；上游版本直接读取 `backend/cmd/server/UPSTREAM_VERSION`。如果 checkout 或版本读取失败，使用 `unknown`，且总结步骤本身不得掩盖原始失败。

推送步骤成功后，通过 registry 中的不可变标签查询 manifest digest。查询失败不会把成功的镜像发布改判为失败，而是在总结中显示 `unavailable`。总结只记录标签、版本、SHA、digest 和步骤状态，不输出 registry 用户名、密码或其他 secret。

## 状态判定

总结按步骤顺序选择第一个 `failure` 或 `cancelled` 状态作为失败阶段。推送成功时状态为 `success`；任何前置步骤失败或推送失败时状态为 `failure`；某步骤因前置失败未运行时显示 `skipped`，同时保留已知的前一失败阶段。若 GitHub 返回的状态无法对应具体步骤，则失败阶段显示 `unknown`。

总结步骤使用 `if: always()`，并对缺失 outputs、缺失文件和 digest 查询失败提供回退值，确保成功和失败场景都能写入 `$GITHUB_STEP_SUMMARY`。

## 测试与验收

扩展 `deploy/tests/ci-workflow-contract-test.sh`，先证明旧 workflow 缺少总结契约，再验证以下约束：

- 总结步骤使用 `if: always()`；
- 总结包含发布状态、应用版本、上游版本、分支、commit、标签和 digest；
- 总结使用现有 `dev`、`latest`、`dev-<SHA>`、`main-<SHA>` 标签规则；
- PR 仍不发布镜像，`publish-image` 仍依赖 `ci-gate`；
- workflow YAML 可以被解析；
- 原有 CI workflow contract 全部继续通过。

最终行为需要由 PR 的 GitHub Actions 验证；合并到 `aiwelink-dev` 后，可在 `publish-image` job 的 Summary 中确认真实标签和 digest。

## 非目标

- 不恢复独立的 `Publish AIWeLink Branch Image` workflow；
- 不修改 registry 地址或凭据 secret；
- 不新增语义版本分支标签；
- 不改变 `aiwelink-dev -> main` 的合并和 review 策略；
- 不自动部署或重启使用该镜像的服务。
