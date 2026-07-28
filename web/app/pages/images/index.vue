<script setup lang="ts">
import type {
  GalleryDiscovery,
  GalleryImagePage,
  GallerySearchSuggestion,
} from "~/types/gallery";
import type { GalleryCatalogSort, GalleryCatalogView } from "~/utils/catalog";

const mobileFiltersOpen = ref(false);
const {
  state: catalogState,
  request: query,
  searchDraft,
  selectedCategories,
  selectedFacets,
  hasFilters,
  apply,
  clear,
  toggleCategory,
  toggleFacet,
  setPage,
  setSort,
  setView,
  setPreview,
  setTag,
  removeSearch,
  removeCategory,
  removeFacet,
  removeTag,
} = useGalleryCatalogState();
const page = computed(() => catalogState.value.page);
const sort = computed(() => catalogState.value.sort);
const view = computed(() => catalogState.value.view);
const preview = computed(() => catalogState.value.preview);
const previewTrigger = shallowRef<HTMLElement>();
const previewScrollY = ref(0);

const [{ data: pageData, error, status, refresh }, { data: discovery }] =
  await Promise.all([
    useFetch<GalleryImagePage>("/api/gallery/images", {
      query,
      watch: [query],
    }),
    useFetch<GalleryDiscovery>("/api/gallery/discovery", {
      query: { seed: "catalog-facets" },
    }),
  ]);

const facetGroups = computed(() => {
  return (pageData.value?.facets || discovery.value?.facets || []).map(
    (facet) => ({
      facet,
      values: facet.values,
    }),
  );
});
const categoryCandidates = computed(
  () => pageData.value?.categories || discovery.value?.categories || [],
);
const pageNumbers = computed(() => {
  const total = pageData.value?.totalPages || 0;
  if (!total) return [];
  const start = Math.max(1, Math.min(page.value - 2, total - 4));
  return Array.from(
    { length: Math.min(5, total) },
    (_, index) => start + index,
  );
});
const sortItems = [
  { label: "最新", value: "newest" },
  { label: "最早", value: "oldest" },
  { label: "标题 A-Z", value: "title_asc" },
  { label: "标题 Z-A", value: "title_desc" },
];
const activeRefinements = computed(() => {
  const items: Array<{
    key: string;
    label: string;
    kind: "search" | "category" | "facet" | "tag";
    value: string;
  }> = [];
  if (catalogState.value.q) {
    items.push({
      key: "search",
      label: `搜索：${catalogState.value.q}`,
      kind: "search",
      value: catalogState.value.q,
    });
  }
  for (const slug of catalogState.value.categories) {
    const category = categoryCandidates.value.find(
      (item) => item.slug === slug,
    );
    items.push({
      key: `category:${slug}`,
      label: category?.name || slug,
      kind: "category",
      value: slug,
    });
  }
  for (const selection of catalogState.value.facets) {
    const [facetSlug, valueSlug] = selection.split(":");
    const group = facetGroups.value.find(
      (item) => item.facet.slug === facetSlug,
    );
    const candidate = group?.values.find((item) => item.slug === valueSlug);
    items.push({
      key: `facet:${selection}`,
      label: group
        ? `${group.facet.name}：${candidate?.name || valueSlug}`
        : selection,
      kind: "facet",
      value: selection,
    });
  }
  if (catalogState.value.tag) {
    items.push({
      key: `tag:${catalogState.value.tag}`,
      label: `标签：${catalogState.value.tag}`,
      kind: "tag",
      value: catalogState.value.tag,
    });
  }
  return items;
});
const searchSuggestions = computed<GallerySearchSuggestion[]>(() => {
  const rawTerm = searchDraft.value.trim();
  const tagTerm = rawTerm.startsWith("#") ? rawTerm.slice(1).trim() : rawTerm;
  const term = tagTerm.toLocaleLowerCase();
  if (!term) return [];
  const suggestions: GallerySearchSuggestion[] = [
    {
      key: `tag:${tagTerm}`,
      label: `#${tagTerm}`,
      context: "按标签精确搜索",
    },
  ];
  for (const category of categoryCandidates.value) {
    if (
      `${category.name} ${category.slug}`.toLocaleLowerCase().includes(term)
    ) {
      suggestions.push({
        key: `category:${category.slug}`,
        label: category.name,
        context: `分类 · ${category.count}`,
      });
    }
  }
  for (const group of facetGroups.value) {
    for (const value of group.values) {
      if (`${value.name} ${value.slug}`.toLocaleLowerCase().includes(term)) {
        suggestions.push({
          key: `facet:${group.facet.slug}:${value.slug}`,
          label: value.name,
          context: `${group.facet.name} · ${value.count}`,
        });
      }
    }
  }
  return suggestions.slice(0, 7);
});

function openMobileFilters() {
  mobileFiltersOpen.value = true;
}
function closeMobileFilters() {
  mobileFiltersOpen.value = false;
}
function applyFilters() {
  void apply();
  mobileFiltersOpen.value = false;
}
function clearFilters() {
  void clear();
}
function changeSort(value: unknown) {
  void setSort(String(value) as GalleryCatalogSort);
}
function changeView(value: GalleryCatalogView) {
  void setView(value);
}
function openPreview(imageId: string, trigger: HTMLElement | null) {
  previewTrigger.value = trigger || undefined;
  previewScrollY.value = import.meta.client ? window.scrollY : 0;
  void setPreview(imageId);
}
function closePreview() {
  void setPreview("");
}
function navigatePreview(imageId: string) {
  void setPreview(imageId);
}
function removeRefinement(item: (typeof activeRefinements.value)[number]) {
  if (item.kind === "search") void removeSearch();
  if (item.kind === "category") void removeCategory(item.value);
  if (item.kind === "facet") void removeFacet(item.value);
  if (item.kind === "tag") void removeTag();
}
function selectSuggestion(key: string) {
  if (key.startsWith("tag:")) {
    void setTag(key.slice("tag:".length));
    return;
  }
  if (key.startsWith("category:")) {
    const slug = key.slice("category:".length);
    if (!selectedCategories.value.includes(slug)) toggleCategory(slug);
  }
  if (key.startsWith("facet:")) {
    const value = key.slice("facet:".length);
    if (!selectedFacets.value.includes(value)) toggleFacet(value);
  }
  void apply();
}

watch(preview, async (current, previous) => {
  if (current || !previous || !import.meta.client) return;
  await nextTick();
  window.scrollTo({ top: previewScrollY.value, behavior: "auto" });
  previewTrigger.value?.focus({ preventScroll: true });
  previewTrigger.value = undefined;
});

useSeoMeta({
  title: "浏览图片",
  description: "按分类、标签和多个维度分页浏览公开图片。",
  robots: () =>
    hasFilters.value ||
    page.value > 1 ||
    sort.value !== "newest" ||
    preview.value
      ? "noindex,follow"
      : "index,follow",
});
</script>

<template>
  <div class="gallery-page">
    <header class="gallery-page-header">
      <div>
        <h1 class="gallery-page-title">浏览图片</h1>
        <p class="gallery-page-copy">
          按分类和维度慢慢看，页码会记住你停下的位置。
        </p>
      </div>
      <p v-if="pageData" class="gallery-count">
        <span>{{ pageData.total }}</span> 张公开图片
      </p>
    </header>

    <div class="gallery-mobile-search lg:hidden">
      <UInput
        v-model="searchDraft"
        class="min-w-0 flex-1"
        icon="i-tabler-search"
        placeholder="搜索标题、说明或标签"
        @keyup.enter="applyFilters"
      />
      <UButton
        color="neutral"
        variant="outline"
        icon="i-tabler-adjustments-horizontal"
        label="筛选"
        @click="openMobileFilters"
      />
    </div>
    <GallerySearchSuggestions
      class="-mt-3 mb-5 lg:hidden"
      :items="searchSuggestions"
      @select="selectSuggestion"
    />

    <div class="grid gap-7 lg:grid-cols-[16.5rem_minmax(0,1fr)] lg:gap-10">
      <aside class="hidden lg:block" aria-label="图片过滤器">
        <div class="gallery-filter-panel sticky top-24 space-y-7">
          <div class="flex items-center justify-between gap-3">
            <h2 class="font-semibold text-highlighted">筛选图片</h2>
            <UBadge
              v-if="activeRefinements.length"
              color="primary"
              variant="soft"
              :label="`${activeRefinements.length} 项`"
            />
          </div>
          <UInput
            v-model="searchDraft"
            icon="i-tabler-search"
            placeholder="搜索标题、说明或标签"
            @keyup.enter="applyFilters"
          />
          <GallerySearchSuggestions
            :items="searchSuggestions"
            @select="selectSuggestion"
          />
          <GalleryCatalogFilterFields
            :categories="categoryCandidates"
            :facets="facetGroups.map((group) => group.facet)"
            :selected-categories="selectedCategories"
            :selected-facets="selectedFacets"
            @toggle-category="toggleCategory"
            @toggle-facet="toggleFacet"
          />
          <div class="flex gap-2">
            <UButton label="应用" size="sm" block @click="applyFilters" />
            <UButton
              v-if="hasFilters"
              color="neutral"
              variant="ghost"
              icon="i-tabler-x"
              aria-label="清除筛选"
              @click="clearFilters"
            />
          </div>
        </div>
      </aside>

      <section aria-live="polite">
        <div class="gallery-results-toolbar">
          <div class="flex min-w-0 flex-wrap gap-2">
            <span v-if="!hasFilters" class="px-1 text-sm text-muted">
              全部图片
            </span>
            <UButton
              v-for="item in activeRefinements"
              :key="item.key"
              class="gallery-refinement"
              color="neutral"
              variant="soft"
              size="xs"
              trailing-icon="i-tabler-x"
              :label="item.label"
              :aria-label="`移除筛选：${item.label}`"
              @click="removeRefinement(item)"
            />
            <UButton
              v-if="hasFilters"
              color="neutral"
              variant="ghost"
              size="xs"
              label="清除"
              @click="clearFilters"
            />
          </div>
          <div class="flex shrink-0 items-center gap-1">
            <USelect
              :model-value="sort"
              :items="sortItems"
              value-key="value"
              class="w-32 shrink-0"
              @update:model-value="changeSort"
            />
            <div class="gallery-view-switch" aria-label="图片布局">
              <UButton
                color="neutral"
                :variant="view === 'grid' ? 'soft' : 'ghost'"
                label="网格"
                aria-label="切换到网格布局"
                @click="changeView('grid')"
              />
              <UButton
                color="neutral"
                :variant="view === 'masonry' ? 'soft' : 'ghost'"
                label="瀑布"
                aria-label="切换到瀑布流布局"
                @click="changeView('masonry')"
              />
            </div>
          </div>
        </div>
        <p v-if="catalogState.q && pageData" class="gallery-search-explanation">
          标题、说明、替代文本或标签中包含“{{ catalogState.q }}”的结果，共
          {{ pageData.total }} 张。
        </p>

        <div v-if="status === 'pending'" class="gallery-grid">
          <USkeleton
            v-for="index in 16"
            :key="index"
            class="aspect-[4/3] rounded-lg"
          />
        </div>
        <UAlert
          v-else-if="error"
          color="error"
          variant="subtle"
          icon="i-tabler-alert-circle"
          title="图片列表加载失败"
          ><template #actions
            ><UButton label="重试" @click="refresh()" /></template
        ></UAlert>
        <GalleryMasonry
          v-else-if="pageData?.items.length && view === 'masonry'"
          :items="pageData.items"
          :priority="page === 1"
          quick-view
          @preview="openPreview"
        />
        <GalleryImageGrid
          v-else-if="pageData?.items.length"
          :items="pageData.items"
          :priority="page === 1"
          quick-view
          @preview="openPreview"
        />
        <div v-else class="gallery-compact-empty">
          <span class="gallery-empty-icon"
            ><UIcon name="i-tabler-filter-off" class="size-6"
          /></span>
          <h2 class="mt-4 text-lg font-semibold text-highlighted">
            没有符合条件的图片
          </h2>
          <p class="mx-auto mt-2 max-w-sm text-sm leading-6 text-muted">
            换一组筛选条件，或者从完整目录重新开始。
          </p>
          <UButton
            v-if="hasFilters"
            class="mt-4"
            color="neutral"
            variant="outline"
            label="清除筛选"
            @click="clearFilters"
          />
        </div>

        <nav
          v-if="pageData?.totalPages && pageData.totalPages > 1"
          class="mt-10 flex items-center justify-center gap-1"
          aria-label="图片分页"
        >
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-tabler-chevron-left"
            aria-label="上一页"
            :disabled="page <= 1"
            @click="setPage(page - 1)"
          />
          <UButton
            v-for="number in pageNumbers"
            :key="number"
            color="neutral"
            :variant="number === page ? 'solid' : 'ghost'"
            :label="String(number)"
            :aria-current="number === page ? 'page' : undefined"
            @click="setPage(number)"
          />
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-tabler-chevron-right"
            aria-label="下一页"
            :disabled="page >= pageData.totalPages"
            @click="setPage(page + 1)"
          />
        </nav>
      </section>
    </div>

    <UDrawer
      v-model:open="mobileFiltersOpen"
      direction="bottom"
      :ui="{
        content: 'max-h-[88dvh] rounded-t-2xl',
        body: 'min-h-0 overflow-y-auto p-0 sm:p-0',
        footer:
          'border-t border-default bg-default p-4 pb-[max(1rem,env(safe-area-inset-bottom))] sm:p-4',
      }"
    >
      <template #header>
        <div class="flex min-w-0 flex-1 items-center justify-between gap-4">
          <div class="min-w-0">
            <h2 class="font-semibold text-highlighted">筛选图片</h2>
            <p class="mt-0.5 text-xs text-muted">分类与维度可组合选择</p>
          </div>
          <div class="flex items-center gap-2">
            <UBadge
              v-if="selectedCategories.length + selectedFacets.length"
              color="primary"
              variant="soft"
              :label="`${selectedCategories.length + selectedFacets.length} 项`"
            />
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-tabler-x"
              aria-label="关闭筛选"
              @click="closeMobileFilters"
            />
          </div>
        </div>
      </template>
      <template #body>
        <div class="px-4 py-5">
          <GalleryCatalogFilterFields
            :categories="categoryCandidates"
            :facets="facetGroups.map((group) => group.facet)"
            :selected-categories="selectedCategories"
            :selected-facets="selectedFacets"
            @toggle-category="toggleCategory"
            @toggle-facet="toggleFacet"
          />
        </div>
      </template>
      <template #footer>
        <div class="flex w-full items-center gap-2">
          <UButton
            v-if="hasFilters"
            color="neutral"
            variant="ghost"
            label="重置"
            @click="clearFilters"
          />
          <UButton
            class="flex-1"
            :label="pageData ? `查看 ${pageData.total} 张图片` : '应用筛选'"
            @click="applyFilters"
          />
        </div>
      </template>
    </UDrawer>

    <GalleryQuickView
      :image-id="preview"
      :items="pageData?.items || []"
      @close="closePreview"
      @navigate="navigatePreview"
    />
  </div>
</template>
