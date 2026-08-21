# CI Merge Optimizations Design

日期：2026-08-22

## 目标

保持现有 `ci-gate` 合并门禁和开发分支触发规则不变，降低任务永久阻塞和主分支镜像重复发布的风险。

## 方案

- 为测试、Lint、CodeQL、聚合门禁和镜像发布 job 增加保守的超时上限；不改变测试命令或 required check 名称。
- 在 Go CodeQL 初始化前设置 Go，避免 CodeQL tracing 路径被 `setup-go` 覆盖，同时保留现有 Go 构建和分析步骤。
- 继续使用 `backend-ci.yml` 中依赖 `ci-gate` 的镜像发布实现；主分支迁移时移除旧的无门禁发布 workflow，避免两个发布者竞争相同标签。
- 暂不提前修改 `main` 的 required checks。等新 CI 文件进入 `main` 后，再把旧状态名切换为 `ci-gate`，避免当前主分支流程被提前阻塞。

## 非目标

- 不改变 `pull_request`/受信任分支 `push` 触发规则。
- 不把定时安全扫描、公网 canary 或 Release workflow 纳入 PR 门禁。
- 不修改业务测试、版本规则、审批数量或生产秘密。

## 验收

- CI workflow 合同测试验证 job 超时、CodeQL 步骤顺序和发布依赖关系。
- YAML 可解析，shell 合同测试通过，工作树只包含本次 CI 配置和文档变更。
