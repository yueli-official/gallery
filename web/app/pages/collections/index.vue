<script setup lang="ts">
import type { GalleryCollection } from "~/types/gallery";

const { data, error, status, refresh } = await useFetch<{
  items: GalleryCollection[];
}>("/api/gallery/collections");
useGalleryMediaRefresh(data, refresh);
useSeoMeta({ title: "专题集合", description: "由运营方整理的公开图片专题。" });
</script>

<template>
  <GalleryPublicPage class="collections-page">
    <GalleryArchiveHero
      kicker="月离图库 / COLLECTION ARCHIVE"
      title="专题集合"
      description="围绕一个主题重新整理图片，让浏览更有方向。"
      :count="data?.items.length"
      count-label="个专题"
    />

    <div v-if="status === 'pending'" class="collections-grid">
      <USkeleton v-for="index in 6" :key="index" class="aspect-[4/5] rounded-lg" />
    </div>
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="专题加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <div v-else-if="data?.items.length" class="collections-grid">
      <NuxtLink
        v-for="(collection, index) in data.items"
        :key="collection.id"
        :to="`/collections/${collection.slug}`"
        class="collection-card group relative block min-w-0 overflow-hidden focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--gallery-accent)]"
        :style="{ backgroundColor: collection.coverColor || undefined }"
      >
        <img
          v-if="collection.coverAssetId"
          v-bind="galleryImageSources(collection, 'grid', false)"
          :alt="collection.coverAltText || collection.name"
          :width="collection.coverWidth"
          :height="collection.coverHeight"
          class="absolute inset-0 size-full object-cover transition-transform duration-700 ease-out group-hover:scale-[1.04]"
        />
        <span v-else class="absolute inset-0 grid place-items-center text-[var(--gallery-accent)]">
          <UIcon name="i-tabler-folders" class="size-10" />
        </span>
        <span class="collection-shade" aria-hidden="true" />
        <span class="collection-no font-numeric"><UIcon name="i-tabler-folders" />{{ String(index + 1).padStart(2, "0") }}</span>
        <span class="collection-count font-numeric">{{ collection.itemCount }} 张</span>
        <div class="collection-copy">
          <h2 class="collection-title">{{ collection.name }}</h2>
          <p v-if="collection.description" class="collection-desc">
            {{ collection.description }}
          </p>
          <span class="collection-link">
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
    <GalleryCompactEmpty
      v-else
      icon="i-tabler-folders"
      title="还没有公开专题"
      description="运营整理的主题集合会出现在这里。"
    />
  </GalleryPublicPage>
</template>

<style scoped>
.collections-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 18rem), 1fr));
  gap: 1rem;
}

.collection-card {
  aspect-ratio: 4 / 5;
  border: 1px solid var(--gallery-line);
  border-radius: .6rem;
  transition:
    transform 240ms ease,
    box-shadow 240ms ease;
}

.collection-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 24px 44px -24px rgb(11 12 18 / 0.5);
}

.collection-card img {
  background: var(--gallery-panel);
}

.collection-shade {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(180deg, rgb(6 7 12 / 0.06) 0%, rgb(6 7 12 / 0) 38%),
    linear-gradient(0deg, rgb(6 7 12 / 0.86) 0%, rgb(6 7 12 / 0.22) 54%, rgb(6 7 12 / 0) 78%);
}

.collection-no {
  display: inline-flex;
  align-items: center;
  gap: .4rem;
  position: absolute;
  top: 0.8rem;
  left: 0.8rem;
  padding: .3rem .5rem;
  border: 1px solid rgb(255 255 255 / .2);
  border-radius: .3rem;
  background: rgb(8 9 14 / .4);
  backdrop-filter: blur(8px);
  font-size: .6rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  color: #fff;
}

.collection-count {
  position: absolute;
  top: 0.8rem;
  right: 0.8rem;
  padding: 0.2rem 0.55rem;
  border-radius: .3rem;
  background: rgb(8 9 14 / 0.34);
  font-size: 0.6rem;
  color: rgb(255 255 255 / 0.9);
  backdrop-filter: blur(8px);
}

.collection-copy {
  position: absolute;
  right: 1.1rem;
  bottom: 1rem;
  left: 1.1rem;
  color: #fff;
}

.collection-title {
  font-size: 1.2rem;
  font-weight: 650;
  line-height: 1.4;
  letter-spacing: -.025em;
}

.collection-desc {
  display: -webkit-box;
  margin-top: 0.3rem;
  overflow: hidden;
  font-size: 0.7rem;
  line-height: 1.55;
  color: rgb(255 255 255 / 0.68);
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.collection-link {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  margin-top: 0.55rem;
  font-size: 0.66rem;
  font-weight: 500;
  color: rgb(255 255 255 / 0.86);
  transition: gap 180ms ease;
}

.collection-link svg {
  width: 0.78rem;
  height: 0.78rem;
}

.collection-card:hover .collection-link {
  gap: 0.45rem;
}

</style>
