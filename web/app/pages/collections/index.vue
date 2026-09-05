<script setup lang="ts">
import type { GalleryCollection } from "~/types/gallery";

const { data, error, status, refresh } = await useFetch<{
  items: GalleryCollection[];
}>("/api/gallery/collections");
useSeoMeta({ title: "专题集合", description: "由运营方整理的公开图片专题。" });
</script>

<template>
  <GalleryPublicPage>
    <GalleryPageHeader
      class="max-w-3xl"
      title="专题集合"
      description="围绕一个主题重新整理图片，让浏览更有方向。"
    />
    <div v-if="status === 'pending'" class="grid gap-5 sm:grid-cols-2">
      <USkeleton v-for="index in 6" :key="index" class="h-56 rounded-lg" />
    </div>
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="专题加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <div
      v-else-if="data?.items.length"
      class="grid gap-x-5 gap-y-9 sm:grid-cols-2 lg:grid-cols-3"
    >
      <NuxtLink
        v-for="collection in data.items"
        :key="collection.id"
        :to="`/collections/${collection.slug}`"
        class="gallery-collection-card group block min-w-0 rounded-2xl focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary"
      >
        <div
          class="gallery-collection-cover relative isolate grid aspect-video place-items-center overflow-hidden rounded-[0.9rem] bg-muted transition-transform duration-200 ease-out group-hover:-translate-y-0.5"
          :style="{ backgroundColor: collection.coverColor || undefined }"
        >
          <img
            v-if="collection.coverAssetId"
            v-bind="galleryImageSources(collection.coverAssetId, 'grid', false)"
            :alt="collection.coverAltText || collection.name"
            :width="collection.coverWidth"
            :height="collection.coverHeight"
            class="relative z-1 size-full object-cover transition-transform duration-300 ease-out group-hover:scale-[1.025]"
          />
          <UIcon
            v-else
            name="i-tabler-folders"
            class="size-10 text-primary transition-transform duration-300 group-hover:-rotate-3 group-hover:scale-105"
          />
          <span
            class="absolute bottom-3.5 right-4 z-1 rounded-md bg-black/70 px-2 py-1 text-[0.72rem] text-white/90"
            >{{ collection.itemCount }} 张</span
          >
        </div>
        <div class="mt-4 flex items-start justify-between gap-4 px-1">
          <div>
            <h2 class="font-semibold text-highlighted">
              {{ collection.name }}
            </h2>
            <p class="mt-1 line-clamp-2 text-sm text-muted">
              {{ collection.description }}
            </p>
          </div>
          <UIcon
            name="i-tabler-arrow-up-right"
            class="mt-1 size-4 shrink-0 text-muted transition group-hover:text-primary"
          />
        </div>
      </NuxtLink>
    </div>
    <GalleryCompactEmpty
      v-else
      icon="i-tabler-folders"
      title="还没有公开专题"
      description="运营整理的主题集合会出现在这里。"
    />
  </GalleryPublicPage>
</template>
