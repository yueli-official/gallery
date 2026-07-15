<script setup lang="ts">
const { brand } = useSiteRuntime();
</script>

<template>
  <div class="platform-app-shell flex min-h-dvh flex-col text-default">
    <GalleryHeader :brand-name="brand" />
    <main id="public-main" tabindex="-1" class="gallery-main flex-1 outline-none">
      <slot />
    </main>
    <footer class="mt-16 border-t border-default/80">
      <div class="mx-auto flex w-full max-w-[112rem] flex-col gap-2 px-4 py-7 text-xs text-muted sm:flex-row sm:items-center sm:justify-between sm:px-6 lg:px-8">
        <p>月离图库 · 公开图片的发现、收藏与整理</p>
        <p>来源可补充，投稿者不公开，举报由运营复核。</p>
      </div>
    </footer>
  </div>
</template>
