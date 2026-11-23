<script setup lang="ts">
import type { GalleryCollection } from "~/types/gallery";

defineProps<{
  items: GalleryCollection[];
}>();

function needle(index: number): string {
  return String(index + 1).padStart(2, "0");
}

const fallbackWashes = [
  "linear-gradient(135deg, #8fcc00 0%, #1a2200 100%)",
  "linear-gradient(135deg, #0b0c12 0%, #343a4a 100%)",
  "linear-gradient(135deg, #1b2a6b 0%, #0b0c12 100%)",
];
</script>

<template>
  <div class="home-collections">
    <NuxtLink
      v-for="(collection, index) in items"
      :key="collection.id"
      :to="`/collections/${collection.slug}`"
      class="home-collection group relative block aspect-[16/10] min-w-0 overflow-hidden rounded-[1.1rem] focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--gallery-accent)] sm:aspect-[4/5]"
      :style="{
        background: collection.coverColor || fallbackWashes[index % fallbackWashes.length],
      }"
    >
      <img
        v-if="collection.coverAssetId"
        v-bind="galleryImageSources(collection.coverAssetId, 'grid')"
        :alt="collection.coverAltText || collection.name"
        :width="collection.coverWidth"
        :height="collection.coverHeight"
        class="absolute inset-0 size-full object-cover transition-transform duration-700 ease-out group-hover:scale-[1.04]"
      />
      <span class="home-collection-shade" aria-hidden="true" />
      <span class="home-collection-no font-numeric" aria-hidden="true">
        {{ needle(index) }}
      </span>
      <span class="home-collection-count font-numeric">
        {{ collection.itemCount }} 张
      </span>
      <div class="home-collection-copy">
        <h3 class="home-collection-title">{{ collection.name }}</h3>
        <p v-if="collection.description" class="home-collection-desc">
          {{ collection.description }}
        </p>
        <span class="home-collection-link">
          查看专题
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path
              d="M4 12h15m-5-5 5 5-5 5"
              fill="none"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </span>
      </div>
    </NuxtLink>
  </div>
</template>

<style scoped>
.home-collections {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.85rem;
}

.home-collection {
  isolation: isolate;
  transition:
    transform 260ms ease,
    box-shadow 260ms ease;
}

.home-collection:hover {
  transform: translateY(-4px);
  box-shadow: 0 24px 44px -22px rgb(11 12 18 / 0.45);
}

.home-collection-shade {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(180deg, rgb(6 7 12 / 0.08) 0%, rgb(6 7 12 / 0) 38%),
    linear-gradient(0deg, rgb(6 7 12 / 0.82) 0%, rgb(6 7 12 / 0.2) 52%, rgb(6 7 12 / 0) 78%);
}

.home-collection-no {
  position: absolute;
  top: 0.8rem;
  left: 0.8rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 1.7rem;
  padding: 0 0.55rem;
  border-radius: 999px;
  background: rgb(255 255 255 / 0.92);
  font-size: 0.62rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  color: #0b0c12;
}

.home-collection-count {
  position: absolute;
  top: 0.8rem;
  right: 0.8rem;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  background: rgb(8 9 14 / 0.32);
  font-size: 0.6rem;
  color: rgb(255 255 255 / 0.9);
  backdrop-filter: blur(8px);
}

.home-collection-copy {
  position: absolute;
  right: 1.1rem;
  bottom: 1rem;
  left: 1.1rem;
  color: #fff;
}

.home-collection-title {
  font-size: clamp(1.3rem, 2.2vw, 1.7rem);
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -0.045em;
}

.home-collection-desc {
  display: -webkit-box;
  margin-top: 0.3rem;
  overflow: hidden;
  font-size: 0.7rem;
  line-height: 1.55;
  color: rgb(255 255 255 / 0.68);
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.home-collection-link {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  margin-top: 0.55rem;
  font-size: 0.66rem;
  font-weight: 500;
  color: rgb(255 255 255 / 0.86);
  transition: gap 180ms ease;
}

.home-collection-link svg {
  width: 0.78rem;
  height: 0.78rem;
}

.home-collection:hover .home-collection-link {
  gap: 0.45rem;
}

@media (min-width: 640px) {
  .home-collections {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0.85rem;
  }
}
</style>
