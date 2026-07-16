<script setup lang="ts">
import type { GalleryCollection } from "~/types/gallery";

const { data, error, status, refresh } = await useFetch<{
  collections: GalleryCollection[];
}>("/api/gallery/collections");
useSeoMeta({ title: "专题集合", description: "由运营方整理的公开图片专题。" });
</script>

<template>
  <div class="gallery-page max-w-7xl">
    <header class="gallery-page-header max-w-3xl">
      <div>
        <p class="gallery-eyebrow">运营精选</p>
        <h1 class="gallery-page-title mt-3">专题集合</h1>
        <p class="gallery-page-copy">
          围绕一个主题重新整理图片，让浏览更有方向。
        </p>
      </div>
    </header>
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
      v-else-if="data?.collections.length"
      class="grid gap-x-6 gap-y-10 sm:grid-cols-2"
    >
      <NuxtLink
        v-for="collection in data.collections"
        :key="collection.id"
        :to="`/collections/${collection.slug}`"
        class="gallery-collection-card group focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary"
      >
        <div
          class="gallery-collection-cover"
          :style="{ backgroundColor: collection.coverColor || undefined }"
        >
          <img
            v-if="collection.coverAssetId"
            v-bind="galleryImageSources(collection.coverAssetId, 'grid', false)"
            :alt="collection.coverAltText || collection.name"
            :width="collection.coverWidth"
            :height="collection.coverHeight"
          />
          <UIcon
            v-else
            name="i-tabler-folders"
            class="size-10 text-primary transition-transform duration-300 group-hover:-rotate-3 group-hover:scale-105"
          />
          <span>{{ collection.itemCount }} 张</span>
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
    <div v-else class="gallery-compact-empty">
      <span class="gallery-empty-icon"
        ><UIcon name="i-tabler-folders" class="size-6"
      /></span>
      <h2 class="mt-4 text-lg font-semibold text-highlighted">
        还没有公开专题
      </h2>
      <p class="mt-2 text-sm text-muted">运营整理的主题集合会出现在这里。</p>
    </div>
  </div>
</template>
