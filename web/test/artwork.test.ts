import { describe, expect, it } from "vitest";

import { artworkAspectRatio, contentRatingLabel } from "../app/utils/artwork";

describe("gallery artwork presentation", () => {
  it("preserves a valid source aspect ratio", () => {
    expect(artworkAspectRatio(1200, 800)).toBe("1200 / 800");
  });

  it("falls back when dimensions are incomplete", () => {
    expect(artworkAspectRatio(0, 800)).toBe("4 / 5");
  });

  it("uses explicit safety labels", () => {
    expect(contentRatingLabel("sensitive")).toBe("敏感内容");
  });
});
