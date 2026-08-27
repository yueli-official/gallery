# Gallery 后台集合与控制台重构计划

## P0 — 基线与差距

- [x] 对照 Blog/Docs Dashboard、CollectionPanel、ManageSortHeader 与行操作实现。
- [x] 盘点 Gallery 管理页面、当前查询参数、排序白名单、编辑 API 和权限边界。
- [x] 建立 dashboard、图片表格和 `sortBy/sortOrder` 的红色回归合同。

## P1 — 查询与编辑合同

- [x] 图片查询统一为 `sortBy/sortOrder`，不再从 API/UI 发出旧参数。
- [x] 复用既有图片可编辑字段、服务端校验和权限，并同步 OpenAPI 与前端类型。
- [x] 行动作提供编辑和打开公开页，危险治理动作保持低频入口。

## P2 — 控制台

- [x] 导航与标题统一为“控制台”。
- [x] 使用真实 Gallery 日聚合提供浏览摘要、趋势和热门图片；无数据时显示明确空态。
- [x] 删除“需要你处理”“继续工作”等非 Dashboard 主任务内容。

## P3 — 图片集合

- [x] 使用单栏 CollectionPanel 工具栏和表格式主视图。
- [x] 列标题点击切换服务端排序并明确当前方向；分页、筛选、选择与批量治理保持稳定。
- [x] 完成图片编辑、公开页快捷入口、桌面/移动响应式与状态验收。

## P4 — 其他集合与收口

- [x] 投稿、处理单、专题及其他增长型集合统一单栏工具栏和表格/列表操作语言；分类继续使用适合治理任务的 Tabs。
- [x] 运行 API/Web 全量测试、typecheck、OpenAPI 校验、detector 和真实登录 Playwright。
- [x] 明暗主题与 390/768/1024/1440 验收完成后更新 Flightdeck 并关闭 Work。
