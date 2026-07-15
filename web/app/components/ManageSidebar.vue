<script setup lang="ts">
const route = useRoute();
const groups = [
  { items: [
    { label: "控制台", icon: "i-tabler-layout-dashboard", to: "/manage" },
    { label: "图片", icon: "i-tabler-photo", to: "/manage/images" },
    { label: "投稿与审核", icon: "i-tabler-photo-check", to: "/manage/submissions" },
  ] },
  { label: "组织与发现", items: [
    { label: "专题集合", icon: "i-tabler-folders", to: "/manage/collections" },
    { label: "分类与维度", icon: "i-tabler-category", to: "/manage/classification" },
    { label: "处理单", icon: "i-tabler-flag", to: "/manage/cases" },
    { label: "随机与排行", icon: "i-tabler-arrows-random", to: "/manage/discovery" },
  ] },
  { label: "系统", items: [
    { label: "资源与设置", icon: "i-tabler-settings", to: "/manage/assets" },
  ] },
];
function isActive(to: string) { return to === "/manage" ? route.path === to : route.path.startsWith(to); }
</script>

<template>
  <div class="flex h-full flex-col bg-elevated/25">
    <NuxtLink to="/manage" class="flex h-16 items-center gap-2 border-b border-default px-5 font-display font-semibold text-highlighted">
      <span class="gallery-mark"><span /></span>图库运营
    </NuxtLink>
    <nav aria-label="图库管理" class="flex-1 space-y-5 overflow-y-auto p-3">
      <div v-for="(group, index) in groups" :key="index">
        <p v-if="group.label" class="mb-1 px-3 text-[.68rem] font-semibold uppercase tracking-[.14em] text-dimmed">{{ group.label }}</p>
        <div class="space-y-1">
          <NuxtLink v-for="item in group.items" :key="item.to" :to="item.to" class="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition" :class="isActive(item.to) ? 'bg-primary/10 text-primary' : 'text-muted hover:bg-elevated hover:text-default'">
            <UIcon :name="item.icon" class="size-5" />{{ item.label }}
          </NuxtLink>
        </div>
      </div>
    </nav>
  </div>
</template>
