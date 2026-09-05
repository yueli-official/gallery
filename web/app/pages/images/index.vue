<script setup lang="ts">
import type {
  GalleryDiscovery,
  GalleryImagePage,
  GallerySearchSuggestion,
} from "~/types/gallery";
import type { GalleryCatalogSort, GalleryCatalogView } from "~/utils/catalog";

const mobileFiltersOpen = ref(false);
let desktopFilterMedia: MediaQueryList | undefined;
const {
  state: catalogState,
  request: query,
  searchDraft,
  selectedCategories,
  selectedFacets,
  selectedTag,
  hasFilters,
  apply,
  clear,
  toggleCategory,
  toggleFacet,
  toggleTag,
  setPage,
  setSort,
  setView,
  setTag,
  removeSearch,
  removeCategory,
  removeFacet,
  removeTag,
} = useGalleryCatalogState();
const page = computed(() => catalogState.value.page);
const sort = computed(() => catalogState.value.sort);
const view = computed(() => catalogState.value.view);
const viewerNavigation = useGalleryViewerNavigation();

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

watch(
  [pageData, query],
  ([currentPage, currentRequest]) => {
    if (!currentPage || currentPage.page !== currentRequest.page) return;
    viewerNavigation.value = createCatalogViewerNavigationSession(
      currentPage.items,
      currentRequest,
      galleryPageCount(currentPage),
    );
  },
  { immediate: true },
);

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
const tagCandidates = computed(() => {
  const contextual = pageData.value?.tags || [];
  return contextual.length ? contextual : discovery.value?.tags || [];
});
const draftFilterCount = computed(
  () =>
    selectedCategories.value.length +
    selectedFacets.value.length +
    (selectedTag.value ? 1 : 0),
);
const pageNumbers = computed(() => {
  const total = galleryPageCount(pageData.value) || 0;
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
    const tag = tagCandidates.value.find(
      (item) => item.slug === catalogState.value.tag,
    );
    items.push({
      key: `tag:${catalogState.value.tag}`,
      label: `标签：${tag?.name || catalogState.value.tag}`,
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
  const suggestions: GallerySearchSuggestion[] = [];
  let exactTagMatch = false;
  for (const tag of tagCandidates.value) {
    if (`${tag.name} ${tag.slug}`.toLocaleLowerCase().includes(term)) {
      suggestions.push({
        key: `tag:${tag.slug}`,
        label: `#${tag.name}`,
        context: `标签 · ${tag.count}`,
      });
    }
    if (
      tag.slug.toLocaleLowerCase() === term ||
      tag.name.toLocaleLowerCase() === term
    )
      exactTagMatch = true;
  }
  if (!exactTagMatch) {
    suggestions.push({
      key: `tag:${tagTerm}`,
      label: `#${tagTerm}`,
      context: "按标签精确搜索",
    });
  }
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
  return suggestions.slice(0, 9);
});

function openMobileFilters() {
  mobileFiltersOpen.value = true;
}
function closeMobileFilters() {
  mobileFiltersOpen.value = false;
}
function handleFilterBreakpoint(event: MediaQueryListEvent) {
  if (event.matches) closeMobileFilters();
}
onMounted(() => {
  desktopFilterMedia = window.matchMedia("(min-width: 64rem)");
  if (desktopFilterMedia.matches) closeMobileFilters();
  desktopFilterMedia.addEventListener("change", handleFilterBreakpoint);
});
onBeforeUnmount(() => {
  desktopFilterMedia?.removeEventListener("change", handleFilterBreakpoint);
});
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

useSeoMeta({
  title: "浏览图片",
  description: "按分类、标签和多个维度分页浏览公开图片。",
  robots: () =>
    hasFilters.value || page.value > 1 || sort.value !== "newest"
      ? "noindex,follow"
      : "index,follow",
});
</script>

<template>
  <GalleryPublicPage>
    <GalleryPageHeader
      title="浏览图片"
      description="按分类、维度和标签组合筛选，页码会记住你停下的位置。"
    >
      <template #actions>
        <p
          v-if="pageData"
          class="gallery-count shrink-0 text-[0.8rem] text-muted max-md:hidden"
        >
          <span
            class="font-display text-[1.8rem] font-semibold text-highlighted"
            >{{ pageData.total }}</span
          >
          张公开图片
        </p>
      </template>
    </GalleryPageHeader>

    <div
      class="gallery-mobile-search mb-6 flex gap-2 rounded-2xl border border-default bg-muted p-2.5 lg:hidden"
    >
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

    <div class="grid gap-7 lg:grid-cols-[18rem_minmax(0,1fr)] lg:gap-8">
      <aside class="hidden lg:block" aria-label="图片过滤器">
        <div
          class="gallery-filter-panel sticky top-24 max-h-[calc(100dvh-7rem)] space-y-7 overflow-y-auto overscroll-contain rounded-2xl border border-default bg-[color-mix(in_srgb,var(--gallery-panel)_94%,transparent)] p-[1.1rem] [scrollbar-gutter:stable]"
        >
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
            :tags="tagCandidates"
            :selected-categories="selectedCategories"
            :selected-facets="selectedFacets"
            :selected-tag="selectedTag"
            @toggle-category="toggleCategory"
            @toggle-facet="toggleFacet"
            @toggle-tag="toggleTag"
          />
          <div
            class="gallery-filter-actions sticky -bottom-[1.1rem] -mx-1 flex gap-2 bg-[linear-gradient(to_bottom,transparent,var(--gallery-panel)_0.75rem)] px-1 pb-[1.1rem] pt-3"
          >
            <UButton label="应用" size="sm" block @click="applyFilters" />
            <UButton
              v-if="hasFilters || draftFilterCount"
              color="neutral"
              variant="ghost"
              icon="i-tabler-x"
              aria-label="清除筛选"
              @click="clearFilters"
            />
          </div>
        </div>
      </aside>

      <section class="min-w-0" aria-live="polite">
        <div
          class="gallery-results-toolbar mb-5 flex min-h-[3.7rem] flex-col items-stretch justify-between gap-3 rounded-2xl border border-default bg-muted px-3.5 py-2.5 md:flex-row md:items-center"
        >
          <div class="flex min-w-0 flex-wrap gap-2">
            <span v-if="!hasFilters" class="px-1 text-sm text-muted">
              全部图片
            </span>
            <UButton
              v-for="item in activeRefinements"
              :key="item.key"
              class="gallery-refinement max-w-[min(20rem,70vw)] rounded-none [clip-path:polygon(0_0,calc(100%_-_0.42rem)_0,100%_0.42rem,100%_100%,0_100%)]"
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
          <div
            class="gallery-results-controls flex w-full shrink-0 items-center gap-1 md:w-auto"
          >
            <USelect
              :model-value="sort"
              aria-label="图片排序"
              :items="sortItems"
              value-key="value"
              class="gallery-results-sort min-w-0 flex-1 basis-32 md:w-32 md:flex-none md:shrink-0"
              @update:model-value="changeSort"
            />
            <div
              class="gallery-view-switch flex gap-0.5 rounded-xl border border-default bg-default p-0.5"
              aria-label="图片布局"
            >
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
        <p
          v-if="catalogState.q && pageData"
          class="gallery-search-explanation -mt-2.5 mb-5 break-words text-[0.8rem] text-muted"
        >
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
        />
        <GalleryImageGrid
          v-else-if="pageData?.items.length"
          :items="pageData.items"
          :priority="page === 1"
        />
        <GalleryCompactEmpty
          v-else
          icon="i-tabler-filter-off"
          title="没有符合条件的图片"
          description="换一组筛选条件，或者从完整目录重新开始。"
        >
          <template v-if="hasFilters" #actions>
            <UButton
              class="mt-4"
              color="neutral"
              variant="outline"
              label="清除筛选"
              @click="clearFilters"
            />
          </template>
        </GalleryCompactEmpty>

        <nav
          v-if="galleryPageCount(pageData) && galleryPageCount(pageData) > 1"
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
            :disabled="page >= galleryPageCount(pageData)"
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
            <p class="mt-0.5 text-xs text-muted">分类、维度与标签可组合选择</p>
          </div>
          <div class="flex items-center gap-2">
            <UBadge
              v-if="draftFilterCount"
              color="primary"
              variant="soft"
              :label="`${draftFilterCount} 项`"
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
            :tags="tagCandidates"
            :selected-categories="selectedCategories"
            :selected-facets="selectedFacets"
            :selected-tag="selectedTag"
            @toggle-category="toggleCategory"
            @toggle-facet="toggleFacet"
            @toggle-tag="toggleTag"
          />
        </div>
      </template>
      <template #footer>
        <div class="flex w-full items-center gap-2">
          <UButton
            v-if="hasFilters || draftFilterCount"
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
  </GalleryPublicPage>
</template>
