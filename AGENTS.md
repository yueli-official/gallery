# Repository instructions
AGENTS version: 1.0.2

通用规则遵循用户级根 AGENTS.md。本文件保留 Gallery 的职责与恢复入口。

- 本仓库拥有图库投稿、图片集合与审核业务；媒体存储、交付和删除保护使用 Asset，用户身份使用 Identity。

- When adjacent `../workspace` exists and its `repos.lock.yaml` includes this repository, read and follow
  `../workspace/docs/multi-project-development.md` before cross-repository work or any local process lifecycle action. Product
  sessions treat Workspace contracts and `.doctor/` as read-only, use the Workspace CLI for generated state, use a separate Git
  worktree for concurrent writes to this repository, and never own shared Provider processes.

## Flightdeck

长期工作使用已安装的 Flightdeck 技能，从 `flightdeck/deck.md` 恢复。
`flightdeck/` 仅作本地记录，`.gitignore` 保留 `/flightdeck/`；维护工作台不代表提交、推送或发布。

## 共用入口

- 共用规则、技能和知识的归属见 [Workspace 内容归属](../workspace/docs/knowledge-placement.md)；产品实现与特有操作留本仓库，不复制根流程。
- 代码部署使用 [Workspace 部署技能](../workspace/.agents/skills/yueli-deploy/SKILL.md)和[发布合同](../workspace/docs/deployment-policy.md)，服务参数与更新命令仍由本仓库拥有。
- 使用共享 UI 前，先读实际前端目录中已安装的 `node_modules/@yueli/ui/AGENTS.md`，按该包执行 inspect、search/docs；不拿邻仓新源码冒充已安装接口。
