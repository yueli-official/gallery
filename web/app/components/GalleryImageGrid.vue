<script setup lang="ts">
import type { GalleryImageCard } from "~/types/gallery";

const props = defineProps<{
  items: GalleryImageCard[];
  priority?: boolean;
}>();

function sourcePolicy(assetId: string, index: number) {
  return galleryImageSources(
    assetId,
    "grid",
    Boolean(props.priority && index < 6),
  );
}

</script>

<template>
  <div
    class="gallery-grid grid grid-cols-2 gap-x-3 gap-y-[1.65rem] sm:grid-cols-3 sm:gap-x-4 lg:grid-cols-4 lg:gap-x-[1.1rem] min-[90rem]:grid-cols-5"
  >
    <article
      v-for="(image, index) in items"
      :key="image.id"
      class="gallery-tile group relative min-w-0 rounded-[0.9rem] active:translate-y-px"
    >
      <NuxtLink
        :to="`/images/${encodeURIComponent(image.id)}`"
        class="gallery-tile-link block rounded-[0.9rem] focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary"
      >
        <div
          class="gallery-tile-media relative aspect-[4/3] overflow-hidden rounded-[0.9rem] bg-muted"
          :style="{ backgroundColor: image.dominantColor || undefined }"
        >
          <img
            v-bind="sourcePolicy(image.assetId, index)"
            :alt="image.altText || image.title"
            :width="image.width || undefined"
            :height="image.height || undefined"
            class="size-full object-cover transition-[transform,filter] duration-200 ease-out group-hover:scale-[1.018] group-hover:brightness-[0.96] group-hover:saturate-[0.94]"
          />
        </div>
        <div
          class="gallery-tile-copy flex min-w-0 items-start justify-between gap-3 px-1 pt-3"
        >
          <div class="min-w-0">
            <h2 class="truncate text-sm font-semibold text-highlighted">
              {{ image.title }}
            </h2>
            <p
              v-if="image.primaryCategory"
              class="mt-0.5 truncate text-xs text-muted"
            >
              {{ image.primaryCategory }}
            </p>
          </div>
          <span
            class="flex shrink-0 items-center gap-1 text-xs text-muted"
            :aria-label="`${image.metrics.favorites} 次收藏`"
          >
            <UIcon name="i-tabler-heart" class="size-3.5" />{{
              compactMetric(image.metrics.favorites)
            }}
          </span>
        </div>
      </NuxtLink>
      <GalleryFavoriteButton :image-id="image.id" :title="image.title" />
    </article>
  </div>
</template>
