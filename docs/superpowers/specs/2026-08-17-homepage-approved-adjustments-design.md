# AIWeLink 首页已批准调整恢复设计

## 背景

`aiwelink-dev` 已经合并过首页文案、登录入口和浅色粒子网络调整，但后续提交 `047f76a8c feat(home): restore AIWeLink application homepage` 使用旧组件重建首页，覆盖了部分已合并行为。当前默认首页的导航、Hero 和底部 CTA 都把未登录访客发送到 `/register`，而紧凑首页已经使用 `/login`；注册关闭时，访客还会进入“注册已关闭”的死路。

恢复依据如下：

- `1434fbc86 feat(home): refine landing page messaging`
- `389138ec3 fix(home): route start actions to login`
- `dbb55d7ff fix(home): deepen light particle network`

导航入口未包含在 `389138ec3` 中，是本次补齐的同类遗漏。

## 目标

1. 将默认首页的导航、Hero 和底部 CTA 三个访客主入口统一到 `/login`。
2. 导航访客按钮显示现有翻译 `home.login`；Hero 和底部 CTA 保持“点击开始”。
3. 恢复已批准的中文首页文案和 Hero 描述最大宽度 `700px`。
4. 仅在浅色主题中把粒子、连接线和指针连接线透明度提高 30%，同时保留现有上限和深色主题表现。
5. 用组件测试和上游同步合同防止下一次同步再次引入旧首页。

## 非目标

- 不修改注册功能、growth registration 或注册开关。
- 不修改 `/r/:code` 推荐码识别、HTTPS 跳转或 SPA bypass。
- 不修改控制台红/金主题、首页结构、粒子聚集爆散阈值或双版本机制。
- 不引入官方 Sub2API 没有的后端业务逻辑。
- 不合并 PR #19 或过期的 PR #28。

## 期望行为

| 场景 | 导航 | Hero | 底部 CTA |
| --- | --- | --- | --- |
| 未登录访客 | 显示“登录”，进入 `/login` | 显示“点击开始”，进入 `/login` | 显示“点击开始”，进入 `/login` |
| 已登录用户 | 显示“控制台”，进入传入的 dashboard 路径 | 进入传入的 dashboard 路径 | 进入传入的 dashboard 路径 |

恢复的中文文案：

- Hero：`一个 API，连接所有主流模型。为开发者、团队、企业提供稳定、统一、透明的模型接入。`
- 使用场景标题：`从 Coding 到 科研 与 Agent 接入`
- 使用场景说明：`让模型成为工作流的一部分。`
- 最终 CTA：`现在，开始你的创造`

## 实现方案

首页入口继续使用组件现有的 `computed` 目标，不引入新路由辅助函数。三个组件只把访客分支从 `/register` 改为 `/login`，认证用户分支保持不变。导航文案直接复用 `home.login`，避免新增重复翻译。

粒子网络增加一个局部的 `LIGHT_VISIBILITY_MULTIPLIER = 1.3` 和统一的透明度换算函数。该函数仅在浅色主题乘以 1.3，并分别沿用连接线、普通粒子和玫红粒子的现有最大透明度，避免溢出或改变深色主题。

`deploy/tests/upstream-sync-contract-test.sh` 将检查三个默认首页组件都含 `/login` 且不含访客 `/register` 目标，并检查导航使用登录文案。这样以后恢复旧组件时，CI 会在同步合同阶段直接失败。

## 验证范围

- `HomepageContent.spec.ts`：三个访客入口、认证用户入口、中文文案和 Hero 宽度。
- `ParticleNetwork.spec.ts`：浅色透明度准确提高 30%，深色透明度不变，指针连接线仍受上限约束。
- `upstream-sync-contract-test.sh`：版本、官方布局、AIWeLink 主题和首页登录入口合同。
- 前端 `typecheck`、相关 ESLint、`git diff --check`。
- 浏览器验证 `/home -> 导航/Hero/底部 CTA -> /login`，覆盖桌面与移动视口，并检查页面、控制台和错误覆盖层。

## 发布策略

从最新 `origin/aiwelink-dev` 创建独立分支，提交后推送到 `origin`，创建 base 为 `aiwelink-dev` 的 ready PR。PR 不指定 reviewer、不请求 approval、不启用自动合并，也不向 `Attillo123` 发送申请。
