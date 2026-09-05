import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const home = readFileSync(
  fileURLToPath(new URL("../app/pages/index.vue", import.meta.url)),
  "utf8",
);
const manageLayout = readFileSync(
  fileURLToPath(new URL("../app/layouts/manage.vue", import.meta.url)),
  "utf8",
);
const publicLayout = readFileSync(
  fileURLToPath(new URL("../app/layouts/default.vue", import.meta.url)),
  "utf8",
);
const publicHeader = readFileSync(
  fileURLToPath(new URL("../app/components/GalleryHeader.vue", import.meta.url)),
  "utf8",
);
const publicContainer = readFileSync(
  fileURLToPath(
    new URL("../app/components/GalleryPublicContainer.vue", import.meta.url),
  ),
  "utf8",
);
const publicPage = readFileSync(
  fileURLToPath(
    new URL("../app/components/GalleryPublicPage.vue", import.meta.url),
  ),
  "utf8",
);
const publicPageHeader = readFileSync(
  fileURLToPath(
    new URL("../app/components/GalleryPageHeader.vue", import.meta.url),
  ),
  "utf8",
);
const styles = readFileSync(
  fileURLToPath(new URL("../app/assets/css/main.css", import.meta.url)),
  "utf8",
);

describe("Gallery surface polish", () => {
  it("owns the public site width through one Tailwind container template", () => {
    expect(publicContainer).toContain("mx-auto w-full max-w-[75rem]");
    expect(publicHeader).toContain("<GalleryPublicContainer");
    expect(publicLayout.match(/<GalleryPublicContainer/g) || []).toHaveLength(2);
    expect(publicPage).toContain(
      "gallery-page w-full px-4 pt-[1.1rem] sm:px-6 sm:pt-5 lg:px-8",
    );
    expect(publicPageHeader).toContain(
      "gallery-page-header mb-4 flex items-start justify-between gap-2",
    );
    expect(publicPageHeader).toContain("md:text-[length:var(--gallery-title-page)]");
    expect(styles).not.toContain("--gallery-public-max");
    expect(styles).not.toContain(".gallery-public-container");
    expect(styles).not.toMatch(/\.gallery-page\s*\{/);
    expect(styles).not.toMatch(/\.gallery-page-header\s*\{/);
    expect(styles).not.toMatch(/\.gallery-page-title\s*\{/);
    expect(styles).not.toContain("max-width: 100rem");
    expect(styles).not.toContain("max-width: 96rem");
  });

  it("keeps fixture provenance out of the public discovery surface", () => {
    expect(home).not.toContain("isSyntheticPreview");
    expect(home).not.toContain("演示数据");
  });

  it("uses a dedicated semantic palette for the management shell", () => {
    expect(manageLayout).toContain(
      'useHead({ bodyAttrs: { class: "gallery-manage-active" } });',
    );
    expect(styles).toContain("body.gallery-manage-active");
    expect(styles).toContain(".dark body.gallery-manage-active");
    expect(styles).toContain("--yueli-admin-shell:");
    expect(styles).toContain("--yueli-admin-canvas:");
    expect(styles).toContain("--yueli-admin-search:");
    expect(styles).toMatch(
      /\.gallery-manage-active \.admin-shell-sidebar[\s\S]*?border-inline-end-color:\s*var\(--ui-border-muted\)/,
    );
  });
});
