<script setup lang="ts">
import type { GalleryArtworkCard } from "~/types/gallery";

defineProps<{
  items: GalleryArtworkCard[];
  priority?: boolean;
}>();
</script>

<template>
  <div
    class="grid grid-cols-2 gap-x-3 gap-y-6 sm:grid-cols-3 sm:gap-x-4 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6"
  >
    <NuxtLink
      v-for="(artwork, index) in items"
      :key="artwork.id"
      :to="`/artworks/${encodeURIComponent(artwork.id)}`"
      class="group min-w-0 rounded-xl focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary"
    >
      <div
        class="overflow-hidden rounded-xl bg-elevated"
        :style="{ aspectRatio: artworkAspectRatio(artwork.width, artwork.height) }"
      >
        <img
          v-if="artwork.coverUrl"
          :src="artwork.coverUrl"
          :alt="artwork.title"
          :width="artwork.width || undefined"
          :height="artwork.height || undefined"
          :loading="priority && index < 2 ? 'eager' : 'lazy'"
          :fetchpriority="priority && index === 0 ? 'high' : 'auto'"
          class="size-full object-cover transition-transform duration-300 group-hover:scale-[1.02]"
        />
        <div v-else class="grid size-full place-items-center text-dimmed">
          <UIcon name="i-tabler-photo-off" class="size-8" />
        </div>
      </div>
      <div class="mt-3 min-w-0">
        <h3 class="truncate text-sm font-semibold text-highlighted">
          {{ artwork.title }}
        </h3>
        <p class="mt-1 truncate text-xs text-muted">
          {{ artwork.creator.displayName }}
        </p>
      </div>
    </NuxtLink>
  </div>
</template>
