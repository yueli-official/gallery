# Gallery 交付级验收测试大纲

## 1. 验收目标

把当前 Gallery 当作一个需要对真实交付质量负责的完整产品，而不是“页面能打开”的演示站。测试需要证明：

1. 匿名访客、Guest、普通登录用户和管理员的核心任务都能完成。
2. 页面在手机、平板和 Windows 桌面尺寸下没有功能降级、遮挡、横向溢出或不可操作控件。
3. 登录、Guest claim、Gallery BFF、Asset 与 Identity 的边界正确，不泄漏令牌、不误用其他服务的 BFF。
4. 空态、错误态、慢请求、依赖故障和重复操作都有确定行为，不出现无限加载、假成功或数据损坏。
5. 视觉、无障碍和性能达到仓库已经声明的标准，而不是只通过宽松 smoke test。
6. 每个失败都有可复现证据；不通过重跑、更新快照或跳过用例掩盖缺陷。

本大纲先覆盖本地 Workspace 真栈。Asset 内置恶意文件扫描已在后续架构决策中删除；上传成功只验证同步
基础文件校验，内容安全、审核与发布由 Gallery 负责。当前主机没有 Docker Engine，因此真实容器拓扑仍只能
标记为 `Blocked`。任何阻塞主路径的 `Blocked` 项都意味着最终验收未通过。

## 2. 发布准入标准

只有同时满足以下条件才允许给出“Gallery 本地验收通过”的结论：

- P0、P1 缺陷为 0。
- 所有关键旅程 100% 通过；不允许用 retry 把首次失败改写为通过。
- 所有目标路由返回预期的 2xx、3xx 或明确 404；无意外 401、403、5xx。
- 所有验收页面无 `console.error`、未捕获 `pageerror`、失败资源和意外 5xx。
- WCAG 2.2 A/AA 自动扫描无 violation；核心流程可只用键盘完成。
- 390、768、1100、1440 四个基准视口的视觉合同通过；320 宽、200% 文本缩放和 1920 宽无结构性破坏。
- Light、Dark 两个主题角色一致，无错误颜色、不可读文本或主题切换残留。
- Gallery 首页复访性能不超过现有预算：
  - LCP ≤ 2200ms
  - CLS ≤ 0.05
  - TTFB ≤ 800ms
  - 单张已传输图片 ≤ 400KiB
  - 可见图片总传输量 ≤ 2MiB
  - 图片自然宽高均 ≤ 1600px
- 视觉快照使用冻结、确定的验收夹具；不得依赖会被后台扫描异步修改的 Submission 状态。
- 测试证据不包含访问令牌、密码、数据库连接串或完整 session cookie。

缺陷等级：

| 等级 | 定义 | 例子 |
| --- | --- | --- |
| P0 | 安全、数据损坏或产品完全不可用 | 越权进入管理端、提交导致数据丢失、首页持续 500 |
| P1 | 主任务失败或 WCAG AA 明确违规 | 无法登录、投稿/收藏失败、移动端关键按钮不可点击 |
| P2 | 有绕行方案但明显影响体验 | 筛选状态丢失、错误文案不能指导重试、局部布局错位 |
| P3 | 不阻挡任务的打磨问题 | 轻微间距、非关键动画或文案一致性问题 |

## 3. 环境与数据前置

### 3.1 运行环境

- 统一从 Workspace `environments/gallery-local/run.ps1` 启动。
- 使用同一组 `LOCAL_*` 变量派生 Gallery、Identity、Account 和 Asset 地址。
- 保留外部 `3000` 项目，Gallery Account 使用 `3010`。
- 每次执行前记录 Workspace session id、五个进程状态和健康 URL。
- 测试后再次检查五个进程，确认测试没有导致服务退出或端口漂移。

### 3.2 浏览器矩阵

最低验收矩阵：

| 类型 | 浏览器 | 视口 |
| --- | --- | --- |
| 手机 | Playwright Chromium | 390×844 |
| 平板 | Playwright Chromium | 768×1024 |
| 中等桌面 | Playwright Chromium | 1100×900 |
| Windows 桌面 | Playwright Chromium | 1440×900 |

结构性补测：

- 320px 窄屏：不允许横向滚动、截断关键动作或重叠。
- 1920px 宽屏：内容不能无限拉伸，图片和信息层级仍成立。
- 200% 文本缩放：导航、表单、筛选和管理工具仍可使用。
- `prefers-reduced-motion: reduce`：状态变化仍可理解，不能靠动画表达唯一信息。
- 若本机已安装 Firefox/WebKit，公开浏览、登录、Viewer、投稿和管理入口各跑一轮跨浏览器 smoke。

### 3.3 身份与数据角色

必须准备四种互不混用的上下文：

| 角色 | 初始状态 | 重点 |
| --- | --- | --- |
| Anonymous | 无 session、无 Guest cookie | 公开浏览、首次进入个人页 |
| Guest | 有 Guest session、未登录 | 投稿、Guest 数据保持与 claim |
| Member | 普通登录用户、无管理权限 | 私人收藏、自己的投稿、403 边界 |
| Admin | 当前测试管理员 | 全部运营功能 |

现有环境只有管理员账号时，需要补一个普通 Member 夹具，否则不能证明管理权限边界。

验收数据必须包含且可重置：

- 公开、隐藏、下架和缺失 Image。
- 空收藏与非空收藏。
- Guest 与 Member 各自拥有的 Submission。
- queued、processing、ready、failed、pending review、approved、rejected、withdrawn 等确定状态。
- 可公开、草稿、空内容和多图片 Collection。
- 可处理与已处理 Case。
- 正常、版本冲突、越权和不存在资源。

异步 worker 不得在视觉截图期间改写这些状态。需要单独的稳定视觉夹具或测试事务，而不是反复更新基线。

## 4. Playwright 测试层次

### Gate A — 预检与快速红灯

目标是在 1–2 分钟内发现“根本不能测”的问题：

- Workspace 五进程全部 `ready`。
- Gallery `/healthz`、API `/readyz`、Identity `/readyz`、Asset 当前选定健康门禁均返回预期状态。
- 首页、目录、详情、登录入口和管理入口 smoke。
- 新浏览器访问 `/submissions` 不得出现“投稿记录加载失败”。
- 管理员登录后 `/me/submissions` 返回非空数据。
- 任何 console error、pageerror 或意外 5xx 立即失败。

Gate A 失败时停止后续大矩阵，先修根因。

### Gate B — 关键业务旅程

测试完整用户任务和真实副作用，不能只断言标题出现：

#### B1. 公共发现与目录

- 首页首屏出现真实图片，图片优先于说明性内容。
- 搜索词、Category、Facet、Tag、排序和分页可独立及组合使用。
- URL query 与 UI 状态双向一致；刷新、前进、后退后状态不丢失。
- 清除筛选恢复默认目录，结果数量和分页同步更新。
- 移动端筛选抽屉可打开、应用、取消和关闭；焦点返回触发按钮。
- 搜索无结果显示明确空态，不出现空白页面或错误提示。
- 图片请求使用合适 rendition，不拉取超规格 master。

#### B2. Collection 与排行

- Collection 列表、详情和不存在 slug 行为正确。
- Collection 图片顺序与后台配置一致，跳转到正确 Image。
- 排行页项目可进入详情，指标和排序稳定。
- 空 Collection 有明确空态，不产生损坏卡片。

#### B3. Image 详情与 Viewer

- 已知图片正常渲染，缺失图片返回产品 404。
- 连续点击下一张至少 12 次，URL 和内容同步，至少访问 10 张不同图片。
- 上一张、下一张、缩略图跳转、返回目录、浏览器前进/后退均正确。
- Fullscreen 可进入、切图后保持、Esc 可退出，退出后焦点返回原触发控件。
- 键盘方向键、Tab、Shift+Tab 和 Enter 可完成同等操作。
- 分享、举报、收藏等动作具有可识别反馈，不发生重复提交。
- 图片 `alt` 来自真实数据；装饰图片不进入无意义朗读。
- 图片加载失败时出现可恢复状态，不保留永久 skeleton。

#### B4. 收藏

- Anonymous 点击收藏后得到明确的登录或 Guest 产品行为。
- Member 首次收藏、重复收藏、取消收藏、重复取消均幂等。
- 收藏按钮、详情页和“我的收藏”状态保持一致。
- 刷新页面和重启 Gallery Web 后收藏仍存在。
- 两个页面同时修改时，版本冲突能刷新为服务器真值并给出明确提示。
- 无权限用户不能读取或修改他人的收藏。

#### B5. Guest 投稿、登录和 claim

- 新 Anonymous 访问“我的投稿”建立一次 Guest actor，会显示空态而不是 401。
- 普通公开目录 GET 不创建 Guest session。
- Guest 创建投稿后能在自己的列表看见，其他新 Guest 看不见。
- Guest 登录后原投稿被 claim 到该 Member，Guest cookie 被清理。
- 登录回调只落到允许的 LAN/localhost origin，不能开放到普通 HTTP 域名或公网 IP。
- 登出后私人数据不可继续通过旧页面或 BFF 获取。
- session 过期、撤销和刷新失败时进入明确登录流程，不无限重定向。

#### B6. 投稿与上传

- 投稿类型、必填字段、单张/批量选择、预览、删除待上传项和重复文件行为正确。
- 客户端拒绝不支持格式、超尺寸、空文件和损坏文件。
- 上传进度、取消、网络中断、重试和部分失败状态明确。
- 提交后 Submission 状态、数量和详情一致。
- processing、review、published、duplicate、rejected、failed、withdrawn 的文案和动作正确。
- 可撤回状态能够撤回，不可撤回状态不展示误导操作。
- 筛选和分页覆盖 outcome、processingState、reviewState 的组合。
- Asset finalize 后不得轮询已删除的扫描状态；每张图片只产生一次 upload-init、PUT 和 finalize。
- 上传成功不等于公开：Gallery 的 processing、safety、review 和 publication 仍需独立验收。

### Gate C — 管理端真实操作

每个管理页至少包含：读取、修改、保存、刷新验证副作用、权限拒绝和失败恢复。

#### C1. 权限与导航

- Anonymous 进入管理路由跳 Account 登录。
- Member 登录后仍不能进入任何管理页或调用管理 API。
- Admin 能进入所有管理路由。
- 左侧/移动导航的活动状态、折叠、回退和直接 URL 访问一致。
- `/manage/authorization` 必须纳入路由、视觉和权限测试，不能成为遗漏页。

#### C2. 投稿审核

- 筛选待审、失败和已完成状态。
- 单条 approve/reject、审核说明和重复提交。
- 全选本页、取消、跨分页选择边界和批量操作数量。
- 批量操作部分失败时不显示整体假成功，结果可追溯。
- 审核后公开可见性与 Image 状态同步。

#### C3. Image 管理

- 搜索、筛选、分类、维度、Tag 和公开状态修改。
- 已发布图片像素不可替换；撤稿后遵守重新投稿规则。
- 版本冲突、越权和不存在 Image 都有明确错误。
- 修改后公开目录和详情真实反映结果。

#### C4. 分类体系

- Category、Facet、Facet Value 和 Tag 的概念、入口与文案不混淆。
- 创建、编辑、排序、禁用/删除限制和被引用资源行为正确。
- 重名、非法 slug、空名称和超长内容有前后端一致校验。

#### C5. Collection 管理

- 创建草稿、修改标题/说明、添加/移除图片、拖动排序、保存和发布。
- 脏数据离开保护：取消离开保留修改，确认离开丢弃修改。
- 发布后的公开页面、顺序和 SEO 数据正确。
- 空 Collection 与删除/下线后访问行为明确。

#### C6. Discovery 设置

- 首页板块开关、顺序、数量和随机发现策略保存后真实反映到公开首页。
- 非法数值、空配置、并发版本冲突和保存失败可恢复。
- 刷新后配置不回退，seed 不覆盖运营设置。

#### C7. Case 与 Asset 设置

- Case 筛选、查看、处理和重复处理行为正确。
- Asset delivery/profile/rendition 设置读取和保存真实生效。
- 危险设置需要明确确认，不泄漏内部 secret。
- Asset 不可用时管理页显示依赖故障，不把错误包装成空数据。

### Gate D — 错误与韧性

使用 Playwright route、服务级故障或超时注入验证：

- Gallery API 401、403、404、409、422、429、500/503 的产品化呈现。
- Identity、Asset、Gallery API 分别不可用时，页面不崩溃、不无限加载。
- 慢请求、请求超时、断网后恢复、点击重试。
- 快速连点提交、重复导航和并发请求不产生重复数据。
- 页面加载中切路由、关闭弹窗或返回，不产生未捕获 promise。
- 刷新、进程重启后已提交数据仍存在。
- 错误提示包含用户下一步，但不泄漏 Go stack、SQL、内部 URL、token 或 trace 之外的敏感信息。

### Gate E — 安全与 BFF 边界

- Gallery 页面只调用具名 `/api/gallery`，不把业务路径发到 Identity 默认 BFF。
- Browser 不持有 access token；localStorage、DOM、URL 和错误消息中无 token。
- session/Guest cookie 的 HttpOnly、SameSite、Path 和生产 Secure 合同正确。
- 公共 GET 不无条件创建持久 Guest；`/me/*` 首次访问才创建。
- 所有 `/admin/*` 在 API 层再次鉴权，不能只靠前端 middleware。
- BFF 拒绝路径穿越、绝对目标 URL、危险 Header、非法 method 和超大请求体。
- OIDC `return_to`、callback、logout redirect 防止开放重定向。
- 文件名、搜索词、Tag、审核说明和管理表单中的 HTML/script 不执行。
- Problem response 不含数据库、文件系统和内部网络细节。
- 失败 trace/video 作为本地敏感制品处理，报告中不得复制 session 或 Authorization。

### Gate F — 视觉、响应式与主题

先冻结数据，再一次性采集桌面和移动端，不边看边反复改基线。

覆盖当前合同中的全部场景：

- public、catalog、mobile-filter、submission-options
- my-favorites、my-submissions
- collections、collection-detail、rankings、viewer
- manage-default、manage-collections、manage-collection-detail
- manage-images、classification、manage-discovery、manage-assets
- review-queue、review-bulk-selection、case-queue
- empty、error

新增遗漏场景：

- `/contribute` 的兼容/跳转行为
- `/manage/authorization`
- Guest 空投稿
- 登录、登出和 session 失效状态
- 表单校验、保存成功、保存失败、脏数据确认
- 依赖不可用和请求重试
- Viewer fullscreen/缩略图/图片失败

每张截图同时检查：

- 图片是否仍是公共页面主角。
- 标题、正文、按钮和状态层级是否清楚。
- 控件对齐、间距、边界、圆角、focus 和 hover 是否一致。
- 无横向溢出、遮挡、裁切、不可读状态色和空白大块。
- Light/Dark 角色一致，矿物蓝仍是唯一主强调色。
- 演示数据不会被误当真实生产内容。

像素差异阈值保留现有 2%，但只有在人工/代理检查实际图像正确后才能接受新基线。当前 Submission
`asset_processing_failed` 与“等待处理”基线的差异必须通过稳定数据解决，不能直接 `--update-snapshots`。

### Gate G — 无障碍

自动检查：

- 所有主要路由在 mobile/desktop 下执行 axe WCAG 2.2 A/AA。
- Light/Dark 均验证文本、控件、焦点和状态对比度。
- 所有表单有可关联 label、错误说明、required 和 disabled 状态。
- heading、main、nav、dialog、region 和列表语义正确。
- 图片替代文本、装饰图隐藏和 icon-only button 名称正确。

任务级键盘检查：

- 导航、目录筛选、Viewer、收藏、投稿表单、管理筛选、批量审核和确认弹窗可完成。
- Tab/Shift+Tab 顺序稳定，Modal/Drawer 焦点受控且关闭后归还。
- Enter/Space/Escape/方向键符合控件语义。
- 所有焦点都有可见指示，不依赖颜色作为唯一状态。
- 200% 文本缩放和 reduced motion 不丢功能。
- 触摸目标至少 44×44px，或具有不重叠的等效可点击区域。

### Gate H — 性能

保留现有暖缓存首页预算，并新增：

- 首页冷启动 mobile/desktop。
- 图片目录首次加载与下一页。
- Viewer 首图、下一张切换和缩略图加载。
- 登录后管理首页和投稿列表。
- 长列表滚动过程无明显 layout thrash 或持续掉帧。
- 无重复请求瀑布、同一图片多规格误加载和不可见 master 预加载。
- 图片 lazy/eager 策略与首屏优先级正确。
- 生产 build 做 bundle 体积对比，异常增长必须能定位到依赖。

开发服务器测量只用于本地回归；最终性能结论需要 production build 或容器环境复验。

### Gate I — 生命周期与部署

本地主机可执行：

- `run.ps1` 的 Up、Status、Logs、Down、SkipPrepare 闭环。
- 端口覆盖后所有健康 URL、OIDC callback 和 E2E URL 一致。
- 停止只影响 Workspace 登记的 Gallery 组合，不影响外部 `3000` 项目。
- 停止/重启后数据库数据保留，重新 prepare 幂等。

需要 Docker 主机和包含扫描移除的新 Asset 不可变版本：

- 完整独立、Attach、Hybrid 三个 Compose 入口真实 `up -d --wait`。
- 登录、Guest、上传、同步基础校验、签名、发布、重启、down 和数据卷保留。
- 接入已有 Identity/Asset 时 binding check 拒绝错误 client、audience 或 profile。
- 销毁命令与数据删除边界符合文档。

## 5. 现有自动化覆盖与必须补齐的缺口

| 领域 | 当前已有 | 必须补齐 |
| --- | --- | --- |
| Journey | 首页、管理登录、空态、404、Guest/登录投稿、Viewer 连续浏览 | 收藏、Guest claim、上传、撤稿、管理真实修改、登出/过期 |
| Visual | 主要公开与管理场景，四视口、双主题 | 稳定数据、授权页、错误/保存状态、320/缩放结构检查 |
| Accessibility | 首页和管理首页 axe、基础双向焦点 | 全路由、表单、Drawer/Modal、Viewer、200% zoom |
| Performance | 桌面暖缓存首页和图片预算 | 冷启动、mobile、目录、Viewer、登录页面、production build |
| Evidence | JUnit、失败截图、trace、video、联系表 | 验收摘要、敏感信息约束、每个缺陷的最小复现 |
| Security | 已有 BFF/会话源码合同 | 运行时权限矩阵、重定向、注入、cookie 和 secret 泄漏 |
| Resilience | 基础错误/空态 | 依赖故障、慢网、超时、并发、重试和重启持久化 |

## 6. 执行顺序

1. 冻结并重置验收数据，补普通 Member 和稳定 Submission 状态。
2. 跑 Gate A；失败则立即停止并修复。
3. 跑 B1–B5 的只读/可逆旅程。
4. 跑收藏、投稿、管理 CRUD 等有副作用旅程，每组后校验数据库/API 真值并清理。
5. 跑安全、错误和依赖故障注入。
6. 一次性采集 mobile/desktop、Light/Dark 视觉矩阵并检查联系表。
7. 跑完整无障碍和性能矩阵。
8. 重启五进程，验证持久化和最终 smoke。
9. 输出验收报告：通过项、Blocked、P0–P3 缺陷、证据路径和最终结论。

## 7. 证据与报告

每次正式运行使用唯一 `GALLERY_E2E_RUN_ID`，保留：

- JUnit XML。
- 失败页面截图和视觉 diff。
- trace 与 video。
- console/pageerror/失败网络请求摘要。
- 性能 JSON。
- 视觉联系表。
- 一份最终验收摘要，列出每个 Gate 的 Pass/Fail/Blocked。

每个缺陷必须包含：

- 严重度和受影响角色。
- 最短复现步骤。
- 预期与实际。
- 路由、视口、主题和浏览器。
- 失败请求、状态码、trace id。
- 截图/trace/video 路径。
- 修复后的回归测试名称。

禁止事项：

- 不因测试失败直接更新视觉基线。
- 不把 retry 后通过当作稳定通过。
- 不把缺少 Docker 的关键项写成 Passed。
- 不使用已有登录状态污染 Anonymous/Guest 场景。
- 不让测试数据永久污染开发环境；有副作用用例必须清理或重新 seed。
