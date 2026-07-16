<script setup lang="ts">
import type {
  GalleryCollection,
  GalleryDiscovery,
  GalleryImagePage,
  GalleryRanking,
} from "~/types/gallery";

const route = useRoute();
const router = useRouter();
const seed = computed(() => String(route.query.seed || ""));
const [
  { data, error, status, refresh },
  { data: collections },
  { data: latest },
  { data: ranking },
] = await Promise.all([
  useFetch<GalleryDiscovery>("/api/gallery/discovery", {
    query: computed(() => ({ seed: seed.value || undefined })),
    watch: [seed],
  }),
  useFetch<{ collections: GalleryCollection[] }>("/api/gallery/collections"),
  useFetch<GalleryImagePage>("/api/gallery/images", {
    query: { sort: "newest", page: 1, size: 8 },
  }),
  useFetch<{ ranking: GalleryRanking }>("/api/gallery/rankings", {
    query: { kind: "trending", window: "7d" },
  }),
]);

const featuredCategories = computed(
  () => data.value?.categories.slice(0, 8) || [],
);
const featuredCollections = computed(
  () => collections.value?.collections.slice(0, 4) || [],
);
const trendingImages = computed(
  () => ranking.value?.ranking.images.slice(0, 5) || [],
);

useSeoMeta({
  title: () => data.value?.site.name || "月离图库",
  description: () =>
    data.value?.site.description || "随机发现、收藏和投稿公开图片。",
  robots: () => (seed.value ? "noindex,follow" : "index,follow"),
});

function nextBatch() {
  const next =
    globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random()}`;
  void router.replace({ path: "/", query: { seed: next } });
}
</script>

<template>
  <div class="gallery-page gallery-home">
    <section class="gallery-home-intro" aria-labelledby="gallery-home-title">
      <div class="gallery-home-intro-copy">
        <p class="gallery-eyebrow">公共图片资料库</p>
        <h1 id="gallery-home-title">寻找下一张值得使用的图片</h1>
        <p>
          从运营整理的分类、专题和多维标签中发现图片，也可以直接搜索标题与说明。
        </p>
      </div>
      <GalleryGlobalSearch />
      <nav
        v-if="featuredCategories.length"
        class="gallery-category-rail"
        aria-label="热门分类"
      >
        <NuxtLink
          v-for="category in featuredCategories"
          :key="category.id"
          :to="{ path: '/images', query: { categories: category.slug } }"
          class="gallery-category-link"
        >
          <span>{{ category.name }}</span>
          <span>{{ category.count }}</span>
        </NuxtLink>
      </nav>
    </section>

    <section class="gallery-section">
      <GallerySectionHeader
        eyebrow="随机发现"
        title="换一个角度浏览"
        description="不依赖热度排序，从不同主题中重新组合一批图片。"
      >
        <template #action>
          <UButton
            class="shrink-0"
            color="neutral"
            variant="outline"
            icon="i-tabler-refresh"
            label="换一批"
            :loading="status === 'pending'"
            @click="nextBatch"
          />
        </template>
      </GallerySectionHeader>

      <div
        v-if="status === 'pending'"
        class="gallery-masonry gallery-home-stream"
        aria-label="正在加载随机图片"
      >
        <USkeleton
          v-for="index in 20"
          :key="index"
          class="mb-3 h-64 break-inside-avoid rounded-lg"
          :style="{ height: `${180 + (index % 4) * 46}px` }"
        />
      </div>

      <UAlert
        v-else-if="error"
        color="error"
        variant="subtle"
        icon="i-tabler-alert-circle"
        title="随机图片暂时没有加载出来"
        description="请确认 Gallery API 与 Asset 服务可用。"
      >
        <template #actions
          ><UButton
            color="error"
            variant="soft"
            label="重试"
            @click="refresh()"
        /></template>
      </UAlert>

      <GalleryMasonry
        v-else-if="data?.images.length"
        :items="data.images"
        priority
      />

      <div v-else class="gallery-empty gallery-home-empty">
        <div class="gallery-empty-visual" aria-hidden="true">
          <span /><span /><span />
        </div>
        <div class="max-w-md">
          <p class="text-sm font-medium text-primary">第一批收藏，从你开始</p>
          <h2
            class="mt-3 text-2xl font-semibold tracking-tight text-highlighted sm:text-3xl"
          >
            这里还没有图片，但已经准备好展示它们
          </h2>
          <p class="mt-3 text-sm leading-6 text-muted">
            图片通过处理和审核后，会进入随机流和分页目录。
          </p>
          <div class="mt-6 flex flex-wrap gap-2">
            <UButton to="/submit" icon="i-tabler-photo-up" label="投稿图片" />
            <UButton
              to="/images"
              color="neutral"
              variant="outline"
              label="浏览目录"
            />
          </div>
        </div>
      </div>
    </section>

    <section v-if="featuredCollections.length" class="gallery-section">
      <GallerySectionHeader
        eyebrow="运营精选"
        title="从专题进入"
        description="沿着一个清晰主题，查看经过整理的图片集合。"
        to="/collections"
      />
      <div class="gallery-editorial-grid">
        <NuxtLink
          v-for="(collection, index) in featuredCollections"
          :key="collection.id"
          :to="`/collections/${collection.slug}`"
          class="gallery-editorial-item"
        >
          <span class="gallery-editorial-index">0{{ index + 1 }}</span>
          <div>
            <h3>{{ collection.name }}</h3>
            <p>{{ collection.description || "查看这个专题中的公开图片。" }}</p>
          </div>
          <span>{{ collection.itemCount }} 张</span>
        </NuxtLink>
      </div>
    </section>

    <section v-if="latest?.items.length" class="gallery-section">
      <GallerySectionHeader
        eyebrow="持续更新"
        title="最新入库"
        description="最近完成处理和审核的公开图片。"
        to="/images"
      />
      <GalleryImageGrid :items="latest.items.slice(0, 8)" />
    </section>

    <section v-if="trendingImages.length" class="gallery-section">
      <GallerySectionHeader
        eyebrow="近 7 天"
        title="正在被发现"
        description="近期获得更多有效浏览的图片。"
        to="/rankings?kind=trending&window=7d"
        action-label="查看排行"
      />
      <GalleryImageGrid :items="trendingImages" />
    </section>
  </div>
</template>
