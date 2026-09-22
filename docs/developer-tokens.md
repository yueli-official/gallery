# 开发者令牌 API

Gallery 复用 Identity 账户中心的个人访问令牌（PAT）。令牌只通过 `Authorization: Bearer pat_…` 使用，不换取浏览器 Cookie；每次请求都检查“令牌勾选权限 ∩ 当前账号在 Gallery 的实时权限”。撤销令牌或撤销运营角色后，下次请求立即失效。

## 权限

普通账号可选择：

| 权限键 | 用途 |
| --- | --- |
| `media.upload` | 通过 Asset 上传 `gallery-submission` 私有原图；还需选择创建投稿 |
| `gallery.submission.create` | 以当前账号创建投稿 |
| `gallery.submission.read_own` | 查询自己的投稿和处理状态 |
| `gallery.submission.withdraw` | 撤回自己的投稿 |
| `gallery.favorite.read` | 读取自己的私有收藏 |
| `gallery.favorite.manage` | 添加或移除自己的收藏成员 |
| `gallery.comment.create` | 以当前账号发表评论或回复 |

当前内容运营者还可按项选择后台概览、图片读取/编辑/下架、投稿读取/审核、专题读取/管理、分类读取、标签提案审核、处理单读取、发现策略读取和评论治理。当前管理员还可显式选择分类治理、处理单处置和站点/发现设置写入；scope 不会把这些权限授予普通账号。Asset 平台策略、授权控制台、角色申请/授予和 Guest Claim 不向 PAT 开放。

## 上传并投稿

API 完整定义见 [OpenAPI](../contracts/openapi/gallery.json)。下面示例中的令牌由调用者放在当前终端，不能写入仓库。开发客户端必须使用 Asset 返回的上传地址、请求头和 `uploadToken`，不得推导对象 Key 或存储后端。

```powershell
$gallery = 'http://127.0.0.1:8091'
$asset = 'http://127.0.0.1:8082'
$headers = @{
  Authorization = "Bearer $env:GALLERY_TOKEN"
  'X-Yueli-Token-Site' = 'gallery-main-web'
}
$file = Get-Item ./picture.webp

$initBody = @{
  filename = $file.Name
  mime = 'image/webp'
  size = $file.Length
  siteKey = 'gallery'
  profileKey = 'gallery-submission'
  category = 'gallery-submission'
  visibility = 'private'
  multipart = $false
} | ConvertTo-Json
$init = Invoke-RestMethod "$asset/api/v1/assets/upload-init" -Method Post -Headers $headers `
  -ContentType 'application/json' -Body $initBody

$uploadHeaders = @{}
$init.uploadHeaders.psobject.Properties | ForEach-Object { $uploadHeaders[$_.Name] = [string]$_.Value }
Invoke-WebRequest $init.uploadUrl -Method Put -Headers $uploadHeaders -InFile $file.FullName | Out-Null
$finalized = Invoke-RestMethod "$asset/api/v1/assets/finalize" -Method Post -Headers $headers `
  -ContentType 'application/json' -Body (@{ uploadToken = $init.uploadToken } | ConvertTo-Json)

# 先读取 submission-options，并按返回的分类约束填写分类与 Facet。
$options = Invoke-RestMethod "$gallery/api/v1/gallery/submission-options" -Headers $headers
$submissionBody = @{
  assetId = $finalized.asset.id
  title = '示例投稿'
  description = '由开发者令牌提交'
  categoryIds = @($options.categories[0].id)
  primaryCategoryId = $options.categories[0].id
  facets = @(@{
    facetId = $options.facets[0].id
    valueIds = @($options.facets[0].values[0].id)
  })
  tags = @()
} | ConvertTo-Json -Depth 6
$created = Invoke-RestMethod "$gallery/api/v1/gallery/submissions" -Method Post `
  -Headers $headers -ContentType 'application/json' -Body $submissionBody
```

Asset 只接受带 `X-Yueli-Token-Site: gallery-main-web` 的本站 PAT，并在线回查 Gallery 的媒体授权端点；最终只授予 `asset.profile.gallery-submission.upload`。上传成功不等于公开，Gallery 仍执行处理和审核流程。

常用入口：

- `GET /api/v1/gallery/me/submissions`：查询自己的投稿。
- `POST /api/v1/gallery/me/submissions/{submissionId}/withdraw`：撤回自己的投稿。
- `GET|PUT|DELETE /api/v1/gallery/me/favorites[/{imageId}]`：读取或管理收藏。
- `POST /api/v1/gallery/images/{imageId}/comments`：发表评论或回复。
- 后台端点沿用 OpenAPI；账号和令牌都必须拥有对应细粒度能力。

成功返回直接 DTO 或空的 204；失败返回 `application/problem+json`。权限不足为 403，无效或已撤销令牌为 401，资源归属、状态和版本冲突沿用现有 404/409 合同。携带失败 PAT 的请求不会回退浏览器登录状态。

## 本地接入

从 Workspace 运行：

```powershell
./environments/gallery-local/run.ps1 -Mode Isolated
```

默认入口：Gallery `http://gallery.dev.yuelili.test:3007`、Account `http://account-gallery.dev.yuelili.test:3000/developer-tokens`、Gallery API `http://127.0.0.1:8091`、Asset API `http://127.0.0.1:8082`。本地环境同时登记 Identity 权限目录、Gallery 在线 PAT 验证和 Asset 媒体授权回查。

独立部署需同时配置：

- Identity `pat.applications`：登记 `gallery-main-web`、权限目录 URL 和 audience。
- Gallery `gallery.personalToken.*`：site ID、Identity `/api/v1/pat/verify` 和可信 HTTP 策略。
- Asset `asset.personalTokens.authorities`：把 `gallery-main-web` 映射到 Gallery `/api/v1/personal-token/media-authorization`。

本地实现不代表生产已部署；使用生产 Origin 前必须先核对正式部署记录。
