# gallery 资产引用生命周期

## Status
Finished

## Goal
补齐 gallery 自有业务素材登记、替换/删除撤销与失败恢复，并如实提交注册声明。

## Current
2026-09-11 本产品已锁定正式 Asset Go v0.4.0；独立工作树无 replace/源码覆盖完成全量 Go 测试、命令构建、go vet 及真实 PostgreSQL 引用事实源回归。业务查询与引用生命周期代码未修改。以下源码候选阶段记录保留，不能理解为本轮重新部署。
2026-09-11 当前本地源码组合 20260911T064430Z-57124 已启用；125 条图片、48 条投稿、6 个头像及 6 张封面已对账，原 85 条公开授权保留。12 张隐藏图片全部保留引用，4 张既有删除图片没有公开授权残留。
真实 API 上传/投稿/撤回通过；另新建投稿经审核发布后，隐藏保留图片引用，删除撤销全部业务引用与公开授权，素材可删除。仅本次测试素材已清理，测试投稿保留业务系统的撤回/删除记录。
CLI Playwright 桌面/移动端引用弹窗、删除保护、布局通过；图库定向 health 全部 ready。修复启动脚本最终健康检查错误覆盖其他组合的问题。未发布 SDK、未部署、未提交。

## Next
None。本站本地验收完成。

## References
- [上下文](context.md)
- [验证和接入](validation.md)
- [共享 SDK 与待接线补丁](../../../../asset/flightdeck/work/2026-09-11-reference-source-sdk/index.md)

## Git 交付（2026-09-11）
用户已授权本地提交。本次纳入本 Work 的实现、相关验证和部署记录；其他工作改动保留，未推送或发布正式 SDK。此前“未提交”为对应阶段的历史状态。提交及范围汇总见 Workspace `flightdeck/work/2026-09-11-target-stop-isolation/commits.md`。

## 正式 SDK 依赖升级

Go 模块改用 `github.com/yueli-official/asset v0.4.0` 并更新 go.sum，依赖图所需的传递版本由 go mod tidy 收敛。真实 PostgreSQL 回归使用事务/连接局部临时表，不变更业务数据。证据目录：`E:/tmp/yueli-asset-sdk-upgrade-20260911/`，本产品的 `gallery-results.json` 与 `gallery-reference-db.log`。此次只升级后端依赖，无页面或业务代码修改；沿用此前 CLI Playwright 的界面验收，不将其称为正式依赖运行中的新页面验收。未重新部署、未推送产品分支。
