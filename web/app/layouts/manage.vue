<script setup lang="ts">
import { ManageShell } from "@platform/manage/components";

const route = useRoute();
const { brand } = useSiteRuntime();
const labels: Record<string, string> = {
  "/manage/reviews": "作品审核",
  "/manage/creators": "创作者准入",
  "/manage/assets": "资源配置",
};
</script>

<template>
  <ManageShell :site-name="brand" :context-label="labels[route.path] || '图库管理'" storage-key="gallery-manage" home-to="/manage/reviews">
    <template #sidebar><ManageSidebar /></template>
    <template #user><ConsumerManageAccountControl home-to="/" /></template>
    <slot />
  </ManageShell>
</template>
