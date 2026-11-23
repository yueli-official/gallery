# Gallery 图片信息补充与单图编辑发布

已上线：`image-information-20261115-1`。API/Web均 healthy、restart 0；数据库 clean26。切换停写18.88秒，677张图片、678条投稿及既有分类/标签关联保留。

本次修正公开页只能补充标签、编辑信息按标题搜索的问题。普通用户和游客通过“补充信息”建议标题、说明、替代文本、来源、主分类、作品角色和六组图片特征，审核后生效；标签页沿用原标签提案。有权限的编辑者直达当前图片 `/manage/images/{imageId}`。

## 来源与制品

- 发布版本 `image-information-20261115-1`，基线 `image-features-20261115-1`；原授权未提交改动保留，没有 commit/push/分支操作。
- Source snapshot：`28528ce31ee9a46724c395e49d58a315e5691fd987f8251d6c4c8682eccae308`。
- API/Web均重建。Go依赖使用已冻结Nov8 context，未改邻仓；Web从批准tarball离线安装，缓存复用768、下载0，安装14m10.8s，5包baseline `b896e8efe51b901bf8877e580cfa0af10de62e8ebe0aac6d2573584ad98f63e5`。
- API继承已验证当前镜像，仅覆盖三个二进制和迁移；本机忽略的运行配置不打包，生产秘密配置保持。
- 上传82,531,198字节，210.02秒，服务端摘要匹配。

## 验证与边界

Go gallery/dao/controller三包、HTTP catalog生成物、Web typecheck和20文件73项原测试通过；正式Web构建通过。正式制品CLI Playwright八项旅程覆盖单图直达、同名图定位、SSR刷新、版本保存/409草稿、隐藏图、375px、普通用户待审提交、429不重试、原标签流程和审核双权限。桌面/手机实际截图已亲自查看，长表单提交按钮固定底部。

原生PostgreSQL18.4副本 clean25→26、六个独立重建场景通过，覆盖原子采纳/防重复、驳回、过期图片、无效目录/身份、目录变更事务回滚和10/min额度。额外带已有标签的副本确认采纳分类建议不丢标签。每场景恢复副本并核对图片、处理单、Facet、目录和标签摘要。限额的浏览器/原生阈值在单图验证，跨图共享锁及SQL范围已静态核对。

认证浏览器旅程使用批准Identity SDK签发的隔离测试会话，API为隔离夹具；数据库业务使用原生PG副本和进程内Asset替身，无真实Identity/Asset写入。尚未声称真实管理员SSO验收。候选/公网仅只读，拦截浏览器统计与下载请求，不提交生产建议。

候选和公网均验证137个JS/CSS资源摘要、SSR/BFF、媒体、8位公开键及旧长地址404、后台登录门禁。两端实际镜像均匹配已接受候选：API `sha256:2306b167bdfaf133f2d74c8a18fbcca61fa23b79e677b4a02f48af8c8d934478`，Web `sha256:53a0bf30b9a8942917695ee39def87999526753803d3279e74833bab16d1eff9`。677项Asset引用无旧地址、无待应用项。

候选浏览器原请求头触发canonical307，曾落回旧生产并出现旧资源404，该结果明确判失败；改用所有Gallery请求转发到候选、只保留明确转发头，199个请求走新候选，桌面/375px表单和六组特征检查通过，无页面错误。公网直接访问同样通过，截图已亲自查看。

## 数据和恢复

迁移0026扩展现有处理单，审核额外要求图片编辑权限，处理单和图片在同一事务采纳；版本过期、目录冲突和权限错误不覆盖图片。历史迁移字节保持。

远端任务/备份入口 `/projects/yuelili.com/.deploy/image-information-20261115-1/`；生产当前真值为 `/projects/yuelili.com/gallery/deployment-release.json`。已停写取得最新可读数据库备份、核对迁移前后全部图片字段/关联和既有处理单摘要，只切换Gallery API/Web，PostgreSQL及其他服务保持。恢复备份为 `backup/production-before.dump`、`backup/compose.yaml`，不得自动向下迁移数据库。

本地小型证据在 `D:/coding/yueli-official/task-artifacts/gallery/image-metadata-20261115/`；唯一source-inputs和正式制品归档保留，可再生构建副本清理后记录真实结果。

## 收尾

自有候选API/Web及独立数据库已停止并删除；本地30,322个文件、2,549个内部链接和480,317,423字节可再生产物清理通过，543.22秒。文档CLI的两个已核对外部junction仅解除链接，原checkout和共享缓存保持。一次性浏览器/Go旅程源码已移除，结果与源码SHA保留；唯一发布归档、数据库/Compose恢复备份及必要回执保留。
