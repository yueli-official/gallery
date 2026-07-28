# 图片站产品

- 生命周期：活跃的中型产品实现
- 权威来源：[领域语言](./CONTEXT.md)、产品 E2E 契约和代码
- 当前实例：`gallery-main`；后续 Gallery 实例复用同一产品源码和独立实例库

Gallery 是运营方维护的公开图片发现站，支持稳定 URL 的目录搜索/分类/Facet/Tag 筛选、Grid/Masonry、共享 Quick View/详情 Viewer、相关图片、专题、排行、私人收藏、投稿记录和批量投稿。它不建立创作者主页、关注、评论或多图帖子。

公开界面采用图片优先的连续发现布局。顶部搜索支持标题、说明、替代文本与标签的模糊搜索；
输入 `#标签` 可按标签精确进入目录。首页不重复放置第二个搜索框，分类、Facet 和标签入口统一
位于图片流前的 refinement 条中。视觉规范见 [`DESIGN.md`](./DESIGN.md)。

管理端覆盖 Dashboard、图片生命周期、投稿审核、处理单、专题编排、发现设置、分类治理和 Asset delivery settings。处理、审核、安全、发布是独立状态；公开资格必须同时满足全部策略，隐藏/删除/安全阻断始终 fail closed。

## 边界

- `api/`：图片、投稿、审核、Collection/收藏、Case、分类投影、事件日聚合和排行快照。
- `web/`：公开发现/目录/Viewer、个人工作区、批量投稿和运营工作台。
- Asset：二进制上传、私有 master、公开 rendition/profile/variant 和撤销；Gallery 只保存 `assetId`。
- Media delivery：cacheable public object/rendition 不消耗 JSON API 的每 IP 请求 bucket，生产由 CDN/edge 控制带宽与滥用；上传、签名和受保护交付继续使用 API limiter。
- Identity：OIDC User/Guest 与权限；Gallery 不复制临时身份系统。
- PostgreSQL 18 + pgvector：Gallery 事实源和当前搜索 Adapter。外部搜索与 Redis 的触发器见工作主题的 [`scale.md`](../../flightdeck/work/2026-07-16-gallery-medium-large-upgrade/scale.md)。

## 仓库自治

Gallery 正在从 Platform 拆分为独立业务服务。当前目录已经具备独立 Go module、独立 pnpm
依赖与锁文件，并通过 `doctor.yaml` 声明 Identity、Asset、PostgreSQL/pgvector、检查任务和
本地进程。API 和 Web 不再依赖 Platform 私有 Go 包、`@platform/*`、pnpm catalog 或
workspace link。

工具链：

- Go `1.25.12`
- Node.js `24.18.x`
- pnpm `10.28.x`

前端首次安装必须忽略上级 Platform workspace，确保验证的是 Gallery 自己的锁文件：

```powershell
cd web
corepack pnpm install --ignore-workspace --frozen-lockfile
```

## 聚焦验证

```powershell
cd api
$env:GOWORK = "off"
go test -timeout 60s ./...

cd ../web
corepack pnpm test
corepack pnpm typecheck
$env:NODE_OPTIONS = "--max-old-space-size=4096"
corepack pnpm build
```

生产构建可能超过一分钟；本地自动化必须设置超时并在超时后清理 Nuxt/esbuild 子进程，不能让
后台构建占用内存。开发服务不使用生产构建流程。

## 本地运行

```bash
doctor check
doctor test
doctor up --detach
doctor status --check
doctor logs gallery-api
doctor logs gallery-web
doctor down
```

运行前将 `api/manifest/config/config.example.yaml` 复制为被 Git 忽略的 `config.yaml`，提供
`GALLERY_DATABASE_URL`，并由统筹仓库注入 `IDENTITY_BASE_URL` 与 `ASSET_BASE_URL`。
Gallery API 默认监听 `8091`，Web 默认监听 `3007`。Web 显式使用 `--host 0.0.0.0`，
同一局域网内的手机和 Windows 设备可通过开发机 IP 访问。

Gallery 开发数据已由 `api/cmd/devseed` 自持，并在一个事务中幂等对账站点分类、128 张图片、48 条投稿、
20 个处理单和 4 个专题。它只写 `GALLERY_DATABASE_URL` 指向的 Gallery 数据库；248 个实际图片对象及
Asset 记录由 Asset 仓的 `fixtureSet: gallery` 准备任务创建。任何产品 seed 都不得跨库写 Asset 表。

## 产品 E2E

产品自己的浏览器契约、旅程、视觉、无障碍和性能测试均位于 `web/test/e2e`，不再依赖 Platform
根目录的 Playwright harness。默认目标为 Gallery `http://127.0.0.1:3007`、Identity
`http://127.0.0.1:8081` 和 Account `http://127.0.0.1:3000`。

先由 Workspace 的 `gallery-local` 环境启动完整依赖，再运行：

```powershell
cd web
$env:PLATFORMCTL_E2E_EMAIL = "本地测试账号"
$env:PLATFORMCTL_E2E_PASSWORD = "本地测试密码"
corepack pnpm test:e2e
```

只检查测试发现而不打开浏览器：

```powershell
corepack pnpm exec playwright test --config test/e2e/playwright.config.ts --list
```

可通过 `PLATFORMCTL_E2E_SUITE` 选择 `journeys`、`visual`、`accessibility` 或
`performance`；默认运行 `all`。实际执行会写入 `web/test-results/e2e`，视觉基准由
`web/test/e2e/screenshots` 自持。
