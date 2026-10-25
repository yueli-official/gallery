import type { Ref } from "vue";

// Refresh the owning DTO before its signed links expire. This stays at the data
// boundary: it never signs links or persists them, and public data has no timer.
// Pass the useFetch/useAsyncData refresh function, which owns its error state.
export function useGalleryMediaRefresh(data: Ref<unknown>, refresh: () => Promise<unknown>) {
  if (import.meta.server) return;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let disposed = false;
  let lastAttempt = 0;

  function expiry(value: unknown): number {
    const pending = [value];
    const seen = new Set<object>();
    let earliest = Infinity;
    while (pending.length) {
      const item = pending.pop();
      if (!item || typeof item !== "object" || seen.has(item)) continue;
      seen.add(item);
      const record = item as Record<string, unknown>;
      if (record.mode === "signed" && typeof record.expiresAt === "string") {
        const expires = Date.parse(record.expiresAt);
        if (Number.isFinite(expires)) earliest = Math.min(earliest, expires);
      } else pending.push(...Object.values(record));
    }
    return earliest;
  }

  function schedule() {
    clearTimeout(timer);
    if (disposed || document.hidden) return;
    const expires = expiry(data.value);
    if (!Number.isFinite(expires)) return;
    const delay = Math.max(1000, lastAttempt + 30_000 - Date.now(), expires - Date.now() - 30_000);
    timer = setTimeout(() => {
      lastAttempt = Date.now();
      void refresh().finally(schedule);
    }, delay);
  }

  watch(data, schedule, { immediate: true });
  document.addEventListener("visibilitychange", schedule);
  onScopeDispose(() => {
    disposed = true;
    clearTimeout(timer);
    document.removeEventListener("visibilitychange", schedule);
  });
}
