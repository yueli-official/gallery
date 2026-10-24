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
const images = computed(() => (data.value?.ranking.images || []).slice(0, 24));
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
  <GalleryPublicPage class="rankings-page">
    <GalleryArchiveHero
      kicker="月离图库 / RANKING"
      title="排行榜"
      description="发现最近受到关注的图片。"
    >
      <template #actions>
        <USelect
          :model-value="windowKey"
          :items="windows"
          value-key="value"
          class="w-36"
          aria-label="排行时间范围"
          @update:model-value="(value) => setQuery({ window: String(value) })"
        />
      </template>
    </GalleryArchiveHero>

    <div class="ranking-toolbar">
    <div class="ranking-tabs" role="group" aria-label="排行类型">
      <button
        v-for="item in kinds"
        :key="item.value"
        type="button"
        class="ranking-tab"
        :class="{ 'is-active': kind === item.value }"
        :aria-pressed="kind === item.value"
        @click="setQuery({ kind: item.value })"
      >
        <UIcon :name="item.value === 'trending' ? 'i-tabler-trending-up' : item.value === 'most_favorited' ? 'i-tabler-heart' : 'i-tabler-eye'" />{{ item.label }}
      </button>
    </div>
    <span class="ranking-total font-numeric">TOP {{ images.length }}</span>
    </div>

    <div v-if="status === 'pending'" class="gallery-grid">
      <USkeleton
        v-for="index in 15"
        :key="index"
        class="aspect-[3/4] rounded-[0.9rem]"
      />
    </div>
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="排行加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <GalleryImageGrid v-else-if="images.length" :items="images" view="grid" priority ranked />
    <GalleryCompactEmpty
      v-else
      icon="i-tabler-chart-bar"
      title="还没有形成排行"
      description="有更多公开浏览和收藏后，这里会开始更新。"
    />
  </GalleryPublicPage>
</template>

<style scoped>
.ranking-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .75rem;
  margin-bottom: 1.1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--gallery-line);
}
.ranking-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: .2rem;
  padding: .25rem;
  border: 1px solid var(--gallery-line);
  border-radius: .5rem;
  background: var(--gallery-panel);
}
.ranking-total{font-size:.67rem;letter-spacing:.08em;color:var(--gallery-muted);white-space:nowrap}
.ranking-tab {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: .4rem;
  flex: none;
  min-height: 2rem;
  padding: .35rem .8rem;
  border: 1px solid transparent;
  border-radius: .3rem;
  font-size: 0.72rem;
  font-weight: 550;
  color: var(--gallery-muted);
  transition:
    background 160ms ease,
    color 160ms ease,
    border-color 160ms ease;
}

.ranking-tab:hover {
  color: var(--gallery-ink);
  background: var(--gallery-accent-soft);
}

.ranking-tab.is-active {
  border-color: var(--gallery-line);
  background: var(--gallery-panel-strong);
  color: var(--gallery-accent);
  box-shadow: 0 1px 3px rgb(0 0 0 / .08);
}

.ranking-tab:focus-visible {
  outline: 2px solid var(--gallery-accent);
  outline-offset: 2px;
}

@media (max-width: 500px) {
  .ranking-toolbar{gap:.4rem}
  .ranking-tabs{flex:1}
  .ranking-tab {
    flex:1;
    padding: .35rem .45rem;
    font-size: 0.68rem;
  }
  .ranking-total{display:none}
}
</style>
