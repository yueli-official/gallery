export const GALLERY_UPLOAD_MAX_BYTES = 20 * 1024 * 1024;
export const GALLERY_UPLOAD_BATCH_LIMIT = 20;
export const GALLERY_UPLOAD_ACCEPT =
  ".jpg,.jpeg,.png,.webp,image/jpeg,image/png,image/webp";
export const GALLERY_UPLOAD_FORMAT_LABEL = "JPEG、PNG 和 WebP";

const allowedExtensions = new Set([
  "jpg",
  "jpeg",
  "png",
  "webp",
]);
const allowedTypes = new Set([
  "image/jpeg",
  "image/png",
  "image/webp",
]);

export function galleryUploadFileError(
  file: Pick<File, "name" | "size" | "type">,
): string {
  const extension = file.name.split(".").pop()?.toLowerCase() || "";
  if (
    !allowedExtensions.has(extension) ||
    (file.type && !allowedTypes.has(file.type))
  ) {
    return `仅支持 ${GALLERY_UPLOAD_FORMAT_LABEL} 静态图片`;
  }
  if (file.size <= 0) return "文件内容为空";
  if (file.size > GALLERY_UPLOAD_MAX_BYTES) return "文件超过 20 MiB";
  return "";
}

export async function galleryUploadAnimationError(file: File): Promise<string> {
  const extension = file.name.split(".").pop()?.toLowerCase() || "";
  if (!new Set(["png", "webp"]).has(extension)) return "";
  const bytes = new Uint8Array(
    await file.slice(0, 2 * 1024 * 1024).arrayBuffer(),
  );
  const marker =
    extension === "png" ? [0x61, 0x63, 0x54, 0x4c] : [0x41, 0x4e, 0x49, 0x4d];
  for (let offset = 0; offset <= bytes.length - marker.length; offset += 1) {
    if (marker.every((value, index) => bytes[offset + index] === value)) {
      return "不支持动画图片，请选择静态文件";
    }
  }
  return "";
}
