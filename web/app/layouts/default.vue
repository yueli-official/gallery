<script setup lang="ts">
const { brand } = useSiteRuntime();
</script>

<template>
  <div class="platform-app-shell flex min-h-dvh flex-col text-default">
    <GalleryHeader :brand-name="brand" />
    <main
      id="public-main"
      tabindex="-1"
      class="gallery-main flex-1 outline-none"
    >
      <slot />
    </main>
    <footer class="gallery-footer">
      <div class="gallery-footer-inner">
        <div>
          <p class="font-display text-sm font-semibold text-highlighted">
            月离图库
          </p>
          <p class="mt-1 max-w-md text-xs leading-5 text-muted">
            公开图片的发现、收藏与整理。投稿者身份不会出现在公开页面。
          </p>
        </div>
        <nav
          class="flex flex-wrap gap-x-5 gap-y-2 text-xs text-muted"
          aria-label="页脚导航"
        >
          <NuxtLink to="/images">浏览图片</NuxtLink>
          <NuxtLink to="/collections">专题集合</NuxtLink>
          <NuxtLink to="/submit">投稿图片</NuxtLink>
        </nav>
      </div>
    </footer>
  </div>
</template>
