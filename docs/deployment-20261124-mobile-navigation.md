# 2026-11-24 图库移动导航部署

手机底栏提供「首页」「图片」「我的」。[「我的」](https://gallery.yuelili.com/me)是独立页面，站群九宫格留在右上角。

| 服务 | 版本 | 运行镜像 ID |
| --- | --- | --- |
| Web | `mobile-navigation-20261124-1` | `sha256:317c4431a92f297323b053519235cb6ffd91778a08600a3252d420189cdf8ed5` |
| API | `image-information-20261115-1` | `sha256:2306b167bdfaf133f2d74c8a18fbcca61fa23b79e677b4a02f48af8c8d934478` |

候选通过手机与桌面浏览器检查。公网复查了九宫格、手机导航和「我的」页面，104 项脚本及样式摘要与制品一致。本次切换 `gallery-web`，复用原 API，数据库无需迁移。

当前回执为 `/projects/yuelili.com/gallery/deployment-release.json`，制品和验证证据在 `/projects/yuelili.com/.deploy/gallery-mobile-navigation-20261124-1/`。恢复配置和前版回执位于 `/projects/yuelili.com/backups/gallery-mobile-navigation-20261124-1/`，前版 Web 镜像为 `yueli/gallery-web:site-launcher-20261123-1`。恢复时使用备份的 Web 镜像配置重建 `gallery-web`，核对公网后恢复对应回执。
