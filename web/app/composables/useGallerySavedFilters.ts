import type { GallerySavedFilter } from "~/types/presentation";

const storageKey = "gallery-saved-filters";

function readLocalFilters(): GallerySavedFilter[] {
  const raw = localStorage.getItem(storageKey);
  if (!raw) return [];
  const parsed: unknown = JSON.parse(raw);
  if (!Array.isArray(parsed)) throw new Error("Invalid saved filters");
  return parsed.filter((item): item is GallerySavedFilter =>
    typeof item?.id === "string" && typeof item.name === "string" &&
    item.query && typeof item.query === "object",
  );
}

export function useGallerySavedFilters() {
  const { user } = useAuth();
  const { call } = useGalleryApi();
  const localItems = useState<GallerySavedFilter[]>("gallery-local-filters", () => []);
  const accountItems = useState<GallerySavedFilter[]>("gallery-account-filters", () => []);
  const accountOwner = useState("gallery-account-filter-owner", () => "");
  const accountLoaded = useState("gallery-account-filters-loaded", () => false);
  const localError = useState("gallery-local-filters-error", () => "");
  const accountError = useState("gallery-account-filters-error", () => "");
  const owner = computed(() => user.value?.sub || "");
  const canSaveToAccount = computed(() => Boolean(owner.value) && owner.value === accountOwner.value && accountLoaded.value);

  function refreshLocal() {
    localError.value = "";
    try { localItems.value = readLocalFilters(); }
    catch { localError.value = "无法读取这个浏览器中的筛选。"; }
  }

  async function refreshAccount(subject: string, isCurrent: () => boolean) {
    accountOwner.value = subject;
    accountLoaded.value = false;
    accountItems.value = [];
    accountError.value = "";
    if (!subject) return;
    try {
      const result = await call<{ items: GallerySavedFilter[] }>("/me/saved-filters");
      if (!isCurrent()) return;
      accountItems.value = result.items;
      accountLoaded.value = true;
    } catch (reason) {
      if (!isCurrent()) return;
      const code = galleryFailureCode(reason);
      if (["gallery.forbidden", "gallery.not_authenticated", "common.forbidden", "common.unauthorized"].includes(code)) return;
      accountError.value = "账号中的筛选暂时无法加载。";
    }
  }

  function trackFilters() {
    const onStorage = (event: StorageEvent) => {
      if (event.key === storageKey || event.key === null) refreshLocal();
    };
    onMounted(() => {
      refreshLocal();
      window.addEventListener("storage", onStorage);
    });
    onBeforeUnmount(() => window.removeEventListener("storage", onStorage));
    if (import.meta.client) watch(owner, (subject, _, onCleanup) => {
      let current = true;
      onCleanup(() => { current = false; });
      void refreshAccount(subject, () => current);
    }, { immediate: true });
  }

  function updateLocal(items: GallerySavedFilter[]) {
    localStorage.setItem(storageKey, JSON.stringify(items));
    localItems.value = items;
    localError.value = "";
  }

  function saveLocal(name: string, query: GallerySavedFilter["query"]) {
    const items = readLocalFilters();
    if (items.some(item => item.name === name)) return false;
    updateLocal([{ id: crypto.randomUUID(), name, query }, ...items]);
    return true;
  }

  async function saveToAccount(name: string, query: GallerySavedFilter["query"]) {
    const subject = owner.value;
    const result = await call<{ filter: GallerySavedFilter }>("/me/saved-filters", {
      method: "POST", body: { name, query },
    });
    if (owner.value === subject) accountItems.value = [result.filter, ...accountItems.value];
  }

  function removeLocal(id: string) {
    updateLocal(readLocalFilters().filter(item => item.id !== id));
  }

  async function removeFromAccount(id: string) {
    const subject = owner.value;
    await call(`/me/saved-filters/${encodeURIComponent(id)}`, { method: "DELETE" });
    if (owner.value === subject) accountItems.value = accountItems.value.filter(item => item.id !== id);
  }

  const items = computed(() => [
    ...localItems.value.map(item => ({ ...item, local: true })),
    ...(canSaveToAccount.value ? accountItems.value.map(item => ({ ...item, local: false })) : []),
  ]);
  const error = computed(() => localError.value || accountError.value);
  return { items, error, canSaveToAccount, trackFilters, saveLocal, saveToAccount, removeLocal, removeFromAccount };
}
