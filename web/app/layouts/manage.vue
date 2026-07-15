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
  <ManageShell :site-name="brand" :context-label="labels[route.path] || '图库管理'" storage-key="gallery-manage" home-to="/manage">
    <template #sidebar><ManageSidebar /></template>
    <template #user><ConsumerManageAccountControl home-to="/" /></template>
    <slot />
  </ManageShell>
</template>
