import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

const stylesheet = readFileSync(
  fileURLToPath(new URL("../app/assets/css/main.css", import.meta.url)),
  "utf8",
);
const header = readFileSync(
  fileURLToPath(
    new URL("../app/components/GalleryHeader.vue", import.meta.url),
  ),
  "utf8",
);
const home = readFileSync(
  fileURLToPath(new URL("../app/pages/index.vue", import.meta.url)),
  "utf8",
);

describe("Gallery mobile header", () => {
  it("keeps the home image stream free of a non-image lead tile", () => {
    expect(home).not.toContain("gallery-home-lead");
    expect(home).toMatch(/<h1[^>]*class="sr-only"[^>]*>/);
  });

  it("renders only operator-managed titles and actions for home sections", () => {
    expect(home).toContain("homeSection.title");
    expect(home).toContain(":label=\"homeSection.actionLabel\"");
    expect(home).not.toContain("homeSection.description");
    expect(home).not.toContain(">随机看看</h2>");
    expect(home).not.toContain('label="换一批"');
    expect(home).not.toContain('title="从专题进入"');
    expect(home).not.toContain('title="最新入库"');
    expect(home).not.toContain('title="正在被发现"');
    expect(home).not.toContain("featuredCategories");
    expect(home).not.toContain("featuredFacets");
    expect(home).not.toContain("gallery-discovery-group");
    expect(home).not.toContain("gallery-tag-search-hint");
  });

  it("moves the primary navigation below the header only under 768px", () => {
    expect(header).toContain(
      'class="gallery-desktop-nav hidden items-center gap-1 md:flex"',
    );
    expect(header).toContain('class="gallery-mobile-nav md:hidden"');
    expect(stylesheet).toMatch(
      /\.gallery-mobile-nav\s*\{[\s\S]*grid-template-columns:\s*repeat\(4,\s*minmax\(0,\s*1fr\)\)/,
    );
    expect(stylesheet).toMatch(
      /\.gallery-mobile-nav a\[aria-current="page"\]::after\s*\{[\s\S]*opacity:\s*1/,
    );
  });

  it("places a quiet submit link immediately before account controls", () => {
    expect(header).not.toMatch(
      /class="gallery-submit-button"[\s\S]*color="primary"/,
    );
    expect(header).toMatch(
      /<UColorModeButton[\s\S]*class="gallery-submit-link"[\s\S]*>投稿<\/NuxtLink>[\s\S]*<ConsumerAccountControl/,
    );
  });

  it("pins compact actions right and collapses search to an icon below 1024px", () => {
    expect(header).toMatch(
      /<GalleryGlobalSearch[\s\S]*class="gallery-header-search"[\s\S]*compact[\s\S]*:placeholder="props\.searchPlaceholder"/,
    );
    expect(header).toContain(
      'class="gallery-header-search-trigger"',
    );
    expect(header).not.toContain("gallery-header-mobile-search");
    expect(stylesheet).toMatch(
      /\.gallery-header-actions\s*\{[\s\S]*margin-left:\s*auto/,
    );
    expect(stylesheet).toMatch(
      /\.gallery-header-search\s*\{[\s\S]*display:\s*none/,
    );
    expect(stylesheet).toMatch(
      /\.gallery-header-search-trigger\s*\{[\s\S]*display:\s*grid/,
    );
    expect(stylesheet).toMatch(
      /@media \(min-width:\s*64rem\)[\s\S]*\.gallery-header-search\s*\{[\s\S]*display:\s*flex[\s\S]*\.gallery-header-search-trigger\s*\{[\s\S]*display:\s*none/,
    );
    expect(home).not.toContain("<GalleryGlobalSearch");
  });
});
