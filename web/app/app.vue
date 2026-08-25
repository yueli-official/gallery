<script setup lang="ts">
import { FeedbackToastRegion } from "@yueli/ui/feedback/pattern";

const { brand } = useSiteRuntime();
const hydrated = useClientHydrated();

useHead({
  htmlAttrs: { lang: "zh-CN" },
  titleTemplate: (title) => (title ? `${title} · ${brand.value}` : brand.value),
});
</script>

<template>
  <div
    class="contents"
    :data-gallery-hydrated="hydrated ? 'true' : 'false'"
  >
    <UApp :toaster="null">
      <NuxtLoadingIndicator color="var(--ui-primary)" :height="3" />
      <NuxtLayout>
        <NuxtPage />
      </NuxtLayout>
      <FeedbackToastRegion />
    </UApp>
  </div>
</template>

