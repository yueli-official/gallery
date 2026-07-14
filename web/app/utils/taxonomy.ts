import type { GalleryCategory, GalleryFacet } from "../types/gallery";

export interface GalleryFacetGroup {
  facet: GalleryFacet;
  categories: GalleryCategory[];
}

export function groupCategoriesByFacet(
  facets: GalleryFacet[],
  categories: GalleryCategory[],
): GalleryFacetGroup[] {
  return facets
    .map((facet) => ({
      facet,
      categories: categories.filter((category) => category.facetId === facet.id),
    }))
    .filter((group) => group.categories.length > 0);
}
