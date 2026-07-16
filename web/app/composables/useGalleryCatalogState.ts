import type {
  GalleryCatalogSort,
  GalleryCatalogState,
  GalleryCatalogView,
} from "~/utils/catalog";

export function useGalleryCatalogState() {
  const route = useRoute();
  const router = useRouter();
  const state = computed(() =>
    parseGalleryCatalogState(route.query as Record<string, unknown>),
  );
  const request = computed(() => galleryCatalogRequest(state.value));
  const searchDraft = ref("");
  const selectedCategories = ref<string[]>([]);
  const selectedFacets = ref<string[]>([]);

  watch(
    state,
    (current) => {
      searchDraft.value = current.q;
      selectedCategories.value = [...current.categories];
      selectedFacets.value = [...current.facets];
    },
    { immediate: true },
  );

  async function pushState(next: GalleryCatalogState): Promise<void> {
    await router.push({ path: "/images", query: galleryCatalogQuery(next) });
  }

  function apply() {
    return pushState({
      ...state.value,
      q: searchDraft.value.trim(),
      categories: [...selectedCategories.value],
      facets: [...selectedFacets.value],
      page: 1,
      preview: "",
    });
  }

  function clear() {
    searchDraft.value = "";
    selectedCategories.value = [];
    selectedFacets.value = [];
    return pushState({
      ...state.value,
      q: "",
      categories: [],
      facets: [],
      tag: "",
      page: 1,
      preview: "",
    });
  }

  function toggleCategory(slug: string) {
    selectedCategories.value = selectedCategories.value.includes(slug)
      ? selectedCategories.value.filter((item) => item !== slug)
      : [...selectedCategories.value, slug];
  }

  function toggleFacet(value: string) {
    selectedFacets.value = selectedFacets.value.includes(value)
      ? selectedFacets.value.filter((item) => item !== value)
      : [...selectedFacets.value, value];
  }

  function setPage(page: number) {
    return pushState({ ...state.value, page: Math.max(1, page), preview: "" });
  }

  function setSort(sort: GalleryCatalogSort) {
    return pushState({ ...state.value, sort, page: 1, preview: "" });
  }

  function setView(view: GalleryCatalogView) {
    return pushState({ ...state.value, view, preview: "" });
  }

  function setPreview(preview: string) {
    return pushState({ ...state.value, preview });
  }

  function removeSearch() {
    searchDraft.value = "";
    return apply();
  }

  function removeCategory(slug: string) {
    selectedCategories.value = selectedCategories.value.filter(
      (item) => item !== slug,
    );
    return apply();
  }

  function removeFacet(value: string) {
    selectedFacets.value = selectedFacets.value.filter(
      (item) => item !== value,
    );
    return apply();
  }

  function removeTag() {
    return pushState({ ...state.value, tag: "", page: 1, preview: "" });
  }

  return {
    state,
    request,
    searchDraft,
    selectedCategories,
    selectedFacets,
    hasFilters: computed(() => galleryCatalogFilterCount(state.value) > 0),
    activeFilterCount: computed(() => galleryCatalogFilterCount(state.value)),
    apply,
    clear,
    toggleCategory,
    toggleFacet,
    setPage,
    setSort,
    setView,
    setPreview,
    removeSearch,
    removeCategory,
    removeFacet,
    removeTag,
  };
}
