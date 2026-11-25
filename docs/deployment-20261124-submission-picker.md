# 2026-11-24 图库投稿选项部署

投稿缩略图的移除按钮居中。分类及作品角色超过三项时用可搜索的下拉。

| 服务 | 版本 | 运行镜像 |
| --- | --- | --- |
| Web | `gallery-picker-20261124-1` | `sha256:a23e1699d3d440f0a64694a7aeba3593b86e87174b1a8e367e87bd05b2343335` |
| API | `image-information-20261115-1` | `sha256:2306b167bdfaf133f2d74c8a18fbcca61fa23b79e677b4a02f48af8c8d934478` |

候选和公网通过浏览器验证不同宽度下的移除按钮、搜索作品和角色以及选择分类。本地用四项分类验证搜索下拉。公网加载的 107 个脚本及样式与制品摘要一致。此轮切换 `gallery-web`，复用原 API 和数据库。

生产回执位于 `/projects/yuelili.com/gallery/deployment-release.json`。发布制品与验证证据在 `/projects/yuelili.com/.deploy/gallery-gallery-picker-20261124-1/`。恢复配置与前版回执在 `/projects/yuelili.com/backups/gallery-gallery-picker-20261124-1/`，前版 Web 镜像为 `yueli/gallery-web:gallery-compact-20261124-1`。
