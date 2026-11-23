# 2026-11-23 图库九宫格部署

[图库](https://gallery.yuelili.com/)已更新至 `site-launcher-20261123-1`。页头已采用共享九宫格。九个入口以浅色底和彩色图标呈现。

- 桌面和手机公网浏览器均通过九宫格验证，97 项脚本与样式的摘要与制品一致。两端运行镜像已核对，重启次数为 0。
- 本次无需数据库迁移。 API 复用生产镜像 `image-information-20261115-1`。
- API 镜像 `sha256:2306b167bdfaf133f2d74c8a18fbcca61fa23b79e677b4a02f48af8c8d934478`，Web 镜像 `sha256:037431de69a5c42f3379f4e69ec2d69815ddb95e20e18592184ea6c9a1554114`。
- 源码快照 `a63cb90fc6be6243095bcba2502226e91cd4551b90a8fff271af3565bc01dc01`。双端验收回执见 `/projects/yuelili.com/gallery/deployment-release.json`，恢复配置见 `/projects/yuelili.com/backups/gallery-site-launcher-20261123-1`。
