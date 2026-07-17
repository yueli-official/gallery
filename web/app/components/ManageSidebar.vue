<script setup lang="ts">
import { ManageSidebarLink } from "@platform/manage/components";

const route = useRoute();
const groups = [
  {
    label: "内容",
    items: [
      { label: "图片", icon: "i-tabler-photo", to: "/manage/images" },
      {
        label: "投稿审核",
        icon: "i-tabler-photo-check",
        to: "/manage/submissions",
      },
      {
        label: "专题策展",
        icon: "i-tabler-folders",
        to: "/manage/collections",
      },
    ],
  },
  {
    label: "目录与分发",
    items: [
      {
        label: "分类与维度",
        icon: "i-tabler-category",
        to: "/manage/classification",
      },
      { label: "发现策略", icon: "i-tabler-sparkles", to: "/manage/discovery" },
    ],
  },
  {
    label: "信任与安全",
    items: [
      { label: "处理单", icon: "i-tabler-shield-check", to: "/manage/cases" },
    ],
  },
  {
    label: "设置",
    items: [
      {
        label: "资源设置",
        icon: "i-tabler-settings",
        to: "/manage/assets",
      },
    ],
  },
];
function isActive(to: string) {
  return to === "/manage" ? route.path === to : route.path.startsWith(to);
}
</script>

<template>
  <div class="flex h-full flex-col bg-elevated/30">
    <NuxtLink
      to="/manage"
      class="flex h-16 items-center gap-3 border-b border-default px-5 text-highlighted"
    >
      <span class="gallery-mark"><span /></span>
      <span class="min-w-0"
        ><span class="block font-display text-sm font-semibold leading-4"
          >月离图库</span
        ><span class="mt-0.5 block text-[.68rem] text-dimmed"
          >运营中心</span
        ></span
      >
    </NuxtLink>

    <nav aria-label="图库管理" class="flex-1 overflow-y-auto px-3 py-4">
      <ManageSidebarLink
        class="mb-5 font-semibold"
        to="/manage"
        label="今日"
        icon="i-tabler-sun-high"
        :active="route.path === '/manage'"
      />

      <section v-for="group in groups" :key="group.label" class="mb-5">
        <h2
          class="mb-1.5 px-3 text-[.68rem] font-semibold tracking-wide text-dimmed"
        >
          {{ group.label }}
        </h2>
        <div class="space-y-0.5">
          <ManageSidebarLink
            v-for="item in group.items"
            :key="item.to"
            :to="item.to"
            :label="item.label"
            :icon="item.icon"
            :active="isActive(item.to)"
          />
        </div>
      </section>
    </nav>
  </div>
</template>
