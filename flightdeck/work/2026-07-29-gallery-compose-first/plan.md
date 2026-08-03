# Gallery Compose-first 试点计划

## P0 — 基线

- [x] 记录既有 diff、运行模块、依赖合同与验证基线。
- [x] 审计 Dockerfile、配置覆盖、migration、healthcheck 和 seed 边界。

## P1 — 运行合同

- [x] 定义 Gallery 自身 API/Web/PostgreSQL 模块。
- [x] 定义可机器校验的能力需求和外部绑定检查。
- [x] 删除独立仓文档中的 Doctor 入口和遗留 Platform 运行命名。

## P2 — 三种 Compose 拓扑

- [x] 完整独立：受管 Identity、Asset、Gallery 及其基础设施。
- [x] Attach：只部署 Gallery，显式接入已有 Identity 与 Asset。
- [x] Hybrid：复用 Identity，部署 Gallery 专属 Asset。
- [x] 提供严格 `.env.example`、部署锁、数据保留和销毁说明。

## P3 — 验收

- [x] Go 测试、vet、staticcheck、govulncheck、deadcode 和 Linux 构建通过。
- [x] Web 测试、typecheck、production build 通过。
- [x] 三种 Compose 模型在可用工具下完成配置验证；记录真实 Docker 限制。
- [x] 同步 Workspace/Platform handoff 并关闭本工作。

## P4 — 登录后消费者回归

- [x] 用真实 OIDC 会话复现“我的投稿”BFF 404，并增加 Gallery 客户端源码合同测试。
- [x] 统一 Gallery BFF、完整 HTTP method 与服务端会话凭据，移除 Gallery 对 Identity 默认客户端的误用。
- [x] 投稿、收藏、管理接口和收藏 PUT/DELETE 真栈通过；Web/API 全量测试通过。

## P5 — 首次 Guest actor 与本地 E2E

- [x] 用新浏览器复现 `/submissions` 首次 GET 401，并增加页面级 Guest 回归测试。
- [x] 仅为 `/api/gallery/me/*` 建立 Guest 会话，公共读取保持无状态；新访客空投稿页通过。
- [x] E2E URL 从 Workspace `LOCAL_*` 变量派生；已登录管理员 48 条投稿回归通过。

## P6 — 交付级代理验收

- [x] 建立功能、视觉、响应式、无障碍、性能、安全和故障恢复的分层验收大纲。
- [x] 冻结验收数据，补普通 Member，并让视觉夹具不依赖异步扫描状态。
- [x] 执行 Gate A 与关键用户旅程；P0/P1 采用停止线策略修复并增加回归。
- [x] 完成管理 CRUD、故障注入、全路由视觉/无障碍、性能与生命周期验收。
- [x] 输出带证据的 Pass/Fail/Blocked 报告；扫描阻塞由 P11 架构变更关闭，Docker 主机验收仍待补。

## P7 — 公共 Viewer 视觉重构

- [x] 审计右侧栏、标题尺度和真实长中文约束。
- [x] 生成三种同视觉世界构图，选定全宽图片展台与低位信息坞。
- [x] 重构桌面、平板、手机与明暗模式，不裁切原图或改变功能语义。
- [x] 完成 Viewer 单测、连续浏览、四项视觉合同、1024px 定点溢出和独立设计审稿。
- [x] 按用户 Pixiv 实页反馈继续降噪：图标化收藏/分享，弱化统计；简介归到标题下，尺寸与治理动作渐进披露。

## P8 — Viewer 动作与全局反馈

- [x] 在真实 Member 与退化 Guest 会话下验证收藏；精确 401/403 改为保留当前页的登录恢复。
- [x] 为无 Web Share、非安全上下文和无 Clipboard API 的 LAN HTTP 页面补分享复制回退。
- [x] 全局关闭 Toast 倒计时进度条，收敛尺寸、阴影、状态图标和并发数量，并完成实拍。

## P9 — LAN HTTP 投稿与 Toast 对齐

- [x] 在真实局域网 HTTP 地址复现投稿选图的 `crypto.randomUUID` 崩溃，并用单变量注入排除输入与预检。
- [x] 增加跨安全上下文的客户端队列 ID 工具和回归测试；真实选图后队列稳定新增一项。
- [x] 在真实 Toast 上测量图标、文字和关闭按钮中心，把 4px 偏差收敛为 0px，并完成定向门禁。

## P10 — 扫描停滞与轮询风暴

- [x] 用 Asset readiness、数据库工作队列和扫描证据确认 worker/ClamAV 缺失，不把 pending 误诊为 HTTP。
- [x] 将固定 500ms 轮询改为封顶指数退避，并用 10 秒调用计数红绿回归固定请求预算。
- [x] 在上传和已有 Asset 等待前探测 `/readyz`；扫描能力不可用时明确失败且不创建新 Asset。
- [x] 修正 Workspace Gallery Asset 模板的 worker 默认值，完成真实页面、Vitest、typecheck 与环境测试。

## P11 — 移除扫描链并同步独立 Gallery

- [x] 删除独立 Gallery 的 Asset readiness 探测、scan 状态轮询、遗留类型和扫描错误文案。
- [x] 投稿在 Asset finalize 成功后直接创建 Submission；补禁止旧扫描消费回归合同。
- [x] 删除独立 Gallery 自持 Compose/config/deployment infrastructure 中的 ClamAV。
- [x] 标准 Workspace prepare/migration/seed 和五进程 ready 通过。
- [x] 真实 Guest BFF 双图以 8 个请求完成，两个 Asset 和两个 Submission 均成功。
- [x] 用仓库本地 Playwright 在 LAN HTTP `/submit` 完成双图选取、分类/场景填写、提交完成态、网络 trace
  和最终整页截图；8 个业务请求全部 200，无 Asset GET 或扫描状态轮询。
- [x] Gallery 部署锁已指向 Asset `v0.3.0`；无 Docker Engine 的真实容器限制已记录，后续部署验收另行处理。
