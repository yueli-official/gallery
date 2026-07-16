<script setup lang="ts">
import { ManageHeader, SkeletonList } from "@platform/manage/components";
import type { GalleryDiscovery } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "发现策略 · 图库管理" });
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
    <ManageHeader title="发现策略">
      <template #subtitle>
        检查首页探索与排行榜是否各自完成任务，不在这里手工干预单张图片。
      </template>
    </ManageHeader>
    <SkeletonList v-if="!hydrated || pending" :rows="5" />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="发现策略加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <div v-else-if="data" class="space-y-6">
      <section
        class="overflow-hidden rounded-xl border border-default bg-default"
      >
        <div class="grid lg:grid-cols-[minmax(0,1.2fr)_minmax(20rem,.8fr)]">
          <div
            class="grid h-80 grid-cols-4 grid-rows-2 gap-1 bg-elevated p-1 sm:h-96"
          >
            <div
              v-for="(item, index) in data.images.slice(0, 4)"
              :key="item.id"
              class="overflow-hidden bg-muted"
              :class="
                index === 0
                  ? 'col-span-2 row-span-2'
                  : index === 1
                    ? 'col-span-2'
                    : ''
              "
            >
              <img
                v-bind="galleryImageSources(item.assetId, 'grid', false)"
                :alt="item.altText || item.title"
                class="size-full object-cover"
              />
            </div>
          </div>
          <div class="flex flex-col justify-between p-5 sm:p-7">
            <div>
              <div
                class="flex items-center gap-2 text-xs font-semibold text-success"
              >
                <span class="size-2 rounded-full bg-success" />
                当前策略可用
              </div>
              <p
                class="mt-5 text-xs font-semibold uppercase tracking-[.14em] text-primary"
              >
                Homepage exploration
              </p>
              <h2
                class="mt-2 text-2xl font-semibold tracking-tight text-highlighted"
              >
                多样性优先的首页探索
              </h2>
              <p class="mt-3 text-sm leading-6 text-muted">
                同一 seed
                保持结果稳定，同时避免单一分类和相似画面占满首屏。热度不会改变这条探索流。
              </p>
              <dl
                class="mt-6 grid grid-cols-3 gap-3 border-y border-default py-4"
              >
                <div>
                  <dt class="text-xs text-dimmed">首批</dt>
                  <dd
                    class="mt-1 text-lg font-semibold tabular-nums text-highlighted"
                  >
                    {{ data.site.randomBatchSize }}
                  </dd>
                </div>
                <div>
                  <dt class="text-xs text-dimmed">候选池</dt>
                  <dd
                    class="mt-1 text-lg font-semibold tabular-nums text-highlighted"
                  >
                    {{ data.site.randomCandidateSize }}
                  </dd>
                </div>
                <div>
                  <dt class="text-xs text-dimmed">维度轴</dt>
                  <dd
                    class="mt-1 text-lg font-semibold tabular-nums text-highlighted"
                  >
                    {{ data.facets.length }}
                  </dd>
                </div>
              </dl>
            </div>
            <div class="mt-6 flex flex-wrap items-center gap-3">
              <UButton
                to="/?seed=operator-preview"
                target="_blank"
                icon="i-tabler-external-link"
                label="打开这批首页"
              />
              <span class="max-w-48 truncate font-mono text-xs text-dimmed"
                >seed {{ data.seed }}</span
              >
            </div>
          </div>
        </div>
      </section>

      <section
        class="grid gap-5 rounded-xl border border-default bg-default p-5 lg:grid-cols-[minmax(0,1fr)_17rem] lg:items-center lg:p-6"
      >
        <div class="min-w-0">
          <p
            class="text-xs font-semibold uppercase tracking-[.14em] text-primary"
          >
            Rankings
          </p>
          <h2 class="mt-2 text-lg font-semibold text-highlighted">
            排行榜只回答“最近值得回看什么”
          </h2>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-muted">
            浏览、唯一访客、收藏与分享按时间窗口聚合。举报只进入信任处理单，不公开展示，也不会仅凭次数自动下架。
          </p>
        </div>
        <div class="flex flex-wrap items-center gap-2 lg:justify-end">
          <span class="rounded-md bg-elevated px-2.5 py-1.5 text-xs text-muted"
            >默认 7 天</span
          >
          <UButton
            to="/rankings"
            target="_blank"
            color="neutral"
            variant="outline"
            icon="i-tabler-external-link"
            label="查看公开排行"
          />
        </div>
      </section>
    </div>
  </div>
</template>
