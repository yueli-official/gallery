<script setup lang="ts">
import { AdminTrendChart } from "@yueli/ui/admin";
import type { GalleryAdminTrafficPoint } from "~/types/gallery";

const props = withDefaults(defineProps<{
  points: readonly GalleryAdminTrafficPoint[];
  metric?: "views" | "favorites";
  compact?: boolean;
}>(), { metric: "views", compact: false });

const primaryMetric = computed(() => props.metric === "views");
const metricLabel = computed(() => primaryMetric.value ? "浏览" : "收藏");
const unit = computed(() => primaryMetric.value ? " 次" : " 次");

const trendPoints = computed(() => props.points.map((point) => ({
  label: point.day,
  axisLabel: point.day.slice(5).replace("-", "/"),
  value: point[props.metric],
})));
</script>

<template>
  <AdminTrendChart
    data-gallery-dashboard-trend-chart
    :points="trendPoints"
    :compact="compact"
    :metric-label="metricLabel"
    :unit="unit"
    :empty-label="`当前周期还没有${metricLabel}记录`"
  />
</template>
