<script setup lang="ts">
import { groupFacetValues } from "@platform/facet";

import type { GalleryFacet, GalleryFacetValue } from "~/types/gallery";

const { facets, values } = defineProps<{
  facets: GalleryFacet[];
  values: GalleryFacetValue[];
}>();

const groups = computed(() => groupFacetValues(facets, values));
</script>

<template>
  <section v-if="groups.length" aria-labelledby="facets-heading">
    <div class="mb-6 max-w-2xl">
      <p class="text-sm font-medium text-primary">不只按一个分类浏览</p>
      <h2
        id="facets-heading"
        class="mt-1 text-2xl font-semibold text-highlighted"
      >
        多维度探索
      </h2>
      <p class="mt-2 text-sm leading-6 text-muted">
        同一件作品可以同时属于不同媒介、题材、风格和用途。
      </p>
    </div>

    <div class="grid gap-4 lg:grid-cols-2">
      <article
        v-for="group in groups"
        :key="group.facet.id"
        class="rounded-2xl border border-default bg-elevated/50 p-5"
      >
        <div class="flex items-start justify-between gap-4">
          <div>
            <h3 class="font-semibold text-highlighted">
              {{ group.facet.name }}
            </h3>
            <p class="mt-1 text-sm leading-6 text-muted">
              {{ group.facet.description }}
            </p>
          </div>
          <UBadge color="neutral" variant="subtle">
            {{ group.values.length }} 项
          </UBadge>
        </div>
        <ul
          class="mt-4 flex flex-wrap gap-2"
          :aria-label="`${group.facet.name}分类`"
        >
          <li v-for="value in group.values" :key="value.id">
            <span
              class="inline-flex items-center gap-2 rounded-full border border-default bg-default px-3 py-1.5 text-sm text-toned"
            >
              {{ value.name }}
              <span v-if="value.count" class="text-xs text-dimmed">
                {{ value.count }}
              </span>
            </span>
          </li>
        </ul>
      </article>
    </div>
  </section>
</template>
