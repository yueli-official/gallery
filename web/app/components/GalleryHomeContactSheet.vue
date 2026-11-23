<script setup lang="ts">
import type { GalleryImageCard } from "~/types/gallery";

const props = defineProps<{
  items: GalleryImageCard[];
  priority?: boolean;
}>();

const failedAssets = ref<Set<string>>(new Set());
const visibleItems = computed(() =>
  props.items.filter((item) => !failedAssets.value.has(item.assetId)),
);
const grid = ref<HTMLElement | null>(null);

function markFailed(assetId: string) {
  failedAssets.value = new Set([...failedAssets.value, assetId]);
}

onMounted(() => {
  for (const image of grid.value?.querySelectorAll<HTMLImageElement>(
    "img[data-gallery-asset]",
  ) || []) {
    if (image.complete && image.naturalWidth === 0 && image.dataset.galleryAsset)
      markFailed(image.dataset.galleryAsset);
  }
});

function sourcePolicy(assetId: string, index: number) {
  const priority = Boolean(props.priority && index < 6);
  return {
    src: galleryRendition(assetId, "grid-lg"),
    srcset: `${galleryRendition(assetId, "grid-sm")} 480w, ${galleryRendition(assetId, "grid-lg")} 960w`,
    sizes:
      "(max-width: 639px) 50vw, (max-width: 1023px) 33vw, (max-width: 1279px) 25vw, 17vw",
    loading: priority ? "eager" : "lazy",
    fetchpriority: priority ? "high" : "auto",
  } as const;
}
</script>

<template>
  <div ref="grid" class="home-contact-sheet">
    <article
      v-for="(image, index) in visibleItems"
      :key="image.id"
      class="home-contact-item group relative min-w-0"
    >
      <NuxtLink
        :to="`/images/${encodeURIComponent(image.publicId)}`"
        class="home-contact-link block focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--gallery-accent)]"
      >
        <div
          class="home-contact-media relative aspect-[3/4] overflow-hidden rounded-[0.9rem]"
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
          <span
            v-else
            class="absolute inset-0 flex items-center justify-center bg-[var(--gallery-panel)] text-[0.6rem] text-[var(--gallery-muted)]"
            role="img"
            :aria-label="`${image.title}暂不可用`"
            >图片暂不可用</span
          >
          <span
            v-if="image.primaryCategory"
            class="home-contact-tag"
            >{{ image.primaryCategory }}</span
          >
        </div>
        <div class="home-contact-copy">
          <div class="home-contact-text min-w-0">
            <h3 class="home-contact-title">{{ image.title }}</h3>
            <p class="home-contact-sub font-numeric">
              ♥ {{ compactMetric(image.metrics.favorites) }}
            </p>
          </div>
          <span class="home-contact-arrow" aria-hidden="true">
            <svg viewBox="0 0 24 24">
              <path
                d="M7 17 17 7M8 7h9v9"
                fill="none"
                stroke="currentColor"
                stroke-width="1.7"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </span>
        </div>
      </NuxtLink>
      <GalleryFavoriteButton :image-id="image.id" :title="image.title" />
    </article>
  </div>
</template>

<style scoped>
.home-contact-sheet {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 0.65rem;
  row-gap: 1.35rem;
}

.home-contact-item {
  transition: transform 220ms ease;
}

.home-contact-item:hover {
  transform: translateY(-4px);
}

.home-contact-media {
  background: var(--gallery-panel);
  box-shadow: 0 1px 2px rgb(11 12 18 / 0.04);
  transition: box-shadow 220ms ease;
}

.home-contact-item:hover .home-contact-media {
  box-shadow: 0 18px 34px -18px rgb(11 12 18 / 0.35);
}

.home-contact-tag {
  position: absolute;
  bottom: 0.45rem;
  left: 0.45rem;
  padding: 0.12rem 0.4rem;
  background: rgb(255 255 255 / 0.9);
  font-size: 0.56rem;
  font-weight: 600;
  line-height: 1.4;
  color: #0b0c12;
  transition:
    background 180ms ease,
    color 180ms ease;
}

.home-contact-item:hover .home-contact-tag {
  background: var(--gallery-accent);
  color: var(--gallery-on-accent);
}

.home-contact-copy {
  display: flex;
  align-items: flex-end;
  gap: 0.4rem;
  padding: 0.5rem 0.05rem 0;
}

.home-contact-text {
  min-width: 0;
  flex: 1;
}

.home-contact-title {
  overflow: hidden;
  font-size: 0.77rem;
  font-weight: 600;
  line-height: 1.35;
  color: var(--gallery-ink);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-contact-sub {
  margin-top: 0.12rem;
  font-size: 0.6rem;
  color: var(--gallery-muted);
}

.home-contact-arrow {
  display: grid;
  flex: none;
  place-items: center;
  width: 1.35rem;
  height: 1.35rem;
  border-radius: 999px;
  background: var(--gallery-panel);
  color: var(--gallery-muted);
  opacity: 0;
  transform: translateX(-3px);
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

.home-contact-arrow svg {
  width: 0.7rem;
  height: 0.7rem;
}

.home-contact-item:hover .home-contact-arrow {
  opacity: 1;
  transform: translateX(0);
}

.home-contact-item :deep(.gallery-favorite-button) {
  top: 0.45rem;
  right: 0.45rem;
  width: 1.9rem;
  height: 1.9rem;
  border: 0;
  border-radius: 999px;
  background: rgb(255 255 255 / 0.86);
  color: #0b0c12;
  backdrop-filter: blur(8px);
}

.home-contact-item :deep(.gallery-favorite-button[class*="border-primary"]) {
  background: var(--gallery-accent);
  color: var(--gallery-on-accent);
}

@media (min-width: 640px) {
  .home-contact-sheet {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    column-gap: 0.75rem;
    row-gap: 1.5rem;
  }

  .home-contact-title {
    font-size: 0.8rem;
  }
}

@media (min-width: 1024px) {
  .home-contact-sheet {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (min-width: 1280px) {
  .home-contact-sheet {
    grid-template-columns: repeat(6, minmax(0, 1fr));
    column-gap: 0.65rem;
    row-gap: 1.35rem;
  }

  .home-contact-title {
    font-size: 0.72rem;
  }
}
</style>
