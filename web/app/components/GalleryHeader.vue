<script setup lang="ts">
import type { PlatformUserMenuAction } from "@platform/ui/components";

defineProps<{ brandName?: string }>();

const route = useRoute();
const nav = [
  { label: "发现", to: "/" },
  { label: "浏览", to: "/images" },
  { label: "专题", to: "/collections" },
  { label: "排行", to: "/rankings" },
];
const contextActions: PlatformUserMenuAction[] = [
  { label: "我的收藏", icon: "i-tabler-heart", to: "/favorites" },
  { label: "我的投稿", icon: "i-tabler-photo-up", to: "/submissions" },
];
</script>

<template>
  <header class="gallery-header">
    <div class="gallery-header-inner">
      <NuxtLink to="/" class="gallery-wordmark" aria-label="月离图库首页">
        <span class="gallery-mark" aria-hidden="true"><span /></span>
        <span>{{ brandName || "月离图库" }}</span>
      </NuxtLink>

      <nav class="hidden items-center gap-1 lg:flex" aria-label="主要导航">
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

      <GalleryGlobalSearch class="ml-auto hidden xl:flex" />

      <div class="ml-auto flex items-center gap-1 sm:gap-1.5 xl:ml-0">
        <UButton
          class="xl:hidden"
          to="/images"
          color="neutral"
          variant="ghost"
          icon="i-tabler-search"
          aria-label="搜索图库"
        />
        <UButton
          class="gallery-submit-button"
          to="/submit"
          icon="i-tabler-plus"
          label="投稿"
          color="primary"
          variant="solid"
          size="sm"
        />
        <UColorModeButton
          color="neutral"
          variant="ghost"
          aria-label="切换颜色模式"
        />
        <ConsumerAccountControl
          :context-actions="contextActions"
          manage-to="/manage"
          manage-label="图库管理"
        />
      </div>
    </div>
    <nav class="gallery-mobile-nav lg:hidden" aria-label="移动端主要导航">
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
