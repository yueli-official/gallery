import { createGalleryNotifier } from "~/utils/feedback";
import { getApiFailure } from "@yueli/http-runtime";
import type { ComputedRef } from "vue";
import type { GalleryImage } from "~/types/gallery";

function actionFailureCode(reason: unknown): string {
  return getApiFailure(reason)?.code || "";
}

function actionFailureStatus(reason: unknown): number {
  const failure = getApiFailure(reason);
  return failure?.kind === "remote" ? failure.status : 0;
}

async function copyShareUrl(url: string): Promise<boolean> {
  if (window.isSecureContext && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(url);
      return true;
    } catch {
      // LAN HTTP and browser policies can reject the modern clipboard path.
    }
  }

  const field = document.createElement("textarea");
  field.value = url;
  field.readOnly = true;
  field.setAttribute("aria-hidden", "true");
  field.style.position = "fixed";
  field.style.inset = "0 auto auto -9999px";
  field.style.opacity = "0";
  document.body.appendChild(field);
  field.focus();
  field.select();
  field.setSelectionRange(0, field.value.length);
  let copied = false;
  try {
    copied = document.execCommand("copy");
  } finally {
    field.remove();
  }
  return copied;
}

export function useGalleryImageActions(
  image: ComputedRef<GalleryImage | undefined>,
) {
  const { loggedIn, login } = useAuth();
  const { call } = useGalleryApi();
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
        `/me/favorites/${encodeURIComponent(image.value.id)}`,
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
      const status = actionFailureStatus(reason);
      if (
        status === 401 ||
        (status === 403 && actionFailureCode(reason) === "gallery.forbidden")
      ) {
        await login();
        return;
      }
      toast.add({
        title: "收藏没有保存",
        description: galleryFailureMessage(reason, "请稍后重试"),
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
        try {
          await navigator.share({ title: image.value?.title, url });
        } catch (reason: any) {
          if (reason?.name === "AbortError") return;
          if (!(await copyShareUrl(url))) throw reason;
          toast.add({ title: "链接已复制", color: "success" });
        }
      } else {
        if (!(await copyShareUrl(url))) throw new Error("copy failed");
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
