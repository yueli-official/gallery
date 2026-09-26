<script setup lang="ts">
import DashboardTrendChart from "~/components/DashboardTrendChart.vue";
import { ManagePage, AdminOverview, AdminMetricCard } from "@yueli/ui/admin";
import type { GalleryAdminOverview } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "控制台 · 图库管理" });

const { can } = useGalleryMe();
const { call } = useGalleryApi();
const hydrated = useClientHydrated();
const period = ref(14);
const periodItems = [
  { label: "7 天", value: 7 },
  { label: "14 天", value: 14 },
  { label: "30 天", value: 30 },
];

const { data, pending, error, refresh } = await useAsyncData(
  "gallery-admin-overview",
  () =>
    call<{ overview: GalleryAdminOverview }>("/admin/overview", {
      query: { days: period.value },
    }),
  {
    server: false,
    watch: [period],
    default: () => ({
      overview: {
        pendingSubmissions: 0,
        openCases: 0,
        publishedImages: 0,
        failedProcessing: 0,
        days: period.value,
        allTimeViews: 0,
        periodViews: 0,
        previousPeriodViews: 0,
        periodFavorites: 0,
        previousPeriodFavorites: 0,
        series: [],
        topImages: [],
      },
    }),
  },
);
const overview = computed(() => data.value.overview);
const formatter = new Intl.NumberFormat("zh-CN");
const topPeak = computed(() =>
  Math.max(1, ...overview.value.topImages.map((image) => image.views)),
);
const favoriteRate = computed(() =>
  overview.value.periodViews
    ? (overview.value.periodFavorites / overview.value.periodViews) * 100
    : 0,
);

function formatNumber(value: number) {
  return formatter.format(value || 0);
}

function comparison(current: number, previous: number) {
  if (!previous)
    return current ? `本期新增 ${formatNumber(current)}` : "与前期持平";
  const percent = Math.round(((current - previous) / previous) * 100);
  return `${percent >= 0 ? "+" : ""}${percent}% 较前 ${period.value} 天`;
}

const metricCards = computed(() => [
  {
    label: "累计浏览",
    value: formatNumber(overview.value.allTimeViews),
    detail: `${formatNumber(overview.value.publishedImages)} 张公开图片`,
    icon: "i-tabler-eye",
    to: "/manage/images?status=published",
  },
  {
    label: `近 ${period.value} 天浏览`,
    value: formatNumber(overview.value.periodViews),
    detail: comparison(
      overview.value.periodViews,
      overview.value.previousPeriodViews,
    ),
    icon: "i-tabler-chart-line",
    to: "/manage",
  },
  {
    label: "收藏率",
    value: `${favoriteRate.value.toFixed(1)}%`,
    detail: `${formatNumber(overview.value.periodFavorites)} 次收藏`,
    icon: "i-tabler-heart",
    to: "/manage/images",
  },
  {
    label: "公开图片",
    value: formatNumber(overview.value.publishedImages),
    detail: "当前可浏览目录",
    icon: "i-tabler-photo",
    to: "/manage/images?status=published",
  },
]);
</script>

<template>
  <ManagePage
    id="dashboard"
    title="控制台"
    icon="i-tabler-dashboard"
    data-gallery-dashboard-analytics
  >
    <template v-if="can('gallery.dashboard.read') && !error" #tools>
      <AdminOverview>
        <template #artwork><ManageOverviewArtwork /></template>
        <div v-if="!hydrated || (pending && !overview.series.length)" data-admin-metrics><USkeleton v-for="n in 4" :key="n" class="h-24 rounded-xl" /></div>
        <div v-else data-admin-metrics><AdminMetricCard v-for="card in metricCards" :key="card.label" :label="card.label" :value="card.value" :icon="card.icon" :detail="card.detail" :to="card.to" data-gallery-dashboard-metric /></div>
      </AdminOverview>
    </template>

    <UAlert
      v-if="!can('gallery.dashboard.read')"
      color="error"
      variant="subtle"
      icon="i-tabler-lock"
      title="当前账号没有图库管理权限"
    />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      icon="i-tabler-alert-circle"
      title="统计暂时不可用"
      description="图片管理不受影响；重试后仍失败时再检查 Gallery API。"
    >
      <template #actions>
        <UButton
          color="error"
          variant="soft"
          icon="i-tabler-refresh"
          label="重试"
          @click="refresh()"
        />
      </template>
    </UAlert>

    <template v-else-if="can('gallery.dashboard.read')">
      <div
        class="grid items-start gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(18rem,1fr)]"
      >
        <UCard
          variant="soft"
          class="divide-y-0 bg-default shadow-sm"
          data-gallery-dashboard-trend
          :aria-busy="pending"
          :ui="{
            header: 'p-5 sm:p-6',
            body: 'px-5 pb-5 pt-0 sm:px-6 sm:pb-6 sm:pt-0',
          }"
        >
          <template #header>
            <div
              class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
            >
              <h2 class="text-base font-semibold text-highlighted">浏览趋势</h2>
              <UTabs
                v-model="period"
                :items="periodItems"
                :content="false"
                :disabled="!hydrated || pending"
                size="sm"
                class="w-fit"
                aria-label="统计时间范围"
              />
            </div>
          </template>
          <DashboardTrendChart :points="overview.series" />
        </UCard>

        <UCard
          title="收藏趋势"
          variant="soft"
          class="divide-y-0 bg-default shadow-sm"
          :ui="{
            header: 'p-5 sm:p-6',
            body: 'space-y-5 px-5 pb-5 pt-0 sm:px-6 sm:pb-6 sm:pt-0',
          }"
        >
          <div class="flex items-end justify-between gap-4">
            <div>
              <p
                class="text-3xl font-semibold tabular-nums tracking-tight text-highlighted"
              >
                {{ formatNumber(overview.periodFavorites) }}
              </p>
              <p class="mt-1 text-xs text-muted">次收藏</p>
            </div>
            <UBadge
              color="neutral"
              variant="soft"
              :label="comparison(overview.periodFavorites, overview.previousPeriodFavorites)"
            />
          </div>
          <DashboardTrendChart
            :points="overview.series"
            metric="favorites"
            compact
          />
        </UCard>
      </div>

      <UCard
        title="热门图片"
        variant="soft"
        class="divide-y-0 bg-default shadow-sm"
        :ui="{ header: 'p-5 sm:p-6', body: 'p-0 sm:p-0' }"
      >
        <div v-if="overview.topImages.length" class="divide-y divide-default">
          <div
            v-for="image in overview.topImages"
            :key="image.id"
            class="grid grid-cols-[minmax(0,1fr)_6rem_auto] items-center gap-3 px-5 py-3 sm:grid-cols-[minmax(0,1fr)_6rem_6rem_auto] sm:px-6"
          >
            <div class="min-w-0">
              <p class="truncate text-sm font-medium text-highlighted">
                {{ image.title }}
              </p>
              <UProgress
                class="mt-2"
                :model-value="image.views"
                :max="topPeak"
                size="2xs"
              />
            </div>
            <div class="text-right">
              <p class="font-semibold tabular-nums text-highlighted">
                {{ formatNumber(image.views) }}
              </p>
              <p class="text-xs text-dimmed">浏览</p>
            </div>
            <div class="hidden text-right sm:block">
              <p class="font-semibold tabular-nums text-highlighted">
                {{ formatNumber(image.favorites) }}
              </p>
              <p class="text-xs text-dimmed">收藏</p>
            </div>
            <div class="flex justify-end gap-1">
              <UTooltip text="查看公开图片">
                <UButton
                  :to="`/images/${image.id}`"
                  target="_blank"
                  rel="noopener"
                  icon="i-tabler-external-link"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  square
                  :aria-label="`查看公开图片：${image.title}`"
                />
              </UTooltip>
              <UTooltip text="管理图片">
                <UButton
                  :to="{ path: '/manage/images', query: { q: image.title } }"
                  icon="i-tabler-pencil"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  square
                  :aria-label="`管理图片：${image.title}`"
                />
              </UTooltip>
            </div>
          </div>
        </div>
        <p v-else class="p-8 text-center text-sm text-muted">
          当前周期还没有有效浏览记录。
        </p>
      </UCard>
    </template>
  </ManagePage>
</template>
