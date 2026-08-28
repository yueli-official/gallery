import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

function source(path: string) {
  return readFileSync(fileURLToPath(new URL(path, import.meta.url)), "utf8");
}

const page = source("../app/pages/images/index.vue");
const fields = source("../app/components/GalleryCatalogFilterFields.vue");
const state = source("../app/composables/useGalleryCatalogState.ts");

describe("Gallery public catalog filters", () => {
  it("renders category, facet and counted tag candidates in both filter surfaces", () => {
    expect(page).toContain(':tags="tagCandidates"');
    expect(page).toContain(':selected-tag="selectedTag"');
    expect(page).toContain('@toggle-tag="toggleTag"');
    expect(fields).toContain("GalleryTagCandidate");
    expect(fields).toContain("热门标签");
    expect(fields).toContain("gallery-tag-filter-count");
    expect(fields).toContain(':aria-pressed="selectedTag === tag.slug"');
  });

  it("stages tag choice with category and facet choices before applying the URL", () => {
    expect(state).toContain('const selectedTag = ref("")');
    expect(state).toContain("tag: selectedTag.value");
    expect(state).toContain("function toggleTag(value: string)");
  });
});
