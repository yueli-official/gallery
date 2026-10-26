<script setup lang="ts">
import { AssetPolicyPage } from "@yueli/asset-nuxt/components";
import { createAssetReferenceSyncAdapter } from "@yueli/asset-nuxt/references";
import { galleryFailureMessage } from "~/utils/galleryFailure";

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "资源策略 · 图库管理" });
const { can } = useGalleryMe();
const {call}=useApi('platform');
const referenceSync=createAssetReferenceSyncAdapter(call,{kindLabels:{'gallery-image':'图片','gallery-submission-image':'投稿','gallery-classification-cover':'分类','gallery-site-artwork':'站点视觉'},describeError:error=>galleryFailureMessage(error,'资源引用操作失败，请稍后重试。')});
</script>

<template>
  <AssetPolicyPage
    expected-namespace="gallery"
    :profile-order="['gallery-submission']"
    :can-edit="can('gallery.asset_settings.manage')"
    :reference-sync="referenceSync"
  />
</template>
