<script setup lang="ts">
const route = useRoute();
const router = useRouter();
const query = ref("");

watch(
  () => route.query.q,
  (value) => {
    query.value = typeof value === "string" ? value : "";
  },
  { immediate: true },
);

async function submit(): Promise<void> {
  const value = query.value.trim();
  await router.push({
    path: "/images",
    query: value ? { q: value } : undefined,
  });
}
</script>

<template>
  <form
    class="gallery-global-search flex"
    role="search"
    @submit.prevent="submit"
  >
    <UInput
      v-model="query"
      class="min-w-0 flex-1"
      icon="i-tabler-search"
      placeholder="搜索图片、主题或标签"
      aria-label="搜索图库"
    />
    <UButton
      type="submit"
      color="neutral"
      variant="ghost"
      icon="i-tabler-arrow-right"
      aria-label="提交搜索"
    />
  </form>
</template>
