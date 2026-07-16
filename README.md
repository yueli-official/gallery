# 图片站产品

- 生命周期：活跃的设计与实现样板
- 权威来源：Catalog 产品类型 `gallery`、当前 Gallery 工作主题和代码
- 消费者：`gallery-main` 与后续 Gallery 实例
- 验证：`pnpm platformctl verify product --file catalog/overlays/local.yaml --root . gallery`

Gallery 是由运营方维护的公开图片发现站，支持用户投稿和私人收藏。登录用户投稿可直接展示但允许下架；匿名或访客投稿需要审核。分类、标签和 facet 筛选消费统一分类模型。

`api/` 负责图片、投稿、审核、集合/收藏和指标；`web/` 负责分页网格、随机发现与管理体验。二进制接收和图片变体属于 Asset。当前共享 E2E 基线位于 `tests/e2e/__screenshots__/gallery`；产品工作流会把契约和基准截图迁移到 `web/test/e2e/`。
