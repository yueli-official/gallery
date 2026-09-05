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
  <div
    class="gallery-masonry columns-2 gap-3 sm:columns-3 sm:gap-[0.8rem] lg:columns-4 lg:gap-4 min-[90rem]:columns-5"
  >
    <slot name="lead" />
    <article
      v-for="(image, index) in items"
      :key="image.id"
      class="gallery-masonry-item group relative mb-3 block break-inside-avoid rounded-[0.9rem] active:translate-y-px sm:mb-[0.8rem] lg:mb-4"
    >
      <NuxtLink
        :to="`/images/${encodeURIComponent(image.id)}`"
        class="gallery-masonry-link block rounded-[0.9rem] focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary"
      >
        <div
          class="gallery-masonry-media relative overflow-hidden rounded-[0.9rem] bg-muted"
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
            class="size-full object-cover transition-[transform,filter] duration-200 ease-out group-hover:scale-[1.018] group-hover:brightness-[0.96] group-hover:saturate-[0.94]"
          />
          <div
            class="gallery-masonry-caption absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/80 to-transparent px-4 pb-4 pt-16 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100 [@media(hover:none)]:pt-8 [@media(hover:none)]:opacity-100"
          >
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
