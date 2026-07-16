<script setup lang="ts">
import type { GalleryCollectionDetail } from "~/types/gallery";

definePageMeta({ middleware: "auth" });
const hydrated = useClientHydrated();
const { call } = useApi();
const { data, error, pending, refresh } = await useAsyncData(
  "gallery-my-favorites",
  () =>
    call<{ collection: GalleryCollectionDetail }>(
      "/api/v1/gallery/me/favorites",
    ),
  { server: false },
);
const collection = computed(() => data.value?.collection);
useSeoMeta({ title: "我的收藏", robots: "noindex,nofollow" });
</script>

<template>
  <div class="gallery-page">
    <header class="gallery-page-header">
      <div>
        <h1 class="gallery-page-title">我的收藏</h1>
        <p class="gallery-page-copy">只对你可见，收藏过的图片都在这里。</p>
      </div>
    </header>
    <div v-if="!hydrated || pending" class="gallery-grid">
      <USkeleton
        v-for="index in 12"
        :key="index"
        class="aspect-[4/3] rounded-lg"
      />
    </div>
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="收藏加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <GalleryImageGrid
      v-else-if="collection?.images.length"
      :items="collection.images"
    />
    <div v-else class="gallery-compact-empty">
      <div>
        <span class="gallery-empty-icon"
          ><UIcon name="i-tabler-heart" class="size-6"
        /></span>
        <h2 class="mt-4 text-lg font-semibold text-highlighted">
          还没有收藏图片
        </h2>
        <p class="mt-2 text-sm text-muted">遇到想再看的图片时，点一下收藏。</p>
        <UButton
          to="/images"
          class="mt-4"
          color="neutral"
          variant="outline"
          label="去浏览"
        />
      </div>
    </div>
  </div>
</template>
