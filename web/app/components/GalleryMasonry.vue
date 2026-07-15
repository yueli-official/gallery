<script setup lang="ts">
import type { GalleryImageCard } from "~/types/gallery";

defineProps<{ items: GalleryImageCard[]; priority?: boolean }>();
</script>

<template>
  <div class="gallery-masonry">
    <NuxtLink
      v-for="(image, index) in items"
      :key="image.id"
      :to="`/images/${encodeURIComponent(image.id)}`"
      class="gallery-masonry-item group"
    >
      <div
        class="relative overflow-hidden bg-elevated"
        :style="{ aspectRatio: imageAspect(image.width, image.height, '4 / 5'), backgroundColor: image.dominantColor || undefined }"
      >
        <img
          :src="galleryRendition(image.assetId, index < 8 ? 'masonry-lg' : 'masonry-sm')"
          :alt="image.altText || image.title"
          :width="image.width || undefined"
          :height="image.height || undefined"
          :loading="priority && index < 4 ? 'eager' : 'lazy'"
          :fetchpriority="priority && index === 0 ? 'high' : 'auto'"
          class="size-full object-cover transition duration-300 group-hover:scale-[1.012] group-hover:brightness-[.96]"
        />
        <div class="gallery-masonry-caption">
          <p class="truncate text-sm font-medium text-white">{{ image.title }}</p>
		  <p v-if="image.primaryCategory" class="mt-0.5 text-xs text-white/70">{{ image.primaryCategory }}</p>
        </div>
      </div>
    </NuxtLink>
  </div>
</template>
