# Project Guidelines

## Scope

Wot Starter 是基于 uni-app、Vue 3 和 Wot UI 的跨平台起手项目。优先做最小改动，保持现有页面、路由、组件和构建行为稳定。

## Architecture

- 页面代码位于 `src/pages` 和 `src/subPages`。
- 通用组件位于 `src/components`；状态位于 `src/store`；请求封装位于 `src/api`。
- `pages.config.ts` 与 `@uni-helper/vite-plugin-uni-pages` 负责页面生成；生成文件不要手工维护重复来源。
- `.agents/skills/` 保存给 Agent 使用的技能说明。

## Build And Test

- 安装依赖：`pnpm install`
- 代码检查：`pnpm lint`
- 类型检查：`pnpm type-check`
- H5 构建：`pnpm build:h5`

## OpenSpec Workflow

涉及新功能、用户可见行为变化、API 或数据契约变化、新页面/路由、公共组件/主题调整、跨端兼容处理或跨模块重构时，必须先使用 OpenSpec 建立变更：

1. 阅读 `openspec/specs/` 中相关规范。
2. 使用 `.agents/skills/openspec-propose` 创建并评审变更提案。
3. 使用 `.agents/skills/openspec-apply-change` 实现代码。
4. 使用 `.agents/skills/openspec-verify-change` 核对规格、任务、实现和测试。
5. 确认无误后使用 `.agents/skills/openspec-archive-change` 归档。

纯拼写、格式化、明显的单点修复和不改变行为的依赖维护可以不创建变更，但仍需说明验证结果。

OpenSpec 的项目上下文和 artifact 规则位于 `openspec/config.yaml`；工作流以 CLI 返回的路径和状态为准，不手工假设 artifact 位置。

项目已锁定 `@fission-ai/openspec@1.9.0`。运行 `pnpm install` 后可使用 `pnpm exec openspec`；共享 OpenSpec Skills 调用的是裸命令 `openspec`，因此 Codex/Claude Code 环境还需要将同一版本安装到 `PATH`：

```bash
npm install -g @fission-ai/openspec@1.9.0
openspec --version
```

## Conventions

- Vue 组件使用 `<script setup lang="ts">` 和 Composition API，保持严格类型。
- 网络流程必须考虑加载、成功、空数据、失败和恢复/重试状态。
- 不得把凭据写入仓库；不得默认使用只在浏览器可用的 API 破坏小程序端。
- 修改 Wot UI、主题、API 或模板相关代码时，优先读取 `.agents/skills/` 中对应的 `SKILL.md`。
- 未经明确要求，不要顺手重构无关模块。
