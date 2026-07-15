<script setup lang="ts">
import type { GalleryDiscovery } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "分类与维度 · 图库管理" });
const { call } = useApi();
const { data, pending, error, refresh } = await useAsyncData("gallery-manage-classification", () => call<GalleryDiscovery>("/api/v1/gallery/discovery?seed=classification"), { server: false });
const groups = computed(() => data.value?.facets || []);
</script>

<template>
  <div>
    <ManageHeader title="分类与维度"><template #subtitle>Category 负责主要浏览入口，Facet 负责可组合的结构化筛选；Tag 保留为扁平长尾词。</template></ManageHeader>
    <SkeletonList v-if="pending" :rows="6" />
    <UAlert v-else-if="error" color="error" variant="subtle" title="分类数据加载失败"><template #actions><UButton label="重试" @click="() => refresh()" /></template></UAlert>
	<div v-else class="space-y-6">
	  <section class="rounded-lg border border-default bg-default p-5">
		<h2 class="font-semibold text-highlighted">Category</h2>
		<p class="mt-1 text-sm text-muted">对象可以归入多个分类，并显式选择一个主分类。</p>
		<div class="mt-4 flex flex-wrap gap-2"><UBadge v-for="category in data?.categories || []" :key="category.id" color="primary" variant="soft" :label="category.name" /></div>
	  </section>
	  <div class="grid gap-4 xl:grid-cols-2">
		<section v-for="facet in groups" :key="facet.id" class="rounded-lg border border-default bg-default p-5">
		  <h2 class="font-semibold text-highlighted">{{ facet.name }}</h2>
		  <div class="mt-4 divide-y divide-default"><div v-for="value in facet.values" :key="value.id" class="py-2 text-sm"><span>{{ value.name }}</span></div></div>
		</section>
	  </div>
	</div>
  </div>
</template>
