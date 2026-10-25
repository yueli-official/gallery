import { assetMediaUrl } from "@yueli/asset-nuxt/media";
import type { GalleryMediaDelivery } from "~/types/gallery";

export type GalleryMediaReference = string | {
  assetId?: string;
  coverAssetId?: string;
  delivery?: GalleryMediaDelivery;
  coverDelivery?: GalleryMediaDelivery;
};

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

export function galleryRendition(
  reference: GalleryMediaReference,
  rendition: GalleryRendition,
): string {
  if (typeof reference !== "string") {
    const delivery = reference.delivery ?? reference.coverDelivery;
    if (delivery) {
      const url = delivery.renditions[rendition];
      if (delivery.mode !== "signed" || !url) throw new TypeError("Missing signed Gallery rendition");
      return url;
    }
  }
  const assetId = typeof reference === "string" ? reference : reference.assetId ?? reference.coverAssetId ?? "";
  // Asset's lossless-alpha revision uses a separate immutable cache identity.
  return assetMediaUrl(assetId, rendition, { version: 20261025 });
}

export function galleryImageSources(
  reference: GalleryMediaReference,
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
      src: galleryRendition(reference, "grid-lg"),
      srcset: `${galleryRendition(reference, "grid-sm")} 480w, ${galleryRendition(reference, "grid-lg")} 960w`,
      sizes: "(max-width: 639px) 50vw, (max-width: 1279px) 33vw, 20vw",
    };
  }
  if (slot === "masonry") {
    return {
      ...base,
      src: galleryRendition(reference, "masonry-lg"),
      srcset: `${galleryRendition(reference, "masonry-sm")} 960w, ${galleryRendition(reference, "masonry-lg")} 1600w`,
      sizes: "(max-width: 639px) 50vw, (max-width: 1279px) 33vw, 20vw",
    };
  }
  return {
    ...base,
    src: galleryRendition(reference, slot),
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
