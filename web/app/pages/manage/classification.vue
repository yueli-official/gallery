<script setup lang="ts">
import type { GalleryDiscovery } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "分类与维度 · 图库管理" });
const { call } = useApi();
const { data, pending, error, refresh } = await useAsyncData("gallery-manage-classification", () => call<GalleryDiscovery>("/api/v1/gallery/discovery?seed=classification"), { server: false });
const groups = computed(() => (data.value?.facets || []).map(facet => ({ facet, values: (data.value?.facetValues || []).filter(value => value.facetId === facet.id) })));
</script>

<template>
  <div>
    <ManageHeader title="分类与维度"><template #subtitle>Category 在界面中保持正常语义，后端映射到保留的 single-select topic Facet；Tag 仍是自由关键词。</template></ManageHeader>
    <UAlert class="mb-5" color="info" variant="subtle" icon="i-tabler-hierarchy" title="平台统一模型稍后迁移" description="Gallery 当前消费共享 Facet 行为，但 Category/Tag/Facet 的全站写模型会在独立 Topic 中统一；这里不复制一套临时治理 API。" />
    <SkeletonList v-if="pending" :rows="6" />
    <UAlert v-else-if="error" color="error" variant="subtle" title="分类数据加载失败"><template #actions><UButton label="重试" @click="() => refresh()" /></template></UAlert>
    <div v-else class="grid gap-4 xl:grid-cols-2"><section v-for="group in groups" :key="group.facet.id" class="rounded-lg border border-default bg-default p-5"><div class="flex items-start justify-between gap-4"><div><h2 class="font-semibold text-highlighted">{{ group.facet.name }}</h2><p class="mt-1 text-sm text-muted">{{ group.facet.description }}</p></div><UBadge color="neutral" variant="soft" :label="group.facet.selectionMode" /></div><div class="mt-4 divide-y divide-default"><div v-for="value in group.values" :key="value.id" class="flex items-center justify-between gap-3 py-2 text-sm"><span>{{ value.name }}</span><span class="tabular-nums text-muted">{{ value.count }}</span></div></div></section></div>
  </div>
</template>
