import type { ComputedRef } from "vue";
import type { GalleryImageCard } from "~/types/gallery";

export function useGalleryQuickView(items: ComputedRef<GalleryImageCard[]>) {
  const route = useRoute();
  const router = useRouter();
  const preview = computed(() => String(route.query.preview || ""));
  const trigger = shallowRef<HTMLElement>();
  const scrollY = ref(0);

  async function setPreview(imageId: string): Promise<void> {
    const query = { ...route.query };
    if (imageId) query.preview = imageId;
    else delete query.preview;
    await router.replace({ query });
  }

  function openPreview(imageId: string, element: HTMLElement | null): void {
    trigger.value = element || undefined;
    scrollY.value = import.meta.client ? window.scrollY : 0;
    void setPreview(imageId);
  }

  async function closePreview(): Promise<void> {
    await setPreview("");
    if (!import.meta.client) return;
    await nextTick();
    window.scrollTo({ top: scrollY.value });
    trigger.value?.focus({ preventScroll: true });
  }

  const activeItems = computed(() => items.value);
  return {
    preview,
    items: activeItems,
    openPreview,
    closePreview,
    navigatePreview: setPreview,
  };
}
