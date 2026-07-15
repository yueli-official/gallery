# Gallery product

- Lifecycle: active design/implementation tracer
- Authority: Catalog product type `gallery`, current Gallery work topic and code
- Consumers: `gallery-main` and future Gallery instances
- Verify: `pnpm platformctl verify product --file catalog/overlays/local.yaml --root . gallery`

Gallery is a public, operator-maintained image discovery site with optional
user submissions and private favorites. Logged-in submissions can publish
directly but remain removable; anonymous/guest submissions require review.
Category, tag and facet filtering consume the shared classification model.

`api/` owns images, submissions, moderation, collections/favorites and metrics;
`web/` owns the paged grid, random discovery and management experience. Binary
ingestion/variants belong to Asset, not Gallery. Product-owned E2E and
multi-resolution screenshots must land below this product when implemented.
