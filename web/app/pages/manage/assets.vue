<script setup lang="ts">
import { AssetPolicyPage } from "@yueli/asset-nuxt/components";

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "资源策略 · 图库管理" });
const { can } = useGalleryMe();
</script>

<template>
  <AssetPolicyPage
    expected-namespace="gallery"
    :profile-order="['gallery-submission']"
    :can-edit="can('gallery.asset_settings.manage')"
  />
</template>
