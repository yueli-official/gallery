# Gallery 本地验收报告

## 结论

当前主机上的非容器本地验收已经覆盖公开旅程、个人区、管理面、安全、韧性、响应式、无障碍、性能与
Workspace 生命周期。Asset 内置扫描链移除后的真实双图上传和投稿已经通过；不能把结果写成“全部通过”：
本机没有 Docker Engine，独立 Asset 尚未发布新的不可变部署 revision；用户叫停了图标打包修复后的
全路由视觉重跑。

## 已通过证据

| 范围 | 结果 | 证据 |
| --- | --- | --- |
| 关键旅程 | 9/9 | `acceptance-journeys-full-20260730-0146` |
| 管理面 | 8/8 | `acceptance-management-full-20260730-0143` |
| 安全回归 | 3/3 | `acceptance-security-regression-20260730-0146` |
| 故障恢复 | 1/1 | `acceptance-resilience-20260730-0052` |
| 无障碍 | 3/3 | `acceptance-accessibility-regression-20260730-0345` |
| 性能预算 | 1/1 | `acceptance-performance-20260730-0347` |
| Web 单元测试 | 61/61 | 16 个测试文件 |
| Go API | Passed | `go test ./...`、`go vet ./...` |
| Web 编译合同 | Passed | Nuxt typecheck 与隔离 `.nuxt-acceptance` production build |
| 全路由视觉与极端响应式 | 105/105 | `acceptance-visual-final-20260730-0356`，发生在最终图标打包修复之前 |
| Workspace 生命周期 | Passed | `down` 只停止登记的五进程；外部 `ae-online` 的 3000 监听保持运行 |
| 扫描移除后双图投稿 | Passed | 两张各 4 个请求；finalize 51ms/36ms；两个 Submission 创建成功 |
| `/submit` 双图真实 UI | Passed | 本地 Playwright 1.61.1；选择两张 PNG、应用“插画 / 人物”并提交，6.4 秒内双“已完成” |
| `/submit` 浏览器网络 | Passed | Trace 为 8 个业务请求且全部 200；Asset GET 与 scan/security/readyz 请求均为 0 |

图标运行时请求噪声已通过 Nuxt Icon 客户端/服务端 bundle 配置消除。修复后的直接页面访问得到
`ICON_WARNING_COUNT=0`、`WARNING_COUNT=0`，typecheck 和 production build 通过；随后开始的全路由
视觉重跑被用户明确叫停，因此不能把该次运行记录为 Passed。

## Viewer 重设计增量

用户指出旧详情页像功能堆砌，标题尺寸与窄侧栏会让十字以上中文标题失去控制。详情页现改为全宽图片展台和
低位信息坞，标题降为 `1.3-1.65rem`，移动端为 `1.25-1.55rem`，并使用 `overflow-wrap: anywhere`。

| 范围 | 结果 |
| --- | --- |
| Viewer 单元合同 | 4/4 |
| 连续浏览并显式返回目录 | 1/1，连续切换 12 张 |
| Viewer light/dark、mobile/desktop 视觉合同 | 4/4 更新后非更新复验 |
| 1024px 精确边界 | 页面、信息坞、事实区和标题溢出均为 0 |
| Impeccable detector | `[]` |
| 独立 finish review | Passed，无 Material finding；唯一 P2 断点问题已 Resolved |

构图证据和实拍位于 `.impeccable/mocks/` 与
`web/test-results/design/viewer-bottom-dock/`；批准构图是
`.impeccable/mocks/viewer-comp-b-bottom-dock.png`。

## Blocked

- Docker Engine 不可用，不能执行三种 Compose 拓扑的真实 `up`、健康链与数据持久化验收。
- 独立 Asset 尚未发布包含扫描移除的新版本，Gallery 部署锁仍需在发布后更新 revision。
- 图标打包修复后的 105 项全路由视觉重跑被用户叫停；只有本次 Viewer 的四项视觉合同在最终代码上复验。

## 当前运行

Workspace session `20260730T041505Z-17308`：

- Identity `8081`
- Account `3010`
- Asset `8082`
- Gallery API `8091`
- Gallery Web `3007`

外部 `ae-online` 继续监听 `127.0.0.1:3000` 且 HTTP 200。
