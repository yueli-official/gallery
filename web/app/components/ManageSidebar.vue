<script setup lang="ts">
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
      <NuxtLink
        to="/manage"
        class="relative mb-5 flex min-h-10 items-center gap-3 rounded-lg px-3 text-sm font-semibold transition-colors"
        :class="
          isActive('/manage')
            ? 'bg-default text-highlighted shadow-sm ring-1 ring-default'
            : 'text-muted hover:bg-default/70 hover:text-default'
        "
      >
        <span
          v-if="isActive('/manage')"
          class="absolute inset-y-2 left-0 w-0.5 rounded-full bg-primary"
        /><UIcon name="i-tabler-sun-high" class="size-[1.125rem]" />今日
      </NuxtLink>

      <section v-for="group in groups" :key="group.label" class="mb-5">
        <h2
          class="mb-1.5 px-3 text-[.68rem] font-semibold tracking-wide text-dimmed"
        >
          {{ group.label }}
        </h2>
        <div class="space-y-0.5">
          <NuxtLink
            v-for="item in group.items"
            :key="item.to"
            :to="item.to"
            class="relative flex min-h-10 items-center gap-3 rounded-lg px-3 text-sm font-medium transition-colors"
            :class="
              isActive(item.to)
                ? 'bg-default text-highlighted shadow-sm ring-1 ring-default'
                : 'text-muted hover:bg-default/70 hover:text-default'
            "
          >
            <span
              v-if="isActive(item.to)"
              class="absolute inset-y-2 left-0 w-0.5 rounded-full bg-primary"
            /><UIcon :name="item.icon" class="size-[1.125rem]" />{{
              item.label
            }}
          </NuxtLink>
        </div>
      </section>
    </nav>

    <div class="border-t border-default p-3">
      <NuxtLink
        to="/manage/assets"
        class="flex min-h-10 items-center gap-3 rounded-lg px-3 text-sm font-medium transition-colors"
        :class="
          isActive('/manage/assets')
            ? 'bg-default text-highlighted shadow-sm ring-1 ring-default'
            : 'text-muted hover:bg-default/70 hover:text-default'
        "
        ><UIcon
          name="i-tabler-settings"
          class="size-[1.125rem]"
        />资源设置</NuxtLink
      >
      <NuxtLink
        to="/"
        class="mt-0.5 flex min-h-10 items-center gap-3 rounded-lg px-3 text-sm text-dimmed transition-colors hover:bg-default/70 hover:text-default"
        ><UIcon
          name="i-tabler-arrow-up-right"
          class="size-[1.125rem]"
        />查看公开站点</NuxtLink
      >
    </div>
  </div>
</template>
