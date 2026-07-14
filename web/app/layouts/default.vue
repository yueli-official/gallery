<script setup lang="ts">
import type { GalleryDiscovery } from "~/types/gallery";

const { data } = await useFetch<GalleryDiscovery>("/api/gallery/discovery", {
  key: "gallery-discovery-shell",
});
</script>

<template>
  <div class="platform-app-shell flex min-h-dvh flex-col text-default">
    <GalleryHeader
      :brand-name="data?.site.name"
      :brand-tagline="data?.site.title"
    />
    <main
      id="public-main"
      tabindex="-1"
      class="mx-auto w-full max-w-screen-2xl flex-1 px-4 py-8 outline-none sm:px-6 lg:px-8"
    >
      <slot />
    </main>
    <footer class="border-t border-default bg-default/80">
      <div
        class="mx-auto flex w-full max-w-screen-2xl flex-col gap-2 px-4 py-6 text-sm text-muted sm:flex-row sm:items-center sm:justify-between sm:px-6 lg:px-8"
      >
        <p>{{ data?.site.footerTagline }}</p>
        <p>尊重原创、版权与创作者声明。</p>
      </div>
    </footer>
  </div>
</template>

