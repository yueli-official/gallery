import type { GalleryUploadedAsset } from "~/types/gallery";
import { waitForGalleryAssetReady } from "~/utils/galleryAssetReadiness";

interface UploadInit {
  uploadUrl: string;
  uploadToken: string;
  uploadHeaders?: Record<string, string>;
}

export function useGalleryAssetUpload() {
  const { call } = useAssetApi();
  const { slug: siteSlug } = useSiteRuntime();

  function put(url: string, file: File, headers: Record<string, string>, onProgress?: (value: number) => void) {
    return new Promise<void>((resolve, reject) => {
      const request = new XMLHttpRequest();
      request.open("PUT", url);
      Object.entries(headers).forEach(([key, value]) => request.setRequestHeader(key, value));
      request.upload.onprogress = (event) => {
        if (event.lengthComputable) onProgress?.(Math.round((event.loaded / event.total) * 95));
      };
      request.onload = () => request.status >= 200 && request.status < 300
        ? resolve()
        : reject(new Error(`上传失败 (HTTP ${request.status})`));
      request.onerror = () => reject(new Error("上传网络错误，请确认资源服务在线"));
      request.send(file);
    });
  }

  async function upload(file: File, onProgress?: (value: number) => void) {
    const init = await call<UploadInit>("/api/v1/assets/upload-init", {
      method: "POST",
      body: {
        filename: file.name,
        mime: file.type || "application/octet-stream",
        size: file.size,
        siteKey: siteSlug.value,
        profileKey: "gallery-submission",
        category: "gallery-submission",
        visibility: "private",
        multipart: false,
      },
    });
    await put(init.uploadUrl, file, init.uploadHeaders ?? {}, onProgress);
    onProgress?.(98);
    const finalized = await call<{ asset: GalleryUploadedAsset }>("/api/v1/assets/finalize", {
      method: "POST",
      body: { uploadToken: init.uploadToken },
    });
    onProgress?.(100);
    return finalized.asset;
  }

  async function waitUntilReady(
    assetId: string,
    initial?: GalleryUploadedAsset,
  ) {
    let first = initial;
    return waitForGalleryAssetReady(assetId, async (id) => {
      if (first) {
        const asset = first;
        first = undefined;
        return asset;
      }
      const response = await call<{ asset: GalleryUploadedAsset }>(
        `/api/v1/assets/${encodeURIComponent(id)}`,
      );
      return response.asset;
    });
  }

  return { upload, waitUntilReady };
}
