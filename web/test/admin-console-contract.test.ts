import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

function source(relativePath: string) {
  return readFileSync(
    fileURLToPath(new URL(relativePath, import.meta.url)),
    "utf8",
  );
}

const layout = source("../app/layouts/manage.vue");
const dashboard = source("../app/pages/manage/index.vue");
const images = source("../app/pages/manage/images.vue");
const submissions = source("../app/pages/manage/submissions.vue");
const cases = source("../app/pages/manage/cases.vue");
const collections = source("../app/pages/manage/collections/index.vue");

describe("Gallery management console contract", () => {
  it("uses the shared console language and a real traffic dashboard", () => {
    expect(layout).toContain('label: "控制台"');
    expect(layout).not.toContain('label: "今日"');
    expect(dashboard).toContain("DashboardTrendChart");
    expect(dashboard).toContain('"/admin/overview"');
    expect(dashboard).not.toContain("今日运营");
    expect(dashboard).not.toContain("需要你处理");
    expect(dashboard).not.toContain("继续工作");
  });

  it("uses sortBy/sortOrder column sorting and direct image actions", () => {
    expect(images).toContain("CollectionSortHeader");
    expect(images).toContain("sortBy");
    expect(images).toContain("sortOrder");
    expect(images).not.toMatch(/\bdirection\b/);
    expect(images).not.toContain("CollectionViewToggle");
    expect(images).toContain("openEdit(image)");
    expect(images).toContain("i-tabler-external-link");
  });

  it("keeps the remaining growing collections on one toolbar language", () => {
    expect(submissions).toContain("CollectionPanel");
    expect(submissions).toContain("CollectionSortHeader");
    expect(submissions).toContain("sortBy");
    expect(submissions).toContain("sortOrder");
    expect(cases).toContain("CollectionTableToolbar");
    expect(cases).toContain("CollectionSortHeader");
    expect(collections).toContain("CollectionTableToolbar");
  });
});
