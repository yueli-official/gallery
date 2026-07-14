import { describe, expect, it } from "vitest";

import type { GalleryCategory, GalleryFacet } from "../app/types/gallery";
import { groupCategoriesByFacet } from "../app/utils/taxonomy";

const facets: GalleryFacet[] = [
  {
    id: "medium",
    slug: "medium",
    name: "媒介",
    description: "",
    selectionMode: "multiple",
  },
  {
    id: "style",
    slug: "style",
    name: "风格",
    description: "",
    selectionMode: "multiple",
  },
];

const illustration: GalleryCategory = {
  id: "illustration",
  facetId: "medium",
  parentId: "",
  slug: "illustration",
  name: "插画",
  description: "",
  artworkCount: 0,
};

describe("gallery taxonomy", () => {
  it("groups categories by controlled facet order", () => {
    expect(groupCategoriesByFacet(facets, [illustration])).toEqual([
      { facet: facets[0], categories: [illustration] },
    ]);
  });

  it("does not render empty taxonomy groups", () => {
    expect(groupCategoriesByFacet(facets, [])).toEqual([]);
  });
});
