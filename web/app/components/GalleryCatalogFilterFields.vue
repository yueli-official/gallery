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
      class="gallery-filter-section-tabs"
      role="tablist"
      aria-label="筛选类型"
    >
      <button
        type="button"
        role="tab"
        :aria-selected="activeSection === 'categories'"
        aria-controls="gallery-filter-categories"
        aria-label="分类筛选"
        :disabled="!categories.length"
        @click="activeSection = 'categories'"
      >
        分类
        <span v-if="selectedCategories.length">
          {{ selectedCategories.length }}
        </span>
      </button>
      <button
        type="button"
        role="tab"
        :aria-selected="activeSection === 'facets'"
        aria-controls="gallery-filter-facets"
        aria-label="维度筛选"
        :disabled="!facets.length"
        @click="activeSection = 'facets'"
      >
        维度
        <span v-if="selectedFacets.length">{{ selectedFacets.length }}</span>
      </button>
      <button
        type="button"
        role="tab"
        :aria-selected="activeSection === 'tags'"
        aria-controls="gallery-filter-tags"
        aria-label="标签筛选"
        :disabled="!tags.length"
        @click="activeSection = 'tags'"
      >
        标签
        <span v-if="selectedTag">1</span>
      </button>
    </div>

    <section
      v-if="activeSection === 'categories' && categories.length"
      id="gallery-filter-categories"
      class="gallery-filter-group"
      role="tabpanel"
      aria-labelledby="catalog-category-label"
    >
      <div class="gallery-filter-group-heading">
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
          class="gallery-filter-option"
          :class="{
            'is-selected': selectedCategories.includes(category.slug),
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
        <div class="gallery-filter-group-heading">
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
            class="gallery-filter-option"
            :class="{
              'is-selected': selectedFacets.includes(
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
      <div class="gallery-filter-group-heading">
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
      <div class="gallery-tag-filters">
        <button
          v-for="tag in tags"
          :key="tag.id"
          type="button"
          class="gallery-tag-filter"
          :class="{ 'is-selected': selectedTag === tag.slug }"
          :aria-pressed="selectedTag === tag.slug"
          :aria-label="`标签：${tag.name}，${tag.count} 张图片`"
          @click="emit('toggleTag', tag.slug)"
        >
          <span>#{{ tag.name }}</span>
          <span class="gallery-tag-filter-count">{{ tag.count }}</span>
        </button>
      </div>
      <p class="mt-2 px-1 text-xs leading-5 text-dimmed">
        一次选择一个标签，可继续与分类和维度组合。
      </p>
    </section>
  </div>
</template>
