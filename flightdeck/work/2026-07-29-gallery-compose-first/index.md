# Gallery Compose-first 试点

## Goal

在不迁移其他消费者的前提下，把 Gallery 作为首个完整试点：本地开发由 Workspace 一键管理，生产部署由
Gallery 仓库自持标准 Docker Compose，并明确支持完整独立、接入已有 Identity + Asset、共享 Identity +
专属 Asset 三种拓扑。

## Status

Finished

## Current

Gallery 已自持完整独立、接入已有 Identity + Asset、共享 Identity + 专属 Asset 三种 Compose 入口，并通过
Workspace 管理 Identity、Account、Asset、Gallery API/Web 五进程。本地交付级验收已覆盖关键旅程、管理面、
安全、故障恢复、响应式、无障碍、性能与生命周期；完整结果和不能声称通过的边界见
[本地验收报告](references/acceptance-report-2026-07-30.md)。

用户叫停图标打包修复后的 105 项全路由视觉重跑，并连续指出公共图片详情仍像功能堆砌、标题过大；随后用
Pixiv 实页说明浏览量、收藏数和尺寸都不应抢占画面。Viewer 已完成两轮图片优先收敛：固定右侧栏和四列事实区
被删除，标题降为 `1.05-1.25rem` 并安全断行；收藏与分享只保留具备无障碍名称的图标，浏览/收藏缩为弱化
图标数字，来源和标签保持低对比；图片记录自身的简介直接放在标题下方且不加重复标签，尺寸、举报和补充来源
进入“更多图片信息”。媒体角标只显示缩放百分比；“更多”支持外部点击、Esc、切图和打开治理对话框时关闭。

真实桌面/手机画面、“更多信息”展开态和 0 横向溢出已检查；四张 Viewer 亮暗/桌面/手机基线已强制刷新，
非更新复验 4/4。Viewer 单测 8/8、真实外部点击回归、typecheck、diff check 与 Impeccable detector 通过。
先前连续切图旅程
1/1、1024px 定点复核和独立设计审稿结论仍有效；没有恢复 105 项全路由视觉重跑。

用户随后实页发现收藏提示 `gallery.forbidden`、LAN HTTP 分享失败，并要求 Toast 去掉倒计时、改善外观。
真栈确认正常 Member 收藏为 200；当前会话若在 Gallery API 退化为 Guest 并返回精确 401/403，会进入登录恢复
且保留当前详情 URL，不再把内部错误码作为 Toast。分享在无 Web Share、非安全上下文且无 Clipboard API 的
局域网环境使用受控旧式复制回退，真实页面从“分享没有完成”转为“链接已复制”。全局 Toast 关闭 progress，
收窄为无边框、轻阴影、小状态图标样式，最多同时显示 3 条并保留手动关闭。反馈与 Viewer 定向 Vitest
11/11、真栈收藏/分享/重登录回归、typecheck、diff check 和 Impeccable detector 通过。

用户继续指出成功 Toast 的图标、文字、关闭按钮没有对齐，并在局域网投稿页触发
`crypto.randomUUID is not a function`。真实 Playwright 在 `secureContext=false`、
`randomUUID=undefined` 时稳定复现选图崩溃；只注入 UUID 后同一路径转绿，确认不是文件输入或图片预检问题。
投稿队列现统一使用 `galleryClientId`：优先原生 UUID，非安全上下文使用 `getRandomValues` 生成 v4 UUID，
极端旧环境才退到时间与随机数组合。Toast 对齐放在全局 `.gallery-toast` 层，绕过 Nuxt UI 方向变体对
`items-center` 的合并覆盖；真实渲染从图标/关闭按钮各偏 4px 收敛为 0px。投稿 LAN HTTP 选图真栈通过，
相关 Vitest 18/18、Nuxt typecheck 和唯一一次最终 Impeccable detector 通过。

用户真实上传曾暴露旧 Asset 扫描链和 Gallery 固定轮询；最终架构决策是删除这条扫描链，而不是继续部署
ClamAV。独立 Asset 已改为同步基础文件校验后直接 finalize，Platform 和 Workspace 已同步。首次标准重启
后发现独立 Gallery 真源仍遗漏旧 `readyz/securityState/scanStatus` 消费逻辑和自持 Compose 的
`asset-clamav`；已用红色源码合同复现并同步删除。Gallery 的内容 safety/review 仍属于业务审核，不受影响。

标准 Workspace prepare/migration/seed 成功，会话 `20260730T041505Z-17308` 五进程 ready，Asset
`/readyz` 为 200。真实 Guest BFF 双图回归严格按每张 `upload-init → PUT → finalize → submission` 执行，
共 8 个请求，没有 Asset 状态 GET 或轮询；两个 finalize 均为 200（51ms/36ms），返回字段无扫描状态；
两个 Submission 均成功创建。首次缺少必填场景维度的负向请求被 400 正确拒绝，没有伪装成成功。

独立 Gallery 新增“不得恢复旧扫描消费”的 Web 源码合同和 Compose 合同；Web 16 文件/61 测试、定向
10/10、Nuxt typecheck、Go deploy tests 与 diff check 通过。随后改用仓库本地 Playwright 1.61.1 与
bundled Chromium 对局域网 `/submit` 做真实 UI 验收：两张本次生成的 PNG 经页面选取，批量应用“插画 /
人物”并提交，6.4 秒内都进入“已完成”，最终页面出现“查看投稿记录”。Trace 中恰好有两组
`upload-init → PUT → finalize → submission`，8 个业务请求全部 200，Asset GET、scan/security/readyz
请求均为 0；最终整页截图和 trace 保存在忽略提交的 `web/test-results/live-submit/artifacts/`。
其他消费者迁移继续冻结；本机仍没有 Docker Engine。

## Next

None。

## Current execution

Completed。本地实现、Compose 合同和真栈验收已收口；跨消费者逐站验收由 Workspace 新 Work 统一承载。

## Constraints

- 不恢复或兼容 `doctor.yaml`。
- Gallery 业务代码只声明自身需求；具体服务实例由 Workspace 或 Compose 绑定。
- 当前主机没有 Docker Engine；不能把静态 Compose 合同表述为真实容器验收。
- `deployment.lock.json` 已指向包含扫描移除的 Asset `v0.3.0 / 7007ec950321...`。

## References

- [执行计划](plan.md)
- [边界与已知状态](context.md)
