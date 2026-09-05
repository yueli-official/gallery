<script setup lang="ts">
import type {
  GalleryClassificationNode,
  GalleryFacet,
  GalleryTagCandidate,
} from "~/types/gallery";

defineProps<{
  categories: GalleryClassificationNode[];
  facets: GalleryFacet[];
  tags: GalleryTagCandidate[];
  selectedCategories: string[];
  selectedFacets: string[];
  selectedTag: string;
}>();

const emit = defineEmits<{
  toggleCategory: [slug: string];
  toggleFacet: [selection: string];
  toggleTag: [slug: string];
}>();

type FilterSection = "categories" | "facets" | "tags";
const activeSection = ref<FilterSection>("categories");
const facetSearch = reactive<Record<string, string>>({});

function filteredValues(facet: GalleryFacet) {
  const term = (facetSearch[facet.slug] || "").trim().toLocaleLowerCase();
  if (!term) return facet.values;
  return facet.values.filter((value) =>
    `${value.name} ${value.slug}`.toLocaleLowerCase().includes(term),
  );
}

function facetKey(facetSlug: string, valueSlug: string) {
  return `${facetSlug}:${valueSlug}`;
}
</script>

<template>
  <div class="space-y-6" data-gallery-catalog-filters>
    <div
      class="gallery-filter-section-tabs grid grid-cols-3 gap-0.5 rounded-xl border border-default bg-default p-0.5"
      role="tablist"
      aria-label="筛选类型"
    >
      <button
        type="button"
        role="tab"
        :aria-selected="activeSection === 'categories'"
        aria-controls="gallery-filter-categories"
      aria-label="分类筛选"
      class="inline-flex min-h-11 min-w-0 cursor-pointer items-center justify-center gap-1 rounded-lg text-[0.8rem] font-semibold text-muted hover:text-highlighted focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary disabled:cursor-not-allowed disabled:opacity-45 md:min-h-9"
      :class="
        activeSection === 'categories'
          ? 'bg-muted text-highlighted shadow-sm'
          : ''
      "
        :disabled="!categories.length"
        @click="activeSection = 'categories'"
      >
        分类
        <span
          v-if="selectedCategories.length"
          class="inline-grid h-[1.15rem] min-w-[1.15rem] place-items-center rounded-full bg-primary/10 text-[0.68rem] tabular-nums text-primary"
        >
          {{ selectedCategories.length }}
        </span>
      </button>
      <button
        type="button"
        role="tab"
        :aria-selected="activeSection === 'facets'"
        aria-controls="gallery-filter-facets"
      aria-label="维度筛选"
      class="inline-flex min-h-11 min-w-0 cursor-pointer items-center justify-center gap-1 rounded-lg text-[0.8rem] font-semibold text-muted hover:text-highlighted focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary disabled:cursor-not-allowed disabled:opacity-45 md:min-h-9"
      :class="
        activeSection === 'facets'
          ? 'bg-muted text-highlighted shadow-sm'
          : ''
      "
        :disabled="!facets.length"
        @click="activeSection = 'facets'"
      >
        维度
        <span
          v-if="selectedFacets.length"
          class="inline-grid h-[1.15rem] min-w-[1.15rem] place-items-center rounded-full bg-primary/10 text-[0.68rem] tabular-nums text-primary"
          >{{ selectedFacets.length }}</span
        >
      </button>
      <button
        type="button"
        role="tab"
        :aria-selected="activeSection === 'tags'"
        aria-controls="gallery-filter-tags"
      aria-label="标签筛选"
      class="inline-flex min-h-11 min-w-0 cursor-pointer items-center justify-center gap-1 rounded-lg text-[0.8rem] font-semibold text-muted hover:text-highlighted focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary disabled:cursor-not-allowed disabled:opacity-45 md:min-h-9"
      :class="
        activeSection === 'tags' ? 'bg-muted text-highlighted shadow-sm' : ''
      "
        :disabled="!tags.length"
        @click="activeSection = 'tags'"
      >
        标签
        <span
          v-if="selectedTag"
          class="inline-grid h-[1.15rem] min-w-[1.15rem] place-items-center rounded-full bg-primary/10 text-[0.68rem] tabular-nums text-primary"
          >1</span
        >
      </button>
    </div>

    <section
      v-if="activeSection === 'categories' && categories.length"
      id="gallery-filter-categories"
      class="gallery-filter-group"
      role="tabpanel"
      aria-labelledby="catalog-category-label"
    >
      <div class="gallery-filter-group-heading mb-2 flex items-center justify-between gap-3 px-1">
        <h3
          id="catalog-category-label"
          class="flex items-center gap-2 text-sm font-semibold text-highlighted"
        >
          <UIcon name="i-tabler-category" class="size-4 text-muted" />
          分类
        </h3>
        <span class="text-xs tabular-nums text-dimmed"
          >{{ categories.length }} 项</span
        >
      </div>
      <div class="space-y-0.5">
        <label
          v-for="category in categories"
          :key="category.id"
          class="gallery-filter-option flex min-h-[2.4rem] cursor-pointer items-center justify-between gap-3 rounded-lg px-2 py-1.5 text-sm transition-colors hover:bg-primary/10"
          :class="{
            'bg-primary/10 text-highlighted': selectedCategories.includes(
              category.slug,
            ),
          }"
        >
          <span class="flex min-w-0 items-center gap-2.5">
            <UCheckbox
              :model-value="selectedCategories.includes(category.slug)"
              :aria-label="category.name"
              @update:model-value="emit('toggleCategory', category.slug)"
            />
            <span class="truncate">{{ category.name }}</span>
          </span>
          <span class="text-xs tabular-nums text-dimmed">{{
            category.count
          }}</span>
        </label>
      </div>
    </section>

    <div
      v-if="activeSection === 'facets' && facets.length"
      id="gallery-filter-facets"
      class="space-y-6"
      role="tabpanel"
    >
      <section
        v-for="group in facets"
        :key="group.id"
        class="gallery-filter-group"
        :aria-labelledby="`catalog-facet-${group.id}`"
      >
        <div class="gallery-filter-group-heading mb-2 flex items-center justify-between gap-3 px-1">
          <h3
            :id="`catalog-facet-${group.id}`"
            class="flex items-center gap-2 text-sm font-semibold text-highlighted"
          >
            <UIcon
              name="i-tabler-adjustments-horizontal"
              class="size-4 text-muted"
            />
            {{ group.name }}
          </h3>
          <span class="text-xs tabular-nums text-dimmed"
            >{{ group.values.length }} 项</span
          >
        </div>
        <UInput
          v-if="group.values.length > 8"
          v-model="facetSearch[group.slug]"
          class="mb-2"
          size="sm"
          icon="i-tabler-search"
          :placeholder="`查找${group.name}`"
          :aria-label="`查找${group.name}选项`"
        />
        <div class="space-y-0.5">
          <label
            v-for="value in filteredValues(group)"
            :key="value.id"
            class="gallery-filter-option flex min-h-[2.4rem] cursor-pointer items-center justify-between gap-3 rounded-lg px-2 py-1.5 text-sm transition-colors hover:bg-primary/10"
            :class="{
              'bg-primary/10 text-highlighted': selectedFacets.includes(
                facetKey(group.slug, value.slug),
              ),
            }"
          >
            <span class="flex min-w-0 items-center gap-2.5">
              <UCheckbox
                :model-value="
                  selectedFacets.includes(facetKey(group.slug, value.slug))
                "
                :aria-label="`${group.name}：${value.name}`"
                @update:model-value="
                  emit('toggleFacet', facetKey(group.slug, value.slug))
                "
              />
              <span class="truncate">{{ value.name }}</span>
            </span>
            <span class="text-xs tabular-nums text-dimmed">{{
              value.count
            }}</span>
          </label>
        </div>
      </section>
    </div>

    <section
      v-if="activeSection === 'tags' && tags.length"
      id="gallery-filter-tags"
      class="gallery-filter-group"
      role="tabpanel"
      aria-labelledby="catalog-tag-label"
    >
      <div class="gallery-filter-group-heading mb-2 flex items-center justify-between gap-3 px-1">
        <h3
          id="catalog-tag-label"
          class="flex items-center gap-2 text-sm font-semibold text-highlighted"
        >
          <UIcon name="i-tabler-tags" class="size-4 text-muted" />
          热门标签
        </h3>
        <span class="text-xs tabular-nums text-dimmed">
          {{ tags.length }} 项
        </span>
      </div>
      <div class="gallery-tag-filters flex flex-wrap gap-1.5">
        <button
          v-for="tag in tags"
          :key="tag.id"
          type="button"
          class="gallery-tag-filter inline-flex min-h-11 cursor-pointer items-center gap-1.5 rounded-lg border border-default bg-default px-3 py-1.5 text-[0.78rem] leading-none text-muted transition-colors hover:border-primary/40 hover:text-highlighted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary md:min-h-[2.15rem] md:px-2"
          :class="
            selectedTag === tag.slug
              ? 'border-primary bg-primary text-white'
              : ''
          "
          :aria-pressed="selectedTag === tag.slug"
          :aria-label="`标签：${tag.name}，${tag.count} 张图片`"
          @click="emit('toggleTag', tag.slug)"
        >
          <span>#{{ tag.name }}</span>
          <span
            class="gallery-tag-filter-count tabular-nums"
            :class="selectedTag === tag.slug ? 'text-white/80' : 'text-dimmed'"
            >{{ tag.count }}</span
          >
        </button>
      </div>
      <p class="mt-2 px-1 text-xs leading-5 text-dimmed">
        一次选择一个标签，可继续与分类和维度组合。
      </p>
    </section>
  </div>
</template>
