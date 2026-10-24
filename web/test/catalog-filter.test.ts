import { describe, expect, it } from "vitest";
import { galleryCatalogQuery, galleryCatalogRequest, parseGalleryCatalogState } from "../app/utils/catalog";

describe("Gallery public catalog filters", () => {
  it("preserves combined filters across URL and API projection", () => {
    const state = parseGalleryCatalogState({categories:"wallpaper",facets:"character:a,character:b,work:game",tag:"night",orientation:"portrait",minLongEdge:"3840",matchAll:"true"});
    expect(galleryCatalogRequest(state)).toMatchObject({categories:"wallpaper",facets:"character:a,character:b,work:game",tag:"night",orientation:"portrait",minLongEdge:3840,matchAll:true});
    expect(parseGalleryCatalogState(galleryCatalogQuery(state))).toEqual(state);
  });
});
