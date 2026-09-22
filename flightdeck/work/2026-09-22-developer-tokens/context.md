# 上下文与权限边界

用户要求按 BVideo、Paste（代码片段）、Gallery、Nav、Distribution 的方向让更多站点支持开发者令牌。仓库规则禁止建立笼统的多站推广 Work，因此本 Work 只交付 Gallery。

Gallery 使用 Identity 作为凭证权威，使用自身 Authorization Runtime 判断当前业务权限，使用 Asset 保存投稿原图。PAT 不能绕过对象归属、乐观并发、投稿状态、审核资格、反滥用和现有管理员授权。

新增普通用户可委派能力：

- `gallery.submission.create`：以当前账号创建投稿。
- `gallery.submission.read_own`：读取当前账号自己的投稿。
- `gallery.submission.withdraw`：撤回当前账号自己的投稿。
- `gallery.favorite.read`：读取当前账号的私有收藏。
- `gallery.favorite.manage`：增删当前账号的收藏成员。
- `gallery.comment.create`：以当前账号在公开图片下发表评论或回复。
- `media.upload`：仅在账号当前拥有投稿创建能力时出现，并只换取 `asset.profile.gallery-submission.upload`。

现有 Gallery 后台 capability 保持权威并可按项委派；管理员可显式委派分类治理、处理单处置和站点/发现设置写入，但 scope 本身不会产生管理员权限。Asset 平台策略、授权控制台、角色申请/授予、Guest Claim、`/api/v1/gallery/me` 与其他未声明路由默认拒绝 PAT。权限目录只接受 `identity-svc` 且带 `personal-token:permissions` scope 的服务身份。

本地 site/application ID 为 `gallery-main-web`，资源 audience 沿用 `gallery-main-web`。Workspace Environment 负责连接 Identity 权限目录、Gallery 在线验证与 Asset 媒体授权回查；产品不复制令牌存储或 Cookie 行为。
