<script setup lang="ts">
import type { PlatformUserMenuAction } from "@platform/ui/components";

defineProps<{ brandName?: string }>();

const route = useRoute();
const nav = [
  { label: "随机看看", to: "/" },
  { label: "浏览", to: "/images" },
  { label: "专题", to: "/collections" },
  { label: "排行榜", to: "/rankings" },
];
const contextActions: PlatformUserMenuAction[] = [
  { label: "我的收藏", icon: "i-tabler-heart", to: "/favorites" },
  { label: "我的投稿", icon: "i-tabler-photo-up", to: "/submissions" },
];
</script>

<template>
  <header class="gallery-header">
    <div class="mx-auto flex h-16 w-full max-w-[112rem] items-center gap-5 px-4 sm:px-6 lg:px-8">
      <NuxtLink to="/" class="gallery-wordmark" aria-label="月离图库首页">
        <span class="gallery-mark" aria-hidden="true"><span /></span>
        <span>{{ brandName || "月离图库" }}</span>
      </NuxtLink>

      <nav class="hidden items-center gap-1 md:flex" aria-label="主要导航">
        <UButton
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          :label="item.label"
          color="neutral"
          :variant="route.path === item.to || (item.to !== '/' && route.path.startsWith(item.to)) ? 'soft' : 'ghost'"
          size="sm"
        />
      </nav>

      <div class="ml-auto flex items-center gap-1.5">
        <UButton to="/submit" icon="i-tabler-plus" label="投稿" color="primary" variant="solid" size="sm" />
        <UColorModeButton color="neutral" variant="ghost" aria-label="切换颜色模式" />
        <ConsumerAccountControl :context-actions="contextActions" manage-to="/manage" manage-label="图库管理" />
      </div>
    </div>
    <nav class="gallery-mobile-nav md:hidden" aria-label="移动端主要导航">
      <NuxtLink
        v-for="item in nav"
        :key="item.to"
        :to="item.to"
        :aria-current="route.path === item.to ? 'page' : undefined"
      >{{ item.label }}</NuxtLink>
    </nav>
  </header>
</template>
