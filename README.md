# 图片站产品

- 生命周期：活跃的中型产品实现
- 权威来源：[领域语言](./CONTEXT.md)、产品 E2E 契约和代码
- 当前实例：`gallery-main`；后续 Gallery 实例复用同一产品源码和独立实例库

Gallery 是运营方维护的公开图片发现站，支持稳定 URL 的目录搜索/分类/Facet/Tag 筛选、Grid/Masonry、
独立详情 Viewer、相关图片、专题、排行、私人收藏、投稿记录和批量投稿。它不建立创作者主页、关注、
评论或多图帖子。

公开界面采用图片优先的连续发现布局。顶部搜索支持标题、说明、替代文本与标签的模糊搜索；
输入 `#标签` 可按标签精确进入目录。首页只保留“随机看看”和“换一批”，分类、Facet 与标签
筛选集中在浏览页。视觉规范见 [`DESIGN.md`](./DESIGN.md)。

管理端覆盖 Dashboard、图片生命周期、投稿审核、处理单、专题编排、发现设置、分类治理和 Asset delivery settings。处理、审核、安全、发布是独立状态；公开资格必须同时满足全部策略，隐藏/删除/安全阻断始终 fail closed。

## 边界

- `api/`：图片、投稿、审核、Collection/收藏、Case、分类投影、事件日聚合和排行快照。
- `web/`：公开发现/目录/Viewer、个人工作区、批量投稿和运营工作台。
- Asset：二进制上传、私有 master、公开 rendition/profile/variant 和撤销；Gallery 只保存 `assetId`。
- Media delivery：cacheable public object/rendition 不消耗 JSON API 的每 IP 请求 bucket，生产由 CDN/edge 控制带宽与滥用；上传、签名和受保护交付继续使用 API limiter。
- Identity：OIDC User/Guest 与权限；Gallery 不复制临时身份系统。
- PostgreSQL 18 + pgvector：Gallery 事实源和当前搜索 Adapter；外部搜索只在目录规模和查询负载达到
  明确迁移阈值后引入。

## 仓库自治

Gallery 已是独立业务服务，具备独立 Go module、pnpm 依赖与锁文件、版本化 migration、
Dockerfile 和产品 E2E。Identity、Asset、PostgreSQL/pgvector 及本地进程由 Workspace 的
`gallery-local` 环境统一组合；API 和 Web 不依赖 Platform 私有 Go 包、`@platform/*`、
pnpm catalog 或 workspace link。

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

## 本地一键开发

本地多仓依赖只由相邻 Workspace 仓库管理；Gallery 不再维护 `doctor.yaml` 或另一份进程清单：

```powershell
cd ..\workspace
go run ./cmd/workspace dev prepare gallery --with-deps `
  --local identity --local gallery `
  --environment environments/gallery-local/environment.yaml --root .
go run ./cmd/workspace dev up gallery --with-deps `
  --local identity --local gallery `
  --environment environments/gallery-local/environment.yaml --detach --root .
go run ./cmd/workspace dev status --check --root .
go run ./cmd/workspace dev logs --root . gallery/gallery-web
go run ./cmd/workspace dev down gallery --root .
```

日常可直接使用 Workspace 的 `environments/gallery-local/run.ps1` 包装入口。环境会生成被忽略的 API
配置、运行 migration/seed，并按 Identity → Account → Asset → Gallery API → Gallery Web 启动。
Gallery API 默认监听 `8091`，Web 默认监听 `3007`；同一局域网设备可通过开发机 IP 访问。

Gallery 开发数据已由 `api/cmd/devseed` 自持，并在一个事务中幂等对账站点分类、128 张图片、48 条投稿、
20 个处理单和 4 个专题。它只写 `GALLERY_DATABASE_URL` 指向的 Gallery 数据库；248 个实际图片对象及
Asset 记录由 Asset 仓的 `fixtureSet: gallery` 准备任务创建。任何产品 seed 都不得跨库写 Asset 表。

## Docker Compose 部署

仓库提供三个标准入口，要求 Docker Compose `2.20.0+`。它们共享 Gallery 自身模块，但不会把
Workspace 变成生产依赖：

| 入口 | Identity | Asset | 启动内容 |
|---|---|---|---|
| `compose.yaml` | 受管 | Gallery 专属受管 | Identity/Account、Asset、Gallery 及各自基础设施 |
| `compose.attach.yaml` | 已有 | 已有 | Gallery API/Web、Gallery PostgreSQL、migration 与 binding check |
| `compose.hybrid.yaml` | 已有 | Gallery 专属受管 | Asset、Gallery 及各自基础设施 |

默认完整独立部署：

```powershell
Copy-Item .env.example .env
# 填写所有空 secret、管理员和数据库值
docker compose config --quiet
docker compose up -d --wait
docker compose down
```

接入已有 Identity 与 Asset：

```powershell
Copy-Item deploy/env/attach.env.example deploy/env/attach.env
# 填写外部 HTTPS endpoint、已注册的 Gallery client 和 secret
docker compose --env-file deploy/env/attach.env -f compose.attach.yaml config --quiet
docker compose --env-file deploy/env/attach.env -f compose.attach.yaml up -d --wait
docker compose --env-file deploy/env/attach.env -f compose.attach.yaml down
```

共享 Identity、部署 Gallery 专属 Asset：

```powershell
Copy-Item deploy/env/hybrid.env.example deploy/env/hybrid.env
docker compose --env-file deploy/env/hybrid.env -f compose.hybrid.yaml config --quiet
docker compose --env-file deploy/env/hybrid.env -f compose.hybrid.yaml up -d --wait
docker compose --env-file deploy/env/hybrid.env -f compose.hybrid.yaml down
```

普通 `down` 保留数据库、对象和 Identity publisher 数据。只有明确销毁实例时才在相同 `-f`/`--env-file`
参数后执行 `down --volumes`；该命令不可恢复。

Gallery 的能力需求位于 `deploy/contracts/requirements.json`，确定的 Identity/Asset 源码版本位于
`deploy/deployment.lock.json`。启动时 `gallery-binding` 会验证 OIDC issuer/discovery/JWKS、机器凭据、
Asset readiness、所需 OpenAPI 路径和真实 Bearer 授权；任何绑定不匹配都会阻止 Gallery API 启动。
完整独立入口从锁定 revision 构建基础服务，不依赖相邻源码；attach 不会创建、重启或关闭外部基础服务。

生产必须使用 HTTPS origin、`GALLERY_COOKIE_SECURE=true` 和至少 32 字节随机 Web seal secret。受管 Asset
试点当前固定发布 `8082` 本地存储地址；需要自定义 CDN/S3/public origin 的生产环境应使用 attach，或在
发布新版 Asset 模块后更新 deployment lock，不能直接改消费者业务代码。

## 产品 E2E

产品自己的浏览器契约、旅程、视觉、无障碍和性能测试均位于 `web/test/e2e`，不再依赖 Platform
根目录的 Playwright harness。默认目标为 Gallery `http://127.0.0.1:3007`、Identity
`http://127.0.0.1:8081` 和 Account `http://127.0.0.1:3000`。

先由 Workspace 的 `gallery-local` 环境启动完整依赖，再运行：

```powershell
cd web
$env:GALLERY_E2E_EMAIL = "本地测试账号"
$env:GALLERY_E2E_PASSWORD = "本地测试密码"
corepack pnpm test:e2e
```

只检查测试发现而不打开浏览器：

```powershell
corepack pnpm exec playwright test --config test/e2e/playwright.config.ts --list
```

可通过 `GALLERY_E2E_SUITE` 选择 `journeys`、`visual`、`accessibility` 或
`performance`；默认运行 `all`。实际执行会写入 `web/test-results/e2e`，视觉基准由
`web/test/e2e/screenshots` 自持。
