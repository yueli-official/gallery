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
const styles = readFileSync(
  fileURLToPath(new URL("../app/assets/css/main.css", import.meta.url)),
  "utf8",
);

describe("Gallery surface polish", () => {
  it("composes the public header and layout with the shared container", () => {
    expect(publicHeader).toContain("<GalleryPublicContainer");
    expect(publicLayout.match(/<GalleryPublicContainer/g) || []).toHaveLength(2);
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
