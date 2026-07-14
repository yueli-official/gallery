<script setup lang="ts">
const { brand } = useSiteRuntime();

useHead({
  titleTemplate: (title) => (title ? `${title} · ${brand.value}` : brand.value),
});
</script>

<template>
  <UApp :toaster="{ duration: 4500 }">
    <NuxtLoadingIndicator color="var(--ui-primary)" :height="3" />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
</template>

