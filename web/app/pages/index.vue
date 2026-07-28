<script setup lang="ts">
import type {
  GalleryCollection,
  GalleryDiscovery,
  GalleryImagePage,
  GalleryRanking,
} from "~/types/gallery";

const route = useRoute();
const router = useRouter();
const isSyntheticPreview = import.meta.dev;
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
  <!--
  THESIS: 图片优先的公共资料库；拒绝让非图片内容占据瀑布流位置。
  OWN-WORLD: 近白画布、矿物蓝索引线、无边框图片、切角筛选标签与 14px 图像圆角。
  STORY: 搜索与细化，连续发现，打开、收藏或进入专题。
  FIRST VIEWPORT: 紧凑顶部搜索、“随机看看 / 换一批”控制行，随后是无干扰的连续图片流。
  FORM: 图库标准答案，纯图片优先构图，Pinterest 式发现叠加 Pexels 式搜索，方案 B 的精简版，seed b0a2f453。
  FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md
  -->
  <div class="gallery-page gallery-home">
    <h1 id="gallery-home-title" class="sr-only">月离图库公共图片资料库</h1>
    <section class="gallery-discovery-tools" aria-labelledby="random-title">
      <div class="gallery-random-heading">
        <h2 id="random-title">随机看看</h2>
        <span v-if="isSyntheticPreview" class="gallery-demo-label">演示数据</span>
      </div>
      <UButton
        class="gallery-next-batch"
        color="neutral"
        variant="ghost"
        icon="i-tabler-refresh"
        label="换一批"
        :loading="status === 'pending'"
        @click="nextBatch"
      />
    </section>

    <section class="gallery-home-discovery" aria-labelledby="gallery-home-title">
      <div
        v-if="status === 'pending'"
        class="gallery-masonry gallery-home-stream"
        aria-label="正在加载随机图片"
      >
        <USkeleton
          v-for="index in 20"
          :key="index"
          class="gallery-stream-skeleton"
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
        class="gallery-home-stream"
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
        title="最新入库"
        description="最近完成处理和审核的公开图片。"
        to="/images"
      />
      <GalleryImageGrid :items="latest.items.slice(0, 8)" />
    </section>

    <section v-if="trendingImages.length" class="gallery-section">
      <GallerySectionHeader
        title="正在被发现"
        description="近期获得更多有效浏览的图片。"
        to="/rankings?kind=trending&window=7d"
        action-label="查看排行"
      />
      <GalleryImageGrid :items="trendingImages" />
    </section>
  </div>
</template>
