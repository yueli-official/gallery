<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    compact?: boolean;
    placeholder?: string;
  }>(),
  {
    compact: false,
    placeholder: "搜索图片、主题或标签",
  },
);
const route = useRoute();
const router = useRouter();
const query = ref("");

watch(
  () => [route.query.q, route.query.tag],
  ([search, tag]) => {
    if (typeof search === "string" && search) {
      query.value = search;
      return;
    }
    query.value = typeof tag === "string" && tag ? `#${tag}` : "";
  },
  { immediate: true },
);

async function submit(): Promise<void> {
  const value = query.value.trim();
  const tag = value.startsWith("#") ? value.slice(1).trim() : "";
  await router.push({
    path: "/images",
    query: tag ? { tag } : value ? { q: value } : undefined,
  });
}
</script>

<template>
  <form
    class="gallery-global-search flex"
    :class="{ 'gallery-global-search--compact': props.compact }"
    role="search"
    @submit.prevent="submit"
  >
    <svg
      class="gallery-search-icon"
      viewBox="0 0 24 24"
      width="17"
      height="17"
      aria-hidden="true"
    >
      <circle cx="11" cy="11" r="6.5" fill="none" stroke="currentColor" stroke-width="1.8" />
      <path d="m16 16 4 4" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
    </svg>
    <input
      v-model="query"
      class="gallery-search-input"
      :placeholder="placeholder"
      aria-label="搜索图库"
    />
    <button
      type="submit"
      class="gallery-search-submit"
      aria-label="提交搜索"
    >
      <svg viewBox="0 0 24 24" width="17" height="17" aria-hidden="true">
        <path d="M5 12h14m-5-5 5 5-5 5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </button>
  </form>
</template>
