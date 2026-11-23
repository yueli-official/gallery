<script setup lang="ts">
import type { GalleryImageCard } from "~/types/gallery";

defineProps<{ items: GalleryImageCard[]; removing?: string }>();
const emit = defineEmits<{
  remove: [imageId: string];
}>();
</script>

<template>
  <div class="favorite-grid">
    <article
      v-for="image in items"
      :key="image.id"
      class="favorite-tile group relative min-w-0"
    >
      <NuxtLink
        :to="`/images/${encodeURIComponent(image.publicId)}`"
        class="block focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--gallery-accent)]"
      >
        <div
          class="favorite-media relative aspect-[3/4] overflow-hidden rounded-[0.9rem]"
          :style="{ backgroundColor: image.dominantColor || undefined }"
        >
          <img
            v-bind="galleryImageSources(image, 'grid')"
            :alt="image.altText || image.title"
            :width="image.width || undefined"
            :height="image.height || undefined"
            class="size-full object-cover transition-transform duration-500 ease-out group-hover:scale-[1.045]"
          />
          <span v-if="image.primaryCategory" class="favorite-tag">
            {{ image.primaryCategory }}
          </span>
        </div>
        <div class="favorite-copy">
          <h2 class="favorite-title">{{ image.title }}</h2>
          <p class="favorite-meta font-numeric">
            ♥ {{ compactMetric(image.metrics.favorites) }}
          </p>
        </div>
      </NuxtLink>
      <button
        type="button"
        class="favorite-remove"
        :class="{ 'is-busy': removing === image.id }"
        :disabled="removing === image.id"
        :aria-label="`取消收藏：${image.title}`"
        @click="emit('remove', image.id)"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path
            d="M12 20.2S4.3 15.4 4.3 9.6A4.1 4.1 0 0 1 12 7.2a4.1 4.1 0 0 1 7.7 2.4c0 5.8-7.7 10.6-7.7 10.6Z"
            fill="currentColor"
            stroke="currentColor"
            stroke-width="1.4"
            stroke-linejoin="round"
          />
        </svg>
      </button>
    </article>
  </div>
</template>

<style scoped>
.favorite-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.4rem 0.65rem;
}

.favorite-tile {
  transition: transform 220ms ease;
}

.favorite-tile:hover {
  transform: translateY(-4px);
}

.favorite-media {
  background: var(--gallery-panel);
}

.favorite-tag {
  position: absolute;
  bottom: 0.45rem;
  left: 0.45rem;
  padding: 0.12rem 0.4rem;
  background: rgb(255 255 255 / 0.9);
  font-size: 0.56rem;
  font-weight: 600;
  line-height: 1.4;
  color: #0b0c12;
}

.favorite-copy {
  padding: 0.5rem 0.05rem 0;
}

.favorite-title {
  overflow: hidden;
  font-size: 0.77rem;
  font-weight: 600;
  line-height: 1.35;
  color: var(--gallery-ink);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.favorite-meta {
  margin-top: 0.12rem;
  font-size: 0.6rem;
  color: var(--gallery-muted);
}

.favorite-remove {
  position: absolute;
  top: 0.45rem;
  right: 0.45rem;
  display: grid;
  place-items: center;
  width: 1.9rem;
  height: 1.9rem;
  border-radius: 999px;
  background: rgb(255 255 255 / 0.88);
  color: #b42318;
  backdrop-filter: blur(8px);
  transition:
    background 160ms ease,
    color 160ms ease;
}

.favorite-remove:hover {
  background: #b42318;
  color: #fff;
}

.favorite-remove.is-busy {
  cursor: wait;
  opacity: 0.6;
}

.favorite-remove svg {
  width: 0.95rem;
  height: 0.95rem;
}

.favorite-remove:focus-visible {
  outline: 2px solid var(--gallery-accent);
  outline-offset: 2px;
}

@media (min-width: 640px) {
  .favorite-grid {
    gap: 1.6rem 0.75rem;
  }

  .favorite-title {
    font-size: 0.8rem;
  }
}

@media (min-width: 701px) {
  .favorite-grid {
    grid-template-columns: repeat(auto-fill, minmax(15rem, 1fr));
  }
}
</style>
