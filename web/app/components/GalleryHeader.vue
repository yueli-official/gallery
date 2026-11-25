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
function isActive(path: string) {
  return route.path === path || (path !== "/" && route.path.startsWith(path));
}
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
    class="gallery-header sticky top-0 z-40 border-b border-[var(--gallery-line)] bg-[color-mix(in_srgb,var(--gallery-canvas)_94%,transparent)] backdrop-blur-md"
  >
    <GalleryPublicContainer
      class="gallery-header-inner flex h-16 items-center gap-3 px-4 sm:px-6 md:gap-6 lg:px-8"
    >
      <NuxtLink
        to="/"
        class="gallery-wordmark inline-flex shrink-0 items-center gap-2.5 text-[var(--gallery-ink)] focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--gallery-accent)]"
        :aria-label="`${props.brandName || '图库'}首页`"
      >
        <span
          class="gallery-mark relative block size-[1.72rem] shrink-0 bg-transparent"
          aria-hidden="true"
          ><span
        /></span>
        <span
          class="gallery-wordmark-text max-w-[8rem] truncate max-[23rem]:hidden sm:max-w-[10rem] lg:max-w-[12rem]"
          >{{ props.brandName || "图库" }}</span
        >
      </NuxtLink>

      <nav
        class="gallery-desktop-nav hidden h-full items-center gap-5 md:flex lg:gap-7"
        aria-label="主要导航"
      >
        <NuxtLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          :aria-current="isActive(item.to) ? 'page' : undefined"
          class="gallery-nav-link relative inline-flex h-full items-center text-[0.8rem] font-medium text-[var(--gallery-muted)] transition-colors hover:text-[var(--gallery-ink)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--gallery-accent)] aria-[current=page]:text-[var(--gallery-ink)]"
          >{{ item.label }}</NuxtLink
        >
      </nav>

      <div
        v-if="route.path !== '/images'"
        class="gallery-header-search ml-auto hidden min-w-0 max-w-[18rem] flex-1 lg:block xl:max-w-[21rem]"
      >
        <GalleryGlobalSearch compact :placeholder="props.searchPlaceholder" />
      </div>

      <div
        class="gallery-header-actions ml-auto flex shrink-0 items-center gap-1 sm:gap-1.5 lg:ml-0"
      >
        <NuxtLink
          v-if="route.path !== '/images'"
          to="/images"
          class="gallery-header-search-trigger grid size-9 shrink-0 place-items-center text-[var(--gallery-muted)] transition-colors hover:text-[var(--gallery-ink)] focus-visible:outline-2 focus-visible:outline-[var(--gallery-accent)] lg:hidden"
          aria-label="搜索图库"
        >
          <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
            <circle
              cx="11"
              cy="11"
              r="6.5"
              fill="none"
              stroke="currentColor"
              stroke-width="1.7"
            />
            <path
              d="m16 16 4 4"
              fill="none"
              stroke="currentColor"
              stroke-width="1.7"
              stroke-linecap="round"
            />
          </svg>
        </NuxtLink>
        <NuxtLink
          to="/submit"
          class="gallery-submit-link inline-flex h-8 items-center gap-1.5 rounded-full bg-[var(--gallery-ink)] px-3.5 text-[0.72rem] font-semibold text-[var(--gallery-canvas)] transition-[background,color,transform] hover:-translate-y-px hover:bg-[var(--gallery-accent)] hover:text-[var(--gallery-on-accent)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--gallery-accent)] sm:px-3.5"
          ><UIcon name="i-tabler-photo-up" class="hidden size-3.5 sm:block" />投稿</NuxtLink
        >
        <UColorModeButton
          color="neutral"
          variant="ghost"
          aria-label="切换颜色模式"
        />
        <ConsumerAccountControl :trigger-mode="loggedIn ? 'collapsed' : 'inline'" :context-actions="contextActions" />
      </div>
    </GalleryPublicContainer>
    <nav
      class="gallery-mobile-nav grid h-9 grid-cols-4 border-t border-[color-mix(in_srgb,var(--gallery-line)_60%,transparent)] px-3 text-[0.73rem] text-[var(--gallery-muted)] md:hidden"
      aria-label="移动端主要导航"
    >
      <NuxtLink
        v-for="item in nav"
        :key="item.to"
        :to="item.to"
        :aria-current="isActive(item.to) ? 'page' : undefined"
        class="relative inline-flex min-w-0 items-center justify-center px-2 transition-colors aria-[current=page]:font-semibold aria-[current=page]:text-[var(--gallery-ink)]"
        >{{ item.label }}</NuxtLink
      >
    </nav>
  </header>
</template>

<style scoped>
.gallery-wordmark-text {
  font-family: var(--font-sans);
  font-size: 1.22rem;
  font-weight: 800;
  line-height: 1;
  letter-spacing: -0.04em;
}

.gallery-nav-link::after {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 2px;
  border-radius: 999px 999px 0 0;
  background: var(--gallery-accent);
  content: "";
  opacity: 0;
  transform: scaleX(0.4);
  transform-origin: left center;
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

.gallery-nav-link[aria-current="page"]::after {
  opacity: 1;
  transform: scaleX(1);
}

.gallery-header-search :deep(.gallery-global-search) {
  border-color: transparent;
  border-radius: 999px;
  background: var(--gallery-panel);
}

.gallery-header-actions :deep(button) {
  border-radius: 999px;
}

@media (max-width: 420px) {
  .gallery-wordmark-text {
    font-size: 1.08rem;
  }
}
</style>
