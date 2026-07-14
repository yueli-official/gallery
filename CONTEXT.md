# Gallery

Gallery 是面向多创作者、多内容形态的通用视觉作品社区。旧个人图片站只提供迁移数据，不定义新产品的主题、品牌或分类边界。

## Language

**Artwork（作品）**:
创作者发布的内容单元，可包含一张或多张有序图片，并拥有统一的标题、说明、权利声明和可见性。
_Avoid_: 图片、帖子、壁纸

**Asset（图片资产）**:
作品内部的一张源图片及其衍生规格；它不是可独立互动或分类的作品。
_Avoid_: 作品、附件

**Creator（创作者）**:
在 Gallery 内拥有公开创作身份并发布作品的平台用户。
_Avoid_: 管理员、上传者

**Creator Application（创作者申请）**:
平台用户申请 Gallery 创作资格的站点内准入记录；它不改变统一账号的全局角色。
_Avoid_: 注册、全局角色申请

**Draft（作品草稿）**:
仅创作者可见且允许信息不完整的 Artwork 编辑状态；进入审核后内容会锁定。
_Avoid_: 未发布图片、临时 Asset

**Review Submission（审核提交）**:
创作者将满足发布要求的 Draft 送入待审状态的动作；提交后必须由运营者作出决定。
_Avoid_: 发布、保存

**Moderation Decision（审核决定）**:
运营者对待审 Artwork 作出的通过或退回结论；通过后才成为公开作品，退回时保留可见的修改说明。
_Avoid_: 编辑精选、自动发布

**Facet（维度）**:
运营方治理的离散浏览轴，例如媒介、题材、风格和用途；每个维度包含一组 Facet Value。
_Avoid_: 标签组、栏目

**Facet Value（维度值）**:
某个 Facet 下可分层的受控候选值；同一作品可跨多个维度关联多个值。
_Avoid_: Category、Tag、唯一主分类

**Artwork Facet Assignment（作品归类）**:
Artwork 与一个 Facet Value 之间的受控关系。
_Avoid_: 标签、作品属性

**Tag（标签）**:
创作者用于描述角色、对象、技法或细节的自由词，并通过别名归一化。
_Avoid_: Category、Facet

**Series（系列）**:
创作者组织多件 Artwork 的有序集合，与单件作品内部的多张 Asset 不同。
_Avoid_: 相册、多图作品

**Editorial Feature（编辑精选）**:
由运营者给出理由和有效期的人工推荐位，不代表算法热度或用户排行。
_Avoid_: 热门、趋势、排行榜
