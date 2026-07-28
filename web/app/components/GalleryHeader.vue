<script setup lang="ts">
import type { AccountMenuAction } from "@yueli/ui/account-menu/pattern";

defineProps<{ brandName?: string }>();

const route = useRoute();
const { loggedIn } = useAuth();
const { canManage, refreshMe } = useGalleryMe();
const nav = [
  { label: "发现", to: "/" },
  { label: "浏览", to: "/images" },
  { label: "专题", to: "/collections" },
  { label: "排行", to: "/rankings" },
];
const contextActions = computed<AccountMenuAction[]>(() => [
  { label: "我的收藏", icon: "i-tabler-heart", to: "/favorites" },
  { label: "我的投稿", icon: "i-tabler-photo-up", to: "/submissions" },
  ...(canManage.value
    ? [
        {
          label: "图库管理",
          icon: "i-tabler-layout-dashboard",
          to: "/manage",
        },
      ]
    : []),
]);

watch(
  loggedIn,
  async (value) => {
    if (value) await refreshMe();
  },
  { immediate: true },
);
</script>

<template>
  <header class="gallery-header">
    <div class="gallery-header-inner">
      <NuxtLink to="/" class="gallery-wordmark" aria-label="月离图库首页">
        <span class="gallery-mark" aria-hidden="true"><span /></span>
        <span>{{ brandName || "月离图库" }}</span>
      </NuxtLink>

      <nav class="gallery-desktop-nav hidden items-center gap-1 md:flex" aria-label="主要导航">
        <UButton
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          :label="item.label"
          color="neutral"
          :variant="
            route.path === item.to ||
            (item.to !== '/' && route.path.startsWith(item.to))
              ? 'soft'
              : 'ghost'
          "
          size="sm"
        />
      </nav>

      <GalleryGlobalSearch class="gallery-header-search" compact />

      <div class="gallery-header-actions flex items-center gap-1 sm:gap-1.5">
        <NuxtLink
          v-if="route.path !== '/images'"
          to="/images"
          class="gallery-header-search-trigger"
          aria-label="搜索图库"
        >
          <svg viewBox="0 0 24 24" width="19" height="19" aria-hidden="true">
            <circle
              cx="11"
              cy="11"
              r="6.5"
              fill="none"
              stroke="currentColor"
              stroke-width="1.8"
            />
            <path
              d="m16 16 4 4"
              fill="none"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
            />
          </svg>
        </NuxtLink>
        <UColorModeButton
          color="neutral"
          variant="ghost"
          aria-label="切换颜色模式"
        />
        <NuxtLink to="/submit" class="gallery-submit-link">投稿</NuxtLink>
        <ConsumerAccountControl :context-actions="contextActions" />
      </div>
    </div>
    <nav class="gallery-mobile-nav md:hidden" aria-label="移动端主要导航">
      <NuxtLink
        v-for="item in nav"
        :key="item.to"
        :to="item.to"
        :aria-current="
          route.path === item.to ||
          (item.to !== '/' && route.path.startsWith(item.to))
            ? 'page'
            : undefined
        "
        >{{ item.label }}</NuxtLink
      >
    </nav>
  </header>
</template>
