import { createGalleryNotifier } from "~/utils/feedback";
import type { ComputedRef } from "vue";
import type { GalleryImage } from "~/types/gallery";

export function useGalleryImageActions(
  image: ComputedRef<GalleryImage | undefined>,
) {
  const { loggedIn, login } = useAuth();
  const { call } = useApi();
  const toast = createGalleryNotifier(useToast());
  const favoritePending = ref(false);

  async function track(type: "qualified_view" | "share"): Promise<void> {
    if (!image.value) return;
    try {
      await $fetch(
        `/api/gallery/images/${encodeURIComponent(image.value.id)}/events`,
        { method: "POST", body: { type, sessionKey: galleryMetricSession() } },
      );
    } catch {
      /* 指标上报永远不能阻断图片浏览。 */
    }
  }

  async function toggleFavorite(): Promise<void> {
    if (!image.value || favoritePending.value) return;
    if (!loggedIn.value) {
      await login();
      return;
    }
    favoritePending.value = true;
    try {
      const next = !image.value.favorited;
      await call(
        `/api/v1/gallery/me/favorites/${encodeURIComponent(image.value.id)}`,
        {
          method: next ? "PUT" : "DELETE",
          body: next ? { version: 0 } : undefined,
        },
      );
      image.value.favorited = next;
      image.value.metrics.favorites = Math.max(
        0,
        image.value.metrics.favorites + (next ? 1 : -1),
      );
    } catch (reason: any) {
      toast.add({
        title: "收藏没有保存",
        description: reason?.data?.message || reason?.message || "请稍后重试",
        color: "error",
      });
    } finally {
      favoritePending.value = false;
    }
  }

  async function shareImage(targetUrl?: string): Promise<void> {
    const url = targetUrl || window.location.href;
    try {
      if (navigator.share) {
        await navigator.share({ title: image.value?.title, url });
      } else {
        await navigator.clipboard.writeText(url);
        // feedback-contract: clipboard sharing has no persistent page surface for confirmation.
        toast.add({ title: "链接已复制", color: "success" });
      }
      await track("share");
    } catch (reason: any) {
      if (reason?.name !== "AbortError") {
        toast.add({ title: "分享没有完成", color: "error" });
      }
    }
  }

  return { favoritePending, toggleFavorite, shareImage, track };
}
