# PR #31 合并报告：恢复原控制台并保留注册与安全改动

## 1. 报告摘要

PR [#31](https://github.com/AIwelink/sub2api-aiwelink-dev/pull/31) 已于 2026-08-14 15:58:27（Asia/Shanghai）合并到 `aiwelink-dev`。本次没有继续接受 PR #29 的控制台/UI overlay，也没有把产品版本标记为 Sub2API `v0.1.176`；处理方式是先完整回退 PR #29，再用独立提交恢复 GitHub 首页、已批准的注册/认证安全改动和 CodeQL 修复。

最终状态：

- PR #31 merge commit：`8e3821377be8c04e377731aa6c7c42d1bb1c2273`；
- 合并人：`AchernarAW`；
- PR 源分支：`codex/revert-pr29-dev-v2`；
- PR 目标分支：`aiwelink-dev`；
- 产品版本：`0.1.170-1`；
- 上游基线标识：`0.1.170`；
- 报告编写时 `origin/aiwelink-dev` 为 `31170f1a2`，该提交是随后合入的 PR #32，不是 PR #31 的 merge commit。

## 2. 相关 PR

| PR | 作用 | 结果 |
| --- | --- | --- |
| [#25](https://github.com/AIwelink/sub2api-aiwelink-dev/pull/25) | 将 GitHub 首页改为 AIWeLink API 品牌内容 | 2026-08-11 合并到 `main`，PR #31 只带入核验过的首页文件 |
| [#29](https://github.com/AIwelink/sub2api-aiwelink-dev/pull/29) | 同步 Sub2API v0.1.176 并叠加 growth/UI 改动 | 2026-08-14 11:45 合并，随后因范围和 UI 决策被 PR #31 回退 |
| [#31](https://github.com/AIwelink/sub2api-aiwelink-dev/pull/31) | 回退 PR #29，选择性恢复首页、注册/认证安全和 CodeQL 修复 | 2026-08-14 15:58 合并到 `aiwelink-dev` |
| [#32](https://github.com/AIwelink/sub2api-aiwelink-dev/pull/32) | 增加镜像发布摘要 | 2026-08-14 17:12 合并，是当前开发分支相对 PR #31 的后续提交 |

## 3. 决策边界

### 保留

- PR #25 的三份 README、AIWeLink Logo 和首页设计说明；
- 注册推荐码在 OAuth pending session 和 HttpOnly cookie 中的服务端衔接；
- 浏览器内存凭据、共享 single-flight 令牌刷新、退出和账号切换隔离；
- 注册验证码重试保护及自定义页面标题清理；
- Go `1.26.6`、nanoid `3.3.18` 和对应 CI 版本一致性约束；
- 纯后端 CodeQL 修复。

### 不保留

- PR #29 引入的 AIWeLink 管理控制台、监控面板和普通 UI overlay；
- 未明确批准的大范围后端自定义逻辑；
- 将当前产品版本声明为 `v0.1.176`；
- 把 `origin/main` 的整个历史直接合并进 `aiwelink-dev`。

后端后续开发遵循 [AIWELINK_GIT_WORKFLOW.md](../../AIWELINK_GIT_WORKFLOW.md)：通用能力尽量使用 Sub2API 官方实现，AIWeLink 自定义集中在注册及其必要衔接。

## 4. 实际提交链

PR #31 的源分支包含以下提交：

```text
de1ec7567 Revert PR #29
67f256a66 Apply PR #25 homepage content
1851555dd Retain approved auth/security updates
c9bc6ea21 Restore CodeQL hardening
```

### 4.1 `de1ec7567`：回退 PR #29

```text
722 files changed, 4776 insertions(+), 62772 deletions(-)
```

该提交回退 merge commit `ec1cd4381`，把代码树恢复到 PR #29 之前的基线。大量删除主要来自移除 PR #29 新增的上游同步内容和 UI/growth overlay，不能把这个数字理解为 PR #31 新增了同等规模的业务修改。

### 4.2 `67f256a66`：带入 PR #25 首页内容

```text
5 files changed, 44 insertions(+), 532 deletions(-)
```

原 PR #25 在 `main` 上的 merge commit 是 `a13429cf4`。PR #31 没有整体合并 `origin/main`，而是在回退后的分支上重放了经核验的 5 个首页文件：

- `README.md`；
- `README_CN.md`；
- `README_JA.md`；
- `assets/aiwelink-logo.png`；
- `docs/superpowers/specs/2026-08-11-aiwelink-api-github-homepage-design.md`。

`67f256a66` 虽保留了 “Merge pull request #25” 的提交标题，但它在 PR #31 分支上只有一个父提交，应视为内容重放，不是再次把 `main` 历史合并进来。

### 4.3 `1851555dd`：选择性保留注册和认证安全改动

```text
52 files changed, 1155 insertions(+), 301 deletions(-)
```

主要范围：

- LinuxDo、OIDC、微信、钉钉和邮件注册流程的 affiliate/pending-session 传递；
- 浏览器敏感凭据改为内存持有；
- 主动刷新和拦截器刷新共用一次请求，避免旋转 refresh token 竞争；
- 退出、账号切换和过期 401 响应不能把旧会话恢复到新账号；
- 登录/注册相关组件和聚焦测试；
- Go、nanoid 和 CI 工具链一致性。

该提交会修改登录、注册和 OAuth 页面，因为这些页面属于批准的注册流程；它没有恢复管理控制台、Channel Monitor 或 PR #29 的普通 UI overlay。

### 4.4 `c9bc6ea21`：恢复 CodeQL 修复

```text
11 files changed, 90 insertions(+), 51 deletions(-)
```

修复集中在后端 Go 和测试文件，处理了：

- 代理 URL 和凭据相关日志泄露；
- 配额重置值在整数转换前的范围检查；
- collection capacity 计算的整数溢出；
- Responses Lite、Codex metadata、计费别名和 WebSocket replay payload 的安全告警。

在 `1851555dd` 上运行的 GitHub CodeQL check 报告 “24 new alerts including 24 high severity security vulnerabilities”，annotations 为 24。加入 `c9bc6ea21` 后，同一 PR 的 CodeQL check 变为 “No new alerts in code changed by this pull request”，annotations 为 0。

这里的 0 只表示本 PR 修改代码中没有新的 CodeQL 告警，不代表整个分支的历史告警总数为 0。修复前的 24 个 annotation 是 24 个告警实例，分布在 9 个后端文件和 3 条规则中。

证据：

- [修复前 CodeQL check](https://github.com/AIwelink/sub2api-aiwelink-dev/runs/94692318458)；
- [修复后 CodeQL check](https://github.com/AIwelink/sub2api-aiwelink-dev/runs/94701771195)。

## 5. 合并和冲突处理方式

本次没有在 PR #29 的 700 多个文件上逐个继续叠加修改，而采用“确定性回退 + 允许清单重放”：

1. 以 `git revert -m 1 ec1cd4381` 的结果完整撤销 PR #29；
2. 只重放 PR #25 的首页文件，不整体合并 `main`；
3. 用独立提交恢复批准的注册/认证安全差异；
4. 用独立提交解决 CodeQL 告警；
5. 通过聚焦测试和 GitHub CI 验证允许清单。

这种方式比对整个仓库使用 `ours` 或 `theirs` 更容易审计，也能明确回答每类功能为什么保留或删除。

需要特别区分两个事实：

- PR #31 的最终整体 diff 必然触及大量控制台/UI 文件，因为第一步是在删除 PR #29 的 UI；
- 后续选择性恢复提交 `1851555dd` 和 `c9bc6ea21` 没有重新引入 PR #29 管理控制台或 Channel Monitor。`1851555dd` 的可见页面改动主要限于登录、注册、OAuth 回调和 `CustomPageView.vue` 标题净化，同时还包含后端、CI、Docker、依赖和文档改动；`c9bc6ea21` 只涉及 11 个后端文件。

PR #31 相对其目标分支的最终变更统计为：

```text
706 files changed, 4872 insertions(+), 62463 deletions(-)
```

## 6. 验证结果

PR 中记录的聚焦验证包括：

- 前端 auth/client 相关 4 个测试文件，共 56 个测试；
- 前端 typecheck；
- 修改过的 auth/client 文件聚焦 ESLint；
- OAuth pending-session cookie 和 affiliate handler 测试；
- 代理脱敏、配额边界、namespace flattening、Responses Lite tools、Codex metadata、billing alias 和 WebSocket payload 的聚焦 Go 测试；
- `go vet ./internal/pkg/apicompat ./internal/repository ./internal/service`；
- `bash deploy/tests/ci-workflow-contract-test.sh`；
- `git diff --check`。

最终 GitHub 状态中 15 个检查通过：

```text
CodeQL
backend-security
ci-gate
codeql (go)
codeql (javascript-typescript)
compose
frontend-checks
frontend-security
frontend-tests
golangci-lint
growth-contract
integration-tests
repository-scan
shell
unit-tests
```

`cla-check`、`cla-lock` 和 PR 场景下的 `publish-image` 按条件跳过，没有失败检查。

## 7. 最终版本状态

PR #31 没有完成可发布的 Sub2API `v0.1.176` 同步，因此保留：

```text
backend/cmd/server/VERSION          0.1.170-1
backend/cmd/server/UPSTREAM_VERSION 0.1.170
backend/go.mod                      go 1.26.6
frontend nanoid override            3.3.18
```

这避免了代码已经回退到旧基线、版本号却声称为 `v0.1.176` 的错误标记。

## 8. 遗留风险

### 8.1 已合并但又回退的上游祖先

`ec1cd4381` 仍是 `aiwelink-dev` 的祖先，`de1ec7567` 只是反向提交。以后普通 merge 可能认为 v0.1.176 的提交已经存在，不会自动恢复被回退的内容。

下一次上游同步必须先做目标 tag 与当前开发树的直接比较，并以允许清单重建官方基线。不能简单回退 `de1ec7567`，否则 PR #29 的控制台 UI 和其他 overlay 也会一起恢复。

### 8.2 产品版本仍是 0.1.170-1，上游基线仍是 0.1.170

安全修复和注册流程已保留不等于完成 v0.1.176 产品同步。后续版本提升必须在新的 `sync/upstream-*` PR 中完成，并通过版本脚本和对应 CI。

### 8.3 旧开发指南存在过时 Git 信息

`DEV_GUIDE.md` 仍包含旧 fork 名称、直接把 `upstream/main` 合并到 `main` 等历史命令。后续成员处理 Git 分支、同步和发布时应以根目录 `AIWELINK_GIT_WORKFLOW.md` 为准。

## 9. 回滚建议

不要直接回退 PR #31 的 merge commit `8e3821377`，因为这会重新引入 PR #29 的大范围内容。需要撤销时按目的做小范围 revert PR：

- 首页需要撤销：评估并回退 `67f256a66`；
- 注册/认证调整需要撤销：评估并回退 `1851555dd`；
- 某个 CodeQL 修复产生回归：只回退对应文件或修复提交，并重新运行 CodeQL；
- 长期分支一律通过 PR 回退，不强制推送、不改写历史。

## 10. 后续同步建议

1. 从最新 `origin/aiwelink-dev` 创建新的 `sync/upstream-v<version>` 分支；
2. 以官方发布 tag 为输入，不以未记录的 `upstream/main` 快照为版本基线；
3. 先建立官方代码、AIWeLink 注册差异、品牌/部署差异和已拒绝 UI 的四类清单；
4. 后端通用功能优先采用官方实现，不复制官方逻辑；
5. 只运行冲突相关的聚焦本地测试，完整门槛交由 PR CI；
6. 版本文件在代码同步和验证完成后统一更新。

完整操作步骤见 [AIWELINK_GIT_WORKFLOW.md](../../AIWELINK_GIT_WORKFLOW.md)。
