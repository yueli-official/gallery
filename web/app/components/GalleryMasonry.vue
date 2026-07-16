<script setup lang="ts">
import type { GalleryImageCard } from "~/types/gallery";

const props = defineProps<{
  items: GalleryImageCard[];
  priority?: boolean;
  quickView?: boolean;
}>();
const emit = defineEmits<{
  preview: [imageId: string, trigger: HTMLElement | null];
}>();

function sourcePolicy(assetId: string, index: number) {
  return galleryImageSources(
    assetId,
    "masonry",
    Boolean(props.priority && index < 6),
  );
}

function openPreview(event: MouseEvent, imageId: string): void {
  if (
    !props.quickView ||
    event.button !== 0 ||
    event.metaKey ||
    event.ctrlKey ||
    event.shiftKey ||
    event.altKey
  ) {
    return;
  }
  event.preventDefault();
  emit("preview", imageId, event.currentTarget as HTMLElement | null);
}
</script>

<template>
  <div class="gallery-masonry">
    <NuxtLink
      v-for="(image, index) in items"
      :key="image.id"
      :to="`/images/${encodeURIComponent(image.id)}`"
      class="gallery-masonry-item group"
      :aria-haspopup="quickView ? 'dialog' : undefined"
      @click="openPreview($event, image.id)"
    >
      <div
        class="relative overflow-hidden bg-elevated"
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
          class="size-full object-cover transition duration-500 group-hover:scale-[1.025] group-hover:brightness-[.94]"
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
  </div>
</template>
