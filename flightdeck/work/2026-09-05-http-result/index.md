# Gallery 基础升级与站点复验

## Goal
升级Foundation/Identity/Asset消费与HTTP Result合同，完成真实图库发现、投稿、收藏、管理和失败反馈验收并提交。

## Status
Open

## Current
2026-09-09 角色授权标签使用共享 AuthorizationGrantBadge，删除独立来源字典，角色标签与中文来源 hover/focus 统一。真实本地桌面和手机验证通过，未部署线上。证据 `E:/tmp/yueli-docs-publish-20260909/grants-gallery-*`。

2026-09-09 用户已接受评论分栏并授权推广。Gallery 继续保留本地验收入口，Blog/Docs/BVideo 已接入同一共享布局；Gallery 未部署服务器。本段取代下文“待用户目检”的历史状态。

2026-09-09 评论分栏试版：按用户要求仅 Gallery 本地启用共享 `CommentModerationCollection layout="columns"`。作者/时间在上、正文在下；右侧来源、完整状态（含已通过）、横向三点菜单对齐，移除图片缩略图。审核通过位于菜单；回复摘要和用户详情、批量选择及统一分页保留。390/768/1440/3840 真实 Playwright 验证布局、菜单和选择，无横向溢出；共享与 Gallery 类型检查、Gallery 正式构建通过。证据 `E:/tmp/yueli-docs-publish-20260909/gallery-columns*`。当前 Session `20260908T035724Z-33104`，验收地址 `/manage/comments`。待用户目检，未推广、部署、提交。

已完成Gallery本地基础升级：70个HTTP operation、声明式错误目录、OpenAPI/Go/TS/i18n生成物、创建201/Location、无正文204、统一分页、批量Problem与安全前端/BFF失败反馈。Vue/Router显式对齐，删除旧Envelope解码与产品重认证逻辑。现有UI成果独立保存在bbd4ae8，未接入底部导航。

## Next
本地入口 http://gallery.dev.yuelili.test:3007，Account 为 account-gallery.dev.yuelili.test:3400，当前运行 Session 20260908T022926Z-33220。紧凑网格、图片列表单张加入专题、简化勾选及编辑内“发布状态”已完成真实浏览器验收，待用户目检；编辑可将合格草稿/隐藏图片公开，或下架已公开图片。专题选图 modal 已按用户要求删除。真正私密直链鉴权仍未实施，下架只隐藏目录。仅本地 Gallery，不部署、不推广其他站点、不提交或推送。首位认领按用户要求延期；保留[三项备忘录](../../../../workspace/flightdeck/knowledge/frontend/site-acceptance-checklist.md)。

## References
- [Context](context.md)

## Verification
- GOWORK=off完整Go测试与最终vet通过；新增批量cause保密、分页wire、validation保密测试和trace修正的定向Go测试通过。
- Web79项单测、最终Nuxt类型检查与独立构建目录生产build通过。
- Project -check验证70个operation；已发布v0.4.1 CLI的Go/TS/i18n生成检查通过。
- CLI Playwright累计47个有效用例通过：Journeys13、Management25、HTTP Result4、Responsive5；初次失败修复后只复跑受影响项，无未解决失败。
- 真实上传经同源Asset代理完成并返回投稿201；验证了专题201/Location、评论201/删除204、事件204、批量逐项Problem及一致traceId、收藏新增/页面显示/恢复、后台3次SSR刷新。
- 390/1440px图片编辑的字段错误、未知字段摘要、技术详情、保留草稿、隐藏raw detail和Axe通过；320/1920px及200%有效缩放检查通过。
- 修复Vue重复运行时导致的SSR/评论水合错误；旧测试的标签游标夹具改为当前DTO，专题测试选择有成员的夹具，工作区宽度按共享壳边界验证。
- 证据：web/test-results/e2e/http-result-*、http-result-fields-final/、http-result-favorites-final/、http-result-initial-evidence/和http-result-build.log。

## Runtime boundary
Gallery使用隔离Identity/Asset组合；原共享Asset源码/prepare指纹不匹配，原共享Provider保持运行。浏览器入口http://gallery.dev.yuelili.test:3007，后台/manage。当前Session为20260905T123000Z-43296，后端仅回环8391/8481/8482，Account3400。正式发布/镜像/远端CI不在本轮已完成范围。


当前工作树已有 api/internal/controller/comments.go、web/app/utils/image.ts、web/test/image.test.ts 的媒体合同相关改动，保留。历史 Runtime boundary 的 Session 已不在运行，不能直接复用。


## 本轮三项预检（2026-09-08）

- 认领：尚不符合。api/cmd/gallery/main.go 在没有 bootstrapAdministratorSubs 时 panic，尚无未认领实例显式认领流程。不得把当前种子管理员登录通过计作认领通过。
- preset：现有产品 image.ts 已有 SDK 调整；本地 Environment 媒体健康 URL 本轮已修正，待真实图片请求验收。
- 权限用户资料：尚不符合。authorization.vue 的申请行直接输出 subject，类型也未包含时间；需要共享展示及 Identity 资料读取方案后接入。
- 公共组件评估：Foundation UI 当前有 AccountMenu，但未发现列表用户资料组件。建议基础 UI 拥有纯展示，Identity 拥有批量公开资料请求与映射，产品传业务时间和操作。本轮用户先要求备忘录，尚未实施公共组件。

CLI Playwright 本轮冒烟通过：/images 加载 24 张有效图片、preset 媒体 200；真实测试用户登录后打开权限页及用户管理，确认当前仍显示 TestA123；桌面/手机截图完成，无 pageerror。证据 E:/tmp/yueli-media-preset-20260908/gallery-local-smoke.json 与 screenshots/gallery-local-*。这只证明入口、图片和登录可用，不代表三项或全部功能验收完成。


## 统一申请组件交付（2026-09-08）

已用 Foundation AuthorizationApplication 替换申请条目，AuthorizationUser 同时用于申请和授权用户行；Identity usePublicUserDirectory 统一读取items并分批去重，页面提供原位错误和重试。真实昵称、头像引用、主页、申请createdAt、申请理由和审批忙碌状态由共享组件呈现。审批仍走Gallery原有API，未改权限与数据库。

Foundation UI与Gallery类型检查、Gallery构建、Identity回归2项、CLI Playwright真实资料/主页/响应式/刷新/失败重试及夹具审批均通过。详情见[共享组件交付](../../../../foundation/flightdeck/work/2026-09-08-authorization-application/index.md)。旧预检中的用户资料缺失已由本次替代。当前为local overlay，未发布新包或部署；服务保留3007。


## 隔离 Account 头像/封面上传修复（2026-09-08）

用户报告上传服务不可用。Workspace CLI日志确认 Identity 转发 Asset 返回 asset.registration.not_found；隔离 Asset 只注册Gallery，Account媒体代理还默认指向共享8082。已修正 gallery-local/run.ps1 读取 Identity 自有 contracts/asset.consumer.json，Environment在Asset prepare中通过devprovision幂等注册account，Account NUXT_ASSET_BASE指向本组合Asset8482。服务均启动并健康，未修改产品实现、权限或数据库结构。

已用独立本地测试账号通过CLI Playwright在Account页面执行头像/封面选图、裁剪、确认、真实上传，两者均200；preset媒体交付均200，刷新后头像解码与封面持久化通过。未替换用户头像或封面。证据 E:/tmp/yueli-media-preset-20260908/gallery-account-media-report.json、gallery-account-media.mjs、screenshots/gallery-account-media-fixed.png。仅重启本会话Gallery隔离组合，共享Provider与Blog保持运行；无服务器部署。


本地431及公开地址保存500已修复，详见[Identity修复证据](../../../../identity/flightdeck/work/2026-09-07-provider-administration/references/profile-header-fix.md)。8个3000字节附加Cookie下头像/封面上传、profile保存和公开主页访问通过。正式站修复尚未部署，保留本地验收范围。


当前本机独立域名与20主机Cookie共存验收完成；此前64KiB请求头临时放宽已移除。[本地浏览器域名交付](../../../../workspace/flightdeck/work/2026-09-08-local-browser-origins/index.md)。账户入口 account-gallery.dev.yuelili.test:3400；旧共享Provider和线上服务未改动。

2026-09-08 共享组合域名迁移时由 CLI 停止并恢复 Gallery，当前 Session 20260907T231429Z-32212，原独立域名/端口与数据保留。真实浏览器 SSO/20 主机 Cookie 隔离复验通过；最大产品 Cookie 1769 字节。

## 投稿流程展示调整（2026-09-08）
用户确认投稿流程使用上传中/处理中/处理失败/等待审核/审核未通过/审核通过。前端共用 submissionStatus 映射，终态审核优先于遗留 safetyState；移除后台“安全不确定”页签、安全筛选和安全状态徽标。后台默认等待审核（ready + pending），提供处理失败、审核未通过、审核通过、全部投稿；全部查询以显式 all 表示并在调用 API 前消去，防止路由 codec 把空条件还原成默认待审。用户投稿页展示同一状态，保留拒绝原因、失败说明、重复收录及撤回事实；上传完成不再误称整个流程已完成。
数据库、API 和后端发布限制没有变动，blocked/unavailable/pending 仍不能批准。不伪造安全检测通过，不迁移历史业务记录。当前只做本地 Gallery，未部署、未提交。
验证：19 项投稿状态/批量/失败测试通过；typecheck 和 Nuxt 生产构建通过。CLI Playwright 真实本地登录、后台 API 200、默认等待审核及全部筛选刷新保持通过；独立浏览器响应夹具验证通过/拒绝、拒绝原因必填、处理失败筛选、投稿人终态和原因，390/1280px 无横向溢出。夹具审核不修改数据库中的真实投稿。首次浏览器夹具未覆盖 review 子路由，修正拦截范围与返回 DTO 后通过；构建触发本地 Nuxt 热重启后重新验收通过。
证据：E:/tmp/yueli-media-preset-20260908/gallery-workflow-report.json、gallery-workflow-real-report.json、gallery-workflow-{typecheck,build}.log 及 mjs/截图。当前 Gallery 原域名与运行组合保持，下一步继续用户本地目检。

## 专题选图弹窗（2026-09-08）
用户反馈添加图片卡片过大、没有点击反馈并要求预览。根因是直系 img 的 size-full 占满卡片高度，挤走标题/添加操作，图片无点击行为。改为独立 4:3 缩略图按钮，桌面四列、手机两列；单击图片/标题或键盘选择，明确边框/勾选状态与底部已选数量，批量确认一次调用既有 members API。预览按钮打开 object-contain 大图（正式 display preset），关闭不改变选择；提交时禁用重复提交与取消，失败保留选择。现有身份与资源协议未改，无 SQL 迁移，仅本地。
真实 CLI Playwright：新建私有“选图弹窗验收”专题 AaB-XesYcVOelCcNsjXjZA，桌面四列、手机两列无横向溢出；单击与 Space 选中两张，大图资源实际加载成功，Escape 返回保持选择；members POST 200，API 复查两条成员，页面刷新通过。未修改用户现有专题。证据 E:/tmp/yueli-media-preset-20260908/gallery-picker-report.json、gallery-picker.mjs 及截图。类型检查在显式 void 点击回调修复后通过，Nuxt 生产构建通过；随后仅缩短弹窗操作提示，避免窄屏标题区拥挤。


## Gallery 紧凑网格（2026-09-08）
用户要求先在本地 Gallery 验收“小而美”网格，满意后才推广。保留 Foundation CollectionPanel，在 Gallery main.css 用 opt-in 类覆盖 data-collection-grid 的固定三列：auto-fill 最小 12rem，窄容器 9rem，避免少量图片被拉伸铺满。图片与专题成员缩减卡片留白；投稿审核新增可持久化 list/grid 切换，复用既有审核权限、操作和筛选。GalleryGridToolbar 是本产品组件，网格显示选择本页/排序依据/方向；列表保留列排序，专题保留人工顺序。未修改 Blog、Foundation 或线上运行。
图片徽标按 publicationState 显示已公开/草稿/已隐藏/已删除，移除覆盖主状态的“待生成”；只有已公开但 publicRenditionReady=false 才额外提示“公开版本尚未就绪”，不把草稿解释为生成任务。
CLI Playwright 真实本地账号和数据库：1280px 四列/219.5px、3840px 十七列/193px、390px 两列/162px，无横向溢出和 pageerror；排序方向、选择24条/清除、编辑打开关闭、列表切换、既有测试专题两成员四列占位、20条真实审核数据网格及刷新保持均通过。未操作真实审核或修改用户图片。首次脚本在编辑弹窗刚打开就 Escape，关闭时序未就绪导致后续找不到背景按钮；等待表单再关闭后通过。证据 E:/tmp/yueli-media-preset-20260908/gallery-compact-report.json、gallery-compact.mjs、screenshots/gallery-compact-*.png。
公开/私密切换尚未实施：目前 Gallery HideImage 只下架目录，Asset publications 只有 Publish、没有撤销公开交付接口。此前已询问用户私密是否需要直链鉴权，尚未收到选择；不得仅把 hidden 改名私密。下一步先由用户目检网格并明确私密访问语义，再实现对应权限链路；不推广其他站点。

最终代码 typecheck、Nuxt 生产构建及 git diff --check 通过；最后的构建未再改动已完成浏览器验收的源代码。构建日志为 E:/tmp/yueli-media-preset-20260908/gallery-compact-{typecheck,build}.log。


## 专题选图真实分页（2026-09-08）
用户发现弹窗只能看到12张。根因是 searchImages 从公开目录扫描后累积到12条就停止，既无页状态也无翻页入口；上一轮只验收了卡片和添加行为，漏验超过一页的数据。
已直接使用公开图片 API 的 page/size/total，每页12条；固定弹窗页脚提供总数、当前页/页数和首页/上一页/下一页/末页。搜索提交回第一页，翻页沿用已提交的搜索词；Set 选择跨页/搜索保留。已知专题成员原位显示“已添加”并禁选，避免前端过滤破坏分页总数；预览仍可用。请求序号防止旧搜索覆盖新结果，失败在弹窗原位提示并可重试，选择不丢失；总数缩小时回到合法末页。
类型检查、Nuxt生产构建、diff check通过。真实CLI Playwright：84条公开图片共7页，第二页与第一页内容不同；跨页选2条返回首屏保持勾选；模拟一次503后原位重试恢复第二页选择；第7页和搜索回第一页验证通过，搜索后选择仍为2；真实members POST 200，API复查专题 AaB-oJiocZ-VQq5T_ADZYA 含2个成员，重开弹窗已有图片禁选。1280/390截图及手机无横向溢出通过，无pageerror。仅本地 Gallery，没有修改用户原有专题或其他站点。
最初测试错误假定 sibling-count=0 会显示数字2，随后误用了会被子组件覆盖的data-slot定位；读取实际DOM后按分页按钮无障碍名称定位，完整验收通过，无产品代码变更。测试过程创建的独立私有“选图分页验收”专题保留（前两次为空，最后一个用于成功验收）。
证据 E:/tmp/yueli-media-preset-20260908/gallery-picker-pagination-report.json、gallery-picker-pagination.mjs、gallery-picker-pagination-{typecheck,build}.log、screenshots/gallery-picker-pagination-{desktop,mobile}.png。公开/私密语义仍是此前待用户明确的独立事项，不影响本次分页交付。


## 从图片列表加入专题与简化选择框（2026-09-08）
用户明确替换交互，不再保留专题内选图 modal。删除该弹窗及其搜索/分页/选择/预览状态和调用代码；此前“选图真实分页”记录为历史，已被本次替代。专题页提供“前往图片列表添加”链接，跳到已公开图片列表。图片 rows/grid 的编辑按钮右侧增加 folder-plus 单张添加入口，选择目标专题后复用既有 members API；非公开图片禁用该入口并说明先公开。既有批量加入专题继续使用。单张操作用独立目标，不替换/清空背景批量勾选；模态内显示失败并防重复提交。未改变 API、数据库或公开/私密权限。
Gallery 本地 CSS 仅覆盖紧凑网格卡片的直属 checkbox root，移除 padding、背景、shadow、backdrop-filter 并收紧位置；不修改复选框本体/键盘行为、列表选择框、共享 Foundation 或 Blog。专题序号缩小并邻接勾选框；图片底部操作允许窄屏换行。
类型检查、Nuxt 构建与 diff check 通过。CLI Playwright 使用现有独立测试专题 AaB-oJiocZ-VQq5T_ADZYA：单张 members 200 并持久化且保留另一个勾选，随后批量 members 200 并持久化（专题从2张增至4张）；列表单张入口、专题链接/旧modal移除、透明背景/零padding/无阴影与模糊的计算样式断言、390px无横向溢出与1280px渲染通过。额外只读截图等待页面稳定并验证勾选/取消，无真实用户图片变更。最初脚本把 USelectMenu 触发器误认为 combobox，检查 DOM 后改用其真实 button 角色通过。证据 E:/tmp/yueli-media-preset-20260908/gallery-image-collection-report.json、gallery-image-collection{,-final}.mjs、gallery-image-collection-{typecheck,build}.log 和 screenshots/gallery-image-collection-*、gallery-collection-clean-selection.png。只在本地 Gallery，未部署、未推广。


## 编辑图片公开与下架（2026-09-08）
此前只有审核自动发布和批量下架，主状态展示修正没有补齐用户主动公开入口。现于编辑图片加入“发布状态”：草稿/隐藏可选择公开，已公开可选择下架（隐藏），保存更改后生效。PATCH AdminImageUpdateInput 增加可选 publicationState（仅 published/hidden），未改变时不传；没有新端点或数据库迁移。Controller 除原 image.update 外，公开还要求 submission.review，下架要求 image.hide；Operator 由登录身份提供，内部 PublicationReady 不进入公开 JSON/OpenAPI。
Service 校验当前版本、未删除、处理 ready、审核 approved/not_required、安全 safe。公开资源尚未就绪时先调用正式 Asset PublishImage 成功后再请求 DAO。DAO 同一事务以版本和公开资格谓词防竞争，更新文案/分类/发布状态/时间/资源标记；下架同时生成操作人处置记录。下架仍不撤销已公开媒体 URL，不假称私密。所有错误复用现有目录和表单反馈。
验证：Go全量测试及vet通过；显式连接192.168.5.5:5433的PG18集成测试通过（发布/公开读取/下架隐藏/处置记录/过期版本/拒绝安全未通过与事务不部分保存），Service新增8种发布资格检查通过。Nuxt类型检查、构建、diff check通过。OpenAPI只新增可选字段，由本地Foundation httpcontract CLI生成并通过project -check（70 operations）；已发布CLI未含-project标志，未更改依赖版本或Foundation实现。
真实CLI Playwright新上传独立测试图 publication-acceptance-1788832979382，submission AaB-wRrjcKKm3dx295gntQ，经处理和审核生成 image AaB-wR-OeWmbKSPW7h3Pjg；仅对这张测试图片用受标题+ID约束的本地fixture脚本设为draft/publicRenditionReady=false。编辑选择公开并保存200，返回published/ready=true，前台API与图片页面可见；下架保存200后前台404，编辑再次公开200；匿名PATCH被拒绝，390px编辑状态可见且无横向溢出。公开页面实际图片已加载并截图；用户原有图片未改。
Workspace CLI只重启Gallery隔离组合至20260908T015919Z-1444，8481/3400/8482/8391/3007健康，共享Identity/Asset及Blog PID保持。证据 E:/tmp/yueli-media-preset-20260908/gallery-publication-report.json、gallery-publication{,-final}.mjs、gallery-publication-{go,all-go,vet,typecheck,build,contract-check,runtime}.log，及screenshots/gallery-publication-{edit,public,mobile}.png。未部署、未提交或推广。


## 批量上架与专题序号（2026-09-08）
图片批量操作新增“上架（公开）”和确认说明，调用既有bulk端点的publish action。每项加载当前版本并复用单张UpdateAdminImage公开条件、资源准备和乐观锁，保留逐项结果、去重/60项上限和失败选择；Controller要求image.update及submission.review，与单张公开相同。专题grid序号从左侧移至右上角，封面标记移到左侧勾选框旁，避免覆盖。仅本地Gallery，没有迁移或改动其他站点。
后端定向测试（批量成功/无效ID逐项失败/重复ID/审核条件）与既有controller测试通过；Nuxt类型检查、构建、70 operation合同check、diff check通过。Workspace CLI重启本地隔离组合至20260908T022926Z-33220并健康，共享Provider及Blog保持。
真实CLI Playwright以先前独立测试图AaB-wR-OeWmbKSPW7h3Pjg下架后批量上架成功，公开API200；浏览器列表附加一个不存在ID作为失败夹具，同一次真实bulk请求返回一成功一失败，成功项取消勾选，失败项继续选中。非真实用户数据变更。桌面截图与位置断言确认专题序号在右上，390px无横向溢出，无pageerror。
证据 E:/tmp/yueli-media-preset-20260908/gallery-bulk-publish-report.json、gallery-bulk-publish.mjs、gallery-bulk-publish-{go,typecheck,build,contract,runtime}.log 及 screenshots/gallery-collection-number-right{,-mobile}.png、gallery-bulk-publish-confirm.png。

## Gallery 评论管理紧凑布局（2026-09-08）
用户授权整理杂乱的评论后台。Foundation CommentModerationCollection 新增可选 compact 布局，默认 table 保持；仅 Gallery 显式启用。头像/昵称/时间集中，正文为主体，图片来源放下方，邮箱默认折叠于“用户详情”，状态与审核/更多操作右侧排列；保留共享搜索、状态导航、排序、选择、批量与分页。回复展示 @原作者和最多160字符原文摘要。
AdminComments 通过同图父评论 LEFT JOIN 提供摘要，排除已删除原文；API仅追加可选 parentAuthorName/parentContent，无数据库迁移。原作者账户资料与当前评论一起批量解析，避免注册用户 author_name 为空误报原评论不可用。原评论不在当前搜索/分页中也能展示。
验证：Foundation UI类型检查及104测试、Gallery类型检查和Nuxt构建通过；Go comments/controller/DAO（显式PG18依赖）通过，新增筛选后父文上下文/摘要限制/已删除原文不返回及跨页父作者资料测试。70 operation合同生成与check通过。首次PG测试夹具漏写公开时间，改为draft测试图后通过；首次真实浏览器发现注册作者空字段，补齐资料解析后重验通过。
CLI Playwright真实本地账号：6条评论紧凑渲染、邮箱展开收起、筛选回复原作者与摘要、全选取消、操作菜单、时间升序、待审核筛选与通过入口、单页分页禁用通过；390/1280/3840无横向溢出与pageerror。没有修改现有评论状态。筛选测试最初误用tab定位，核对共享组件为button后通过。截图已目检桌面与手机。
Workspace CLI仅重启Gallery隔离组合，当前session=20260908T025402Z-35924，五进程健康；共享Provider与Blog保持。证据 E:/tmp/yueli-media-preset-20260908/gallery-comments-compact-report.json、gallery-comments-compact.mjs、gallery-comments-controls.mjs、gallery-comments-compact-{go,typecheck,build,contract-check,runtime}.log 和 screenshots/gallery-comments-compact-{desktop,mobile}.png。未部署、提交或启用其他站点布局。下一步用户本地目检。

## 评论顶部工具栏试版（2026-09-08）
用户提供参考图，要求借鉴右上搜索/筛选布局但用色少、干净，先只看本地Gallery。Foundation新增CommentModerationToolbar，复用CollectionTableToolbar原有搜索与中文输入法处理；TableToolbar增加header展示与externalControls选项，CollectionPanel透传外置控制选项，默认行为不变。Gallery把搜索、评论状态下拉、时间排序放ManagePage actions；移除重复状态导航和列表搜索行，列表顶部只保留选择本页与结果数。清空回收站保留在对应状态的列表头，批量区仍复用共享实现。评论行保持头像/昵称/时间/正文/来源/回复摘要结构，未引入渐变、彩色边条、大胶囊或虚构等级。
本次仅UI，无API/数据库修改、无部署。Foundation类型检查与104测试通过，Gallery最终类型检查、Nuxt构建与diff check通过。CLI Playwright真实登录：标题右侧排列、状态过滤、时间排序、搜索及回复摘要、批量选择、清空回收站入口通过；390/768/1440/3840验收，无pageerror。首轮截图发现移动操作区收缩裁切，补齐Gallery标题操作区响应式宽度后最终检查每个输入/筛选/排序控件完整位于视口内，桌面与手机截图复核通过；测试未改评论数据。
证据 E:/tmp/yueli-media-preset-20260908/comments-header-report.json、comments-header.mjs、comments-header-{ui-typecheck,ui-tests,typecheck,build}.log 和 screenshots/comments-header-{desktop,390,768,3840}.png。共享能力仅Gallery启用，其他站点保留原布局。运行Session沿用20260908T025402Z-35924，前端热更新，等待用户本地目检后再决定推广。

## 评论工具高度与右侧状态操作（2026-09-08）
按用户截图继续修正：评论头部搜索、状态下拉、排序按钮统一36px高度（仅header展示）；compact评论状态采用24px高、12px文字和小圆点的淡色标签，与横向三点菜单同行，通过按钮独立放下一行。AdminRowActions仅新增可选overflowIcon，默认竖向菜单保持，不改变其他站点。保留真实状态含义，回收站不改成已完成。
Foundation类型检查、104测试和Gallery类型检查/构建通过，相关diff check通过。CLI Playwright实测390/768/1440三种宽度，三个控件均为36px；标签字号至少12px、标签菜单垂直居中、通过位于菜单下方、控件完整在屏幕内、菜单开关正常且无pageerror。图标首轮测试错误假定CSS类名为i-tabler-dots，实际DOM为i-tabler:dots，检查真实DOM与桌面手机截图确认横向图标。未修改评论数据或部署。
证据 E:/tmp/yueli-media-preset-20260908/comments-controls-fix-report.json、comments-controls-fix.mjs、comments-controls-fix-{ui,tests,typecheck,build}.log、screenshots/comments-controls-fix-{390,768,1440}.png。仍仅本地Gallery启用。

## 评论状态移至时间后（2026-09-08）
用户确认右侧信息过密，将compact状态标签移动到左侧作者/时间信息行、紧跟time，右侧只保留横向菜单和下一行通过操作。标签高度24→18px、文字11px、圆点4px、水平留白6px；窄屏允许自然换行，不缩成不可辨认的半尺寸。默认table和其他站点不变。
Gallery类型检查、Nuxt构建和相关diff check通过。真实CLI Playwright 1440/390：标签位于time之后且不在右侧操作区，实测18px高/11px字，桌面时间后同行，手机无溢出；菜单打开关闭和通过位于菜单下方均通过，无pageerror。已目检两种截图，未修改评论数据。证据E:/tmp/yueli-media-preset-20260908/comments-badge-left-report.json、comments-badge-left.mjs、comments-badge-left-{typecheck,build}.log、screenshots/comments-badge-left-{1440,390}.png。仍本地Gallery验收，未部署。

## 评论阅读层级新版（2026-09-08）
用户确认参考Ghost内容层级（https://ghost.org/changelog/comment-moderation/）重新做本地Gallery。compact行改为32px头像、加粗昵称、14px正文/12px次要信息；状态继续在时间后，回复摘要细线弱化。正文下方集中来源、通过和更多操作，邮箱资料按需展开；桌面右侧80×60来源缩略图，手机隐藏。沿用右上搜索/状态/排序与共享批量/分页，默认table和其他站点不变。
AdminComments现有图片JOIN追加imageAssetId可选字段，仅已准备公开派生且未删除的图片提供引用；复用Gallery现有galleryRendition媒体入口，没有复制媒体编码或新建上传逻辑，无数据库迁移。Foundation来源Adapter只新增thumbnailUrl，不感知Gallery资产协议。
Gallery类型检查/Nuxt构建、Foundation类型检查与104测试、Go comments/controller/DAO PG18测试、70 operation合同生成/check和diff check通过。Workspace CLI重启仅Gallery隔离组合至20260908T035724Z-33104，服务健康；共享Provider及Blog保持。
CLI Playwright真实账号及现有6条评论：1440/390/3840尺寸，正文14px、操作位于正文下方、桌面6张80×60缩略图实际加载且来源链接可用，手机隐藏且无横向溢出；用户详情展开收起、全选取消/批量区域、更多菜单、待审核过滤及通过入口均通过，无pageerror。已目检桌面和手机截图，未改变评论数据。证据E:/tmp/yueli-media-preset-20260908/comments-reading-report.json、comments-reading.mjs、comments-reading-{typecheck,build,go,ui,ui-tests,contract,runtime}.log、screenshots/comments-reading-{1440,390,3840}.png。未部署或推广，等待用户目检新版。

## 评论移动分页收紧（2026-09-08）
用户确认新版其余布局可用，要求手机分页缩小，目检后才推广。CollectionPanel新增compactPagination可选项，当前由compact评论布局启用，默认其他集合不变。小于640px翻页按钮28×28、每页数量控件28px高/72px宽，控件同一行弹性换行，统计行在上；页码siblingCount=0保留首尾/前后操作，避免页数增多撑宽。
Gallery/Foundation类型检查、Gallery生产构建、Foundation104测试和diff check通过。CLI Playwright真实6条评论验证390px按钮与选择器均28px且同行、单页下一页禁用，截图目检通过；只读66条浏览器响应夹具验证下一页21号起、末页6条、末页禁用、每页40条路由更新，无溢出和pageerror，不修改数据库。证据E:/tmp/yueli-media-preset-20260908/comments-pagination-report.json、comments-pagination.mjs、comments-pagination-{typecheck,build,ui,tests}.log、screenshots/comments-pagination-mobile.png。仅本地Gallery，等待用户确认再推广。

## 移动分页图标居中（2026-09-08）
用户反馈28px分页图标未居中。真实DOM检查发现去padding后按钮justify-content=normal，图标/数字靠左，纵向原本已居中。compactPagination窄屏规则明确inline-flex、align-items:center、justify-content:center，仅改三条CSS。Nuxt构建、diff check和CLI Playwright通过；390px五个按钮的图标/数字中心与28×28按钮中心偏差dx=0/dy=0，截图复核通过。证据E:/tmp/yueli-media-preset-20260908/pagination-centered-report.json、pagination-centered.mjs、pagination-align-build.log、screenshots/pagination-centered.png。仍仅本地Gallery，未推广。

## 评论提交与图片顶部工具栏试版（2026-09-08）
用户确认评论并要求提交，之后把图片搜索筛选移到右上，目检满意才推广。已按范围提交Foundation b159c38（compact评论、顶部工具栏、分页）与Gallery b37c7de（评论API/回复上下文/来源缩略图/页面）。Gallery OpenAPI只暂存评论字段两个hunk，未把其他未提交发布字段混入；其他历史修改保留，未push。
图片页复用CollectionTableToolbar header展示，ManagePage actions容纳搜索、筛选、排序依据/方向、列表网格切换与投稿入口。公开状态/分类/维度使用原controls与changeControl，收进筛选Popover；默认表格列排序保留，grid表头去重复排序只留选择本页/总数。CollectionPanel使用externalControls和compactPagination，批量条仍可用。仅Gallery图片页改动，未推广。
类型检查最初发现UButton点击赋值返回字符串，显式void后类型检查与Nuxt构建通过；最终diff check通过。CLI Playwright真实账号：公开状态筛选与刷新保持、分类/维度入口、清除、标题排序/方向、搜索、列表/网格切换、批量全选取消通过；390/768/1440/3840控件在视口内、搜索筛选排序均36px、无横向溢出/pageerror，桌面与手机截图目检通过。测试最初错误期待默认列表写view=list，实际codec省略默认值，改按无grid参数与无grid卡片检查后通过。未修改任何图片数据。
证据E:/tmp/yueli-media-preset-20260908/images-header-report.json、images-header.mjs、images-header-{typecheck,build}.log、screenshots/images-header-{desktop,list,390,768,3840}.png。图片试版尚未提交，等待用户目检确认；全站推广仍未授权开始。

## 图片工具栏改为按钮弹窗（2026-09-08）
用户否定常驻排序下拉/方向的重复展示，明确参考Eagle改成搜索→筛选→排序→列表/网格组，筛选与排序按钮点击modal。Gallery移除顶部常驻排序控件与筛选Popover，增加筛选图片、图片排序两个紧凑UModal。筛选使用公开状态/分类/维度草稿，排序使用依据与方向单选草稿，应用才更新统一CollectionQuery，取消丢弃，重置仅改变草稿；保留原表头快捷排序，弹窗打开读取其当前值。视图切换不会重置筛选/排序。仅图片页面，无共享库或API新变化。
类型检查、Nuxt构建、diff check通过。CLI Playwright真实登录检查筛选改值未应用无路由变更/取消、应用公开筛选、排序标题升序、网格列表切换、表头改降序后弹窗同步；390/768/1440/3840控件完整在视口，36px一致，手机modal及应用按钮可见，无pageerror，截图目检通过。最初测试错用表头可见label而非动态aria-label，并误以为默认desc保留URL参数，核对代码修正测试后通过。未改图片数据。证据E:/tmp/yueli-media-preset-20260908/images-modal-tools-report.json、images-modal-tools.mjs、images-modal-tools-{typecheck,build}.log和screenshots/images-modal-tools-*、images-sort-modal.png、images-filter-modal-mobile.png。未提交图片试版、未推广，等待用户目检。

## 图片视图切换选中态（2026-09-08）
用户澄清列表/网格实际切换正常，但视觉上看不出哪个被选中，并要求白色底。CollectionViewToggle新增可选appearance=surface，仅Gallery图片页启用：默认表面底座，选中项浅primary底、primary图标与内描边，未选中中性透明；32px内部按钮居中，避免图片header的通用36px规则覆盖，保留aria-pressed和路由状态。其他消费者默认不变。
Gallery类型检查、构建、相关diff check通过。CLI Playwright390/1440分别切换list/grid，验证唯一aria-pressed=true、选中与未选中的计算颜色/背景不同、浅色底座white、实际对应视图渲染；刷新grid保持，无横向溢出/pageerror，截图复核通过。证据E:/tmp/yueli-media-preset-20260908/view-selection-report.json、view-selection.mjs、view-selection-{typecheck,build}.log、screenshots/view-selection-{390,1440}-{list,grid}.png。仍本地图片试版，未提交或推广。

## 窄屏投稿入口对齐标题（2026-09-08）
用户要求低分辨率将投稿图片放右上。Gallery图片header小于1280px使用两列grid，actions容器display:contents，投稿链接放标题同排右侧，搜索工具栏占下一整行；DOM内投稿先于工具栏，桌面order保持工具后投稿。没有复制链接或绝对定位。
Nuxt构建、diff check通过。CLI Playwright390/648/1100/1440测量投稿与标题纵向中心偏差小于5px、投稿始终在右侧，小于1280px搜索位于下一行、唯一投稿链接目标/submit，无溢出/pageerror，390/648截图复核通过。证据E:/tmp/yueli-media-preset-20260908/images-submit-position-report.json、images-submit-position.mjs、images-submit-position-build.log和screenshots/images-submit-position-*.png。未提交图片试版或推广。

## 紧凑分页恢复数字页码（2026-09-08）
用户反馈图片分页只有1。真实API105张/每页24/总5页且下一页启用，根因是compactPagination将siblingCount=0并关闭边缘页码，视觉只呈现当前页；上一轮只验证箭头操作未覆盖页码可发现性。恢复siblingCount=1，compact显示边缘页码与总页数，隐藏重复首尾箭头（保留数字首尾页和前后箭头），保留28px居中与窄屏换行。
Gallery类型检查、Nuxt构建、Foundation104测试与diff check通过。真实CLI Playwright显示1–5全部数字与共5页，直接点击第2页内容不同、点击第5页返回9张且下一页禁用、数字1返回首屏；320/390/1440截图与无横向溢出、无pageerror通过。证据E:/tmp/yueli-media-preset-20260908/image-pages-report.json、image-pages.mjs、image-pages-{typecheck,build,ui-tests}.log、screenshots/image-pages-*.png。没有修改数据，仍本地Gallery，未提交或推广。

## Gallery五类列表统一分页条（2026-09-08）
用户要求图片/评论/投稿/专题/处理单统一分页，抽公共组件，删除“显示范围、总数、总页数”，每页数量放左。本轮已明确询问是否同时统一各页顶部工具栏，尚无回复；先完成明确的分页范围，不将其推断成全站推广。
Foundation新增CollectionPaginationBar，统一左侧数量/右侧数字分页与窄屏换行；复用CollectionPagination增加compact选项，集中拥有28px居中、首尾数字/省略号、前后箭头与合法页码判断。CollectionPanel compactPagination委托该组件，移除其重复分页样式与统计文案；旧非compact消费者维持现状。
Gallery图片/评论沿既有compact接入，投稿启用compact；处理单改用公共条、新增20/40/60每页数量并随query请求/重置页码；专题列表增加本地切片分页及搜索/筛选/数量变更回首屏；专题内图片也接入，先完整读取服务端60条分段，再本地分页/搜索，防止超过60条缺失。其全选只针对当前页，保留跨页已选集合；已有大专题排序权限限制保持。
Foundation类型检查、原104测试和新增2项分页边界/合法数量测试通过，Gallery最终类型检查/Nuxt构建/diff check通过。CLI Playwright实测六入口（含专题内图片）390px公共条、无统计文案、左侧数量/右侧页码、每页数量切换及无溢出/pageerror；45专题只读夹具验证第3页5条/第41条起，100页图片夹具只显示当前邻页与首尾/省略号；66成员夹具验证按服务端1/2页取全、用户第4页可见第66张。没有写入业务数据，截图目检通过。
证据E:/tmp/yueli-media-preset-20260908/unified-pagination-report.json、unified-pagination-members-report.json、unified-pagination{,-members}.mjs、unified-pagination-{ui-typecheck,ui-tests,typecheck,build}.log、screenshots/unified-pagination-*.png。未提交本轮改动或推广其他站点。
