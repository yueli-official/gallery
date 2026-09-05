<script setup lang="ts">
import type { GallerySite } from "~/types/gallery";

const { brand } = useSiteRuntime();
const gallerySite = useGallerySiteSettings();
const { data: siteResponse } = await useFetch<{ site: GallerySite }>(
  "/api/gallery/site",
  { key: "gallery-public-site-settings" },
);

watch(
  () => siteResponse.value?.site,
  (value) => {
    if (value) gallerySite.value = value;
  },
  { immediate: true },
);

const siteName = computed(() => gallerySite.value?.name || brand.value);
</script>

<template>
  <div class="yueli-app-shell flex min-h-dvh flex-col text-default">
    <GalleryHeader
      :brand-name="siteName"
      :search-placeholder="gallerySite?.searchPlaceholder"
    />
    <GalleryPublicContainer
      as="main"
      id="public-main"
      tabindex="-1"
      class="gallery-main min-w-0 flex-1 outline-none"
    >
      <slot />
    </GalleryPublicContainer>
    <footer class="gallery-footer mt-24 border-t border-default">
      <GalleryPublicContainer
        class="gallery-footer-inner flex flex-col gap-6 px-4 py-8 sm:flex-row sm:items-end sm:justify-between sm:px-6 lg:px-8"
      >
        <div>
          <p class="font-display text-sm font-semibold text-highlighted">
            {{ siteName }}
          </p>
          <p
            v-if="gallerySite?.footerTagline"
            class="mt-1 max-w-md text-xs leading-5 text-muted"
          >
            {{ gallerySite.footerTagline }}
          </p>
        </div>
        <nav
          class="flex flex-wrap gap-x-5 gap-y-2 text-xs text-muted [&_a]:transition-colors [&_a:hover]:text-highlighted"
          aria-label="页脚导航"
        >
          <NuxtLink to="/images">浏览图片</NuxtLink>
          <NuxtLink to="/collections">专题集合</NuxtLink>
          <NuxtLink to="/submit">投稿图片</NuxtLink>
        </nav>
      </GalleryPublicContainer>
    </footer>
  </div>
</template>
