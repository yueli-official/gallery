<script setup lang="ts">
import { ManageDashboardLayout } from "@platform/manage/components";
import type { GalleryAdminOverview } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "图库控制台" });
const { user, isAdmin } = useAuth();
const { call } = useApi();
const hydrated = useClientHydrated();
const { data, pending, error, refresh } = await useAsyncData(
  "gallery-admin-overview",
  () =>
    call<{ overview: GalleryAdminOverview }>("/api/v1/gallery/admin/overview"),
  { server: false },
);
const overview = computed(() => data.value?.overview);
const cards = computed(() => [
  {
    label: "待审投稿",
    value: overview.value?.pendingSubmissions ?? 0,
    icon: "i-tabler-photo-check",
    to: "/manage/submissions",
  },
  {
    label: "待处理 Case",
    value: overview.value?.openCases ?? 0,
    icon: "i-tabler-flag",
    to: "/manage/cases",
  },
  {
    label: "公开图片",
    value: overview.value?.publishedImages ?? 0,
    icon: "i-tabler-photo",
    to: "/manage/images",
  },
  {
    label: "处理失败",
    value: overview.value?.failedProcessing ?? 0,
    icon: "i-tabler-alert-triangle",
    to: "/manage/submissions",
  },
]);
</script>

<template>
  <ManageDashboardLayout
    title="控制台"
    :description="`你好，${user?.name || user?.email || '运营者'}。这里集中处理投稿、图片、来源修正与发现策略。`"
    pending-title="待处理队列"
    recent-title="运营入口"
  >
    <template #metrics>
      <div
        v-if="!hydrated || pending"
        class="grid grid-cols-2 gap-3 lg:grid-cols-4"
      >
        <USkeleton v-for="item in 4" :key="item" class="h-24 rounded-lg" />
      </div>
      <div v-else class="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <NuxtLink
          v-for="card in cards"
          :key="card.label"
          :to="card.to"
          class="rounded-lg border border-default bg-default p-4 transition hover:border-primary/40"
          ><div class="flex items-center gap-3">
            <span
              class="grid size-10 place-items-center rounded-md bg-primary/10 text-primary"
              ><UIcon :name="card.icon" class="size-5"
            /></span>
            <div>
              <p class="text-2xl font-semibold tabular-nums text-highlighted">
                {{ card.value }}
              </p>
              <p class="text-xs text-muted">{{ card.label }}</p>
            </div>
          </div></NuxtLink
        >
      </div>
    </template>
    <template #pending>
      <div
        v-if="overview?.pendingSubmissions || overview?.openCases"
        class="grid gap-3 sm:grid-cols-2"
      >
        <UButton
          v-if="overview.pendingSubmissions"
          to="/manage/submissions"
          color="neutral"
          variant="soft"
          icon="i-tabler-photo-check"
          :label="`${overview.pendingSubmissions} 条投稿待审核`"
          block
        /><UButton
          v-if="overview.openCases"
          to="/manage/cases"
          color="neutral"
          variant="soft"
          icon="i-tabler-flag"
          :label="`${overview.openCases} 个处理单待跟进`"
          block
        />
      </div>
      <UAlert
        v-else
        color="success"
        variant="subtle"
        icon="i-tabler-circle-check"
        title="当前没有待处理事项"
      />
    </template>
    <template #recent
      ><div class="grid gap-2 p-4 sm:grid-cols-2">
        <UButton
          to="/manage/images"
          color="neutral"
          variant="outline"
          icon="i-tabler-photo"
          label="管理已公开图片"
          block
        /><UButton
          to="/manage/collections"
          color="neutral"
          variant="outline"
          icon="i-tabler-folders"
          label="检查专题集合"
          block
        /><UButton
          to="/manage/classification"
          color="neutral"
          variant="outline"
          icon="i-tabler-category"
          label="查看分类与维度"
          block
        /><UButton
          to="/manage/discovery"
          color="neutral"
          variant="outline"
          icon="i-tabler-arrows-random"
          label="查看发现策略"
          block
        /></div
    ></template>
    <template #health
      ><UAlert
        v-if="!isAdmin"
        color="error"
        variant="subtle"
        icon="i-tabler-lock"
        title="当前账号没有图库管理权限" /><UAlert
        v-else-if="error"
        color="error"
        variant="subtle"
        icon="i-tabler-alert-circle"
        title="图库服务暂时不可用"
        ><template #actions
          ><UButton label="重试" @click="refresh()" /></template></UAlert
      ><UAlert
        v-else
        color="success"
        variant="subtle"
        icon="i-tabler-circle-check"
        title="图库 API 可用"
    /></template>
    <template #quickActions
      ><div class="grid gap-2">
        <UButton
          to="/manage/submissions"
          icon="i-tabler-photo-check"
          label="审核投稿"
          color="neutral"
          variant="soft"
          block
        /><UButton
          to="/manage/cases"
          icon="i-tabler-flag"
          label="处理举报与来源"
          color="neutral"
          variant="soft"
          block
        /><UButton
          to="/manage/assets"
          icon="i-tabler-settings"
          label="资源与设置"
          color="neutral"
          variant="soft"
          block
        /><UButton
          to="/"
          icon="i-tabler-external-link"
          label="查看站点"
          color="neutral"
          variant="ghost"
          block
        /></div
    ></template>
  </ManageDashboardLayout>
</template>
