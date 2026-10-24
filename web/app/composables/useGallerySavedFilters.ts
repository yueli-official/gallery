import type { GallerySavedFilter } from "~/types/presentation";

export function useGallerySavedFilters() {
  const { loggedIn } = useAuth();
  const { call } = useGalleryApi();
  const items = useState<GallerySavedFilter[]>("gallery-saved-filters", () => []);
  const error = useState("gallery-saved-filters-error", () => "");
  async function refresh() {
    if (!loggedIn.value) { items.value = []; return; }
    error.value = "";
    try { items.value = (await call<{ items: GallerySavedFilter[] }>("/me/saved-filters")).items; }
    catch (reason) { error.value = galleryFailureMessage(reason, "常用筛选加载失败"); }
  }
  async function remove(id: string) {
    await call(`/me/saved-filters/${encodeURIComponent(id)}`, { method: "DELETE" });
    items.value = items.value.filter(item => item.id !== id);
  }
  return { items, error, refresh, remove };
}
