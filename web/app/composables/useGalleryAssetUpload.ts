import { failureFromProblemResponse } from "@yueli/http-runtime";
import type { GalleryUploadedAsset } from "~/types/gallery";
import { assetUploadURL } from "@yueli/asset-nuxt/upload";
import { useAssetDraftSession } from "@yueli/asset-nuxt/references";

interface UploadInit {
  uploadUrl: string;
  uploadToken: string;
  uploadHeaders?: Record<string, string>;
}

export function useGalleryAssetUpload(draftKey = "gallery-media-editor") {
  const { call } = useAssetApi();
  const { assetNamespace } = useSiteRuntime();
  const draft=useAssetDraftSession(call,draftKey);

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
    // Normalize orientation/color in the browser before transmission. Asset
    // independently enforces the registered WebP storage target.
    const bitmap = await createImageBitmap(file, { imageOrientation: "from-image" });
    try {
      const scale = Math.min(1, 4096 / Math.max(bitmap.width, bitmap.height));
      const canvas = document.createElement("canvas");
      canvas.width = Math.max(1, Math.round(bitmap.width * scale));
      canvas.height = Math.max(1, Math.round(bitmap.height * scale));
      const context = canvas.getContext("2d", { colorSpace: "srgb" });
      if (!context) throw new Error("image_processing_unavailable");
      context.imageSmoothingEnabled = true;
      context.imageSmoothingQuality = "high";
      context.drawImage(bitmap, 0, 0, canvas.width, canvas.height);
      const blob = await new Promise<Blob>((resolve, reject) =>
        canvas.toBlob(value => value?.type === "image/webp" ? resolve(value) : reject(new Error("image_processing_failed")), "image/webp", 1),
      );
      file = new File([blob], file.name.replace(/\.[^.]+$/, "") + ".webp", { type: "image/webp" });
      canvas.width = canvas.height = 1;
    } finally {
      bitmap.close();
    }
    return uploadPrepared(file, onProgress);
  }

  // The shared image workbench already produces normalized, metadata-free
  // WebP. Preserve its output instead of encoding it a second time.
  async function uploadPrepared(file: File, onProgress?: (value: number) => void) {
    const session=await draft.ensure();
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
        imageProcess: { execution: "auto", target: "stored", preprocessed: true },
      },
    });
    await put(init.uploadUrl, file, { "Content-Type": file.type, ...init.uploadHeaders }, onProgress);
    onProgress?.(98);
    const finalized = await call<{ asset: GalleryUploadedAsset }>(
      "/api/v1/assets/finalize",
      {
        method: "POST",
        body: { uploadToken: init.uploadToken,draftSessionId:session.id },
      },
    );
    onProgress?.(100);
    return finalized.asset;
  }

  return { upload, uploadPrepared,draft };
}
