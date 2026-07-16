<script setup lang="ts">
import type { GalleryCollectionDetail } from "~/types/gallery";

const route = useRoute("/collections/[slug]");
const { data, error, status } = await useFetch<{
  collection: GalleryCollectionDetail;
}>(() => `/api/gallery/collections/${route.params.slug}`);
const collection = computed(() => data.value?.collection);
useSeoMeta({
  title: () => collection.value?.name || "专题",
  description: () => collection.value?.description || "公开图片专题。",
});
</script>

<template>
  <div class="gallery-page">
    <header v-if="collection" class="gallery-page-header">
      <div>
        <NuxtLink
          to="/collections"
          class="mb-4 inline-flex items-center gap-1.5 text-sm text-muted transition-colors hover:text-highlighted"
        >
          <UIcon name="i-tabler-arrow-left" class="size-4" />
          全部专题
        </NuxtLink>
        <h1 class="gallery-page-title">{{ collection.name }}</h1>
        <p v-if="collection.description" class="gallery-page-copy text-base">
          {{ collection.description }}
        </p>
      </div>
      <p class="gallery-count">
        <span>{{ collection.itemCount }}</span> 张图片
      </p>
    </header>
    <div v-if="status === 'pending'" class="gallery-grid">
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
      title="专题不存在或尚未公开"
    />
    <GalleryImageGrid
      v-else-if="collection?.images.length"
      :items="collection.images"
      priority
    />
    <div
      v-else-if="collection"
      class="gallery-compact-empty grid place-items-center"
    >
      <div>
        <span class="gallery-empty-icon">
          <UIcon name="i-tabler-stack-2" class="size-6" />
        </span>
        <h2 class="mt-4 font-semibold text-highlighted">专题正在整理</h2>
        <p class="mt-2 text-sm text-muted">这个专题还没有图片。</p>
      </div>
    </div>
  </div>
</template>
