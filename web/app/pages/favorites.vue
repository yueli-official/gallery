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
    <header class="mb-7">
      <h1 class="text-3xl font-semibold tracking-tight text-highlighted">
        我的收藏
      </h1>
      <p class="mt-1.5 text-sm text-muted">只对你可见的单例收藏夹。</p>
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
    <div
      v-else
      class="grid min-h-72 place-items-center border-y border-dashed border-default text-center"
    >
      <div>
        <UIcon name="i-tabler-heart" class="mx-auto size-8 text-dimmed" />
        <h2 class="mt-3 font-semibold text-highlighted">还没有收藏图片</h2>
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
