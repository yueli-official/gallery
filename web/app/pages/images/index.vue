<script setup lang="ts">
import type { GalleryDiscovery, GalleryImagePage } from "~/types/gallery";

const route = useRoute();
const router = useRouter();
const mobileFiltersOpen = ref(false);
const searchDraft = ref(String(route.query.q || ""));
const selectedCategories = ref(String(route.query.categories || "").split(",").filter(Boolean));
const selectedFacets = ref(String(route.query.facets || "").split(",").filter(Boolean));

const page = computed(() => Math.max(1, Number(route.query.page || 1)));
const sort = computed(() => String(route.query.sort || "newest"));
const query = computed(() => ({
  q: String(route.query.q || "") || undefined,
  sort: sort.value,
  page: page.value,
  size: 24,
	categories: String(route.query.categories || "") || undefined,
  facets: String(route.query.facets || "") || undefined,
  tag: String(route.query.tag || "") || undefined,
}));

const [{ data: pageData, error, status, refresh }, { data: discovery }] = await Promise.all([
  useFetch<GalleryImagePage>("/api/gallery/images", { query, watch: [query] }),
  useFetch<GalleryDiscovery>("/api/gallery/discovery", { query: { seed: "catalog-facets" } }),
]);

const facetGroups = computed(() => {
	return (pageData.value?.facets || discovery.value?.facets || []).map(facet => ({
    facet,
		values: facet.values,
  }));
});
const categoryCandidates = computed(() => pageData.value?.categories || discovery.value?.categories || []);
const pageNumbers = computed(() => {
  const total = pageData.value?.totalPages || 0;
  if (!total) return [];
  const start = Math.max(1, Math.min(page.value - 2, total - 4));
  return Array.from({ length: Math.min(5, total) }, (_, index) => start + index);
});
const hasFilters = computed(() => Boolean(query.value.q || query.value.categories || query.value.facets || query.value.tag));
const sortItems = [
  { label: "最新", value: "newest" },
  { label: "最早", value: "oldest" },
  { label: "标题 A–Z", value: "title_asc" },
  { label: "标题 Z–A", value: "title_desc" },
];

function replaceQuery(patch: Record<string, string | number | undefined>) {
  void router.push({ path: "/images", query: { ...route.query, ...patch } });
}
function applyFilters() {
	replaceQuery({ q: searchDraft.value.trim() || undefined, categories: selectedCategories.value.join(",") || undefined, facets: selectedFacets.value.join(",") || undefined, page: 1 });
  mobileFiltersOpen.value = false;
}
function clearFilters() {
  searchDraft.value = "";
	selectedCategories.value = [];
  selectedFacets.value = [];
  void router.push({ path: "/images", query: { sort: sort.value } });
}
function toggleFacet(id: string) {
  selectedFacets.value = selectedFacets.value.includes(id)
    ? selectedFacets.value.filter(item => item !== id)
    : [...selectedFacets.value, id];
}
function toggleCategory(slug: string) {
	selectedCategories.value = selectedCategories.value.includes(slug)
		? selectedCategories.value.filter(item => item !== slug)
		: [...selectedCategories.value, slug];
}

useSeoMeta({
  title: "浏览图片",
  description: "按分类、标签和多个维度分页浏览公开图片。",
  robots: () => hasFilters.value || page.value > 1 || sort.value !== "newest" ? "noindex,follow" : "index,follow",
});
</script>

<template>
  <div class="gallery-page">
    <header class="mb-6 flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 class="text-3xl font-semibold tracking-tight text-highlighted">浏览图片</h1>
        <p class="mt-1.5 text-sm text-muted">固定网格与页码，适合持续浏览和回到原位置。</p>
      </div>
      <p v-if="pageData" class="text-sm tabular-nums text-muted">{{ pageData.total }} 张</p>
    </header>

    <div class="mb-6 flex gap-2 lg:hidden">
      <UInput v-model="searchDraft" class="min-w-0 flex-1" icon="i-tabler-search" placeholder="搜索标题或说明" @keyup.enter="applyFilters" />
      <UButton color="neutral" variant="outline" icon="i-tabler-adjustments-horizontal" label="筛选" @click="() => { mobileFiltersOpen = true; }" />
    </div>

    <div class="grid gap-8 lg:grid-cols-[15rem_minmax(0,1fr)]">
      <aside class="hidden lg:block">
        <div class="sticky top-24 space-y-6">
          <UInput v-model="searchDraft" icon="i-tabler-search" placeholder="搜索图片" @keyup.enter="applyFilters" />
		  <div v-if="categoryCandidates.length">
			<h2 class="mb-2 text-xs font-semibold uppercase tracking-[.14em] text-muted">分类</h2>
			<div class="space-y-1">
			  <label v-for="category in categoryCandidates" :key="category.id" class="flex cursor-pointer items-center justify-between gap-3 rounded-md px-2 py-1.5 text-sm hover:bg-elevated/60">
				<span class="flex min-w-0 items-center gap-2">
				<input type="checkbox" class="size-4 accent-[var(--ui-primary)]" :checked="selectedCategories.includes(category.slug)" @change="toggleCategory(category.slug)" />
				<span class="truncate">{{ category.name }}</span>
				</span>
				<span class="text-xs tabular-nums text-dimmed">{{ category.count }}</span>
			  </label>
			</div>
		  </div>
          <div v-for="group in facetGroups" :key="group.facet.id">
            <h2 class="mb-2 text-xs font-semibold uppercase tracking-[.14em] text-muted">{{ group.facet.name }}</h2>
            <div class="space-y-1">
			  <label v-for="value in group.values" :key="value.id" class="flex cursor-pointer items-center justify-between gap-3 rounded-md px-2 py-1.5 text-sm hover:bg-elevated/60">
                <span class="flex min-w-0 items-center gap-2">
				  <input type="checkbox" class="size-4 accent-[var(--ui-primary)]" :checked="selectedFacets.includes(`${group.facet.slug}:${value.slug}`)" @change="toggleFacet(`${group.facet.slug}:${value.slug}`)" />
                  <span class="truncate">{{ value.name }}</span>
                </span>
				<span class="text-xs tabular-nums text-dimmed">{{ value.count }}</span>
              </label>
            </div>
          </div>
          <div class="flex gap-2">
            <UButton label="应用" size="sm" block @click="applyFilters" />
            <UButton v-if="hasFilters" color="neutral" variant="ghost" icon="i-tabler-x" aria-label="清除筛选" @click="clearFilters" />
          </div>
        </div>
      </aside>

      <section aria-live="polite">
        <div class="mb-5 flex items-center justify-between gap-3 border-b border-default pb-4">
          <div class="flex min-w-0 flex-wrap gap-2">
            <UBadge v-if="query.q" color="neutral" variant="soft" :label="`搜索：${query.q}`" />
			<UBadge v-if="selectedCategories.length" color="primary" variant="soft" :label="`${selectedCategories.length} 个分类`" />
            <UBadge v-if="selectedFacets.length" color="primary" variant="soft" :label="`${selectedFacets.length} 个筛选`" />
            <UButton v-if="hasFilters" color="neutral" variant="ghost" size="xs" label="清除" @click="clearFilters" />
          </div>
          <USelect :model-value="sort" :items="sortItems" value-key="value" class="w-32 shrink-0" @update:model-value="value => replaceQuery({ sort: String(value), page: 1 })" />
        </div>

        <div v-if="status === 'pending'" class="gallery-grid"><USkeleton v-for="index in 16" :key="index" class="aspect-[4/3] rounded-lg" /></div>
        <UAlert v-else-if="error" color="error" variant="subtle" icon="i-tabler-alert-circle" title="图片列表加载失败"><template #actions><UButton label="重试" @click="() => refresh()" /></template></UAlert>
        <GalleryImageGrid v-else-if="pageData?.items.length" :items="pageData.items" :priority="page === 1" />
        <div v-else class="border-y border-dashed border-default py-20 text-center">
          <UIcon name="i-tabler-filter-off" class="mx-auto size-8 text-dimmed" />
          <h2 class="mt-3 font-semibold text-highlighted">没有符合条件的图片</h2>
          <UButton v-if="hasFilters" class="mt-4" color="neutral" variant="outline" label="清除筛选" @click="clearFilters" />
        </div>

        <nav v-if="pageData?.totalPages && pageData.totalPages > 1" class="mt-10 flex items-center justify-center gap-1" aria-label="图片分页">
          <UButton color="neutral" variant="ghost" icon="i-tabler-chevron-left" aria-label="上一页" :disabled="page <= 1" @click="replaceQuery({ page: page - 1 })" />
          <UButton v-for="number in pageNumbers" :key="number" color="neutral" :variant="number === page ? 'solid' : 'ghost'" :label="String(number)" :aria-current="number === page ? 'page' : undefined" @click="replaceQuery({ page: number })" />
          <UButton color="neutral" variant="ghost" icon="i-tabler-chevron-right" aria-label="下一页" :disabled="page >= pageData.totalPages" @click="replaceQuery({ page: page + 1 })" />
        </nav>
      </section>
    </div>

    <USlideover v-model:open="mobileFiltersOpen" title="筛选图片" description="选择分类与维度，应用后会写入网址。">
      <template #body>
        <div class="space-y-7 pb-24">
          <UInput v-model="searchDraft" icon="i-tabler-search" placeholder="搜索标题或说明" />
		  <div v-if="categoryCandidates.length">
			<h2 class="mb-3 font-semibold text-highlighted">分类</h2>
			<div class="grid grid-cols-2 gap-2"><UButton v-for="category in categoryCandidates" :key="category.id" color="neutral" :variant="selectedCategories.includes(category.slug) ? 'solid' : 'outline'" :label="`${category.name} · ${category.count}`" block @click="toggleCategory(category.slug)" /></div>
		  </div>
          <div v-for="group in facetGroups" :key="group.facet.id">
            <h2 class="mb-3 font-semibold text-highlighted">{{ group.facet.name }}</h2>
            <div class="grid grid-cols-2 gap-2">
			  <UButton v-for="value in group.values" :key="value.id" color="neutral" :variant="selectedFacets.includes(`${group.facet.slug}:${value.slug}`) ? 'solid' : 'outline'" :label="value.name" block @click="toggleFacet(`${group.facet.slug}:${value.slug}`)" />
            </div>
          </div>
        </div>
        <div class="fixed inset-x-0 bottom-0 flex gap-2 border-t border-default bg-default p-4 pb-[calc(1rem+env(safe-area-inset-bottom))]">
          <UButton v-if="hasFilters" color="neutral" variant="outline" label="清除" @click="clearFilters" />
		  <UButton class="flex-1" :label="`应用${selectedCategories.length + selectedFacets.length ? ` ${selectedCategories.length + selectedFacets.length} 项` : ''}`" @click="applyFilters" />
        </div>
      </template>
    </USlideover>
  </div>
</template>
