import { describe, expect, it } from "vitest";

import { galleryRendition, imageAspect } from "../app/utils/image";

describe("gallery image projection", () => {
  it("keeps rendition policy behind one helper", () => {
    expect(galleryRendition("asset id", "grid-lg")).toContain("asset%20id/image/@960x720");
  });

  it("preserves source aspect ratios for masonry", () => {
    expect(imageAspect(1200, 800)).toBe("1200 / 800");
    expect(imageAspect(0, 0)).toBe("4 / 3");
  });
});
