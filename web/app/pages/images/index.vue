<script setup lang="ts">
import type {
  GalleryDiscovery,
  GalleryFacet,
  GalleryImagePage,
  GallerySearchSuggestion,
} from "~/types/gallery";
import type { GalleryCatalogSort, GalleryCatalogView } from "~/utils/catalog";

const mobileFiltersOpen = ref(false);
const facetSearch = reactive<Record<string, string>>({});
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
  const term = searchDraft.value.trim().toLocaleLowerCase();
  if (!term) return [];
  const suggestions: GallerySearchSuggestion[] = [];
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
  return suggestions.slice(0, 6);
});

function filteredFacetValues(facet: GalleryFacet) {
  const term = (facetSearch[facet.slug] || "").trim().toLocaleLowerCase();
  if (!term) return facet.values;
  return facet.values.filter((value) =>
    `${value.name} ${value.slug}`.toLocaleLowerCase().includes(term),
  );
}

function openMobileFilters() {
  mobileFiltersOpen.value = true;
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

    <div class="gallery-mobile-search xl:hidden">
      <UInput
        v-model="searchDraft"
        class="min-w-0 flex-1"
        icon="i-tabler-search"
        placeholder="搜索标题或说明"
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
      class="-mt-3 mb-5 xl:hidden"
      :items="searchSuggestions"
      @select="selectSuggestion"
    />

    <div class="grid gap-7 xl:grid-cols-[16.5rem_minmax(0,1fr)] xl:gap-10">
      <aside class="hidden xl:block">
        <div class="gallery-filter-panel sticky top-24 space-y-7">
          <UInput
            v-model="searchDraft"
            icon="i-tabler-search"
            placeholder="搜索图片"
            @keyup.enter="applyFilters"
          />
          <GallerySearchSuggestions
            :items="searchSuggestions"
            @select="selectSuggestion"
          />
          <div v-if="categoryCandidates.length">
            <h2 class="mb-3 text-sm font-semibold text-highlighted">分类</h2>
            <div class="space-y-1">
              <label
                v-for="category in categoryCandidates"
                :key="category.id"
                class="gallery-filter-option"
              >
                <span class="flex min-w-0 items-center gap-2">
                  <input
                    type="checkbox"
                    class="size-4 accent-[var(--ui-primary)]"
                    :checked="selectedCategories.includes(category.slug)"
                    @change="toggleCategory(category.slug)"
                  />
                  <span class="truncate">{{ category.name }}</span>
                </span>
                <span class="text-xs tabular-nums text-dimmed">{{
                  category.count
                }}</span>
              </label>
            </div>
          </div>
          <div v-for="group in facetGroups" :key="group.facet.id">
            <h2 class="mb-3 text-sm font-semibold text-highlighted">
              {{ group.facet.name }}
            </h2>
            <UInput
              v-if="group.values.length > 8"
              v-model="facetSearch[group.facet.slug]"
              class="mb-2"
              size="xs"
              icon="i-tabler-search"
              :placeholder="`查找${group.facet.name}`"
              :aria-label="`查找${group.facet.name}选项`"
            />
            <div class="space-y-1">
              <label
                v-for="value in filteredFacetValues(group.facet)"
                :key="value.id"
                class="gallery-filter-option"
              >
                <span class="flex min-w-0 items-center gap-2">
                  <input
                    type="checkbox"
                    class="size-4 accent-[var(--ui-primary)]"
                    :checked="
                      selectedFacets.includes(
                        `${group.facet.slug}:${value.slug}`,
                      )
                    "
                    @change="toggleFacet(`${group.facet.slug}:${value.slug}`)"
                  />
                  <span class="truncate">{{ value.name }}</span>
                </span>
                <span class="text-xs tabular-nums text-dimmed">{{
                  value.count
                }}</span>
              </label>
            </div>
          </div>
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
                icon="i-tabler-layout-grid"
                aria-label="网格布局"
                @click="changeView('grid')"
              />
              <UButton
                color="neutral"
                :variant="view === 'masonry' ? 'soft' : 'ghost'"
                icon="i-tabler-layout-columns"
                aria-label="瀑布流布局"
                @click="changeView('masonry')"
              />
            </div>
          </div>
        </div>
        <p v-if="catalogState.q && pageData" class="gallery-search-explanation">
          标题或说明中包含“{{ catalogState.q }}”的结果，共
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

    <USlideover
      v-model:open="mobileFiltersOpen"
      title="筛选图片"
      description="选择分类与维度，应用后会写入网址。"
    >
      <template #body>
        <div class="space-y-7 pb-24">
          <UInput
            v-model="searchDraft"
            icon="i-tabler-search"
            placeholder="搜索标题或说明"
          />
          <GallerySearchSuggestions
            :items="searchSuggestions"
            @select="selectSuggestion"
          />
          <div v-if="categoryCandidates.length">
            <h2 class="mb-3 font-semibold text-highlighted">分类</h2>
            <div class="grid grid-cols-2 gap-2">
              <UButton
                v-for="category in categoryCandidates"
                :key="category.id"
                color="neutral"
                :variant="
                  selectedCategories.includes(category.slug)
                    ? 'solid'
                    : 'outline'
                "
                :label="`${category.name} ${category.count}`"
                block
                @click="toggleCategory(category.slug)"
              />
            </div>
          </div>
          <div v-for="group in facetGroups" :key="group.facet.id">
            <h2 class="mb-3 font-semibold text-highlighted">
              {{ group.facet.name }}
            </h2>
            <UInput
              v-if="group.values.length > 8"
              v-model="facetSearch[group.facet.slug]"
              class="mb-3"
              icon="i-tabler-search"
              :placeholder="`查找${group.facet.name}`"
              :aria-label="`查找${group.facet.name}选项`"
            />
            <div class="grid grid-cols-2 gap-2">
              <UButton
                v-for="value in filteredFacetValues(group.facet)"
                :key="value.id"
                color="neutral"
                :variant="
                  selectedFacets.includes(`${group.facet.slug}:${value.slug}`)
                    ? 'solid'
                    : 'outline'
                "
                :label="value.name"
                block
                @click="toggleFacet(`${group.facet.slug}:${value.slug}`)"
              />
            </div>
          </div>
        </div>
        <div
          class="fixed inset-x-0 bottom-0 flex gap-2 border-t border-default bg-default p-4 pb-[calc(1rem+env(safe-area-inset-bottom))]"
        >
          <UButton
            v-if="hasFilters"
            color="neutral"
            variant="outline"
            label="清除"
            @click="clearFilters"
          />
          <UButton
            class="flex-1"
            :label="`应用${selectedCategories.length + selectedFacets.length ? ` ${selectedCategories.length + selectedFacets.length} 项` : ''}`"
            @click="applyFilters"
          />
        </div>
      </template>
    </USlideover>

    <GalleryQuickView
      :image-id="preview"
      :items="pageData?.items || []"
      @close="closePreview"
      @navigate="navigatePreview"
    />
  </div>
</template>
