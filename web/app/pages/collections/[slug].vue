<script setup lang="ts">
import type { GalleryCollectionDetail } from "~/types/gallery";

const route = useRoute("/collections/[slug]");
const { data, error, status } = await useFetch<{ collection: GalleryCollectionDetail }>(() => `/api/gallery/collections/${route.params.slug}`);
const collection = computed(() => data.value?.collection);
useSeoMeta({ title: () => collection.value?.name || "专题", description: () => collection.value?.description || "公开图片专题。" });
</script>

<template>
  <div class="gallery-page">
    <header v-if="collection" class="mb-8 max-w-3xl">
      <NuxtLink to="/collections" class="text-sm text-muted hover:text-default">← 全部专题</NuxtLink>
      <h1 class="mt-4 text-3xl font-semibold tracking-tight text-highlighted sm:text-4xl">{{ collection.name }}</h1>
      <p v-if="collection.description" class="mt-3 leading-7 text-toned">{{ collection.description }}</p>
      <p class="mt-3 text-sm text-muted">{{ collection.itemCount }} 张图片</p>
    </header>
    <div v-if="status === 'pending'" class="gallery-grid"><USkeleton v-for="index in 12" :key="index" class="aspect-[4/3] rounded-lg" /></div>
    <UAlert v-else-if="error" color="error" variant="subtle" title="专题不存在或尚未公开" />
    <GalleryImageGrid v-else-if="collection?.images.length" :items="collection.images" priority />
    <div v-else-if="collection" class="border-y border-dashed border-default py-16 text-center text-sm text-muted">这个专题还没有图片。</div>
  </div>
</template>
