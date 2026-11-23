# Gallery 开发者 API

> 自动生成，请勿手改；从运行时 PAT 声明和产品 OpenAPI 派生。

[使用指南](../developer-tokens.md) · [机器能力目录](developer-api.json) · [PAT OpenAPI](developer-openapi.json)

本文件描述当前源码允许的范围。生产部署：**未知**；真实调用验证：**未知**。实际生效情况查产品部署记录。

请求携带 `Authorization: Bearer <PAT>` 与 `X-Yueli-Token-Site: <实际站点 ID>`。
同一分支内 scope 全部需要，分支之间任选其一；仍须通过当前账号与资源权限检查。目录不是授权凭证。

## 可申请权限

| Scope | 名称 | 当前业务能力 | 说明 |
| --- | --- | --- | --- |
| gallery.case.read | 查看处理单 | gallery.case.read | 读取纠错和举报处理单；仍需当前账号拥有对应权限。 |
| gallery.case.resolve | 处置处理单 | gallery.case.resolve | 推进、解决或驳回纠错和举报处理单；仍需当前 Gallery 管理员权限。 |
| gallery.classification.govern | 治理分类体系 | gallery.classification.govern | 创建、修改和执行分类治理计划；仍需当前 Gallery 管理员权限。 |
| gallery.classification.proposal_review | 审核标签提案 | gallery.classification.proposal_review | 处理投稿产生的标签提案；仍需当前账号拥有对应权限。 |
| gallery.classification.read | 查看分类体系 | gallery.classification.read | 读取完整分类与标签目录；仍需当前账号拥有对应权限。 |
| gallery.collection.manage | 管理专题集合 | gallery.collection.manage | 创建、编辑和编排专题；仍需当前账号拥有对应权限。 |
| gallery.collection.read | 查看专题集合 | gallery.collection.read | 读取私有和公开专题；仍需当前账号拥有对应权限。 |
| gallery.comment.create | 发表评论 | gallery.comment.create | 以当前账号在公开图片下发表评论或回复。 |
| gallery.comment.delete | 删除评论 | gallery.comment.delete | 删除评论及其回复；仍需当前账号拥有对应权限。 |
| gallery.comment.moderate | 审核评论 | gallery.comment.moderate | 修改评论审核状态；仍需当前账号拥有对应权限。 |
| gallery.comment.read | 查看评论后台 | gallery.comment.read | 读取评论审核队列；仍需当前账号拥有对应权限。 |
| gallery.dashboard.read | 查看运营概览 | gallery.dashboard.read | 读取 Gallery 运营概览；仍需当前账号拥有对应权限。 |
| gallery.discovery.manage | 管理发现与站点设置 | gallery.discovery.manage | 修改站点公开信息与发现策略；仍需当前 Gallery 管理员权限。 |
| gallery.discovery.read | 查看发现策略 | gallery.discovery.read | 读取后台站点与发现设置；仍需当前账号拥有对应权限。 |
| gallery.favorite.manage | 管理自己的收藏 | gallery.favorite.manage | 向当前账号收藏中添加或移除图片。 |
| gallery.favorite.read | 读取自己的收藏 | gallery.favorite.read | 读取当前账号的私有收藏。 |
| gallery.image.hide | 下架或删除图片 | gallery.image.hide | 下架、删除或批量隐藏图片；仍需当前账号拥有对应权限。 |
| gallery.image.read | 查看图片后台 | gallery.image.read | 查询后台图片和详情；仍需当前账号拥有对应权限。 |
| gallery.image.update | 编辑图片 | gallery.image.update | 编辑图片元数据与分类；发布等动作仍需额外当前权限。 |
| gallery.submission.create | 创建投稿 | gallery.submission.create | 以当前账号提交已经上传的图片，仍受分类、反滥用和处理规则限制。 |
| gallery.submission.read | 查看投稿审核队列 | gallery.submission.read | 读取全站投稿与预览；仍需当前账号拥有对应权限。 |
| gallery.submission.read_own | 读取自己的投稿 | gallery.submission.read_own | 列出当前账号拥有的投稿和处理状态。 |
| gallery.submission.review | 审核投稿 | gallery.submission.review | 批准或拒绝投稿；仍需当前账号拥有对应权限。 |
| gallery.submission.withdraw | 撤回自己的投稿 | gallery.submission.withdraw | 撤回当前账号拥有的投稿，并沿用现有状态限制。 |
| media.upload | 上传投稿图片 | gallery.submission.create | 只上传 Gallery 投稿原图；还需选择创建投稿权限。 |

## 媒体 Profile

| Namespace / Profile | 所需 Scope |
| --- | --- |
| gallery / gallery-submission | `media.upload` |

## 允许调用的操作

| 方法 / 路由 | 操作 | 所需 Scope |
| --- | --- | --- |
| `GET /api/v1/gallery/admin/cases` | List and filter Gallery moderation and correction cases | `gallery.case.read` |
| `POST /api/v1/gallery/admin/cases/{caseId}/resolve` | Move a Gallery case to reviewing, resolved or dismissed | `gallery.case.resolve` |
| `GET /api/v1/gallery/admin/classification` | Get the authoritative classification catalog including inactive identities | `gallery.classification.read` |
| `POST /api/v1/gallery/admin/classification/governance/execute` | Execute a freshly previewed classification governance plan | `gallery.classification.govern` |
| `POST /api/v1/gallery/admin/classification/governance/preview` | Preview an immutable classification governance plan | `gallery.classification.govern` |
| `POST /api/v1/gallery/admin/classification/identities` | Create a classification category, facet, facet value or tag | `gallery.classification.govern` |
| `PATCH /api/v1/gallery/admin/classification/identities/{identityId}` | Update classification identity fields, icon and cover | `gallery.classification.govern` |
| `GET /api/v1/gallery/admin/classification/tag-proposals` | List tag proposals awaiting or completing governance | `gallery.classification.read` |
| `POST /api/v1/gallery/admin/classification/tag-proposals/{proposalId}/review` | Resolve, create, alias or reject a tag proposal | `gallery.classification.proposal_review` |
| `GET /api/v1/gallery/admin/classification/tags` | List classification tags with a stable keyset cursor | `gallery.classification.read` |
| `GET /api/v1/gallery/admin/collections` | List private and public editorial collections | `gallery.collection.read` |
| `POST /api/v1/gallery/admin/collections` | Create an editorial image collection | `gallery.collection.manage` |
| `GET /api/v1/gallery/admin/collections/{collectionId}` | Get one editorial collection with ordered members | `gallery.collection.read` |
| `PATCH /api/v1/gallery/admin/collections/{collectionId}` | Update editorial metadata, SEO, cover and visibility | `gallery.collection.manage` |
| `POST /api/v1/gallery/admin/collections/{collectionId}/members` | Atomically add and remove editorial collection images | `gallery.collection.manage` |
| `PUT /api/v1/gallery/admin/collections/{collectionId}/order` | Replace the manual order for all editorial collection members | `gallery.collection.manage` |
| `GET /api/v1/gallery/admin/comments` | List image comments for moderation | `gallery.comment.read` |
| `DELETE /api/v1/gallery/admin/comments/{commentId}` | Delete an image comment and its replies | `gallery.comment.delete` |
| `PATCH /api/v1/gallery/admin/comments/{commentId}` | Moderate an image comment | `gallery.comment.moderate` |
| `GET /api/v1/gallery/admin/images` | List all Gallery images for lifecycle operations | `gallery.image.read` |
| `POST /api/v1/gallery/admin/images/bulk` | Apply a lifecycle or classification action to multiple images with per-item outcomes | `gallery.image.hide` 或 `gallery.image.update` |
| `POST /api/v1/gallery/admin/images/bulk-hide` | Hide multiple published images with per-item outcomes | `gallery.image.hide` |
| `DELETE /api/v1/gallery/admin/images/{imageId}` | Mark a Gallery image as deleted with optimistic concurrency | `gallery.image.hide` |
| `GET /api/v1/gallery/admin/images/{imageId}` | Get a reviewed Gallery image with editable classification | `gallery.image.read` |
| `PATCH /api/v1/gallery/admin/images/{imageId}` | Optimistically update Gallery image metadata | `gallery.image.hide` 或 `gallery.image.update` |
| `POST /api/v1/gallery/admin/images/{imageId}/hide` | Hide a published image | `gallery.image.hide` |
| `GET /api/v1/gallery/admin/overview` | Get actionable Gallery operations counts | `gallery.dashboard.read` |
| `GET /api/v1/gallery/admin/site-settings` | Get editable Gallery site and homepage settings | `gallery.discovery.read` |
| `PATCH /api/v1/gallery/admin/site-settings` | Update Gallery site and homepage settings | `gallery.discovery.manage` |
| `GET /api/v1/gallery/admin/submissions` | List Gallery submissions across processing, review, safety and outcome states | `gallery.submission.read` |
| `POST /api/v1/gallery/admin/submissions/bulk-review` | Approve or reject multiple ready submissions with per-item outcomes | `gallery.submission.review` |
| `GET /api/v1/gallery/admin/submissions/{submissionId}/preview` | Get a short-lived signed submission thumbnail URL | `gallery.submission.read` |
| `POST /api/v1/gallery/admin/submissions/{submissionId}/review` | Approve or reject a ready submission | `gallery.submission.review` |
| `GET /api/v1/gallery/collections` | List public editorial collections | 无额外 scope |
| `GET /api/v1/gallery/collections/{slug}` | Get a public editorial collection | 无额外 scope |
| `GET /api/v1/gallery/discovery` | Get a stable seeded random image batch | 无额外 scope |
| `GET /api/v1/gallery/images` | List eligible public images with page-based filtering | 无额外 scope |
| `GET /api/v1/gallery/images/by-key/{key}` | Get an image by its eight-character public URL key | 无额外 scope |
| `GET /api/v1/gallery/images/{imageId}` | Get one eligible public image | 无额外 scope |
| `GET /api/v1/gallery/images/{imageId}/comments` | List approved comments for a public image | 无额外 scope |
| `POST /api/v1/gallery/images/{imageId}/comments` | Comment on a public image | `gallery.comment.create` |
| `GET /api/v1/gallery/images/{imageId}/related` | List explainable related public images | 无额外 scope |
| `GET /api/v1/gallery/me/favorites` | Get the current user's private singleton favorites collection | `gallery.favorite.read` |
| `DELETE /api/v1/gallery/me/favorites/{imageId}` | Idempotently remove an image from favorites | `gallery.favorite.manage` |
| `PUT /api/v1/gallery/me/favorites/{imageId}` | Idempotently add an image to favorites | `gallery.favorite.manage` |
| `GET /api/v1/gallery/me/submissions` | List submissions owned by the current user or guest | `gallery.submission.read_own` |
| `POST /api/v1/gallery/me/submissions/{submissionId}/withdraw` | Withdraw an owned submission and hide its published image | `gallery.submission.withdraw` |
| `GET /api/v1/gallery/rankings` | Get a cached or aggregated public image ranking | 无额外 scope |
| `GET /api/v1/gallery/site` | Get public Gallery site and homepage settings | 无额外 scope |
| `GET /api/v1/gallery/submission-options` | Get active classification choices for a submission | 无额外 scope |
| `POST /api/v1/gallery/submissions` | Submit one uploaded image | `gallery.submission.create` |
| `POST /api/v1/personal-token/media-authorization` | Authorize constrained Gallery submission upload for the current personal token | `media.upload` |

字段、请求体、响应与引用 Schema 见生成的 OpenAPI。未列入的接口不代表允许 PAT 调用。
