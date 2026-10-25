# 分类、维度与视觉展示

分类规则、层级、维度分配和治理继续由 Foundation Classification 提供。Gallery 按 Foundation Admin Platform 的产品扩展合同，拥有图标、封面及其媒体引用；其他产品不会被迫增加 Gallery 字段。

管理入口为“分类与维度”的分类、维度、维度值、标签编辑表单。复用 `@yueli/asset-nuxt/components` 的 `AssetVisualAssetField`，由共享组件提供完整 Tabler 图标选择和图片采集/裁剪，Gallery 仅负责上传、保存及展示方式。站点设置保留导航背景；分类图标和封面不再在两个页面分别编辑。

每个展示对象可保留一个图标和一张封面，`mode` 决定公开展示使用哪一种。清除封面会清空媒体引用并改回图标；仅切换展示方式仍保留封面，便于再次启用。新上传封面输出 WebP、最长边 1024px；后端读取 Asset 可信元数据复核尺寸。已有从图库选取的封面继续保留其作品引用和在用保护。

## 数据和接口

迁移 `0015_classification_visual_assets` 为现有展示表增加 `asset_id`，从旧的图库图片引用一次性填充，支持 `category/facet/facet_value/tag`。不修改已执行的旧迁移。

分类创建和更新 API 接受可选 `visual`：

```json
{"mode":"image","icon":"device-gamepad-2","imageId":"","assetId":"Asset 返回的资源 ID"}
```

不提交 `visual` 保留当前展示；提交图标模式和空 `assetId/imageId` 清除封面。`kind/nodeId` 由目标身份确定，输入中的对应值不改变身份。`delivery` 是响应中的临时授权，写入时清除，不存入数据库。后台目录和标签响应提供 `visual`；前台筛选、投稿和详情继续消费同一份站点展示数据。

封面由 Asset 接受的 Gallery 图片用途上传，不新增 Gallery 投稿或图片记录。服务端在保存前校验机器客户端的 namespace/profile 授权，并读取媒体元数据；已有隐藏/不合格图片不能借封面入口重新公开。封面使用签名交付，不保存短期 URL、对象 Key 或存储直链。新的独立封面要求启用 Gallery 的签名交付配置。

`gallery-classification-cover` 通过统一 referencesync 快照登记；替换、清空和删除身份后，下轮对账撤销不再使用的关系，不直接删除可能被复用的资源。图标模式仍保留的封面继续算作业务引用。

侧栏设置只提交 `appearance.sidebar`，不覆盖分类编辑表单刚保存的展示；API 中省略 `appearance.classification` 表示保持，显式空数组表示清空。

## 详情页导航

移除与推荐网格重复的横向缩略图条，保留上一张/下一张及键盘切换。推荐区域链接优先进入同角色筛选，再选同作品、同分类；没有分类上下文时不显示笼统的“查看全部”。
