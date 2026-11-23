<script setup lang="ts">
import type { GalleryImageCard } from "~/types/gallery";

defineProps<{
  items: GalleryImageCard[];
}>();

const failedAssets = ref<Set<string>>(new Set());

function markFailed(assetId: string) {
  failedAssets.value = new Set([...failedAssets.value, assetId]);
}

function needle(index: number): string {
  return String(index + 1).padStart(2, "0");
}
</script>

<template>
  <ol class="home-trending">
    <li v-for="(image, index) in items" :key="image.id" class="home-trending-item">
      <NuxtLink
        :to="`/images/${encodeURIComponent(image.publicId)}`"
        class="home-trending-link group flex items-center gap-3 rounded-xl px-2 py-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--gallery-accent)]"
      >
        <span
          class="home-trending-rank font-numeric"
          :class="{ 'is-top': index < 3 }"
          >{{ needle(index) }}</span
        >
        <span
          class="home-trending-thumb relative aspect-[3/4] w-[3.2rem] shrink-0 overflow-hidden rounded-lg"
          :style="{ backgroundColor: image.dominantColor || undefined }"
        >
          <img
            v-if="!failedAssets.has(image.assetId)"
            :src="galleryRendition(image.assetId, 'grid-sm')"
            :alt="image.altText || image.title"
            loading="lazy"
            class="size-full object-cover transition-transform duration-500 group-hover:scale-[1.05]"
            @error="markFailed(image.assetId)"
          />
        </span>
        <span class="min-w-0 flex-1">
          <h3 class="home-trending-title">{{ image.title }}</h3>
          <p v-if="image.primaryCategory" class="home-trending-category">
            {{ image.primaryCategory }}
          </p>
        </span>
        <span class="home-trending-views font-numeric">
          {{ compactMetric(image.metrics.views) }}
          <span>浏览</span>
        </span>
      </NuxtLink>
    </li>
  </ol>
</template>

<style scoped>
.home-trending {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.15rem;
}

.home-trending-link {
  transition: background 160ms ease;
}

.home-trending-link:hover {
  background: var(--gallery-panel);
}

.home-trending-rank {
  width: 1.95rem;
  flex: none;
  font-size: 1.02rem;
  font-weight: 500;
  text-align: center;
  color: var(--gallery-muted);
}

.home-trending-rank.is-top {
  padding: 0 0.28rem;
  border-radius: 0.15rem;
  background: var(--gallery-accent);
  font-size: 1.32rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--gallery-on-accent);
}

.home-trending-thumb {
  background: var(--gallery-panel);
}

.home-trending-title {
  display: block;
  overflow: hidden;
  font-size: 0.79rem;
  font-weight: 600;
  line-height: 1.35;
  color: var(--gallery-ink);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-trending-category {
  display: block;
  margin-top: 0.12rem;
  overflow: hidden;
  font-size: 0.62rem;
  color: var(--gallery-muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-trending-views {
  flex: none;
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--gallery-ink);
}

.home-trending-views span {
  margin-left: 0.12rem;
  font-size: 0.55rem;
  font-weight: 400;
  color: var(--gallery-muted);
}

@media (min-width: 900px) {
  .home-trending {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    column-gap: 1.2rem;
  }
}
</style>
