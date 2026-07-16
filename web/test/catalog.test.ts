import { describe, expect, it } from "vitest";

import {
  galleryCatalogFilterCount,
  galleryCatalogQuery,
  galleryCatalogRequest,
  parseGalleryCatalogState,
} from "../app/utils/catalog";

describe("gallery catalog state", () => {
  it("normalizes route input into one stable state", () => {
    const state = parseGalleryCatalogState({
      q: "  海边  ",
      sort: "unknown",
      page: "-4",
      categories: "wallpaper,photo,wallpaper",
      facets: "scene:sea,orientation:wide",
    });

    expect(state).toMatchObject({
      q: "海边",
      sort: "newest",
      page: 1,
      categories: ["wallpaper", "photo"],
      facets: ["scene:sea", "orientation:wide"],
    });
    expect(galleryCatalogFilterCount(state)).toBe(5);
  });

  it("omits defaults from shareable URLs but keeps them in API requests", () => {
    const state = parseGalleryCatalogState({ q: "山", preview: "image-id" });

    expect(galleryCatalogQuery(state)).toEqual({
      q: "山",
      preview: "image-id",
    });
    expect(galleryCatalogRequest(state)).toEqual({
      q: "山",
      sort: "newest",
      page: 1,
      size: 24,
      categories: undefined,
      facets: undefined,
      tag: undefined,
    });
  });

  it("only persists non-default view preference", () => {
    const state = parseGalleryCatalogState({ view: "masonry" });
    expect(state.view).toBe("masonry");
    expect(galleryCatalogQuery(state)).toEqual({ view: "masonry" });
    expect(parseGalleryCatalogState({ view: "unknown" }).view).toBe("grid");
  });
});
