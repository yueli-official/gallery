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

describe("Gallery mobile header", () => {
  it("renders the primary navigation as one equal-width mobile track", () => {
    expect(stylesheet).toMatch(
      /\.gallery-mobile-nav\s*\{[\s\S]*grid-template-columns:\s*repeat\(4,\s*minmax\(0,\s*1fr\)\)/,
    );
    expect(stylesheet).toMatch(
      /\.gallery-mobile-nav a\[aria-current="page"\]::after\s*\{[\s\S]*opacity:\s*1/,
    );
  });

  it("keeps the header search as a single icon at every breakpoint", () => {
    expect(header).not.toContain("<GalleryGlobalSearch");
    expect(header).toMatch(
      /to="\/images"[\s\S]*icon="i-tabler-search"[\s\S]*aria-label="搜索图库"/,
    );
    expect(header).not.toMatch(
      /class="xl:hidden"[\s\S]*to="\/images"[\s\S]*icon="i-tabler-search"/,
    );
  });
});
