# Gallery 作品与角色树发布

2026-11-04，正式入口：[月离图库](https://gallery.yuelili.com)。用户明确授权 Gallery 提交、部署、迁移和本次目录新增；代码提交 `7922e1b`，当前分支 `main`。

## 已上线行为

作品／系列与角色合并为同一 `work` 维度，角色按主要所属作品建立单父级关系；游戏、动漫、原创仍是独立分类。普拉娜保留身份并归入碧蓝档案，“独立角色”和“待归属角色”作为作品树分组。投稿、筛选、后台目录、图片编辑与详情同步采用该树；图片编辑保留多个角色。

通过真实分类治理接口新增并启用以下作品：原神、明日方舟：终末地、崩坏：星穹铁道、绝区零、鸣潮、明日方舟、崩坏3、战双帕弥什、尘白禁区。碧蓝档案下增加阿罗娜、小春、睦月，与原有普拉娜共用同一父级。目录最终 revision 为 47。空作品已可用于投稿；公开浏览的候选继续依照现有规则仅展示有图片的项。

个人油猴脚本 1.2.0 已具备配套服务器条件，可以覆盖安装旧版。私有安装文件留在用户 `.secret/yueli/`，仓库只提交不含令牌的通用版。

## 制品与来源

| 服务 | 镜像 | 实际 image ID |
| --- | --- | --- |
| Gallery API | `yueli/gallery-api:work-characters-20261104-1` | `sha256:8346e4a9c057c892548967e0c57b25a63ed3c41265cc666f9c046e4e9e945779` |
| Gallery Web | `yueli/gallery-web:work-characters-20261104-1` | `sha256:47fa870ab27ac7e9c64fa74b48b8a88bb49c7376afc64616fc56c9425e809843` |

构建从可核验的当前生产归档恢复，365 个 Web 文件与原构建逐文件清单一致，仅叠加明确列出的 3 个 API 文件和 9 个 Web 文件；没有带入其他本地未提交改动。Go 使用线上冻结依赖上下文，不使用邻仓未发布源码。API 重编译后二进制 SHA256 仍为 `2bdc105fe33b0c040d4ed63c9eba2c1bf49e91922eb28522d7318c3fcd39eddf`，与原生产程序一致；新镜像携带更新后的迁移、运行清单和迁移工具。

前端批准基线为 `b896e8efe51b901bf8877e580cfa0af10de62e8ebe0aac6d2573584ad98f63e5`。源快照身份为 `80fdd91032a1fc445e1c750a8c8ebe54443d8baf6c3bd3a580ed6e7e63c4dda8`；完整源包 SHA256 为 `6b4fa70e366e6f8ee0972a844340647bef140c9b85282668437a63076b9b4bc2`。双端运行包、源码包、实际 Go Workspace 及依赖回执由发布清单记录。

## 验证与切换

冻结安装、正式依赖门禁、Nuxt 类型检查、73 项前端测试、Web 构建及 Linux API 构建通过。候选使用生产库副本执行真实迁移与治理操作；修正副本恢复时须保留原 owner/ACL 的问题后，版本从 clean 22 升为 clean 23。正式切换前暂停 Gallery API/Web，重新生成并验证可读数据库备份，再迁移和启动精确的两个服务；未重跑初始化或修改共享 Provider。

正式库保留 672 张图片、672 条投稿、1344 条图片 Facet 关联和 1344 条投稿 Facet 关联。API/Web 均 healthy、restart 0。候选及公网检查通过：作品和角色筛选、父级归属、只读图片关联、投稿选项、SSR/BFF、登录门禁、135 个 JS/CSS 摘要和 WebP 图片读取。旧角色筛选依据原发现协议返回 `facet.unknown` 和空结果，不新增兼容入口。

浏览器 IAB 绑定候选页面超时，实际画面与油猴视觉验收仍未完成；这一限制先前已向用户说明，随后用户明确要求部署。HTTP 和静态资源验证不算视觉验收。

实测冻结安装 33.9 秒，复用 768 包、下载 0；Nuxt 客户端构建 9442ms、服务端构建 10395ms；API/Web 镜像组装 3.87/2.08 秒；暂停写入至新服务健康的切换阶段为 17.02 秒。API 交叉构建总时长未单独采集，不作估算。

## 回执与恢复

- 当前真值：服务器 `/projects/yuelili.com/gallery/deployment-release.json`。
- 正式发布归档：`/projects/yuelili.com/gallery/releases/work-characters-20261104-1`。
- 执行、运行包与备份：`/projects/yuelili.com/.deploy/work-characters-20261104-1`。
- 数据库备份：上述目录 `backup/production-before.dump`；备份 SHA256、大小及可读验证在 `backup/production-backup.json`。
- 原配套 Compose 与回执：`backup/compose.yaml`、`backup/deployment-release.json`。需要恢复应用时使用已验证旧配套镜像与配置，不自动回退数据库。

候选容器、独立副本数据库、临时连接凭据和展开构建目录已移除；正式运行包、源码归档和恢复备份保留。Compose 语义比较确认只改变 Gallery API/Web 的镜像引用；原 PAT 配置更新记录保留，Identity、Asset 及其他站点均未重建。
