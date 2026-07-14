import { describe, expect, it } from "vitest";

import { groupFacetValues } from "@platform/facet";

import type { GalleryFacet, GalleryFacetValue } from "../app/types/gallery";

const facets: GalleryFacet[] = [
  {
    id: "medium",
    slug: "medium",
    name: "媒介",
    description: "",
    selectionMode: "multiple",
    requiredOnPublish: true,
    filterable: true,
    status: "active",
    sortOrder: 10,
  },
  {
    id: "style",
    slug: "style",
    name: "风格",
    description: "",
    selectionMode: "multiple",
    requiredOnPublish: false,
    filterable: true,
    status: "active",
    sortOrder: 20,
  },
];

const illustration: GalleryFacetValue = {
  id: "illustration",
  facetId: "medium",
  slug: "illustration",
  name: "插画",
  description: "",
  status: "active",
  sortOrder: 10,
  count: 0,
};

describe("gallery taxonomy", () => {
  it("groups facet values by controlled facet order", () => {
    expect(groupFacetValues(facets, [illustration])).toEqual([
      { facet: facets[0], values: [illustration] },
    ]);
  });

  it("does not render empty taxonomy groups", () => {
    expect(groupFacetValues(facets, [])).toEqual([]);
  });
});
