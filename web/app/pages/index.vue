<script setup lang="ts">
import type {
  GalleryCollection,
  GalleryDiscovery,
  GalleryHomeSection,
  GalleryHomeSectionKey,
  GalleryImagePage,
  GalleryRanking,
} from "~/types/gallery";

const route = useRoute();
const router = useRouter();
const seed = computed(() => String(route.query.seed || ""));
const gallerySite = useGallerySiteSettings();
const { data, error, status, refresh } = await useFetch<GalleryDiscovery>(
  "/api/gallery/discovery",
  {
    query: computed(() => ({ seed: seed.value || undefined })),
    watch: [seed],
  },
);

watch(
  () => data.value?.site,
  (value) => {
    if (value) gallerySite.value = value;
  },
  { immediate: true },
);

const sections = computed(() =>
  [...(data.value?.site.homeSections || [])].sort(
    (left, right) => left.position - right.position,
  ),
);

function section(key: GalleryHomeSectionKey): GalleryHomeSection | undefined {
  return sections.value.find((item) => item.key === key);
}

const latestSize = computed(() => section("latest")?.itemLimit || 1);
const [{ data: collections }, { data: latest }, { data: ranking }] =
  await Promise.all([
    useFetch<{ items: GalleryCollection[] }>("/api/gallery/collections"),
    useFetch<GalleryImagePage>("/api/gallery/images", {
      query: computed(() => ({
        sort: "newest",
        page: 1,
        size: latestSize.value,
      })),
      watch: [latestSize],
    }),
    useFetch<{ ranking: GalleryRanking }>("/api/gallery/rankings", {
      query: { kind: "trending", window: "7d" },
    }),
  ]);

const featuredCollections = computed(() => {
  const limit = section("collections")?.itemLimit || 0;
  return collections.value?.items.slice(0, limit) || [];
});
const trendingImages = computed(() => {
  const limit = section("trending")?.itemLimit || 0;
  return ranking.value?.ranking.images.slice(0, limit) || [];
});

useSeoMeta({
  title: () => data.value?.site.title || data.value?.site.name,
  description: () => data.value?.site.description,
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
  FIRST VIEWPORT: 紧凑顶部搜索、运营配置的发现控制行，随后是无干扰的连续图片流。
  FORM: 图库标准答案，纯图片优先构图，Pinterest 式发现叠加 Pexels 式搜索，方案 B 的精简版，seed b0a2f453。
  FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md
  -->
  <GalleryPublicPage class="gallery-home">
    <h1 id="gallery-home-title" class="sr-only">
      {{ data?.site.title || data?.site.name }}
    </h1>

    <template v-for="homeSection in sections" :key="homeSection.key">
      <template v-if="homeSection.enabled && homeSection.key === 'random'">
        <section
          class="gallery-discovery-tools flex min-h-13 items-center justify-between gap-3.5 overflow-x-auto border-b border-default [scrollbar-width:none]"
          :aria-labelledby="`${homeSection.key}-title`"
        >
          <div class="gallery-random-heading flex items-baseline gap-2.5">
            <div>
              <h2
                :id="`${homeSection.key}-title`"
                class="text-[0.95rem] font-bold tracking-[-0.02em] text-highlighted"
              >
                {{ homeSection.title }}
              </h2>
            </div>
          </div>
          <UButton
            class="gallery-next-batch shrink-0"
            color="neutral"
            variant="ghost"
            icon="i-tabler-refresh"
            :label="homeSection.actionLabel"
            :loading="status === 'pending'"
            @click="nextBatch"
          />
        </section>

        <section
          class="gallery-home-discovery pt-4"
          :aria-labelledby="`${homeSection.key}-title`"
        >
          <div
            v-if="status === 'pending'"
            class="gallery-masonry columns-2 gap-3 pt-0 sm:columns-3 sm:gap-[0.8rem] lg:columns-4 lg:gap-4 min-[90rem]:columns-5"
            aria-label="正在加载随机图片"
          >
            <USkeleton
              v-for="index in homeSection.itemLimit"
              :key="index"
              class="gallery-stream-skeleton mb-3 break-inside-avoid rounded-[0.9rem]"
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
            :items="data.images.slice(0, homeSection.itemLimit)"
            priority
            class="gallery-home-stream pt-0"
          />

          <div
            v-else
            class="gallery-empty gallery-home-empty grid grid-cols-1 items-center gap-8 rounded-2xl border border-default bg-[color-mix(in_srgb,var(--gallery-panel)_92%,transparent)] p-5 shadow-[0_24px_80px_rgb(29_47_81_/_0.07)] md:min-h-88 md:grid-cols-[minmax(0,.9fr)_minmax(18rem,1.1fr)] md:gap-[clamp(2rem,6vw,7rem)] md:p-[clamp(1.5rem,5vw,4.5rem)]"
          >
            <div
              class="gallery-empty-visual grid aspect-[4/3] max-h-48 -rotate-2 grid-cols-[1.2fr_.8fr] grid-rows-2 gap-2.5 md:max-h-none"
              aria-hidden="true"
            >
              <span
                class="row-span-2 block rounded-2xl bg-[linear-gradient(145deg,var(--gallery-accent),#87a8ff)]"
              /><span
                class="block rounded-2xl bg-[color-mix(in_srgb,var(--gallery-ink)_12%,var(--gallery-panel))]"
              /><span
                class="block rounded-2xl bg-[color-mix(in_srgb,var(--gallery-accent)_20%,var(--gallery-panel))]"
              />
            </div>
            <div class="max-w-md">
              <p class="text-sm font-medium text-primary">
                第一批收藏，从你开始
              </p>
              <h2
                class="mt-3 text-2xl font-semibold tracking-tight text-highlighted sm:text-3xl"
              >
                这里还没有图片，但已经准备好展示它们
              </h2>
              <p class="mt-3 text-sm leading-6 text-muted">
                图片通过处理和审核后，会进入随机流和分页目录。
              </p>
              <div class="mt-6 flex flex-wrap gap-2">
                <UButton
                  to="/submit"
                  icon="i-tabler-photo-up"
                  label="投稿图片"
                />
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
      </template>

      <section
        v-else-if="
          homeSection.enabled &&
          homeSection.key === 'collections' &&
          featuredCollections.length
        "
        class="gallery-section pt-[clamp(2rem,4vw,3.75rem)]"
      >
        <GallerySectionHeader
          :title="homeSection.title"
          to="/collections"
          :action-label="homeSection.actionLabel"
        />
        <div class="gallery-editorial-grid grid border-t border-default">
          <NuxtLink
            v-for="(collection, index) in featuredCollections"
            :key="collection.id"
            :to="`/collections/${collection.slug}`"
            class="gallery-editorial-item grid grid-cols-[2rem_minmax(0,1fr)] items-start gap-4 border-b border-default px-1 py-[1.35rem] transition-[background,padding] hover:bg-muted hover:px-3 focus-visible:bg-muted focus-visible:px-3 md:grid-cols-[2.5rem_minmax(0,1fr)_auto]"
          >
            <span
              class="gallery-editorial-index font-display text-xs font-semibold text-primary"
              >{{ String(index + 1).padStart(2, "0") }}</span
            >
            <div>
              <h3 class="font-semibold text-highlighted">
                {{ collection.name }}
              </h3>
              <p
                v-if="collection.description"
                class="mt-1.5 max-w-[46rem] text-[0.85rem] leading-[1.55] text-muted"
              >
                {{ collection.description }}
              </p>
            </div>
            <span
              class="col-start-2 text-xs tabular-nums text-dimmed md:col-auto"
              >{{ collection.itemCount }} 张</span
            >
          </NuxtLink>
        </div>
      </section>

      <section
        v-else-if="
          homeSection.enabled &&
          homeSection.key === 'latest' &&
          latest?.items.length
        "
        class="gallery-section pt-[clamp(2rem,4vw,3.75rem)]"
      >
        <GallerySectionHeader
          :title="homeSection.title"
          to="/images"
          :action-label="homeSection.actionLabel"
        />
        <GalleryImageGrid
          :items="latest.items.slice(0, homeSection.itemLimit)"
        />
      </section>

      <section
        v-else-if="
          homeSection.enabled &&
          homeSection.key === 'trending' &&
          trendingImages.length
        "
        class="gallery-section pt-[clamp(2rem,4vw,3.75rem)]"
      >
        <GallerySectionHeader
          :title="homeSection.title"
          to="/rankings?kind=trending&window=7d"
          :action-label="homeSection.actionLabel"
        />
        <GalleryImageGrid :items="trendingImages" />
      </section>
    </template>
  </GalleryPublicPage>
</template>
