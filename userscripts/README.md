# Gallery 图片发布

[通用油猴脚本](gallery-publish.user.js)，适用于桌面 Tampermonkey。页面图片悬停后右上角出现“发布”，点击打开预览和分类面板，依次取图、上传、投稿、等待图片处理并审核公开。没有网页图片的“已投稿”标记，也不扫描 Gallery 的历史记录。

## 安装和使用

1. 在油猴中新建脚本，粘贴完整脚本并保存。更新时覆盖同名脚本，不与旧版同时启用。
2. 在油猴菜单选择“Gallery 发布设置”，保存 Gallery 开发者令牌。个人安装版已预填初始令牌；之后修改、清除都使用油猴存储。
3. 令牌选择“月离图库” (`gallery-main-web`)，需要 `media.upload`、`gallery.submission.create`、`gallery.submission.read_own`、`gallery.submission.review`。账号本身也须有审核资格。
4. 把鼠标移到图片上，点“发布”。确认预览、标题、分类、标签及来源页面，再点弹窗底部“发布”；这会上传并通过审核公开图片。
5. 勾选“记住本次分类”后，下一张图片沿用分类和 Facet 选择；每次仍从 Gallery 读取可用分类及约束。

## 新增分类、作品和角色（1.1.0）

“分类”和每个现有维度标题右侧提供“＋ 新增”。例如在“作品”旁新增作品、在“角色”旁新增角色；不是创建新的维度类型。标签继续使用原有自由输入和服务端标签提案流程。

输入名称后自动生成可编辑的链接标识；点“保存并选中”会将项目加入 Gallery 的全站目录并选中，不只是保存在脚本本地。此功能额外需要 `gallery.classification.read` 和 `gallery.classification.govern`，账号也须有相应治理权限。

新增遵循正式三步合同：读取分类目录检查同范围名称/标识；`POST /api/v1/gallery/admin/classification/identities` 创建草稿；对该 ID 以 `set_status → active` 调用 `governance/preview`，把返回的 catalog revision / request token / impact token 原样传入 `governance/execute`。只有刷新公开投稿选项确认项目可用后，才选中新项目。保留其他已选分类、维度、标题等字段。

已启用的同名项目直接选中；标识属于其他名称时要求修改。已有未启用同名项目提示到管理目录处理，不擅自启用。当前任务刚创建的草稿若启用失败，可以“继续完成”，重新读取状态与治理计划，避免重复创建；创建结果未知时停止重发，先在管理目录核对。取消不删除已经保存的目录草稿。

1.2.0 起，作品与角色使用同一个 `work` 维度：作品／系列是根节点，角色放在其所属作品下。新增角色必须选择所属作品；没有所属作品的选“独立角色”，出处暂不明确的选“待归属角色”。作品和角色都可以单独选中，多角色图片可以多选。“游戏/动漫”等分类保持独立，联动不改变角色的主要归属。新增分类和其他维度值也可选择父级。

此版本需要先应用 Gallery 的 `0023_work_character_tree` 迁移并更新网站；检测到旧的独立 `character` 维度会停止发布并提示先更新。不要在旧站上用新版脚本新增角色。升级与数据处理见[作品与角色树](../docs/work-character-tree.md)。

支持主页面中显示尺寸至少 80×80 的 `img` / `picture` 图片和单一 `background-image`。使用浏览器当前加载的图片 URL，不猜测网站原图地址；不处理 iframe、canvas、伪元素或多层背景。源图片支持 JPEG、PNG、WebP 静态图片，GIF/SVG 和动画 PNG/WebP 会拒绝。客户端上限 20 MiB，Asset Profile 可有更严格限制；遵循 Gallery 当前主路径，转为长边不超过 4096px、去除元数据的 WebP 后上传。

## 请求和凭据边界

`@connect *` 用于读取用户所点图片的来源及 Asset 返回的对象存储上传地址。图片下载不带 Gallery PAT、登录 Cookie 或自定义网站凭据；需要登录、防盗链或其他站点限制导致无法下载时，显示失败，不绕过限制。

PAT 只发送到固定 `https://gallery.yuelili.com/api/v1/*` 与 `https://asset.yuelili.com/api/v1/*`，附带 `X-Yueli-Token-Site: gallery-main-web`；这些 API 请求遇到重定向即停止。上传二进制使用 Asset 返回的请求头和地址，不附加 PAT；Asset 自有 `/api/v1/assets/blob/` 按其公共上传合同归一到 Asset 正式 HTTPS Origin，外部存储的预签名地址原样使用。脚本使用封闭 Shadow DOM 承载设置，不把令牌写入网页 Cookie、localStorage 或日志。

公开脚本不内置任何凭据。个人安装版保存在用户 `.secret/yueli/`，包含初始令牌，不分享、不提交；当前凭据来源和个人安装入口由该分类 README 维护。已有 `.secret/yueli/gallery-publishing.env` 仍是本机凭据真值；更新令牌后可重新生成个人副本，已安装脚本需在油猴设置中更新。

## 真实接口顺序

- 读取 Gallery `GET /api/v1/gallery/submission-options`，按返回的分类和 Facet 数量约束组装正文。
- 匿名下载所选图片、验证格式/动画并转换。
- Asset `POST /api/v1/asset-references/drafts` 创建临时上传草稿。
- Asset `POST /api/v1/assets/upload-init`，使用 `siteKey=gallery`、`profileKey/category=gallery-submission`、`visibility=private` 和预处理 WebP。
- PUT 二进制，再 Asset `POST /api/v1/assets/finalize`，带 `uploadToken` 与 `draftSessionId`。
- Gallery `POST /api/v1/gallery/submissions`，带正式 `assetId`、标题、说明、来源、分类、Facet 和标签。
- 每两秒从 `GET /api/v1/gallery/me/submissions?page=1&size=100` 读取当前已知投稿，最多等待两分钟；按正式审核合同，`ready` 且安全状态为 `safe` 或 `uncertain` 后调用 `POST /api/v1/gallery/admin/submissions/{id}/review`，`decision=approve`。是否准许发布仍由服务端判断；处理失败、安全阻断不会发起审核。
- 只有服务端确认同一投稿 `published` 或 `duplicate` 才显示成功并提供公开图片链接。

准确合同见[开发者令牌指南](../docs/developer-tokens.md)、[开发者 API 目录](../docs/developer-api/index.md)与 Asset [对象引用和临时草稿](../../asset/docs/object-reference-lifecycle.md)。网络能力以 [Tampermonkey GM_xmlhttpRequest 文档](https://www.tampermonkey.net/documentation.php?q=GM_xmlhttpRequest)为准；脚本独立计时中止请求，避免 Chromium 的 fetch 模式忽略原生 timeout。

## 失败和继续

已保存投稿后处理较慢或审核失败，可在同一弹窗点“继续发布”，只继续当前投稿，不重新上传。审核权限不足时，可在弹窗内更换令牌再继续；仍须使用对该投稿有权限的同一账户。

上传、保存或创建投稿发生网络超时且结果未知时，停止后续写入并禁用再次发布，先到 Gallery 确认，避免重复创建。明确的 4xx 拒绝允许修正后手动重试；已确认的 Asset 不重新上传。关闭弹窗会结束本次内存任务，未保存的上传按 Asset 草稿期限处理；关闭前后都不要对结果未知的操作盲目重发。

## 验证范围（2026-11-04）

- 当前 Gallery PAT 读取审核列表 HTTP 200；媒体授权回查 HTTP 200，返回一个受限上传 scope。无生产图片投稿或审核写入。
- 一次性模拟运行十项检查通过：悬停到分类/上传/处理/审核完整链路及重复点击拦截、Asset 自有上传路径归一、审核 403 后继续且不重复创建、创建超时阻止重发、明确拒绝后复用已上传 Asset、拒绝不支持格式、安全阻断不审核、服务端重复结果、无令牌不写入、正式 `uncertain` 状态允许进入审核。
- 语法、公开源码无凭据、个人版与通用版主体一致性检查通过。
- 没有在真实油猴网页中完成视觉检查和真实跨站图片上传闭环；模拟环境不证明网站防盗链、油猴权限提示或真实页面渲染已通过。

1.1.0 新增流程的一次性模拟检查十项通过：创建/预览/启用/刷新选中并保留其他字段、新增根分类、新增角色、启用冲突后继续且不重复创建、同名已启用作品复用、标识冲突不覆盖、治理权限不足、创建超时不重复发送、启用后刷新失败继续而不重复启用、无效输入与取消。只读核对当前 PAT 可读取真实分类目录（HTTP 200）；没有在生产创建测试分类或维度值，新增界面的实际视觉未验收。

1.2.0 树形角色流程的一次性模拟检查十二项通过：作品下显示角色、必选父级、所属作品和独立角色创建、新增作品根节点、按已选角色预选作品、不同归属不复用、启用重试锁定归属且不重复创建、创建结果未知停止重发、权限失败可修正、旧目录停止发布，以及管理目录 UUID／投稿短 ID 的创建和复用。没有生产目录写入或真实网页视觉验收。
