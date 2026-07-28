import type { GallerySite } from "~/types/gallery";

export function useGallerySiteSettings() {
  return useState<GallerySite | undefined>("gallery-site-settings");
}
