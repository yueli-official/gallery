# Gallery WebP 存储与 PLN 迁移

2026-10-25，正式入口：https://gallery.yuelili.com。

同日后续已发布[图片加载与令牌性能修复](performance-20261025.md)。当前运行版本以该记录和服务器根发布清单为准；下文制品表保留迁移阶段的版本。

## 存储与分类

- 主分类：游戏、动漫、原创；独立维度：作品、角色。方向、比例与尺寸由文件计算。默认不创建分类封面。
- 674 份有效 PLN 来源归入游戏 / 碧蓝档案 / 普拉娜。排除 674 preview、674 thumbnail 和 SQLite 中 3 条软删除记录。
- WebP 转换后两组相同内容（来源 302/596、433/559）合并，最终 672 张图片、672 个 Asset，保留全部来源映射。
- 主图高质量 WebP，最长边 4096，等比缩小、不放大小图、不另存原文件。38 份来源缩小，81 份透明图片保留透明通道。
- 最终主图 298,707,866 bytes；来源图 1,157,956,119 bytes，减少约 74%。公开预览仍由 Asset 按注册规格生成。
- WebP 下载返回已存文件；JPG 在浏览器按需转换，透明区域铺白底，不持久存 JPG 副本。
- 旧 PLN uploads 的 2,022 个文件保留不动；有效来源 SHA256 已核对。旧站目录尚未释放磁盘空间。

## 正式制品

| 组件 | 镜像 | 镜像 ID |
| --- | --- | --- |
| Gallery API | `yueli/gallery-api:webp-20261025-02` | `sha256:640356501d220d5089ffb20bea914ba4591505dc1842dd2b7cbf1f10942f1863` |
| Gallery Web | `yueli/gallery-web:webp-20261025-01` | `sha256:b36bdfb619fbeefb2a668f49f25301b6f78ab6e848752091f4f84be7149eb4f3` |
| Asset | `yueli/asset:webp-prepared-20261025-02` | `sha256:f1f4756a74b816f47f5dbb564e71dfd6c7e33db0992fb1df728ae5ddab14689a` |

Gallery 归档及回执位于服务器 `gallery/releases/webp-20261025-01`、`gallery/releases/webp-20261025-02`；根 `gallery/deployment-release.json` 为最终版本。第二次发布以原完整来源归档及两个 SHA256 固定的替换文件描述来源，复用 Web 制品。

Asset 从实际生产来源恢复，仅叠加 JPEG 尺寸读取、注册固定 WebP 目标和完整校验后的已处理 WebP 字节复用，沿用 Go 1.26.10；没有带入 checkout 的其他未发布功能。`asset/deployment-release.json` 保存最终来源归档、二进制、镜像及运行回执。服务继续 API-only，用户明确批准本次无网页构建回执的发布例外；没有提高输入像素限制或改变共享运行配置。

前端依赖基线：`6ef835526d6e357e14a638594c1ae3fc03826d53fff61a05f5fd7c3973892199`。

## 验证与收尾

- Asset Go tests/vet、生产补丁真实 PostgreSQL 上传流程通过：长元数据 JPEG、尺寸限制、固定 WebP 目标、不留原文件、干净 WebP 字节复用、伪造元数据清理、损坏像素拒绝。
- Gallery 前端 73 tests/typecheck/build、相关 Go tests/vet 通过；临时真实 PostgreSQL 探针验证 672 个完成发布不再补齐，一个未完成发布只处理一次，以及既有发布/隐藏事务。临时测试已从仓库删除。
- 浏览器验收覆盖实际 WebP/JPG 下载与尺寸、作品/角色筛选，以及网络拦截的完整上传：5000×300 PNG 变为 4096×246 WebP，透明区域保持；未生成生产测试投稿。
- 手动查看桌面、手机、浅色详情、投稿页截图与正式列表。正式列表为 672 张。
- 首次发布通过 8 项 HTTPS 检查及 73 份 JS/CSS 内容摘要匹配。最终 API 修复保持同一 Web 镜像，API/Web healthy，数据库容器和运行环境不变，schema migrations 14 clean。
- 正式下载复查发现旧补齐任务重放整个合集，触发 Asset 每 IP 限流；已修正只查未完成发布，保持限流设置，正式 WebP/JPG 下载复验通过。
- 672 个最终 Asset 的实际主图字节哈希和事实全部匹配，无原文件归档；259 个迁移旧副本经引用检查及 SHA256 固定计划清理。
- 专用迁移客户端、Asset binding、临时凭据已撤销。普通 Gallery 客户端保留。验证后清理临时转换文件、构建目录和候选服务，保留来源、回执、备份和正式制品。

私有回执位于 `/projects/yuelili.com/gallery/imports/pln-20261025-01`，包含逐来源映射、转换和最终审计、旧副本清理计划、凭据撤销记录。

备份在 `/projects/yuelili.com/backups/`：`gallery-pln-import-20261025-01`、`asset-webp-ingest-20261025-01`、`asset-webp-prepared-20261025-02`、`gallery-webp-20261025-01`、`gallery-webp-20261025-02`，数据库 dump 已验证可读取清单。

代码保持本地未提交；本轮授权不包含 commit、push 或分支切换。
