# Gallery

Gallery 是运营方维护为主、允许用户投稿的公开单图展示与收藏站。它不建立创作者身份、粉丝关系或内容社交。

## Language

**Image（图片）**:
公开目录中的单张图片记录；一条 Image 永远只引用一个逻辑 Asset，并拥有标题、说明、可选来源地址和分类信息。
_Avoid_: Artwork、作品、帖子、多图内容

**Submission（投稿）**:
User、Guest 或运营者提交一张图片的处理记录；它记录所有权、处理、审核和结果，但不一定创建新的 Image。
_Avoid_: Image、作品草稿、创作者发布

**Subject（主体）**:
Submission 的所有者，可以是 Identity User 或 Guest。Subject 只参与权限、认领和审计，不进入公开 Image 投影。
_Avoid_: Creator、作者、公开上传者

**Processing State（处理状态）**:
媒体校验和预处理进度：queued、processing、ready、failed。
_Avoid_: 审核状态、发布状态

**Review State（审核状态）**:
人工审核要求与结论：not_required、pending、approved、rejected。
_Avoid_: 安全扫描、是否公开

**Publication State（发布状态）**:
运营可见性：draft、published、hidden、deleted。
_Avoid_: Processing State、Review State

**Publication Eligibility（公开资格）**:
由处理就绪、审核通过或免审、已发布、安全未阻断和公开 rendition 就绪共同计算的策略结果。
_Avoid_: 单一 status 字段、审核通过

**Collection（集合）**:
同一种资源的组织容器。Gallery 使用 editorial 组织公开专题，使用 favorites 表达每个 User 的私人单例“我的收藏”。
_Avoid_: Category、Tag、Series、多图 Image

**Collection Kind（集合种类）**:
消费者以命名空间注册的集合行为，例如 gallery.editorial、gallery.favorites。
_Avoid_: 全局可任意写字符串、资源类型

**Membership（成员关系）**:
Collection 与 Image 之间的幂等关系；删除关系或集合永远不删除 Image。
_Avoid_: Image 所有权、分类指派

**Facet（维度）**:
运营方治理的结构化筛选轴，例如 topic、orientation、color、style、medium。
_Avoid_: Tag、Collection

**Category（分类）**:
公开界面中的主主题分类；后端映射到保留的 single-select `topic` Facet。
_Avoid_: 所有 Facet、Tag

**Tag（标签）**:
用于描述图片细节的自由关键词，可通过别名归一化。
_Avoid_: Category、Facet Value

**Case（处理单）**:
运营者处理举报、来源修正、安全不确定、近重复或下架调查的工作单元；Case 数量不会自动改变 Image 发布状态。
_Avoid_: 自动下架规则、Submission

**Rendition（衍生版本）**:
Asset 从私有 master 预处理出的命名公开图片版本；Gallery 只保存 assetId，不保存 rendition URL。
_Avoid_: 原图、任意实时 transform

**Tombstone（墓碑）**:
永久删除的公开 Image 标识保留记录，用同一 ID 返回 410，防止标识被误复用。
_Avoid_: hidden、软删除 Image

## Invariants

- 一个 Image 对应一个逻辑 Asset；需要多图时使用 Collection。
- 公开 Image 永远不展示投稿 Subject。
- 已发布 Image 不允许替换像素；撤稿后只能重新投稿。
- 用户收藏只有 private singleton“我的收藏”，没有多个相册或社交分享。
- 首页是 seeded random discovery；主目录是分页固定网格。
- source URL 可空、仅作元数据，服务器不抓取外部来源。
