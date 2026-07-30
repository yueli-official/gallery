<script setup lang="ts">
import type { GalleryImageCard } from "~/types/gallery";

defineProps<{ items: GalleryImageCard[]; removing?: string }>();
const emit = defineEmits<{
  remove: [imageId: string];
}>();
</script>

<template>
  <div class="gallery-grid">
    <article
      v-for="image in items"
      :key="image.id"
      class="gallery-tile group relative"
    >
      <NuxtLink
        :to="`/images/${encodeURIComponent(image.id)}`"
        class="block"
      >
        <div
          class="gallery-tile-media"
          :style="{ backgroundColor: image.dominantColor || undefined }"
        >
          <img
            v-bind="galleryImageSources(image.assetId, 'grid')"
            :alt="image.altText || image.title"
            :width="image.width || undefined"
            :height="image.height || undefined"
            class="size-full object-cover transition duration-500 group-hover:scale-[1.025]"
          />
        </div>
        <div class="gallery-tile-copy pr-10">
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
        </div>
      </NuxtLink>
      <UButton
        class="absolute bottom-0 right-0"
        color="error"
        variant="ghost"
        size="xs"
        icon="i-tabler-heart-minus"
        :loading="removing === image.id"
        :aria-label="`取消收藏：${image.title}`"
        @click="emit('remove', image.id)"
      />
    </article>
  </div>
</template>
