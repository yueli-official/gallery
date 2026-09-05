import { useApi as useFoundationApi } from "@yueli/nuxt-runtime/runtime";
import {
  toGalleryApiRequest,
  type GalleryApiCallOptions,
} from "../utils/apiCompat";
export function useApi(target = "platform") {
  const request = useFoundationApi(target);
  async function call<T>(
    url: string,
    options?: GalleryApiCallOptions,
  ): Promise<T> {
    const prepared = toGalleryApiRequest<T>(url, options);
    return request.request(prepared.path, prepared.options);
  }
  return { call, request: request.request };
}
