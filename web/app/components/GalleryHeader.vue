<script setup lang="ts">
import type { AccountMenuAction } from "@yueli/ui/account-menu/pattern";

const props = defineProps<{
  brandName?: string;
  searchPlaceholder?: string;
}>();

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
  <header
    class="gallery-header sticky top-0 z-40 border-b border-default bg-[color-mix(in_srgb,var(--gallery-canvas)_96%,transparent)] backdrop-blur-md"
  >
    <GalleryPublicContainer
      class="gallery-header-inner flex h-15 items-center gap-4 px-4 sm:px-6 md:h-17 lg:px-8"
    >
      <NuxtLink
        to="/"
        class="gallery-wordmark font-display inline-flex shrink-0 items-center gap-2.5 rounded-[0.65rem] text-[1.05rem] font-bold tracking-[-0.035em] text-highlighted focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary"
        :aria-label="`${props.brandName || '图库'}首页`"
      >
        <span
          class="gallery-mark relative block size-[1.9rem] shrink-0 bg-transparent"
          aria-hidden="true"
          ><span
        /></span>
        <span class="max-[24rem]:hidden">{{ props.brandName || "图库" }}</span>
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

      <GalleryGlobalSearch
        class="gallery-header-search mx-auto hidden flex-1 lg:flex"
        compact
        :placeholder="props.searchPlaceholder"
      />

      <div
        class="gallery-header-actions ml-auto flex shrink-0 items-center gap-1 sm:gap-1.5"
      >
        <NuxtLink
          v-if="route.path !== '/images'"
          to="/images"
          class="gallery-header-search-trigger grid size-11 shrink-0 place-items-center rounded-lg text-muted transition-colors hover:bg-muted hover:text-primary focus-visible:bg-muted focus-visible:text-primary focus-visible:outline-2 focus-visible:outline-transparent lg:hidden"
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
        <NuxtLink
          to="/submit"
          class="gallery-submit-link inline-flex min-h-11 items-center rounded-lg px-2.5 text-[0.82rem] font-semibold text-muted transition-colors hover:bg-muted hover:text-highlighted focus-visible:bg-muted focus-visible:text-highlighted focus-visible:outline-2 focus-visible:outline-transparent"
          >投稿</NuxtLink
        >
        <ConsumerAccountControl :context-actions="contextActions" />
      </div>
    </GalleryPublicContainer>
    <nav
      class="gallery-mobile-nav grid h-11 grid-cols-4 px-3 text-[0.82rem] text-muted md:hidden"
      aria-label="移动端主要导航"
    >
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
        class="relative inline-flex min-w-0 items-center justify-center px-2 transition-colors active:bg-primary/8"
        :class="
          route.path === item.to ||
          (item.to !== '/' && route.path.startsWith(item.to))
            ? 'font-semibold text-primary'
            : ''
        "
        >{{ item.label }}</NuxtLink
      >
    </nav>
  </header>
</template>
