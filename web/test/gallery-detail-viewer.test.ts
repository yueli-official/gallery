import fs from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";

const root = path.resolve(import.meta.dirname, "..");
const viewer = fs.readFileSync(
  path.join(root, "app/components/GalleryViewer.vue"),
  "utf8",
);
const styles = fs.readFileSync(
  path.join(root, "app/assets/css/main.css"),
  "utf8",
);

describe("gallery image detail viewer", () => {
  it("offers exactly one previous and one next control", () => {
    expect(viewer.match(/aria-label="查看上一张图片"/g) || []).toHaveLength(1);
    expect(viewer.match(/aria-label="查看下一张图片"/g) || []).toHaveLength(1);
  });

  it("fullscreens the persistent document root so route navigation keeps fullscreen", () => {
    expect(viewer).toContain("document.documentElement.requestFullscreen()");
    expect(viewer).not.toContain("media.value.requestFullscreen()");
    expect(styles).toContain("html:fullscreen .gallery-detail-media");
    expect(styles).not.toContain(".gallery-detail-media:fullscreen");
  });

  it("shows a small-rendition filmstrip in fullscreen album mode", () => {
    expect(viewer).toContain('class="gallery-detail-filmstrip"');
    expect(viewer).toContain(
      "galleryImageSources(item.assetId, 'thumbnail', false)",
    );
    expect(styles).toContain(
      "html:fullscreen .gallery-detail-filmstrip",
    );
  });

  it("keeps the image dominant and long titles subordinate", () => {
    expect(viewer).toContain('class="gallery-detail-titleline"');
    expect(styles).toMatch(
      /\.gallery-detail-stage\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\)/s,
    );
    expect(styles).toMatch(
      /\.gallery-detail-heading h1\s*\{[^}]*font-size:\s*clamp\(1\.05rem,\s*1\.35vw,\s*1\.25rem\)/s,
    );
    expect(styles).toMatch(
      /\.gallery-detail-heading h1\s*\{[^}]*overflow-wrap:\s*anywhere/s,
    );
    expect(styles).not.toContain("font-size: clamp(2rem, 3vw, 3.35rem)");
  });

  it("keeps the image description below the title without a redundant label", () => {
    expect(viewer).not.toContain('class="gallery-detail-facts"');
    expect(viewer).toContain(
      'class="gallery-detail-description"',
    );
    expect(viewer).not.toContain("<strong>图片说明</strong>");
    expect(viewer.indexOf('class="gallery-detail-description"')).toBeLessThan(
      viewer.indexOf('class="gallery-detail-stats"'),
    );
  });

  it("distills dimensions and governance behind an explicit disclosure", () => {
    expect(viewer).toContain('class="gallery-detail-more"');
    expect(viewer).toContain("<dt>图片尺寸</dt>");
    expect(viewer.indexOf("<dt>图片尺寸</dt>")).toBeGreaterThan(
      viewer.indexOf('class="gallery-detail-more-panel"'),
    );
    expect(viewer).not.toContain(
      '<span aria-hidden="true">·</span>',
    );
  });

  it("dismisses more information outside, on Escape, and on image changes", () => {
    expect(viewer).toContain('ref="moreDetails"');
    expect(viewer).toContain(
      'document.addEventListener("pointerdown", onDocumentPointerDown)',
    );
    expect(viewer).toContain(
      'document.removeEventListener("pointerdown", onDocumentPointerDown)',
    );
    expect(viewer).toContain('event.key === "Escape"');
    expect(viewer).toContain("details.contains(target)");
    expect(viewer.match(/closeMoreDetails\(\)/g) || []).toHaveLength(6);
  });

  it("uses quiet icon-only actions and compact icon statistics", () => {
    expect(viewer).toContain('class="gallery-detail-stats"');
    expect(viewer).toContain('aria-label="分享图片"');
    expect(viewer).toContain(
      ':aria-label="image.favorited ? \'取消收藏\' : \'收藏图片\'"',
    );
    expect(viewer).not.toContain('label="分享"');
    expect(viewer).not.toContain(
      ':label="image.favorited ? \'已收藏\' : \'收藏图片\'"',
    );
  });
});
