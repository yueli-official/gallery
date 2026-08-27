# 工作上下文

## 用户决策

- 后台首项与页面名称统一为“控制台”，不再使用“今日”。
- 控制台以真实浏览指标和趋势为主，不用待办卡片、继续工作或说明性废话填充页面。
- 图片集合采用 Docs 式单栏工具栏与表格；搜索、筛选和视图入口在同一栏，排序直接点击列标题。
- 请求合同使用 `sortBy=<field>&sortOrder=asc|desc`，停止扩展 `direction` 等平行参数。
- 图片行需要编辑入口和打开公开图片页的快捷操作；其他集合页也要统一到同一操作语言。

## 领域边界

- Gallery 是运营维护的公共单图资料库，不展示创作者身份或社交关系。
- Image 像素发布后不可替换；允许编辑的范围必须来自现有治理合同或新增的显式元数据合同。
- Dashboard 只能展示真实事件或聚合数据；没有事实来源的增长和趋势不能伪造。
- Category、Facet、Tag、Collection、Submission 与 Case 继续保持独立概念。

## 设计与实现真值

- Workspace Knowledge 要求 Dashboard 使用真实 summary/trend，集合使用 `PageHeader + CollectionPanel`。
- 表格排序由共享 `ManageSortHeader` 发布 `sortBy/sortOrder`，筛选按 0/1/多项渐进披露。
- Blog 控制台提供 Traffic 趋势与排行组织；Docs 文档集合提供单栏 Toolbar、表格列标题排序和行快捷动作。
- Gallery 已使用 `YAdminConsoleLayout` 和共享反馈区，本工作不重新设计应用壳。

## 运行边界

- 当前 Gallery Attach 会话为 `3007/8091`，复用 Blog 已运行的 Identity/Account/Asset。
- Blog 进程和数据库不属于本工作；只允许读源码和健康状态。
- 当前共享 Asset 已幂等补齐 Gallery consumer registration 与 248 个本地 fixture，公开媒体返回 200。

## 已交付合同

- Dashboard `GET /admin/overview?days=7|14|30` 返回累计/当前/前期浏览、收藏、逐日序列和热门图片。
- 管理图片、投稿与处理单只发布 `sortBy/sortOrder`；OpenAPI 不再暴露这些端点的旧 `sort` 参数。
- 图片列表是唯一编辑入口，编辑不替换已发布像素；公开图片和公开专题行均提供新窗口快捷入口。
- 设置页保存位于 PageHeader，脏数据保护继续阻止误离开；不再显示重复底部保存 Dock。
