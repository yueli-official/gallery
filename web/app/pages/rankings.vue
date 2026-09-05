<script setup lang="ts">
import type { GalleryRanking } from "~/types/gallery";

const route = useRoute();
const router = useRouter();
const kind = computed(() => String(route.query.kind || "trending"));
const windowKey = computed(() => String(route.query.window || "7d"));
const { data, error, status, refresh } = await useFetch<{
  ranking: GalleryRanking;
}>("/api/gallery/rankings", {
  query: computed(() => ({ kind: kind.value, window: windowKey.value })),
  watch: [kind, windowKey],
});
const kinds = [
  { label: "趋势", value: "trending" },
  { label: "收藏最多", value: "most_favorited" },
  { label: "浏览最多", value: "most_viewed" },
];
const windows = [
  { label: "24 小时", value: "24h" },
  { label: "7 天", value: "7d" },
  { label: "30 天", value: "30d" },
  { label: "全部", value: "all" },
];
const images = computed(() => data.value?.ranking.images || []);
function setQuery(patch: Record<string, string>) {
  const query = { ...route.query, ...patch };
  void router.replace({ query });
}
useSeoMeta({
  title: "排行榜",
  description: "按合格浏览和私人收藏聚合的图片排行。",
});
</script>

<template>
  <GalleryPublicPage>
    <GalleryPageHeader
      title="排行榜"
      description="看看最近被反复浏览和收藏的图片，首页仍然保持随机。"
    >
      <template #actions>
        <USelect
          :model-value="windowKey"
          :items="windows"
          value-key="value"
          class="w-32"
          @update:model-value="(value) => setQuery({ window: String(value) })"
        />
      </template>
    </GalleryPageHeader>
    <div
      class="gallery-ranking-tabs mb-7 flex w-max max-w-full gap-1 overflow-x-auto rounded-xl border border-default bg-muted p-1.5"
    >
      <UButton
        v-for="item in kinds"
        :key="item.value"
        color="neutral"
        :variant="kind === item.value ? 'soft' : 'ghost'"
        :label="item.label"
        @click="setQuery({ kind: item.value })"
      />
    </div>
    <div v-if="status === 'pending'" class="gallery-grid">
      <USkeleton
        v-for="index in 15"
        :key="index"
        class="aspect-[4/3] rounded-lg"
      />
    </div>
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="排行加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <GalleryImageGrid
      v-else-if="images.length"
      :items="images"
      priority
    />
    <GalleryCompactEmpty
      v-else
      icon="i-tabler-chart-bar"
      title="还没有形成排行"
      description="有更多公开浏览和收藏后，这里会开始更新。"
    />
  </GalleryPublicPage>
</template>
