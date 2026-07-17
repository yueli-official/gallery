<script setup lang="ts">
import { ManageShell } from "@platform/manage/components";

const route = useRoute();
const { brand } = useSiteRuntime();
const labels: Record<string, string> = {
  "/manage": "控制台",
  "/manage/images": "图片",
  "/manage/submissions": "投稿与审核",
  "/manage/collections": "专题集合",
  "/manage/classification": "分类与维度",
  "/manage/cases": "处理单",
  "/manage/discovery": "随机与排行",
  "/manage/assets": "资源与设置",
};
</script>

<template>
  <ManageShell
    :site-name="brand"
    :context-label="labels[route.path] || '图库管理'"
    storage-key="gallery-manage"
    home-to="/manage"
  >
    <template #sidebar><ManageSidebar /></template>
    <template #topbar-actions>
      <UButton
        to="/"
        target="_blank"
        color="neutral"
        variant="ghost"
        icon="i-tabler-external-link"
        aria-label="在新标签页打开公开站点"
      >
        <span class="hidden sm:inline">公开站点</span>
      </UButton>
    </template>
    <template #user><ConsumerManageAccountControl home-to="/" /></template>
    <slot />
  </ManageShell>
</template>
