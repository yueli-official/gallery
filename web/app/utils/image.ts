export type GalleryRendition =
  | "thumbnail"
  | "grid-sm"
  | "grid-lg"
  | "masonry-sm"
  | "masonry-lg"
  | "preview"
  | "display"
  | "og";

export type GalleryImageSlot =
  "thumbnail" | "grid" | "masonry" | "preview" | "display" | "og";

export interface GalleryImageSourcePolicy {
  src: string;
  srcset?: string;
  sizes?: string;
  loading: "eager" | "lazy";
  fetchpriority: "high" | "auto";
}

const compatibilitySpecs: Record<GalleryRendition, string> = {
  thumbnail: "@240x180_mode=cover_type=webp_q=82.webp",
  "grid-sm": "@480x360_mode=cover_type=webp_q=84.webp",
  "grid-lg": "@960x720_mode=cover_type=webp_q=86.webp",
  "masonry-sm": "@960x960_mode=fit_type=webp_q=85.webp",
  "masonry-lg": "@1600x1600_mode=fit_type=webp_q=88.webp",
  preview: "@1600x1600_mode=fit_type=webp_q=88.webp",
  display: "@2560x2560_mode=fit_type=webp_q=90.webp",
  og: "@1200x630_mode=cover_type=webp_q=86.webp",
};

// Asset 后续会以 released named rendition 替换兼容 transform URL；映射集中在此，
// 避免资源平台升级泄漏进每个页面。
export function galleryRendition(
  assetId: string,
  rendition: GalleryRendition,
): string {
  if (!assetId) return "";
  return `/asset-api/api/v1/assets/${encodeURIComponent(assetId)}/image/${compatibilitySpecs[rendition]}`;
}

export function galleryImageSources(
  assetId: string,
  slot: GalleryImageSlot,
  priority = false,
): GalleryImageSourcePolicy {
  const base = {
    loading: priority ? "eager" : "lazy",
    fetchpriority: priority ? "high" : "auto",
  } as const;

  if (slot === "grid") {
    return {
      ...base,
      src: galleryRendition(assetId, "grid-lg"),
      srcset: `${galleryRendition(assetId, "grid-sm")} 480w, ${galleryRendition(assetId, "grid-lg")} 960w`,
      sizes: "(max-width: 639px) 50vw, (max-width: 1279px) 33vw, 20vw",
    };
  }
  if (slot === "masonry") {
    return {
      ...base,
      src: galleryRendition(assetId, "masonry-lg"),
      srcset: `${galleryRendition(assetId, "masonry-sm")} 960w, ${galleryRendition(assetId, "masonry-lg")} 1600w`,
      sizes: "(max-width: 639px) 50vw, (max-width: 1279px) 33vw, 20vw",
    };
  }
  return {
    ...base,
    src: galleryRendition(assetId, slot),
    sizes: slot === "display" ? "100vw" : undefined,
  };
}

export function imageAspect(
  width: number,
  height: number,
  fallback = "4 / 3",
): string {
  return width > 0 && height > 0 ? `${width} / ${height}` : fallback;
}

export function compactMetric(value: number): string {
  return new Intl.NumberFormat("zh-CN", {
    notation: "compact",
    maximumFractionDigits: 1,
  }).format(value || 0);
}

export function galleryMetricSession(): string {
  if (import.meta.server) return "";
  const key = "gallery.metric.session";
  const current = sessionStorage.getItem(key);
  if (current) return current;
  const created =
    globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random()}`;
  sessionStorage.setItem(key, created);
  return created;
}
