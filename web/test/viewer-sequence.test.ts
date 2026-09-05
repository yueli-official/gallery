import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

import {
  createViewerSequence,
  createCatalogViewerNavigationSession,
  extendViewerSequence,
  moveViewerSequence,
  prependViewerSequence,
  viewerCloseTarget,
} from "../app/utils/viewerSequence";

describe("Gallery Viewer 连续浏览", () => {
  it("耗尽首批后追加去重候选并保持前后轨迹", () => {
    let state = createViewerSequence("image-1", [
      "image-1",
      "image-2",
      "image-3",
    ]);
    state = moveViewerSequence(state, "image-2");
    state = moveViewerSequence(state, "image-3");
    state = extendViewerSequence(state, ["image-2", "image-4", "image-5"]);

    expect(state.ids).toEqual([
      "image-1",
      "image-2",
      "image-3",
      "image-4",
      "image-5",
    ]);
    expect(state.ids[state.index - 1]).toBe("image-2");
    expect(state.ids[state.index + 1]).toBe("image-4");
  });

  it("从历史轨迹返回时不会复制图片", () => {
    let state = createViewerSequence("image-1", ["image-1", "image-2"]);
    state = moveViewerSequence(state, "image-2");
    state = moveViewerSequence(state, "image-1");
    expect(state).toEqual({ ids: ["image-1", "image-2"], index: 0 });
  });

  it("目录上下页追加后保持页序与当前图片位置", () => {
    const cards = (ids: string[]) =>
      ids.map((id) => ({
        id,
        assetId: `${id}-asset`,
        title: id,
        altText: id,
        width: 100,
        height: 100,
        dominantColor: "",
        primaryCategory: "",
        primaryCategorySlug: "",
        metrics: { views: 0, favorites: 0 },
      }));
    const session = createCatalogViewerNavigationSession(
      cards(["image-3", "image-4"]),
      { sort: "newest", page: 2, size: 2 },
      3,
    );
    let state = moveViewerSequence(session.sequence, "image-3");
    state = prependViewerSequence(state, ["image-1", "image-2"]);
    state = extendViewerSequence(state, ["image-5", "image-6"]);

    expect(state).toEqual({
      ids: [
        "image-1",
        "image-2",
        "image-3",
        "image-4",
        "image-5",
        "image-6",
      ],
      index: 2,
    });
    expect(session.catalog).toMatchObject({ previousPage: 1, nextPage: 3 });
  });

  it("关闭目标不会落到另一张详情图", () => {
    expect(viewerCloseTarget("/")).toBe("/");
    expect(viewerCloseTarget("/images?sort=newest")).toBe(
      "/images?sort=newest",
    );
    expect(viewerCloseTarget("/images/image-previous")).toBe("/images");
    expect(viewerCloseTarget(undefined)).toBe("/images");
  });
});

describe("Gallery Viewer 主题与标题尺度", () => {
  const css = readFileSync(
    fileURLToPath(new URL("../app/assets/css/main.css", import.meta.url)),
    "utf8",
  );
  const detail = readFileSync(
    fileURLToPath(
      new URL("../app/pages/images/[imageId].vue", import.meta.url),
    ),
    "utf8",
  );

  it("Viewer 不锁死暗色并且关闭不用 history back", () => {
    expect(css).not.toContain("background: #0b0f16");
    expect(css).not.toContain("background: #111824");
    expect(detail).not.toContain("router.back()");
  });

  it("常规页面标题使用产品级 type ramp", () => {
    expect(css).toContain("--gallery-title-page: clamp(1.75rem, 2.2vw, 2rem)");
    expect(css).toContain(
      "--gallery-title-section: clamp(1.25rem, 1.8vw, 1.5rem)",
    );
  });
});
