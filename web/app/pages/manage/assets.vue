<script setup lang="ts">
import { ManageAssetSettings } from "@platform/asset/components";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "资源与设置 · 图库管理" });

const { slug, brand } = useSiteRuntime();
const { isAdmin } = useAuth();
</script>

<template>
  <ManageAssetSettings
    :site-key="slug"
    :site-name="brand"
    :can-manage="isAdmin"
    permission-description="资源配置仅对 Catalog 声明的当前图库运营者开放。"
    description="配置投稿私有 master 与 Gallery 衍生版本。当前 Asset 只支持单格式惰性衍生；多格式预处理、公开 rendition 撤销和 Worker 状态必须在 Asset 升级后启用。"
    :profile-order="['gallery-submission']"
    :exclude-profiles="['gallery-artwork', 'gallery-creator-cover']"
  />
</template>
