# Gallery 生产开发者令牌配置

2026-11-03，用户明确授权对现有 Gallery API、共享 Identity、共享 Asset 补齐 PAT 配置并重建容器，复用当前镜像，保留其他站点配置。

## 变更

- Gallery API：设置 `GF_GALLERY_PERSONALTOKEN_SITEID=gallery-main-web`、`GF_GALLERY_PERSONALTOKEN_VERIFYURL=http://identity:8081/api/v1/pat/verify`、`GF_GALLERY_PERSONALTOKEN_ALLOWHTTP=true`。
- Identity：在 `GF_PAT_APPLICATIONS` 追加“月离图库”，ID/audience 均为 `gallery-main-web`，权限目录为 `http://gallery-api:8091/api/v1/internal/personal-token/permissions`；在 `NO_PROXY` 追加 `gallery-api`。
- Asset：在 `GF_ASSET_PERSONALTOKENS_AUTHORITIES` 追加 `gallery-main-web` → `http://gallery-api:8091/api/v1/personal-token/media-authorization`。

只更新环境配置。Compose 解析后的完整对象比较确认其余字段不变，已有应用与 authority 保持。Gallery → Asset → Identity 依次以 `docker compose up -d --no-deps --wait --wait-timeout 120 <精确服务名>` 生效。

| 服务 | 复用镜像 | 生效耗时 |
| --- | --- | --- |
| gallery-api | `yueli/gallery-api:single-image-20261027-1` | 7.007 秒 |
| asset | `yueli/asset-api:preview-readiness-20261028-3` | 7.540 秒 |
| identity | `yueli/identity-api:lightweight-site-auth-20261025-02` | 7.478 秒 |

三者实际 image ID 与变更前一致，均 healthy、restart 0。Web 镜像未变；没有安装、构建和制品上传，仅上传本次配置操作脚本。Gallery 数据库 `schema_migrations=22`、dirty=false，本次不执行迁移或修改业务内容。

## 验证与限制

- 三份候选 Compose 与生产配置均解析成功，语义差异仅包含声明的六个环境变量键。
- 核对运行容器实际环境中的 Gallery verifier、Identity 应用及代理排除项、Asset authority。
- Gallery 公网首页、Account 开发者令牌页面均 HTTP 200；Gallery 无效 PAT 请求为 401，Asset 无效 PAT 上传请求为 403。
- 未取得生产账号会话。首次浏览器连接超时；未验证真实账号权限展示、创建/编辑令牌、真实 PAT 上传或视觉效果。
- 本地 galleryauthz/runtime/appconfig 测试通过；controller 测试因现有依赖缺少 Asset referencesync/goframe、delivery 与 Foundation oauthclient 而无法编译，未改动依赖，也未以此替代线上验收。

## 回执与恢复

服务器 `bt` 上 `/projects/yuelili.com/.deploy/gallery-pat-20261103-1/receipt.json` 保存三份 Compose 前后 SHA256、原镜像 ID、实际容器 ID、生效耗时和验证边界。三个产品的 `deployment-release.json` 已追加本次 `configurationUpdates`，保留原制品与来源记录。

每个服务原 Compose、原部署回执及存在的 `.env`/`config/runtime.yaml` 备份在同任务目录的 `<gallery|identity|asset>/backup/`，目录仅服务端当前用户可访问。恢复时先核对当前配置 SHA256 仍匹配本次 afterSha256、镜像和后续变更，再恢复对应原 Compose，按精确服务重建；无需恢复数据库。备份含私密配置，不下载或提交。
