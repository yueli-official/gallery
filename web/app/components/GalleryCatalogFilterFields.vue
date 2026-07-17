<script setup lang="ts">
import type { GalleryClassificationNode, GalleryFacet } from "~/types/gallery";

defineProps<{
  categories: GalleryClassificationNode[];
  facets: GalleryFacet[];
  selectedCategories: string[];
  selectedFacets: string[];
}>();

const emit = defineEmits<{
  toggleCategory: [slug: string];
  toggleFacet: [selection: string];
}>();

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
  <div class="space-y-7">
    <section v-if="categories.length" aria-labelledby="catalog-category-label">
      <div class="mb-2 flex items-center justify-between gap-3 px-1">
        <h3
          id="catalog-category-label"
          class="text-sm font-semibold text-highlighted"
        >
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

    <section
      v-for="group in facets"
      :key="group.id"
      :aria-labelledby="`catalog-facet-${group.id}`"
    >
      <div class="mb-2 flex items-center justify-between gap-3 px-1">
        <h3
          :id="`catalog-facet-${group.id}`"
          class="text-sm font-semibold text-highlighted"
        >
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
</template>
