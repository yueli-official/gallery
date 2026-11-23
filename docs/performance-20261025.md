# Gallery 图片加载与后台令牌性能修复

2026-10-25，正式入口：https://gallery.yuelili.com。

## 已定位的问题

- Asset 引用同步 SDK 每个快照都申请机器令牌，后台每15秒持续调用。Identity 采样96.4% CPU在 bcrypt 服务密钥校验；20秒观察到22次令牌申请，来源是本站服务。当前虚拟机系统实际识别2个CPU，容器没有额外CPU配额。
- 图片发布只准备thumbnail，首次访问列表才排队生成瀑布流预览。第22页首次24图18.706秒，第二次1.672秒。
- WebP 编码参数将透明通道保存为原始字节。第28页24张合计6,081,282 bytes；一张984,884 bytes中921,601 bytes属于未压缩ALPH。该页后续测得5.829、6.456秒。

## 已上线

- Asset 发布前准备已声明的公开预览规格，成功后才返回发布完成。
- 新交付版本`v=20261025`使用透明通道无损压缩，独立派生对象身份；旧版URL及其字节保留，沿用线上一年immutable缓存。分辨率、RGB质量参数、4K主图及存储策略不变。
- 真实三例984884→65520、709522→58072、707626→77924 bytes。临时960²探针923234→1706 bytes，逐像素验证透明度不变；临时测试从仓库移除，诊断源码保存在任务制品目录。
- Gallery 和 Blog 后台已采用 expires_in 令牌复用，提前续期并串行化并发刷新。SDK真实快照调用探针20并发申请从20降到1，过期、取消和失败路径通过。正式Go SDK v0.4.0尚不含该修复；本次使用摘要固定的源码依赖，没有修改模块缓存或宣称发布新版本。

| 组件 | 当前镜像 | 镜像ID |
| --- | --- | --- |
| Gallery API | `yueli/gallery-api:media-performance-20261025-01` | `sha256:820c86747611f2136836cf70bd9762cf20969483d50070063a3798a301d6a485` |
| Gallery Web | `yueli/gallery-web:media-performance-20261025-02` | `sha256:25e76de24678d139b6faa537964208d30957b4d5669e0a7ddd4ab9853a72b5ad` |
| Asset | `yueli/asset-api:alpha-compressed-20261025-02` | `sha256:14b99aabda47eeefffa90256676c73d8c34d4e24a6fc0cec9a724f3c74ab3226` |

来源使用各服务实际生产基准及本次补丁，排除当前checkout的其他未发布功能。Gallery采用当前批准前端基线`32339ab7339bbcccbcecbb920cd556e5d689b309b906e2ee6979ba83b3de78ad`，没有纳入未跟踪首页原型。Asset保留已获准的API-only部署形式、原配置和输入限制。

## 验证

- 冻结安装、73项前端测试、类型检查、构建通过；完整生产Asset Go测试、受影响checkout测试与并发race探针通过。
- 候选与正式桌面/手机、深浅色四组合通过CLI Playwright功能检查，覆盖列表、详情下载入口、专题集合、排行榜和投稿页；无图片失败、脚本异常或水合警告。
- 人工查看实际CUA候选桌面浅色/深色及390×844手机截图，以及正式第28页截图；未使用CLI截图进行视觉判断。正式API/Web运行镜像匹配，82份网页资源摘要匹配；环境不变，数据库容器不变，schema 14 clean。
- 正式第28页，Chrome新上下文禁用浏览器缓存，24图全部解码完成用时**1373、2163ms**，无失败请求。该数字来自当前网络测量，不代表任何网络的统一耗时。
- 实际发布后8种预览规格服务端读取4–5ms。旧版三个示例哈希在新编码生成后仍一致。

## 尚未结束

Identity、Docs、WWW、Distribution、Forum 的令牌复用候选已运行通过，但现有网页仍为旧共享包基线，通用门禁拒绝仅后台发布。已请求用户确认本次保留网页的发布例外，尚未收到答复，不能将Identity CPU整改标为完成。没有降低bcrypt成本、取消鉴权或停掉引用同步。

672个masonry-sm和672个grid-sm已完成。剩余规格继续单并发准备；网页切换曾使BFF连接中断一次，旧进度保留，已通过相同Asset公开媒体接口续接。完整批次完成以服务器回执为准，不能把启动作业当作完成。

## 来源和恢复

- Gallery正式清单：服务器`/projects/yuelili.com/gallery/deployment-release.json`，发布目录`gallery/releases/media-performance-20261025-02`。
- Asset来源及回执：`/projects/yuelili.com/.deploy/asset-alpha-compression-20261025-02`；Gallery预览准备：`.deploy/gallery-image-performance-20261025-02`。
- 数据库和配置备份：`/projects/yuelili.com/backups/gallery-media-performance-20261025-02`、`asset-alpha-compression-20261025-02`、`media-performance-20261025-01`。
- 本地冻结来源及诊断制品：`D:/coding/yueli-official/task-artifacts/gallery/media-performance-20261025-02`、`asset/alpha-compression-20261025-02`、`asset/machine-token-reuse-20261025-01`。
- 来源与正式输出已归档，代码保持本地未提交；没有commit、push、tag或Release授权。
