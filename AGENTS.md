# Repository instructions
AGENTS version: 1.0.1

通用规则遵循用户级根 AGENTS.md。本文件保留 Gallery 的职责与恢复入口。

- When adjacent `../workspace` exists and its `repos.lock.yaml` includes this repository, read and follow
  `../workspace/docs/multi-project-development.md` before cross-repository work or any local process lifecycle action. Product
  sessions treat Workspace contracts and `.doctor/` as read-only, use the Workspace CLI for generated state, use a separate Git
  worktree for concurrent writes to this repository, and never own shared Provider processes.

## Flightdeck

长期工作使用已安装的 Flightdeck 技能，从 `flightdeck/deck.md` 恢复。
`flightdeck/` 仅作本地记录，`.gitignore` 保留 `/flightdeck/`；维护工作台不代表提交、推送或发布。
