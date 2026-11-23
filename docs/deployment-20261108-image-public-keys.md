# Gallery 图片公开短地址发布

2026-11-08 已部署 `image-public-keys-20261108-1`。用户明确授权当前 Gallery 全部已有改动一同纳入并迁移图片；未 commit、push 或切换 Git 分支，未发布其他站点或重建共享 Provider。

## 已上线结果

图片网页使用 Foundation CompactURLV1 的 8 位、大小写敏感 Base58 键，持久保存在 public_id。673 张已有图片全部回填，唯一键 673、无效键 0；UUID 主键、674 条投稿、673 条分类关联和 1346 条 Facet 关联保持。迁移前后逐行图片摘要和关联摘要一致。卡片、分享、前后导航、相关推荐、投稿和后台查看使用公开键。旧 UUID/22 位网页地址直接 404，不保留别名、跳转或兜底；内部 API 资源 ID 继续保持。

用户原图的新地址：[okWbLqZR](https://gallery.yuelili.com/images/okWbLqZR)。作品与角色父子关系保持。

迁移 clean 23 → clean 24，回填重跑 0。引用地址通过共享 referencesync.Change 登记新版本和任务，673 项均已应用，旧地址 0、待应用 0；沿用实体引用和资源关联。生产现有配置使用相对短路径 `/images/<publicId>`。没有重跑 devseed、初始化管理员或创建公开测试图片。

## 制品与验证

| 服务 | 镜像 | 实际 image ID |
| --- | --- | --- |
| API | `yueli/gallery-api:image-public-keys-20261108-1` | `sha256:dcb64d5aaeab4af69174eb965ad5d78ab0a002654857d2744284cc1999066799` |
| Web | `yueli/gallery-web:image-public-keys-20261108-1` | `sha256:001a2d1f4257ad453370253b70f19b7245a445363c185906b4a5109c2f421e43` |

API/Web 均 healthy、RestartCount=0，配套切换停写入阶段实测 22.55 秒。Compose 语义比较确认只修改这两个镜像引用。当前生产源码归档和冻结 Go 依赖上下文先核对 SHA；现有 SQL 迁移原始字节保留，仅增加 0024。Web 冻结安装复用 768 包、下载 0，耗时 19m7.4s；共享依赖门禁、73 项前端测试、类型检查、相关 Go 检查及双端正式构建通过。批准基线 `b896e8efe51b901bf8877e580cfa0af10de62e8ebe0aac6d2573584ad98f63e5`。

来源快照 `fca077305b147b6644110a4d265fb60156f43636a6d6db2e39d0b3557e419351`；完整源码归档 SHA256 `6bac887aae619456e2470f0e2620d722e8f04361d3f2af34bba3f4f72f8bfa59`。来源包括已授权的现有首页文件、导入工具、部署校验、文档与油猴改动；没有将相邻源码依赖用于正式构建。上传共 91,969,803 字节、200.5 秒，没有重复上传嵌套源码归档。

生产数据库副本候选先完成真实迁移及原字段/关联摘要核对。候选与生产的 API/SSR/BFF、135 个 JS/CSS 摘要、实际 WebP 媒体、匿名后台登录门禁通过。CLI Playwright 在候选和真实 HTTPS 验证短地址、分享、下一张、相关推荐和旧地址 404，无 pageerror 或 5xx；已亲自查看桌面和手机运行截图。未创建生产测试投稿，后台写入和真实 Tampermonkey 扩展安装不属于本次覆盖。

配套油猴通用版和个人安装版为 1.3.0，使用 imagePublicId 生成成功链接；此前 1.2.1 的 WebP 400 修复保留。个人文件只在用户 .secret/yueli，未提交凭据或覆盖用户已安装的扩展。

## 回执和恢复

- 当前真值：服务器 `/projects/yuelili.com/gallery/deployment-release.json`。
- 正式来源归档及发布记录：`/projects/yuelili.com/gallery/releases/image-public-keys-20261108-1/`。
- 制品与备份：`/projects/yuelili.com/.deploy/image-public-keys-20261108-1/`。
- 数据库备份：`backup/production-before.dump`，1,195,825 字节，已验证可读；SHA256 `22e8b4332ca5267f5c0928e5e96727c00c4affc53436e35d49fd6e672d4d9237`。
- 原 Compose 和生产回执在同一 backup 目录。新增 NOT NULL 后旧应用不能直接恢复写入；恢复须显式评估配套镜像、配置和完整数据库备份，不自动降级 schema。
- 本地证据：`D:/coding/yueli-official/task-artifacts/gallery/image-public-keys-20261108/`，包含迁移、备份、候选/生产 HTTP、浏览器和运行身份回执。

候选容器、副本数据库、临时维护环境和展开构建目录已停止/清理；来源归档、发布制品、备份和必要证据保留。
