import { readFileSync, readdirSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

const styles = readFileSync(
  fileURLToPath(new URL("../app/assets/css/main.css", import.meta.url)),
  "utf8",
);
const appRoot = fileURLToPath(new URL("../app/", import.meta.url));

function findVueStyleBlocks(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = `${directory}/${entry.name}`;
    if (entry.isDirectory()) return findVueStyleBlocks(path);
    if (!entry.isFile() || !entry.name.endsWith(".vue")) return [];
    return /<style(?:\s|>)/u.test(readFileSync(path, "utf8")) ? [path] : [];
  });
}

describe("Gallery Tailwind ownership", () => {
  it("keeps the legacy global CSS baseline monotonically decreasing", () => {
    const lines = styles.trimEnd().split(/\r?\n/u).length;
    const selectorBlocks = (
      styles.match(/^[ \t]*[^@/\s][^{}\r\n]*\{[ \t]*$/gmu) || []
    ).length;
    const mediaQueries = (styles.match(/^[ \t]*@media\b/gmu) || []).length;

    expect(lines).toBeLessThanOrEqual(331);
    expect(selectorBlocks).toBeLessThanOrEqual(38);
    expect(mediaQueries).toBeLessThanOrEqual(2);
    expect(styles).not.toContain("@apply");
    expect(findVueStyleBlocks(appRoot)).toEqual([]);
  });

  it("does not return migrated layout ownership to global selectors", () => {
    for (const selector of [
      ".gallery-header",
      ".gallery-page",
      ".gallery-public-container",
      ".gallery-global-search",
      ".gallery-filter-panel",
      ".gallery-results-toolbar",
      ".gallery-grid",
      ".gallery-masonry",
      ".gallery-compact-empty",
      ".gallery-detail-comments",
      ".gallery-detail-info",
      ".gallery-detail-related",
      ".gallery-empty",
    ]) {
      expect(styles).not.toMatch(
        new RegExp(`(?:^|\\n)\\s*${selector.replaceAll(".", "\\.")}\\s*\\{`, "u"),
      );
    }
  });
});
