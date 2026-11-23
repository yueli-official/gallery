# 2026-11-24 图库筛选和手机投稿部署

「我的筛选」可保存在本浏览器中并支持离线保存。手机投稿队列采用三列缩略图。

| 服务 | 版本 | 运行镜像 |
| --- | --- | --- |
| Web | `gallery-compact-20261124-1` | `sha256:b4ef913c20d39980baf3c6dbb83d56548f9036d17a1cdb05f8784355b5c507f7` |
| API | `image-information-20261115-1` | `sha256:2306b167bdfaf133f2d74c8a18fbcca61fa23b79e677b4a02f48af8c8d934478` |

候选和公网通过浏览器验证离线保存、刷新恢复、删除筛选以及手机一屏显示六张缩略图。账号同步和权限错误使用测试会话验证。公网加载的 108 个脚本及样式与制品摘要一致。此轮切换 `gallery-web`，复用原 API 和数据库。

生产回执位于 `/projects/yuelili.com/gallery/deployment-release.json`。发布制品与验证证据在 `/projects/yuelili.com/.deploy/gallery-gallery-compact-20261124-1/`。恢复配置与前版回执在 `/projects/yuelili.com/backups/gallery-gallery-compact-20261124-1/`，前版 Web 镜像为 `yueli/gallery-web:traffic-charts-20261124-2`。
