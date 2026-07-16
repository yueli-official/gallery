import { describe, expect, it } from "vitest";

import {
  galleryImageSources,
  galleryRendition,
  imageAspect,
} from "../app/utils/image";

describe("gallery image projection", () => {
  it("keeps rendition policy behind one helper", () => {
    expect(galleryRendition("asset id", "grid-lg")).toContain(
      "asset%20id/image/@960x720",
    );
    expect(galleryRendition("asset id", "grid-lg")).toContain("mode=fill");
  });

  it("preserves source aspect ratios for masonry", () => {
    expect(imageAspect(1200, 800)).toBe("1200 / 800");
    expect(imageAspect(0, 0)).toBe("4 / 3");
  });

  it("projects responsive sources from a named display slot", () => {
    const sources = galleryImageSources("asset", "grid", true);

    expect(sources.srcset).toContain("480w");
    expect(sources.srcset).toContain("960w");
    expect(sources.loading).toBe("eager");
    expect(sources.fetchpriority).toBe("high");
  });
});
