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
    class="gallery-global-search"
    role="search"
    @submit.prevent="submit"
  >
    <svg
      class="gallery-search-icon ml-2 shrink-0 text-[var(--gallery-muted)]"
      viewBox="0 0 24 24"
      width="15"
      height="15"
      aria-hidden="true"
    >
      <circle cx="11" cy="11" r="6.5" fill="none" stroke="currentColor" stroke-width="1.6" />
      <path d="m16 16 4 4" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
    </svg>
    <input
      v-model="query"
      type="search"
      class="gallery-search-input"
      :placeholder="resolvedPlaceholder"
      aria-label="搜索图库"
    />
    <button
      type="submit"
      class="gallery-search-submit grid size-[1.65rem] shrink-0 place-items-center text-[var(--gallery-muted)] transition-colors hover:text-[var(--gallery-ink)] focus-visible:outline-2 focus-visible:outline-[var(--gallery-accent)]"
      aria-label="提交搜索"
    >
      <svg viewBox="0 0 24 24" width="15" height="15" aria-hidden="true">
        <path d="M5 12h14m-5-5 5 5-5 5" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </button>
  </form>
</template>

<style scoped>
.gallery-global-search{display:flex;align-items:center;gap:.25rem;width:100%;max-width:54rem;min-width:0;padding:.1rem .15rem;border:1px solid var(--gallery-line);border-radius:.4rem;background:var(--gallery-panel-strong);transition:border-color .16s,box-shadow .16s}
.gallery-global-search:focus-within{border-color:var(--gallery-accent);box-shadow:0 0 0 3px var(--gallery-accent-soft)}
.gallery-search-input{flex:1;min-width:0;border:0;border-radius:0;background:transparent;padding:.55rem .45rem;color:var(--gallery-ink);font-size:.74rem;line-height:1.5}
.gallery-search-input::placeholder{color:var(--gallery-muted)}
.gallery-search-input:focus-visible{outline:none;box-shadow:none}
.gallery-search-submit{border-radius:.2rem}
</style>
