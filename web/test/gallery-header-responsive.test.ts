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
  it("renders the primary navigation as one equal-width mobile track", () => {
    expect(stylesheet).toMatch(
      /\.gallery-mobile-nav\s*\{[\s\S]*grid-template-columns:\s*repeat\(4,\s*minmax\(0,\s*1fr\)\)/,
    );
    expect(stylesheet).toMatch(
      /\.gallery-mobile-nav a\[aria-current="page"\]::after\s*\{[\s\S]*opacity:\s*1/,
    );
  });

  it("keeps one responsive search surface in the header and none in the home body", () => {
    expect(header).toContain(
      '<GalleryGlobalSearch class="gallery-header-search" compact />',
    );
    expect(header).toMatch(
      /v-if="route\.path !== '\/images'"[\s\S]*class="gallery-header-mobile-search lg:hidden"[\s\S]*<GalleryGlobalSearch compact \/>/,
    );
    expect(stylesheet).toMatch(
      /\.gallery-header-search\s*\{[\s\S]*display:\s*none/,
    );
    expect(stylesheet).toMatch(
      /@media \(min-width:\s*64rem\)[\s\S]*\.gallery-header-search\s*\{[\s\S]*display:\s*flex[\s\S]*\.gallery-header-mobile-search\s*\{[\s\S]*display:\s*none/,
    );
    expect(home).not.toContain("<GalleryGlobalSearch");
  });
});
