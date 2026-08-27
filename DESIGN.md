---
name: 月离图库
description: 图片优先、可持续细化的公共图片索引
colors:
  mineral-blue: "#2458ff"
  mineral-blue-soft: "#edf1ff"
  paper: "#fdfdfb"
  quiet-surface: "#f4f4f0"
  strong-surface: "#ffffff"
  ink: "#171817"
  hairline: "rgb(23 24 23 / 0.12)"
typography:
  display:
    fontFamily: "DM Sans Variable, PingFang SC, Microsoft YaHei UI, ui-sans-serif, system-ui, sans-serif"
    fontSize: "clamp(1.7rem, 2.4vw, 2.45rem)"
    fontWeight: 680
    lineHeight: 0.98
    letterSpacing: "-0.055em"
  body:
    fontFamily: "DM Sans Variable, ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.65
  label:
    fontFamily: "DM Sans Variable, ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.72rem"
    fontWeight: 700
    lineHeight: 1.2
    letterSpacing: "0.05em"
rounded:
  control: "0.55rem"
  search: "0.75rem"
  media: "0.9rem"
spacing:
  compact: "0.5rem"
  regular: "1rem"
  section: "clamp(2rem, 4vw, 3.75rem)"
components:
  search:
    backgroundColor: "{colors.quiet-surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.search}"
    height: "2.65rem"
  media-card:
    backgroundColor: "{colors.quiet-surface}"
    rounded: "{rounded.media}"
  primary-button:
    backgroundColor: "{colors.mineral-blue}"
    textColor: "{colors.strong-surface}"
    rounded: "{rounded.control}"
---

# Design System: 月离图库

## Overview

**Creative North Star: “公共图片索引”**

月离图库首先是一套可用的图片索引，而不是创作者社区或营销首页。界面以近白画布、紧密图片阵列和少量矿物蓝索引信号组织内容；标题与工具服从图片，不在首屏制造第二个视觉中心。

它采用成熟图库的可预期结构，同时用折面标记、切角 refinement 和标题嵌入瀑布流首块形成自己的识别度。Pinterest 式连续发现负责浏览节奏，Pexels 式分类、维度与标签负责收窄范围。

**Key Characteristics:**

- 图片先于说明，首屏立即出现可用内容。
- 矿物蓝只标记状态、索引与关键动作。
- 图片无外框、轻圆角，文本与工具保持紧凑。
- 桌面多列、移动双列，信息层级不因断点而降级。

## Colors

配色由单一高识别蓝与暖白中性色组成；深色模式保持同样角色关系，不另造第二套品牌色。

### Primary

- **矿物蓝：** 用于活动导航、索引线、焦点与主要动作。
- **浅蓝索引面：** 用于 hover、筛选反馈和低强度选中状态。

### Neutral

- **纸张白：** 页面连续画布。
- **安静表面：** 搜索、标题块和筛选分组。
- **墨色：** 标题和主要内容。
- **发丝线：** 只承担结构分隔，不把每张图片框成卡片。

**The One Blue Rule.** 同一视口只允许矿物蓝成为强调色，不能再加入竞争性的装饰色。

## Typography

**Display Font:** DM Sans Variable，并以苹方、微软雅黑 UI 和系统无衬线承接中文。  
**Body Font:** DM Sans Variable 与系统无衬线。

**Character:** 标题粗而紧，正文安静、可扫描。英文只用于必要的机器标识，不作为装饰性眉题。

### Hierarchy

- **Display：** 只用于页面主标题和首页瀑布流首块；紧行高、负字距。
- **Title：** 用于区段与卡片标题，保持一到两行。
- **Body：** 用于说明和筛选解释，建议控制在 42rem 以内。
- **Label：** 用于索引与数量；小字号、高字重、轻字距。

**The No Eyebrow Rule.** 不在每个区段标题上重复“运营精选”“持续更新”之类的小眉题。

## Layout

公共页面使用最大 100rem 的居中画布。首页在搜索下方放置单行 refinement，随后直接进入瀑布流；标题本身是流中的第一块，而不是独立 Hero。瀑布流从移动双列逐步扩展到桌面五列，列间距保持在约 0.75–1rem。

目录页由桌面固定筛选栏和结果区组成；移动端使用页面内搜索与底部筛选抽屉，并隐藏顶部重复搜索。区段间距可以放大，图片内部与工具条密度保持紧凑。

## Elevation & Depth

系统默认扁平。层级主要由中性色面、发丝线和图片自身形成；阴影只用于 Viewer 图片、必要浮层或交互反馈，不能成为普通图片卡片的常驻装饰。

**The Flat-by-Default Rule.** 图片流和专题卡片静止时不使用悬浮卡片阴影。

## Shapes

媒体统一使用约 14px 的轻圆角；搜索使用稍小圆角；小控件可以更紧。折面品牌标记和 refinement 的右上切角是签名几何，不能退化为全站同一种胶囊。

## Components

### Buttons

- **Primary:** 矿物蓝实底、白字，用于投稿和确认动作。
- **Secondary / Ghost:** 使用中性色面或透明背景；hover 才增加表面色。
- **Focus:** 必须保留可见的矿物蓝焦点环。

### Chips

- **Style:** refinement 使用浅表面与右上切角；数量使用低对比文本。
- **State:** 活动筛选保留可移除动作，不依赖颜色作为唯一线索。

### Cards / Containers

- **Media:** 无边框、无常驻阴影，图片充满容器。
- **Lead tile:** 只承载紧凑主标题和极短身份说明。
- **Editorial list:** 使用分隔线而非一组独立浮卡。

### Inputs / Fields

- **Search:** 安静表面、内联 SVG 搜索与箭头，聚焦时显示蓝色边界和焦点环。
- **Filter fields:** 可以成组放入浅表面，但不重复首页级搜索。

### Navigation

桌面顶栏由折面标记、主导航、唯一宽搜索和账户动作构成。移动端保留等宽四项导航；浏览目录使用自己的搜索时，顶栏搜索隐藏。

### Image Stream

瀑布流是签名组件。标题覆盖层在指针设备 hover/focus 时出现，在触摸设备常驻；书签操作使用可辨识的内联 SVG，并位于图片右上角。

## Do's and Don'ts

### Do:

- **Do** 让首屏在 refinement 之后立即出现图片。
- **Do** 用分类、Facet 和 `#标签` 共同增强搜索。
- **Do** 在移动端保持双列连续发现，并为触摸设备显示必要操作。
- **Do** 把 fixture 来源保留在测试证据和开发日志中，不泄漏到公开浏览界面。

### Don't:

- **Don't** 在首页正文再放第二个大型搜索框。
- **Don't** 用长介绍、英文眉题或多行动作重新制造 Hero。
- **Don't** 给每张图片增加边框、阴影和说明卡壳。
- **Don't** 把渐变演示素材记录为正式摄影材质规范。
