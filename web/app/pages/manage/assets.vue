<script setup lang="ts">
import { ManageAssetSettings } from "@platform/asset/components";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "资源设置 · 图库管理" });

const { slug, brand } = useSiteRuntime();
const { isAdmin } = useAuth();
</script>

<template>
  <ManageAssetSettings
    :site-key="slug"
    :site-name="brand"
    :can-manage="isAdmin"
    permission-description="只有具备图库资源管理权限的运营者可以修改这些规则。"
    description="管理投稿原图的私有存储、公开图片衍生版本与上传限制。日常审核不需要进入这里，只有存储策略变化或上传异常时才调整。"
    :profile-order="['gallery-submission']"
    :exclude-profiles="['gallery-artwork', 'gallery-creator-cover']"
  />
</template>
