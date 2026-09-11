# gallery 资产引用生命周期

## Status
Finished

## Goal
补齐 gallery 自有业务素材登记、替换/删除撤销与失败恢复，并如实提交注册声明。

## Current
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
