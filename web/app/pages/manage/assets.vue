<script setup lang="ts">
import { AssetRegistrationSummary } from "@yueli/asset-nuxt/components";
import { PageHeader } from "@yueli/ui/dashboard/pattern";

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "资源策略 · 图库管理" });
const { can } = useGalleryMe();
</script>

<template>
  <div class="space-y-5">
    <PageHeader title="资源策略" />
    <ClientOnly>
      <AssetRegistrationSummary expected-namespace="gallery" :profile-order="['gallery-submission']" :can-edit="can('gallery.asset_settings.manage')" />
      <template #fallback><USkeleton class="h-56 w-full rounded-xl" /></template>
    </ClientOnly>
  </div>
</template>
