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
    "grid",
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
  <div class="gallery-grid">
    <article
      v-for="(image, index) in items"
      :key="image.id"
      class="gallery-tile group"
    >
      <NuxtLink
        :to="`/images/${encodeURIComponent(image.id)}`"
        class="gallery-tile-link"
        :aria-haspopup="quickView ? 'dialog' : undefined"
        @click="openPreview($event, image.id)"
      >
        <div
          class="gallery-tile-media"
          :style="{ backgroundColor: image.dominantColor || undefined }"
        >
          <img
            v-bind="sourcePolicy(image.assetId, index)"
            :alt="image.altText || image.title"
            :width="image.width || undefined"
            :height="image.height || undefined"
            class="size-full object-cover"
          />
        </div>
        <div class="gallery-tile-copy">
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
