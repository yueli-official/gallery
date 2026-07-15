<script setup lang="ts">
import type { GalleryImageCard } from "~/types/gallery";

defineProps<{ items: GalleryImageCard[]; priority?: boolean }>();
</script>

<template>
  <div class="gallery-grid">
    <NuxtLink
      v-for="(image, index) in items"
      :key="image.id"
      :to="`/images/${encodeURIComponent(image.id)}`"
      class="gallery-tile group"
    >
      <div class="gallery-tile-media" :style="{ backgroundColor: image.dominantColor || undefined }">
        <img
          :src="galleryRendition(image.assetId, index < 8 ? 'grid-lg' : 'grid-sm')"
          :alt="image.altText || image.title"
          :width="image.width || undefined"
          :height="image.height || undefined"
          :loading="priority && index < 4 ? 'eager' : 'lazy'"
          :fetchpriority="priority && index === 0 ? 'high' : 'auto'"
          class="size-full object-cover transition-transform duration-300 group-hover:scale-[1.015]"
        />
        <span v-if="image.topic" class="gallery-topic">{{ image.topic }}</span>
      </div>
      <div class="mt-2.5 flex min-w-0 items-start justify-between gap-3">
        <h2 class="truncate text-sm font-medium text-highlighted">{{ image.title }}</h2>
        <span class="flex shrink-0 items-center gap-1 text-xs text-muted" :aria-label="`${image.metrics.favorites} 次收藏`">
          <UIcon name="i-tabler-heart" class="size-3.5" />{{ compactMetric(image.metrics.favorites) }}
        </span>
      </div>
    </NuxtLink>
  </div>
</template>
