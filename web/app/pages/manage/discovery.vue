<script setup lang="ts">
import { ManageHeader, SkeletonList } from "@platform/manage/components";
import type { GalleryDiscovery } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "随机与排行 · 图库管理" });
const { call } = useApi();
const hydrated = useClientHydrated();
const { data, pending, error, refresh } = await useAsyncData(
  "gallery-manage-discovery",
  () =>
    call<GalleryDiscovery>("/api/v1/gallery/discovery?seed=operator-preview"),
  { server: false },
);
</script>

<template>
  <div>
    <ManageHeader title="随机与排行"
      ><template #subtitle
        >首页随机与排行是两套策略：随机优先多样性，排行只帮助回看，不互相污染。</template
      ></ManageHeader
    >
    <SkeletonList v-if="!hydrated || pending" :rows="5" />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="发现策略加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <div v-else-if="data" class="grid gap-5 lg:grid-cols-2">
      <section class="rounded-lg border border-default bg-default p-5">
        <p
          class="text-xs font-semibold uppercase tracking-[.14em] text-primary"
        >
          Random discovery
        </p>
        <h2 class="mt-2 text-lg font-semibold text-highlighted">随机首页</h2>
        <dl class="mt-5 space-y-3 text-sm">
          <div class="flex justify-between gap-4">
            <dt class="text-muted">每批图片</dt>
            <dd class="tabular-nums">{{ data.site.randomBatchSize }}</dd>
          </div>
          <div class="flex justify-between gap-4">
            <dt class="text-muted">候选池</dt>
            <dd class="tabular-nums">{{ data.site.randomCandidateSize }}</dd>
          </div>
          <div class="flex justify-between gap-4">
            <dt class="text-muted">本次 seed</dt>
            <dd class="max-w-60 truncate font-mono text-xs">{{ data.seed }}</dd>
          </div>
        </dl>
        <p class="mt-5 text-sm leading-6 text-muted">
          eligible pool → deterministic hash score → Facet diversity rerank →
          seed cache。不是随机 offset，也不按热度。
        </p>
        <UButton
          to="/?seed=operator-preview"
          target="_blank"
          class="mt-4"
          color="neutral"
          variant="outline"
          icon="i-tabler-external-link"
          label="预览这批"
        />
      </section>
      <section class="rounded-lg border border-default bg-default p-5">
        <p
          class="text-xs font-semibold uppercase tracking-[.14em] text-primary"
        >
          Rankings
        </p>
        <h2 class="mt-2 text-lg font-semibold text-highlighted">排行榜</h2>
        <p class="mt-5 text-sm leading-6 text-muted">
          合格详情浏览、唯一访客、收藏和分享通过事件与日报聚合；Image
          热行不自增。趋势默认 7 天，并支持 24h / 30d / all-time。
        </p>
        <p class="mt-3 text-sm leading-6 text-muted">
          举报指标只进入 Case 和风控，不对公众显示，也不会因次数自动下架。
        </p>
        <UButton
          to="/rankings"
          target="_blank"
          class="mt-4"
          color="neutral"
          variant="outline"
          icon="i-tabler-external-link"
          label="查看公开排行"
        />
      </section>
    </div>
  </div>
</template>
