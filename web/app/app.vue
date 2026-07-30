<script setup lang="ts">
const { brand } = useSiteRuntime();
const hydrated = useClientHydrated();

useHead({
  titleTemplate: (title) => (title ? `${title} · ${brand.value}` : brand.value),
});
</script>

<template>
  <div
    class="contents"
    :data-gallery-hydrated="hydrated ? 'true' : 'false'"
  >
    <UApp
      :toaster="{
        duration: 4500,
        progress: false,
        max: 3,
        position: 'bottom-right',
      }"
    >
      <NuxtLoadingIndicator color="var(--ui-primary)" :height="3" />
      <NuxtLayout>
        <NuxtPage />
      </NuxtLayout>
    </UApp>
  </div>
</template>

