# AIWeLink 命名规范

本文件规定 AIWeLink Sub2API 二次开发仓库中的分支、提交、版本、镜像、代码、接口、配置和数据库命名方式。

## 1. 使用原则

1. 名称必须表达业务意图，避免使用 `test`、`tmp`、`new`、`final`、`misc` 等无语义名称。
2. 默认使用 ASCII 字符、英文和数字。需要分隔时优先使用 `-` 或 `_`，不要混用空格、全角符号和拼音缩写。
3. 同一类对象使用同一种格式。已有公共名称、外部协议字段和上游名称不能为了统一而随意改名。
4. 名称一旦进入 API、数据库、镜像或发布流程，就视为兼容性契约。改名必须说明迁移或兼容方案。
5. 标识符中的产品名写作 `AIWeLink`；API 产品标题写作 `AIWeLink API`。仓库名使用 GitHub 已登记的大小写形式，镜像路径一律小写。

本文件与 [AIWELINK_GIT_WORKFLOW.md](./AIWELINK_GIT_WORKFLOW.md) 配套使用。分支保护、合并方式和发布顺序以 Git 工作流文件为准；本文件负责名称格式。

## 2. Git 分支

### 2.1 长期分支

| 分支 | 用途 | 直接提交 |
| --- | --- | --- |
| `upstream/main` | Sub2API 官方远端跟踪分支 | 只读跟踪，不向其推送 |
| `main` | AIWeLink 生产基线 | 禁止普通功能直接提交 |
| `aiwelink-dev` | 开发、集成和 CI 验证 | 通过 PR 合并 |

### 2.2 短期分支

格式为 `<type>/<short-description>`，全部使用小写 kebab-case：

| 类型 | 起点 | PR 目标 | 使用场景 |
| --- | --- | --- | --- |
| `feature/` | `aiwelink-dev` | `aiwelink-dev` | 新功能 |
| `fix/` | `aiwelink-dev` | `aiwelink-dev` | 普通缺陷修复 |
| `refactor/` | `aiwelink-dev` | `aiwelink-dev` | 不改变业务契约的重构 |
| `docs/` | `aiwelink-dev` | `aiwelink-dev` | 文档变更 |
| `test/` | `aiwelink-dev` | `aiwelink-dev` | 测试和测试基础设施 |
| `chore/` | `aiwelink-dev` | `aiwelink-dev` | 构建、依赖和维护 |
| `sync/upstream-` | `aiwelink-dev` | `aiwelink-dev` | 同步 Sub2API 官方更新 |
| `hotfix/` | `main` | `main` | 生产紧急修复 |

推荐格式：

```text
feature/api-key-expiration
fix/rollback-version-filter
refactor/update-service-release-source
docs/naming-conventions
sync/upstream-v0.1.171
hotfix/payment-callback-timeout
```

需要关联 Issue 时，将编号放在描述前：

```text
feature/123-api-key-expiration
fix/456-rollback-timeout
```

分支名称应控制在 60 个字符以内，并使用动词或明确结果，例如 `add`、`fix`、`remove`、`align`、`validate`、`publish`。短期分支合并后删除，不重复使用旧分支名。

## 3. Commit 和 PR

### 3.1 Commit

使用 Conventional Commits：

```text
<type>(<scope>): <imperative summary>
```

`scope` 可省略，但涉及多个模块时应填写。标题使用英文小写开头、祈使语气、不加句号，建议不超过 72 个字符。

允许的类型：

| 类型 | 用途 |
| --- | --- |
| `feat` | 新功能 |
| `fix` | 缺陷修复 |
| `refactor` | 行为不变的重构 |
| `perf` | 性能优化 |
| `test` | 测试变更 |
| `docs` | 文档变更 |
| `build` | 构建或打包变更 |
| `ci` | CI 工作流变更 |
| `chore` | 维护性变更 |
| `revert` | 回滚提交 |

示例：

```text
feat(version): add AIWeLink release version parser
fix(update): filter draft releases from rollback candidates
docs(workflow): define upstream synchronization flow
ci(release): scope private registry credentials
```

不允许把多个无关目的合并到一个提交中。破坏兼容性时在类型后添加 `!`，并在正文说明迁移方式：

```text
feat(api)!: rename public settings response field
```

### 3.2 Pull Request

PR 标题沿用 Commit 格式，或在自动化发布/维护场景使用明确前缀：

```text
feat(version): add AIWeLink release versioning
[release] publish AIWeLink 0.1.176-1
```

PR 描述至少包含：目的、主要文件或模块、用户影响、配置或部署变化、验证命令、已知风险和回滚方式。普通功能 PR 指向 `aiwelink-dev`；发布 PR 指向 `main`；官方同步 PR 使用 `sync/upstream-*` 并保留合并提交。

## 4. 版本、Tag 和 Release

### 4.1 AIWeLink 版本

AIWeLink 版本由 Sub2API 上游基线和 AIWeLink 修订号组成：

```text
<upstream-baseline>-<aiwelink-revision>[.<subrevision>...]
```

当前基线示例：

| 含义 | 值 |
| --- | --- |
| 上游基线 | `0.1.176` |
| 第一版 AIWeLink 修订 | `0.1.176-1` |
| 后续修订 | `0.1.176-2.4` |
| 上游进入 `0.1.171` 后重新计数 | `0.1.171-1` |

规则：

1. Git Tag 使用 `v` 前缀，例如 `v0.1.176-1`。
2. `backend/cmd/server/VERSION` 不包含 `v`，必须是完整 AIWeLink 版本。
3. `backend/cmd/server/UPSTREAM_VERSION` 只写三段上游版本，例如 `0.1.176`。
4. AIWeLink 修订号从 `1` 开始，禁止 `0`、前导零和无意义的 `-dev` 后缀。
5. 更新、回滚和安装器只接受完整 AIWeLink 版本，不把上游三段版本当作可安装版本。
6. 发布 Tag、VERSION、镜像版本和 Release 元数据必须一致；发布流程不得自动回写 VERSION。

### 4.2 Docker 镜像 Tag

镜像仓库路径使用小写：

```text
docker.aiwelink.cc/sub2api-aiwelink-dev
```

推荐 Tag：

| Tag | 用途 |
| --- | --- |
| `0.1.176-1` | 可追溯的正式版本，不带 `v` |
| `latest` | 仅由生产发布流程维护 |
| `dev` | 开发环境浮动版本 |
| `dev-<short-sha>` | 开发构建的不可变定位版本 |

不要把密码、Token、分支全名或随机时间戳写入镜像 Tag。需要同时表达架构时使用 `-amd64`、`-arm64`，多架构 manifest 使用无架构后缀的版本 Tag。

## 5. Go 后端

| 对象 | 规则 | 示例 |
| --- | --- | --- |
| package | 全小写，单词优先，不使用下划线 | `versioninfo`、`handler` |
| 文件 | 小写 kebab 不适用于 Go 文件；使用小写或现有模块名 | `update_service.go` |
| 导出类型/函数 | PascalCase | `BuildInfo`、`Parse` |
| 非导出类型/函数 | camelCase | `compareVersions` |
| 常量 | Go 标准 PascalCase 或全小写，不强制 UPPER_SNAKE | `maxRollbackVersions` |
| 错误变量 | `Err` + PascalCase | `ErrNoUpdateAvailable` |
| 接口 | 行为名，必要时以 `-er` 结尾 | `GitHubClient`、`Reader` |
| 测试 | 同包 `_test.go`；测试名 `Test<Subject>_<Behavior>` | `TestParse_RejectsDraft` |
| 构建变量 | 明确区分 `Version`、`UpstreamVersion`、`Commit`、`Date` | `main.UpstreamVersion` |

包名不重复模块名，不使用 `utils`、`common`、`misc` 作为新功能的默认归属；应按业务职责命名。

## 6. Vue、TypeScript 和前端资源

| 对象 | 规则 | 示例 |
| --- | --- | --- |
| Vue 组件 | PascalCase | `VersionBadge.vue` |
| 页面 | PascalCase + `View` | `SettingsView.vue` |
| 弹窗 | PascalCase + `Modal` 或 `Dialog` | `ConfirmDialog.vue` |
| Composable | `use` + PascalCase | `useVersionInfo.ts` |
| Store | 现有模块使用 camelCase 文件名 | `adminCompliance.ts` |
| 类型/接口 | PascalCase | `BuildInfo`、`SettingsResponse` |
| 变量/函数 | camelCase | `upstreamVersion`、`loadSettings` |
| 常量 | 模块内 camelCase；跨模块固定值可用 UPPER_SNAKE_CASE | `versionLabel` |
| 测试 | 被测文件名 + `.spec.ts` | `VersionBadge.spec.ts` |
| CSS class | kebab-case；组件私有样式优先 scoped | `.version-badge` |
| 路由 | 小写 kebab-case，使用业务名词 | `/admin/settings` |

组件文件名、导出组件名和测试描述应使用同一业务名词。新建文件不要使用 `NewXxx`、`TempXxx`、`TestXxx` 作为正式名称。

## 7. API、JSON 和 HTTP

1. URL 路径使用小写 kebab-case 和复数资源名，例如 `/v1/api-keys`、`/v1/releases`。
2. URL 版本使用 `/v1`、`/v2`，不要把产品版本或 Git Tag 放入路径。
3. JSON 字段使用 snake_case，并与后端 DTO、数据库字段保持一致，例如 `upstream_version`。
4. 查询参数使用 snake_case；布尔值使用 `true` / `false`，不要混用 `0` / `1`。
5. HTTP Header 使用标准首字母大写形式，例如 `Authorization`、`X-Request-ID`。
6. 新的错误码使用稳定的机器可读标识，例如 `version_mismatch`；展示给用户的文案放在前端 i18n。
7. API 字段改名必须保留兼容读取期，或在 PR 中提供迁移和版本策略。

## 8. 配置、环境变量和密钥

| 类型 | 规则 | 示例 |
| --- | --- | --- |
| 环境变量 | UPPER_SNAKE_CASE | `UPSTREAM_VERSION` |
| Secret 环境变量 | UPPER_SNAKE_CASE，使用 `_TOKEN`、`_PASSWORD`、`_KEY` 后缀 | `AIWELINK_REGISTRY_PASSWORD` |
| YAML 配置键 | snake_case | `required_status_checks` |
| 配置文件 | 按用途命名，环境后缀放末尾 | `docker-compose.aiwelink-dev.yml` |
| GitHub Workflow | 动词-对象，小写 kebab-case | `publish-aiwelink-dev-image.yml` |

禁止把真实密钥、用户 Token、生产 URL 中的临时凭据写进文件名、Tag、Commit、PR 或日志。示例值使用 `CHANGE_ME`、`example-token` 等明显占位符。

## 9. 数据库和迁移

1. 表名、列名、索引名和约束名使用小写 snake_case。
2. 表名使用复数名词，关联表按 `<left>_<right>` 排列，例如 `user_api_keys`。
3. 主键统一使用项目既有类型和字段名；新表默认沿用现有 `id` 约定，不在同一项目中引入 `uuid`、`key_id` 等第二套默认名。
4. 索引格式为 `idx_<table>_<columns>`，唯一索引为 `uniq_<table>_<columns>`，外键为 `fk_<table>_<referenced_table>`。
5. 迁移文件使用项目现有编号格式；描述部分使用小写 snake_case，例如 `20260805_add_upstream_version.sql`。
6. 迁移描述必须表达动作和对象，避免 `update_schema.sql`、`fix.sql` 等模糊名称。

## 10. 命名检查清单

提交或创建 PR 前确认：

- [ ] 分支类型、起点和 PR 目标符合 Git 工作流。
- [ ] 分支、文件、镜像和 Workflow 名称没有空格、临时词或无意义编号。
- [ ] Commit 标题符合 Conventional Commits，且一个提交只有一个主要目的。
- [ ] 版本 Tag、VERSION、UPSTREAM_VERSION 和镜像 Tag 一致。
- [ ] API、JSON、配置和数据库字段没有混用命名风格。
- [ ] 新增 Secret 没有出现在源码、日志、Commit、PR 或镜像 Tag 中。
- [ ] 破坏性改名已经写明兼容或迁移方案。

### 快速参考

```text
分支:     feature/api-key-expiration
Commit:   feat(version): add AIWeLink release version parser
Tag:      v0.1.176-1
镜像:     docker.aiwelink.cc/sub2api-aiwelink-dev:0.1.176-1
Go:       backend/internal/versioninfo/version.go
Vue:      frontend/src/components/common/VersionBadge.vue
API:      /v1/api-keys, upstream_version
环境变量: UPSTREAM_VERSION
```
