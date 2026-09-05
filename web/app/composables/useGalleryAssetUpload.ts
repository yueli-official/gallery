import { failureFromProblemResponse } from "@yueli/http-runtime";
import type { GalleryUploadedAsset } from "~/types/gallery";
import { assetUploadURL } from "@yueli/asset-nuxt/upload";

interface UploadInit {
  uploadUrl: string;
  uploadToken: string;
  uploadHeaders?: Record<string, string>;
}

export function useGalleryAssetUpload() {
  const { call } = useAssetApi();
  const { assetNamespace } = useSiteRuntime();

  function put(
    url: string,
    file: File,
    headers: Record<string, string>,
    onProgress?: (value: number) => void,
  ) {
    return new Promise<void>((resolve, reject) => {
      const request = new XMLHttpRequest();
      request.open("PUT", assetUploadURL(url));
      Object.entries(headers).forEach(([key, value]) =>
        request.setRequestHeader(key, value),
      );
      request.upload.onprogress = (event) => {
        if (event.lengthComputable)
          onProgress?.(Math.round((event.loaded / event.total) * 95));
      };
      request.onload = async () => {
        if (request.status >= 200 && request.status < 300) {
          resolve();
          return;
        }
        if (request.status < 400) {
          reject({
            kind: "network",
            code: "foundation.request.network",
            reauth: "not-attempted",
          });
          return;
        }
        const response = new Response(request.responseText, {
          status: request.status,
          headers: {
            "content-type": request.getResponseHeader("content-type") || "",
            "x-trace-id": request.getResponseHeader("x-trace-id") || "",
          },
        });
        reject(await failureFromProblemResponse(response, 64 * 1024));
      };
      request.onerror = () =>
        reject({
          kind: "network",
          code: "foundation.request.network",
          reauth: "not-attempted",
        });
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
        siteKey: assetNamespace.value,
        profileKey: "gallery-submission",
        category: "gallery-submission",
        visibility: "private",
        multipart: false,
      },
    });
    await put(init.uploadUrl, file, init.uploadHeaders ?? {}, onProgress);
    onProgress?.(98);
    const finalized = await call<{ asset: GalleryUploadedAsset }>(
      "/api/v1/assets/finalize",
      {
        method: "POST",
        body: { uploadToken: init.uploadToken },
      },
    );
    onProgress?.(100);
    return finalized.asset;
  }

  return { upload };
}
