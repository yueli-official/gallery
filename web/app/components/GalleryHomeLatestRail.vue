<script setup lang="ts">
import type { GalleryImageCard } from "~/types/gallery";

const props = defineProps<{
  items: GalleryImageCard[];
}>();

const failedAssets = ref<Set<string>>(new Set());
const visibleItems = computed(() =>
  props.items.filter((item) => !failedAssets.value.has(item.assetId)),
);
const rail = ref<HTMLElement | null>(null);

function markFailed(assetId: string) {
  failedAssets.value = new Set([...failedAssets.value, assetId]);
}

onMounted(() => {
  for (const image of rail.value?.querySelectorAll<HTMLImageElement>(
    "img[data-gallery-asset]",
  ) || []) {
    if (image.complete && image.naturalWidth === 0 && image.dataset.galleryAsset)
      markFailed(image.dataset.galleryAsset);
  }
});

function sourcePolicy(assetId: string, index: number) {
  return {
    src: galleryRendition(assetId, "grid-lg"),
    srcset: `${galleryRendition(assetId, "grid-sm")} 480w, ${galleryRendition(assetId, "grid-lg")} 960w`,
    sizes: "(max-width: 639px) 44vw, 15vw",
    loading: index < 6 ? "eager" : "lazy",
    fetchpriority: index < 4 ? "high" : "auto",
  } as const;
}
</script>

<template>
  <div ref="rail" class="home-latest-rail" role="list">
    <article
      v-for="(image, index) in visibleItems"
      :key="image.id"
      class="home-latest-item group min-w-0"
      role="listitem"
    >
      <NuxtLink
        :to="`/images/${encodeURIComponent(image.publicId)}`"
        class="block focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--gallery-accent)]"
      >
        <div
          class="home-latest-media relative aspect-[3/4] overflow-hidden rounded-[0.9rem]"
          :style="{ backgroundColor: image.dominantColor || undefined }"
        >
          <img
            v-if="!failedAssets.has(image.assetId)"
            v-bind="sourcePolicy(image.assetId, index)"
            :alt="image.altText || image.title"
            :width="image.width || undefined"
            :height="image.height || undefined"
            :data-gallery-asset="image.assetId"
            class="size-full object-cover transition-transform duration-500 ease-out group-hover:scale-[1.045]"
            @error="markFailed(image.assetId)"
          />
          <span v-else class="absolute inset-0 grid place-items-center text-[0.6rem] text-[var(--gallery-muted)]"
            >图片暂不可用</span
          >
          <span v-if="index < 3" class="home-latest-new font-numeric">NEW</span>
        </div>
        <h3 class="home-latest-title">{{ image.title }}</h3>
        <p v-if="image.primaryCategory" class="home-latest-category">
          {{ image.primaryCategory }}
        </p>
      </NuxtLink>
    </article>
  </div>
</template>

<style scoped>
.home-latest-rail {
  display: grid;
  grid-auto-columns: clamp(8.5rem, 22vw, 11.5rem);
  grid-auto-flow: column;
  gap: 0.7rem;
  overflow-x: auto;
  padding-bottom: 0.5rem;
  scroll-snap-type: x mandatory;
  scrollbar-width: none;
  overscroll-behavior-inline: contain;
}

.home-latest-rail::-webkit-scrollbar {
  display: none;
}

.home-latest-item {
  scroll-snap-align: start;
}

.home-latest-media {
  background: var(--gallery-panel);
}

.home-latest-new {
  position: absolute;
  top: 0.45rem;
  left: 0.45rem;
  padding: 0.12rem 0.4rem;
  border-radius: 999px;
  background: var(--gallery-accent);
  font-size: 0.52rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  color: var(--gallery-on-accent);
}

.home-latest-title {
  margin-top: 0.5rem;
  overflow: hidden;
  font-size: 0.76rem;
  font-weight: 600;
  line-height: 1.35;
  color: var(--gallery-ink);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-latest-category {
  margin-top: 0.12rem;
  font-size: 0.62rem;
  color: var(--gallery-muted);
}

@media (min-width: 768px) {
  .home-latest-rail {
    grid-auto-columns: minmax(10rem, 1fr);
    gap: 0.8rem;
  }

  .home-latest-title {
    font-size: 0.8rem;
  }
}

@media (min-width: 1280px) {
  .home-latest-rail {
    grid-auto-columns: minmax(10.5rem, 1fr);
  }
}
</style>
