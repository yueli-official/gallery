import { describe, expect, it } from "vitest";

import {
  galleryImageSources,
  galleryRendition,
  imageAspect,
} from "../app/utils/image";

describe("gallery image projection", () => {
  it("keeps rendition policy behind one helper", () => {
    expect(
      galleryRendition(
        "019b0000-0000-7000-9000-000000000030",
        "grid-lg",
      ),
    ).toBe(
      "/media/31Pj0mXv7cfR5fdZIUvra?format=webp&name=grid-lg&v=1",
    );
  });

  it("preserves source aspect ratios for masonry", () => {
    expect(imageAspect(1200, 800)).toBe("1200 / 800");
    expect(imageAspect(0, 0)).toBe("4 / 3");
  });

  it("projects responsive sources from a named display slot", () => {
    const sources = galleryImageSources(
      "019b0000-0000-7000-9000-000000000030",
      "grid",
      true,
    );

    expect(sources.srcset).toContain("480w");
    expect(sources.srcset).toContain("960w");
    expect(sources.srcset).not.toContain("_mode=");
    expect(sources.loading).toBe("eager");
    expect(sources.fetchpriority).toBe("high");
  });
});
