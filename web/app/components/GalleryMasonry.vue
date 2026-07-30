<script setup lang="ts">
import type { GalleryImageCard } from "~/types/gallery";

const props = defineProps<{
  items: GalleryImageCard[];
  priority?: boolean;
}>();

function sourcePolicy(assetId: string, index: number) {
  return galleryImageSources(
    assetId,
    "masonry",
    Boolean(props.priority && index < 6),
  );
}

</script>

<template>
  <div class="gallery-masonry">
    <slot name="lead" />
    <article
      v-for="(image, index) in items"
      :key="image.id"
      class="gallery-masonry-item group"
    >
      <NuxtLink
        :to="`/images/${encodeURIComponent(image.id)}`"
        class="gallery-masonry-link"
      >
        <div
          class="gallery-masonry-media"
          :style="{
            aspectRatio: imageAspect(image.width, image.height, '4 / 5'),
            backgroundColor: image.dominantColor || undefined,
          }"
        >
          <img
            v-bind="sourcePolicy(image.assetId, index)"
            :alt="image.altText || image.title"
            :width="image.width || undefined"
            :height="image.height || undefined"
            class="size-full object-cover"
          />
          <div class="gallery-masonry-caption">
            <p class="truncate text-sm font-semibold text-white">
              {{ image.title }}
            </p>
            <p v-if="image.primaryCategory" class="mt-1 text-xs text-white/75">
              {{ image.primaryCategory }}
            </p>
          </div>
        </div>
      </NuxtLink>
      <GalleryFavoriteButton :image-id="image.id" :title="image.title" />
    </article>
  </div>
</template>
