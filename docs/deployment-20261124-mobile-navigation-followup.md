# 2026-11-24 图库手机导航调整

手机底栏中间的「投稿」直达投稿页。「我的收藏」仍可从「我的」打开。

候选和公网通过手机与桌面浏览器复查，104 项脚本及样式与制品摘要一致。仅更新 `gallery-web` 到 `mobile-navigation-20261124-4`，API 复用 `image-information-20261115-1`，数据库无需迁移。

当前回执为 `/projects/yuelili.com/gallery/deployment-release.json`，制品和证据位于 `/projects/yuelili.com/.deploy/gallery-mobile-navigation-20261124-4/`。恢复配置和前版回执位于 `/projects/yuelili.com/backups/gallery-mobile-navigation-20261124-4/`。恢复时使用 `yueli/gallery-web:mobile-navigation-20261124-1` 的备份配置重建 `gallery-web`，核验后恢复对应回执。
