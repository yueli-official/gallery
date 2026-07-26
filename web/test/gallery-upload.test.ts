import { describe, expect, it } from "vitest";
import {
  GALLERY_UPLOAD_ACCEPT,
  GALLERY_UPLOAD_FORMAT_LABEL,
  GALLERY_UPLOAD_MAX_BYTES,
  galleryUploadAnimationError,
  galleryUploadFileError,
} from "../app/utils/galleryUpload";

describe("gallery upload preflight", () => {
  it("accepts a supported static image", () => {
    expect(
      galleryUploadFileError({
        name: "rain.webp",
        size: 1024,
        type: "image/webp",
      }),
    ).toBe("");
  });

  it("rejects unsupported and animated formats before upload", () => {
    expect(
      galleryUploadFileError({
        name: "loop.gif",
        size: 1024,
        type: "image/gif",
      }),
    ).toContain("仅支持");
    expect(
      galleryUploadFileError({
        name: "vector.svg",
        size: 1024,
        type: "image/svg+xml",
      }),
    ).toContain("仅支持");
    expect(
      galleryUploadFileError({
        name: "camera.heic",
        size: 1024,
        type: "image/heic",
      }),
    ).toContain("仅支持");
  });

  it("shares one truthful format contract with the file picker", () => {
    expect(GALLERY_UPLOAD_ACCEPT).toBe(
      ".jpg,.jpeg,.png,.webp,image/jpeg,image/png,image/webp",
    );
    expect(GALLERY_UPLOAD_FORMAT_LABEL).toBe("JPEG、PNG 和 WebP");
  });

  it("detects APNG and animated WebP markers", async () => {
    const apng = new File(
      [new Uint8Array([0x89, 0x50, 0x4e, 0x47, 0x61, 0x63, 0x54, 0x4c])],
      "loop.png",
      { type: "image/png" },
    );
    const webp = new File(
      [new Uint8Array([0x52, 0x49, 0x46, 0x46, 0x41, 0x4e, 0x49, 0x4d])],
      "loop.webp",
      { type: "image/webp" },
    );
    await expect(galleryUploadAnimationError(apng)).resolves.toContain("动画");
    await expect(galleryUploadAnimationError(webp)).resolves.toContain("动画");
  });

  it("rejects empty and oversized files", () => {
    expect(
      galleryUploadFileError({ name: "empty.png", size: 0, type: "image/png" }),
    ).toContain("为空");
    expect(
      galleryUploadFileError({
        name: "huge.jpg",
        size: GALLERY_UPLOAD_MAX_BYTES + 1,
        type: "image/jpeg",
      }),
    ).toContain("20 MiB");
  });
});
