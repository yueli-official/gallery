# Gallery 后台集合与控制台重构

## Goal

以站群管理后台 Knowledge 和 Blog/Docs 已验收实现为真值，重构 Gallery 控制台与运营集合页：控制台展示真实浏览趋势，图片及其他集合使用单栏工具栏和表格式布局，列标题通过 `sortBy/sortOrder` 服务端排序，并补齐图片编辑与公开页快捷入口。

## Status

Finished

## Current

Gallery 后台首项、面包屑和页面标题已统一为“控制台”。Dashboard 直接聚合 `gallery_image_metrics_daily`，提供 7/14/30 天浏览与收藏曲线、真实摘要和热门图片；旧待办卡片与继续工作区已删除。

图片、投稿、处理单和专题均使用共享单栏工具栏与表格式行。图片、投稿、处理单 API/UI 已统一为 `sortBy/sortOrder`，列标题直接切换方向；图片编辑与公开页动作位于行尾，真实保存、检索和恢复已通过。授权页遗漏的 ClientBoundary 已收口，设置保存回到 PageHeader 唯一入口。

## Next

None。

## Current execution

Completed。

## Progress

- API 全量测试、Web 62 项测试、Nuxt typecheck、OpenAPI `sortBy/sortOrder` 导出校验与 Impeccable detector 通过。
- Gallery management Playwright 9/9，通过全后台无 5xx/水合错误、站点设置、图片编辑、专题、资源与分类治理。
- 自定义桌面/移动、明暗 Playwright 验收确认零横向溢出、可见图片无裂图、Axe serious/critical 为零；Blog 进程未重启。

## Constraints

- 保持 Image 单图与已发布像素不可替换不变量；“编辑”只修改允许治理的元数据、分类与发布属性。
- 直接复用 Foundation `PageHeader`、`CollectionPanel`、`ManageSortHeader` 等深模块，不复制 Blog/Docs 的 Tailwind 外形。
- 不修改 Blog/Docs 产品代码或进程；它们只作为只读实现真值。
- 页面必须完成 390/768/1024/1440、明暗主题、真实登录和 CLI Playwright 验收。

## References

- [执行计划](plan.md)
- [工作上下文](context.md)
