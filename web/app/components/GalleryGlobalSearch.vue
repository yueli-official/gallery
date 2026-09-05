<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    compact?: boolean;
    placeholder?: string;
  }>(),
  {
    compact: false,
  },
);
const route = useRoute();
const router = useRouter();
const gallerySite = useGallerySiteSettings();
const query = ref("");
const resolvedPlaceholder = computed(
  () =>
    props.placeholder ||
    gallerySite.value?.searchPlaceholder ||
    "搜索图库",
);

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
    class="gallery-global-search flex w-full max-w-[46rem] items-center gap-1 rounded-xl border border-default bg-muted px-[0.22rem] py-[0.18rem] shadow-none focus-within:border-primary/55"
    role="search"
    @submit.prevent="submit"
  >
    <svg
      class="gallery-search-icon ml-2.5 shrink-0 text-muted"
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
      class="gallery-search-input min-w-0 flex-1 border-0 bg-transparent px-[0.3rem] py-2 text-[0.84rem] text-highlighted outline-none placeholder:text-dimmed"
      :placeholder="resolvedPlaceholder"
      aria-label="搜索图库"
    />
    <button
      type="submit"
      class="gallery-search-submit grid size-[2.2rem] shrink-0 place-items-center rounded-lg text-muted transition-colors hover:bg-default hover:text-primary focus-visible:bg-default focus-visible:text-primary focus-visible:outline-2 focus-visible:outline-transparent"
      aria-label="提交搜索"
    >
      <svg viewBox="0 0 24 24" width="17" height="17" aria-hidden="true">
        <path d="M5 12h14m-5-5 5 5-5 5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </button>
  </form>
</template>
