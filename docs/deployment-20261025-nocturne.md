# Gallery Nocturne 正式发布

2026-10-25，用户批准提交与部署，确认正式域名，并在 1Panel 创建站点和 HTTPS 后授权配置反向代理。

## 入口与运行

- 正式入口：<https://gallery.yuelili.com/>，HTTP 自动 301 跳转 HTTPS。
- 1Panel 保留用户原证书，OpenResty 的本站 proxy include 转发至 `http://127.0.0.1:13007`。应用与数据库只监听 Docker 私网或宿主机回环，不公开后端端口。
- 产品目录：`/projects/yuelili.com/gallery`；Compose 项目 `gallery`，声明文件 `compose.yaml`。
- Web 回环端口 13007，API 回环端口 18091；专属 pgvector PostgreSQL 18，无宿主机数据库端口，数据位于 `data/postgres`。
- 私有配置在服务器 `config/`，目录 0700、文件 0600。没有在 Git 或发布记录保存密钥。
- 代理配置：`/opt/1panel/apps/openresty/openresty/www/sites/gallery.yuelili.com/proxy/gallery.conf`，上传传输上限 32 MiB；具体图片规则仍归 Asset profile。

## 固定制品

应用源码提交：`8877da1df347accbd6e9c9c729464cf517e6256f`。本地已有设计提交 `29b452d`，迁移原始字节恢复 `14286b7`，共享管理员认领 `a46577e` 与旧布局断言清理 `8877da1`。未 push、未切换分支。本文件属于后续发布记录，不改变上述制品来源。

| 项目 | 值 |
| --- | --- |
| 源快照 SHA256 | `400581e28b8a4220f4a69d2bbdc022a832a89912faef37a60274e5843b43ecc8` |
| API 包 SHA256 | `ba401fbddd34cf7d486c0615b602ee98114480f20c59fb61432e74587ad01a4e` |
| Web 包 SHA256 | `493579e58704db87a00bc6674b71a66fbc3baff990b6c9390307953dcfe8fd91` |
| API 镜像 | `yueli/gallery-api:nocturne-20261025-01` |
| API Image ID | `sha256:2681b2effc76bd355ab7e61d861a9585c0bb90bd4ac683f7966b903da4623024` |
| Web 镜像 | `yueli/gallery-web:nocturne-20261025-01` |
| Web Image ID | `sha256:5c3f4a601bdd2f28f6e5283ad02f19b0c896ad483da79ad54964d5ac6903ac43` |

正式冻结依赖安装后，以 Go 1.26.10、`GOWORK=off`、Linux amd64 / CGO=0 构建 API 与运维命令；Node 24.22.1 运行网页。服务器镜像只打包已构建的固定制品。Workspace 当前导出的前端基线与依赖回执通过 package/candidate 双阶段发布门禁。

服务器发布证据目录：`releases/nocturne-20261025-01`，包含源快照、API/Web 包、release.json、迁移文件摘要、候选与正式 HTTP 检查及代理回执。本机制品与日志：`D:/coding/yueli-official/task-artifacts/gallery/release-nocturne-20261024`。

## 初始化与基础服务绑定

新建专属数据库与非超级用户 `gallery`，按原始文件应用迁移 1–14，并运行产品 `bootstrap`。没有执行 devseed，没有复制本地图片、投稿、测试用户或管理员权限。

注册精确公开客户端 `gallery-main-web`，回调 `https://gallery.yuelili.com/auth/callback`，登出回调 `https://gallery.yuelili.com/`；注册最小媒体服务客户端 `gallery-asset-svc`，仅授予 `asset-api` 的 `asset:sign`。Asset 消费者声明使用仓库 `deploy/asset.consumer.json`，命名空间 `gallery`，原件私有、公开派生图；由 Asset 所属 provisioner 应用。服务器旧 helper 不支持 referenceLifecycle，改用 Asset 仓库对应工具完成注册，未升级或重启 Identity/Asset。

管理员初始化复用 Foundation PostgreSQL 原子认领。启动允许未认领状态，必须由已登录用户在 `/manage` 进入初始化页并显式确认。发布验收后仍为 `claimed=false`，没有代用户认领；匿名请求返回 401，个人令牌与服务主体不能认领。

注册前备份 Identity 与 Asset，初始 Gallery 数据库也已备份，均通过 `pg_restore --list`。备份位置：`/projects/yuelili.com/backups/gallery-nocturne-20261025-01`，包括代理切换前配置；正式配置私有副本在同一备份目录内。

## 验证结果与限制

- Go 全包测试、vet、HTTP 合同生成通过；一次性真实 PostgreSQL 检查覆盖 20 用户并发仅一个赢家、幂等重试、重新建立 Runtime 后权限持久、匿名/游客/服务主体拒绝。临时测试已删除。
- 前端 20 个测试文件、73 项检查通过；类型检查、生产构建通过。清理了锁定旧版导航/CSS/组件源码结构的断言，保留业务检查。
- 正式 API/Web/数据库健康，生产 Identity/Asset bindingcheck 通过；公开页、BFF、网站图标返回 200，HTTP → HTTPS 301，证书使用默认信任链校验。
- 正式域名 JS/CSS 的响应 SHA256 与固定制品一致。
- CLI Playwright：首页、图片、专题、排行、投稿 SSR 与 hydration，排行排序查询、搜索查询、管理入口跳转真实 Account 登录、匿名认领保护和手机导航通过，pageerror=[]。未代用户进行正式账号登录、管理员认领或投稿写入。
- CUA 实际截图人工检查正式桌面与 390×844 手机首页：导航、搜索、空状态、间距及按钮正常。正式库当前为空，带真实图片的 UI 验证沿用本 Work 已获用户批准的本地运行证据；首次认领页尚未取得正式登录后的截图。

本轮本地 Gallery 预览 Session `20261024T131835Z-125876` 已通过 Workspace CLI 停止；共享 Identity/Asset Provider 保留。正式服务持续运行。截图与详细执行证据保留在本地 Flightdeck 的该设计 Work。
