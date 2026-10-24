export interface GalleryPresentationRule {
  kind: "category" | "facet_value";
  nodeId: string;
  mode: "icon" | "image";
  icon: string;
  imageId: string;
  assetId: string;
}
export interface GalleryAppearance {
  sidebar: { mode: "plain" | "image"; imageId: string; assetId: string; overlay: number };
  classification: GalleryPresentationRule[];
}
export interface GallerySavedFilter {
  id: string;
  name: string;
  query: { q: string; categories: string; facets: string; tag: string; matchAll: boolean; orientation: string; minLongEdge: number };
}
export function createGalleryAppearance(): GalleryAppearance {
  return { sidebar: { mode: "plain", imageId: "", assetId: "", overlay: .82 }, classification: [] };
}
export const galleryPresentationIcons = [
  "photo", "device-desktop", "camera", "palette", "user", "mountain", "building",
  "device-gamepad-2", "book", "sparkles", "category", "device-mobile", "leaf", "heart", "tag", "square",
] as const;
