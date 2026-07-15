<script setup lang="ts">
import type { GalleryRanking } from "~/types/gallery";

const route = useRoute();
const router = useRouter();
const kind = computed(() => String(route.query.kind || "trending"));
const windowKey = computed(() => String(route.query.window || "7d"));
const { data, error, status, refresh } = await useFetch<{ ranking: GalleryRanking }>("/api/gallery/rankings", { query: computed(() => ({ kind: kind.value, window: windowKey.value })), watch: [kind, windowKey] });
const kinds = [{ label: "趋势", value: "trending" }, { label: "收藏最多", value: "most_favorited" }, { label: "浏览最多", value: "most_viewed" }];
const windows = [{ label: "24 小时", value: "24h" }, { label: "7 天", value: "7d" }, { label: "30 天", value: "30d" }, { label: "全部", value: "all" }];
function setQuery(patch: Record<string, string>) { void router.replace({ query: { ...route.query, ...patch } }); }
useSeoMeta({ title: "排行榜", description: "按合格浏览和私人收藏聚合的图片排行。" });
</script>

<template>
  <div class="gallery-page">
    <header class="mb-7 flex flex-wrap items-end justify-between gap-5">
      <div><h1 class="text-3xl font-semibold tracking-tight text-highlighted">排行榜</h1><p class="mt-1.5 text-sm text-muted">排行帮助回看，不影响首页的随机发现。</p></div>
      <USelect :model-value="windowKey" :items="windows" value-key="value" class="w-32" @update:model-value="value => setQuery({ window: String(value) })" />
    </header>
    <div class="mb-6 flex gap-1 overflow-x-auto border-b border-default pb-3">
      <UButton v-for="item in kinds" :key="item.value" color="neutral" :variant="kind === item.value ? 'soft' : 'ghost'" :label="item.label" @click="setQuery({ kind: item.value })" />
    </div>
    <div v-if="status === 'pending'" class="gallery-grid"><USkeleton v-for="index in 15" :key="index" class="aspect-[4/3] rounded-lg" /></div>
    <UAlert v-else-if="error" color="error" variant="subtle" title="排行加载失败"><template #actions><UButton label="重试" @click="() => refresh()" /></template></UAlert>
    <GalleryImageGrid v-else-if="data?.ranking.images.length" :items="data.ranking.images" priority />
    <div v-else class="border-y border-dashed border-default py-16 text-center text-sm text-muted">当前时间范围还没有足够数据。</div>
  </div>
</template>
