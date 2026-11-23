# 2026-11-24 图库统计图表部署

后台访问趋势采用共享图表，横轴显示简短日期，选中数据点时显示完整日期和数值。

| 服务 | 版本 | 运行镜像 |
| --- | --- | --- |
| Web | `traffic-charts-20261124-2` | `sha256:8fa60191820185acb6045e18f98485937a3e05d93c7273d121cc7f42dc9e6aaa` |
| API | `image-information-20261115-1` | `sha256:2306b167bdfaf133f2d74c8a18fbcca61fa23b79e677b4a02f48af8c8d934478` |

候选和公网后台用浏览器测试数据检查桌面与手机宽度、键盘选点和完整日期。公网加载的 108 个脚本及样式与发布制品哈希一致。此轮切换 `gallery-web`，复用原 API 和数据库。

生产回执位于 `/projects/yuelili.com/gallery/deployment-release.json`。发布制品与验证证据在 `/projects/yuelili.com/.deploy/gallery-traffic-charts-20261124-2/`。恢复配置与前版回执在 `/projects/yuelili.com/backups/gallery-traffic-charts-20261124-2/`，前版 Web 镜像为 `yueli/gallery-web:mobile-navigation-20261124-4`。
