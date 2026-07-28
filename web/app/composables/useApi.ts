import { useApi as useFoundationApi } from "@yueli/nuxt-runtime/runtime";
import {
  toGalleryApiRequest,
  type GalleryApiCallOptions,
} from "../utils/apiCompat";

const REAUTH_COOLDOWN_MS = 10_000;
let reauthFlight: Promise<boolean> | undefined;

function mayReauth(): boolean {
  try {
    const last = Number(sessionStorage.getItem("reauth-at") || 0);
    if (Date.now() - last < REAUTH_COOLDOWN_MS) return false;
    sessionStorage.setItem("reauth-at", String(Date.now()));
  } catch {
    // 浏览器存储不是必需条件；单次并发保护仍然有效。
  }
  return true;
}

async function restoreSession() {
  const reauth = getOptionalGalleryReauth();
  if (!reauth) return false;
  if (!reauthFlight) {
    const current = reauth({ requireLoggedIn: true }).finally(() => {
      if (reauthFlight === current) reauthFlight = undefined;
    });
    reauthFlight = current;
  }
  return reauthFlight;
}

/** 兼容 Gallery 现有 call 调用；新代码可直接使用 request。 */
export function useApi(target = "platform") {
  const request = useFoundationApi(target);

  async function call<T>(
    url: string,
    options?: GalleryApiCallOptions,
  ): Promise<T> {
    const prepared = toGalleryApiRequest<T>(url, options);
    try {
      return await request.request(prepared.path, prepared.options);
    } catch (error: unknown) {
      const status = (error as { statusCode?: number })?.statusCode;
      if (import.meta.client && status === 401 && mayReauth())
        await restoreSession();
      throw error;
    }
  }

  return { call, request };
}
