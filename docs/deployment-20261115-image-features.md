# Gallery 图片特征发布（2026-11-15）

Gallery 已发布 `image-features-20261115-1`，API/Web 均 healthy、restart 0，数据库 clean 25。仅切换 Gallery API/Web；共享 Identity、Asset、代理及其他站点未重启。用户已授权功能部署、目录创建及此前同范围既有改动；没有 commit、push 或分支操作。

## 已交付行为

单一“图片特征”维度内维护人物外观、服饰装扮、动作表情、画面构图、场景背景、表现形式六组，15个子目录、115个叶值。可选0–32项，目录不可赋值。网站筛选默认收起，支持路径搜索、移除、多特征同时满足；投稿、批量复制和后台图片编辑可保存多个叶值。后台分类管理可增删、改名及启停具体值，操作语义见[功能说明](image-features.md)。

公开筛选沿用共享 Classification 的已使用候选投影：尚未用于图片的维度和值不默认展示，已选择的零结果条件仍可显示、移除。现有图片未自动打标；目录已可在投稿及后台编辑中使用。标注图片后，相应特征才进入公开筛选候选。

油猴通用源为 `userscripts/gallery-publish.user.js`，版本1.5.0。个人安装版只保存在用户私有凭据目录 `C:/Users/yl/.secret/yueli/gallery-publish-personal-1.5.0.user.js`，未部署到公开站点或收入源码归档。覆盖安装旧版；保留既有令牌/偏好，只有未配置令牌时预填已有本机令牌。保留原有后台队列与完成通知，无新增 grant。真实Tampermonkey安装和系统通知未在本次验收中执行。

## 来源与制品

- 源码快照：`9fee58e261cc33ea80b5600a72902ec98c202a7100ace486dd6774fade714c86`。
- 当前Workspace批准前端基线：`b896e8efe51b901bf8877e580cfa0af10de62e8ebe0aac6d2573584ad98f63e5`，5个正式tarball依赖通过构建门禁。
- API镜像：`sha256:a100584c47612cd9963c8751ae29c3a1c95fb71e7f4ecd879c0a15786ac14f10`；运行归档SHA256：`7ce816402a8bbba233a6a23ff41f28a42cd0da1c4f759e1631e0988b27e212d6`。
- Web镜像：`sha256:28de86d07dc46c9194a0598068538a5034ab93e1986031c8fd0fc3c0a198bc57`；运行归档SHA256：`39aeea6f4d043b75f295b67775a4c638f77a884a68cd44b1ce579792ab4acf9f`。
- 原生产来源为`image-public-keys-20261108-1`。API/Web业务差异限于本次图片特征；19份历史迁移仅有换行差异，已恢复为生产执行的原始字节。未纳入相邻仓库未发布源码，Go使用已验证的11月8日冻结依赖上下文；Web在独立目录通过frozen lock和既有pnpm缓存构建。
- API、bootstrap和本轮24→25专用迁移器交叉构建为Linux/amd64、CGO=0；未变工具继承已部署制品。迁移器源已收入源码归档。Nuxt输出没有原生`.node`，归档只含制品内部相对链接。

## 迁移与验收

原生PostgreSQL18.4生产副本完成24→25、幂等种子、六组115叶值及所有profile策略检查。API/BFF共8次真实筛选查询验证单值、合取、无结果和候选交集计数；副本临时标记完成后删除，图片及关联摘要恢复一致。桌面1280与手机375实际候选浏览器验证折叠、六组、搜索、选择与URL；沿用已通过且源码/依赖匹配的11月14日投稿、批量、后台编辑和油猴完整旅程及视觉证据。

生产迁移前暂停写入并备份，切换耗时19.07秒；677张图片、678条投稿、677条分类关联、1354条原作品角色关联、28个标签和1个集合保留。迁移前后原始图片字段与分类/Facet/标签关联摘要一致；新增136个Facet目录/叶节点，不对旧图片推断。

上线后公网健康、投稿选项、BFF、图片详情/相关图片、登录门禁、8位公开键及旧长链接404通过；135个JS/CSS摘要与当前制品一致，实际WebP媒体200。公网桌面/手机浏览器验证预选特征零结果可搜索/移除并恢复677张图片，无pageerror；已查看实际截图。未向生产创建测试图片、修改已有图片特征或执行初始化/认领。

本地原完整开发组合因开发库192.168.5.5不可用，11月14日采用隔离夹具验收；此次已补充真实原生PG和线上组合验证。未冒充真实管理员登录投稿或真实Tampermonkey/系统通知验收。

## 备份与证据

线上权威回执：`/projects/yuelili.com/gallery/deployment-release.json`；发布归档：`/projects/yuelili.com/gallery/releases/image-features-20261115-1/`。

停写后可读数据库备份：`/projects/yuelili.com/.deploy/image-features-20261115-1/backup/production-before.dump`，1477329字节、SHA256 `c4f5edeb5b9d1baa98b02b9fb3b46b62050a66109844c212a6fafebb276c2209`；同目录保留原Compose与原发布回执（私有权限）。备份已通过pg_restore目录验证；不要自动降级数据库，恢复须基于实际25版本及备份评估。

本地证据：`D:/coding/yueli-official/task-artifacts/gallery/image-features-deploy-20261115/` 的最终release.json、candidate/production-migration、candidate-features、candidate/production-http与browser、browser-fixture-cleanup、switch、backup、传输和清理回执；原本地验收保留在 `image-features-20261114/`。

依赖安装13分52.4秒（命中768、下载0、安装773）；API/Web远端镜像组装分别5.05秒/1.71秒，本轮迁移器修正后API组装0.75秒。本机构建总耗时未单独计时。首次上传103704456字节/266.39秒；旧一次性迁移器在副本被版本前置条件拒绝后，仅补传更新API与来源记录86088926字节/189.70秒，Web逐字节复用。两次均校验服务端SHA256。生产迁移前置条件24→25已在副本通过，随后应用同一制品。候选容器、数据库、临时秘密配置及本地隧道均已清理，生产与恢复资料保留。
