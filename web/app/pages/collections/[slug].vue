<script setup lang="ts">
import type { GalleryCollectionDetail } from "~/types/gallery";

const route = useRoute("/collections/[slug]");
const router = useRouter();
const page = computed(() => Math.max(1, Number(route.query.page) || 1));
const { data, error, status, refresh } = await useFetch<{
  collection: GalleryCollectionDetail;
}>(() => `/api/gallery/collections/${route.params.slug}`, {
  query: computed(() => ({ page: page.value, size: 24 })),
  watch: [page],
});
if (import.meta.server && error.value) setResponseStatus(404);
const collection = computed(() => data.value?.collection);
const images = computed(() => collection.value?.images || []);

useSeoMeta({
  title: () => collection.value?.seoTitle || collection.value?.name || "专题",
  description: () =>
    collection.value?.seoDescription ||
    collection.value?.description ||
    "公开图片专题。",
  robots: () => (error.value ? "noindex, nofollow" : "index, follow"),
  ogImage: () =>
    collection.value?.coverAssetId
      ? galleryRendition(collection.value.coverAssetId, "og")
      : undefined,
});

function setPage(value: number): void {
  const query = { ...route.query };
  if (value <= 1) delete query.page;
  else query.page = String(value);
  void router.push({ query });
}
</script>

<template>
  <div class="gallery-page">
    <header v-if="collection" class="gallery-collection-hero">
      <div class="gallery-collection-hero-copy">
        <NuxtLink
          to="/collections"
          class="mb-5 inline-flex items-center gap-1.5 text-sm text-muted transition-colors hover:text-highlighted"
        >
          <UIcon name="i-tabler-arrow-left" class="size-4" />
          全部专题
        </NuxtLink>
        <h1 class="gallery-page-title">{{ collection.name }}</h1>
        <p v-if="collection.description" class="gallery-page-copy text-base">
          {{ collection.description }}
        </p>
        <p class="mt-5 text-sm text-muted">
          {{ collection.itemCount }} 张图片
          <span v-if="collection.updatedAt">
            ，更新于
            {{ new Date(collection.updatedAt).toLocaleDateString("zh-CN") }}
          </span>
        </p>
      </div>
      <div
        class="gallery-collection-hero-cover"
        :style="{ backgroundColor: collection.coverColor || undefined }"
      >
        <img
          v-if="collection.coverAssetId"
          v-bind="galleryImageSources(collection.coverAssetId, 'preview', true)"
          :alt="collection.coverAltText || collection.name"
          :width="collection.coverWidth"
          :height="collection.coverHeight"
        />
        <UIcon v-else name="i-tabler-folders" class="size-14 text-primary" />
      </div>
    </header>

    <div v-if="status === 'pending'" class="gallery-grid">
      <USkeleton
        v-for="index in 12"
        :key="index"
        class="aspect-[4/3] rounded-lg"
      />
    </div>
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="专题不存在或尚未公开"
    >
      <template #actions><UButton label="重试" @click="refresh()" /></template>
    </UAlert>
    <GalleryImageGrid
      v-else-if="images.length"
      :items="images"
      priority
    />
    <div
      v-else-if="collection"
      class="gallery-compact-empty grid place-items-center"
    >
      <div>
        <span class="gallery-empty-icon"
          ><UIcon name="i-tabler-stack-2" class="size-6"
        /></span>
        <h2 class="mt-4 font-semibold text-highlighted">专题正在整理</h2>
        <p class="mt-2 text-sm text-muted">这个专题还没有图片。</p>
      </div>
    </div>

    <nav
      v-if="collection && collection.totalPages > 1"
      class="mt-10 flex items-center justify-center gap-3"
      aria-label="专题分页"
    >
      <UButton
        color="neutral"
        variant="outline"
        icon="i-tabler-arrow-left"
        label="上一页"
        :disabled="page <= 1"
        @click="setPage(page - 1)"
      />
      <span class="text-sm tabular-nums text-muted"
        >{{ page }} / {{ collection.totalPages }}</span
      >
      <UButton
        color="neutral"
        variant="outline"
        trailing-icon="i-tabler-arrow-right"
        label="下一页"
        :disabled="page >= collection.totalPages"
        @click="setPage(page + 1)"
      />
    </nav>
  </div>
</template>
