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
const images = computed(() => collection.value?.items || []);

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
  <GalleryPublicPage>
    <header
      v-if="collection"
      class="gallery-collection-hero mb-[clamp(2rem,4vw,3.5rem)] grid items-center gap-6 md:grid-cols-[minmax(0,.9fr)_minmax(20rem,1.1fr)] md:gap-[clamp(1.5rem,4vw,4rem)]"
    >
      <div class="gallery-collection-hero-copy py-4">
        <NuxtLink
          to="/collections"
          class="mb-5 inline-flex items-center gap-1.5 text-sm text-muted transition-colors hover:text-highlighted"
        >
          <UIcon name="i-tabler-arrow-left" class="size-4" />
          全部专题
        </NuxtLink>
        <h1
          class="gallery-page-title font-display text-[1.75rem] font-bold leading-[1.15] tracking-[-0.04em] text-highlighted md:text-[length:var(--gallery-title-page)]"
        >
          {{ collection.name }}
        </h1>
        <p
          v-if="collection.description"
          class="gallery-page-copy mt-1 max-w-[38rem] text-base leading-[1.7] text-muted"
        >
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
        class="gallery-collection-hero-cover row-start-1 grid aspect-[16/10] place-items-center overflow-hidden rounded-[0.9rem] bg-muted md:col-start-2 md:row-auto"
        :style="{ backgroundColor: collection.coverColor || undefined }"
      >
        <img
          v-if="collection.coverAssetId"
          v-bind="galleryImageSources(collection.coverAssetId, 'preview', true)"
          :alt="collection.coverAltText || collection.name"
          :width="collection.coverWidth"
          :height="collection.coverHeight"
          class="size-full object-cover"
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
    <template v-else-if="collection">
      <GalleryCompactEmpty
        icon="i-tabler-stack-2"
        title="专题正在整理"
        description="这个专题还没有图片。"
      />
    </template>

    <nav
      v-if="collection && galleryPageCount(collection) > 1"
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
        >{{ page }} / {{ galleryPageCount(collection) }}</span
      >
      <UButton
        color="neutral"
        variant="outline"
        trailing-icon="i-tabler-arrow-right"
        label="下一页"
        :disabled="page >= galleryPageCount(collection)"
        @click="setPage(page + 1)"
      />
    </nav>
  </GalleryPublicPage>
</template>
