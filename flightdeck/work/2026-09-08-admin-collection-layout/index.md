# gallery 后台集合布局

## Goal
将本站管理集合接入用户确认的共享分页、网格、评论和标题工具规则，保留权限与查询行为。

## Status
Finished

## Current
已接入共享分页与标题工具区。CLI Playwright 已验证 /manage/cases, /manage/collections, /manage/submissions，覆盖 390/1440 宽度，无 pageerror。本地类型检查和生产构建通过。

## Next
None

## References
- [共享规则](../../../../foundation/flightdeck/knowledge/frontend/compact-admin-collections.md)
- 本次浏览器证据：`E:/tmp/yueli-media-preset-20260908/gallery-all-admin-browser.json` 与 screenshots。

## Final verification
- 最终生产构建通过：`gallery-admin-build.log`；本轮仅推广后台布局，未重新验收支付等无关业务。
- 类型检查、CLI Playwright 页面与相关交互、改动文件空白检查通过。额外验收组合已通过 Workspace CLI 停止，原有共享服务与 Gallery / Blog 保留。
- [页面验收结果](references/browser-layout.json)；截图及交互日志保存在本轮外部制品目录。
- 专题成员页已额外验证标题搜索、网格选中态和两种宽度；Gallery 独立快照构建没有影响原有开发进程。共享网格另覆盖 3840px、分页首页末页与筛选取消。

## 无主操作标题栏修复（2026-09-09）

共享 PageHeader 按 actions 插槽是否存在选择网格列数。投稿审核原来固定三列产生 0px 空列加 24px gap；现在没有主操作时不留空列。CLI Playwright 对投稿审核、图片、专题三页验证 390/1024/1440，投稿审核工具右侧差值为 0，有按钮页面保持原位置；共享 UI 类型检查通过。仅本地，未更新在线制品。

[标题栏验收](references/header-actions-layout.json)。
