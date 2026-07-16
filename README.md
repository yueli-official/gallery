# 图片站产品

- 生命周期：活跃的中型产品实现
- 权威来源：Catalog 产品类型 `gallery`、[领域语言](./CONTEXT.md)、[公开图片站契约](../../flightdeck/knowledge/gallery/public-image-site-contract.md)和代码
- 当前实例：`gallery-main`；后续 Gallery 实例复用同一产品源码和独立实例库
- 当前工作记录：[Gallery 中大型图片库升级](../../flightdeck/work/2026-07-16-gallery-medium-large-upgrade/index.md)

Gallery 是运营方维护的公开图片发现站，支持稳定 URL 的目录搜索/分类/Facet/Tag 筛选、Grid/Masonry、共享 Quick View/详情 Viewer、相关图片、专题、排行、私人收藏、投稿记录和批量投稿。它不建立创作者主页、关注、评论或多图帖子。

管理端覆盖 Dashboard、图片生命周期、投稿审核、处理单、专题编排、发现设置、分类治理和 Asset delivery settings。处理、审核、安全、发布是独立状态；公开资格必须同时满足全部策略，隐藏/删除/安全阻断始终 fail closed。

## 边界

- `api/`：图片、投稿、审核、Collection/收藏、Case、分类投影、事件日聚合和排行快照。
- `web/`：公开发现/目录/Viewer、个人工作区、批量投稿和运营工作台。
- Asset：二进制上传、私有 master、公开 rendition/profile/variant 和撤销；Gallery 只保存 `assetId`。
- Media delivery：cacheable public object/rendition 不消耗 JSON API 的每 IP 请求 bucket，生产由 CDN/edge 控制带宽与滥用；上传、签名和受保护交付继续使用 API limiter。
- Identity：OIDC User/Guest 与权限；Gallery 不复制临时身份系统。
- PostgreSQL 18 + pgvector：Gallery 事实源和当前搜索 Adapter。外部搜索与 Redis 的触发器见工作主题的 [`scale.md`](../../flightdeck/work/2026-07-16-gallery-medium-large-upgrade/scale.md)。

## 本地运行

```bash
pnpm provision:gallery
pnpm dev:gallery
```

Local provision 幂等生成 248 个真实 Asset 对象、128 张图片、48 条投稿、20 个处理单和 4 个专题，并只重置共享 E2E 账户的 Gallery 可变工作区。生产 migration 不包含演示业务数据。

## 验证

```bash
go test ./products/gallery/api/...
pnpm --dir products/gallery/web test
pnpm --dir products/gallery/web typecheck
pnpm --dir products/gallery/web build
pnpm contracts:check
pnpm migration:check
pnpm performance:check

go run ./cmd/platformctl verify e2e \
  --file catalog/overlays/local.yaml --root . gallery-main
```

产品 browser contract、journey adapter 和 light/dark mobile/desktop 基准都在 [`web/test/e2e`](./web/test/e2e/)；根 Playwright harness 从 Catalog 展开 visual、WCAG 2.2 AA、双向键盘焦点、performance 和 Asset settings 合同。
