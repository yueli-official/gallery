export type GalleryRendition =
  | "thumbnail"
  | "grid-sm"
  | "grid-lg"
  | "masonry-sm"
  | "masonry-lg"
  | "preview"
  | "display"
  | "og";

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

// Asset will replace this compatibility transform URL with released named
// renditions. Keeping the mapping here prevents that upgrade leaking into pages.
export function galleryRendition(assetId: string, rendition: GalleryRendition): string {
  if (!assetId) return "";
  return `/asset-api/api/v1/assets/${encodeURIComponent(assetId)}/image/${compatibilitySpecs[rendition]}`;
}

export function imageAspect(width: number, height: number, fallback = "4 / 3"): string {
  return width > 0 && height > 0 ? `${width} / ${height}` : fallback;
}

export function compactMetric(value: number): string {
  return new Intl.NumberFormat("zh-CN", { notation: "compact", maximumFractionDigits: 1 }).format(value || 0);
}

export function galleryMetricSession(): string {
  if (import.meta.server) return "";
  const key = "gallery.metric.session";
  const current = sessionStorage.getItem(key);
  if (current) return current;
  const created = globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random()}`;
  sessionStorage.setItem(key, created);
  return created;
}
