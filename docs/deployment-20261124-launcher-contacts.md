# 2026-11-24 站群入口与联系方式部署

页头九宫格放在账号入口左侧，已有主题按钮时放在主题按钮右侧。共享菜单保留站点官方图标，在底部提供 QQ 群、GitHub 和哔哩哔哩入口。

| 服务 | 版本 | 运行镜像 |
| --- | --- | --- |
| Web | `launcher-contacts-20261124-1` | `sha256:ef056d4d8142e4b886ce78eee2acabe69792507f50e930174bf92836ba1893fa` |
| API | `image-information-20261115-1` | `sha256:2306b167bdfaf133f2d74c8a18fbcca61fa23b79e677b4a02f48af8c8d934478` |

[正式站点](https://gallery.yuelili.com/)的桌面和手机浏览器检查通过，104 项静态资源摘要与发布制品一致。共享 UI 使用 `0.10.4-preview.20261124.contacts.1`。API 复用原镜像，数据库无需迁移。

生效回执位于 `/projects/yuelili.com/gallery/deployment-release.json`，制品与核验记录位于 `/projects/yuelili.com/.deploy/gallery-launcher-contacts-20261124-1/`。恢复配置及旧回执位于 `/projects/yuelili.com/backups/gallery-launcher-contacts-20261124-1/`，上一版 Web 镜像为 `yueli/gallery-web:gallery-picker-20261124-1`。
