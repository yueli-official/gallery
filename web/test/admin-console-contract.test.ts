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
const collectionEditor = source(
  "../app/pages/manage/collections/[collectionId].vue",
);
const categoryPanel = source(
  "../app/components/ClassificationCategoryPanel.vue",
);
const facetPanel = source("../app/components/ClassificationFacetPanel.vue");
const tagPanel = source("../app/components/ClassificationTagPanel.vue");
const proposalPanel = source(
  "../app/components/ClassificationProposalPanel.vue",
);
const nuxtConfig = source("../nuxt.config.ts");

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
    expect(images).toContain("CollectionViewToggle");
    expect(images).toContain("data-gallery-image-grid-item");
    expect(images).toContain('value: "add_to_collection"');
    expect(images).toContain("batchCollectionId");
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

  it("uses the shared collection and inspector patterns for topic editing", () => {
    expect(layout).toContain(':immersive="immersive"');
    expect(collectionEditor).toContain("EditorInspector");
    expect(collectionEditor).toContain("CollectionPanel");
    expect(collectionEditor).toContain("CollectionViewToggle");
    expect(collectionEditor).toContain("data-gallery-collection-commandbar");
    expect(collectionEditor).toContain('aria-label="专题设置"');
    expect(collectionEditor).toContain("data-gallery-collection-list-item");
    expect(collectionEditor).toContain("data-gallery-collection-grid-item");
    expect(collectionEditor).toContain("startMemberDrag");
    expect(collectionEditor).toContain('label: "移到顶部"');
    expect(collectionEditor).toContain('label: "移到底部"');
    expect(collectionEditor).toContain('label="批量移出专题"');
    expect(collectionEditor).toContain("memberSortPreset");
    expect(collectionEditor).not.toContain("gallery-manage-panel");
    expect(collectionEditor).toContain(
      '{ label: "SEO", value: "search", icon: "i-tabler-search" }',
    );
    const inspector = collectionEditor.slice(
      collectionEditor.indexOf("<EditorInspector"),
      collectionEditor.indexOf("</EditorInspector>"),
    );
    const fields = [
      ...inspector.matchAll(
        /<U(?:Input|Textarea|Select|SelectMenu)\b[\s\S]*?>/g,
      ),
    ].map((match) => match[0]);
    expect(fields).toHaveLength(7);
    for (const field of fields) expect(field).toMatch(/class="[^"]*w-full/);
  });

  it("keeps image and submission states concise and actionable", () => {
    expect(images).toContain('label: "待生成"');
    expect(images).toContain("imageStatusBadge(image)");
    expect(images).not.toContain('label: "公开版本未就绪"');
    expect(submissions).toContain('label="批量通过"');
    expect(submissions).toContain('label="通过"');
    expect(submissions).toContain('title="拒绝投稿"');
    expect(submissions).not.toContain("批准进入目录");
    expect(submissions).not.toContain("备注或拒绝");
  });

  it("renders classification identities as compact three-column rows", () => {
    for (const panel of [categoryPanel, facetPanel, tagPanel]) {
      expect(panel).toContain("AdminRowActions");
      expect(panel).toContain("grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]");
      expect(panel).not.toContain("<details");
      expect(panel).not.toContain("font-mono");
    }
  });

  it("keeps case stages and classification review tasks distinct", () => {
    expect(cases).toContain('{ label: "待结论", value: "reviewing" }');
    expect(cases).not.toContain('{ label: "处理中", value: "reviewing" }');
    expect(cases).toContain('label="填写结论"');
    expect(facetPanel).toContain('label="新增"');
    expect(facetPanel).not.toContain('label="新增值"');
    expect(proposalPanel).not.toContain("<details");
    expect(proposalPanel).not.toContain("submissionId");
    expect(proposalPanel).not.toContain("font-mono");
    expect(proposalPanel).toContain('size="xs"');
    expect(proposalPanel).not.toContain("min-h-11");
  });

  it("bundles shared dynamic controls for the first frame", () => {
    for (const icon of [
      "i-tabler-dots-vertical",
      "i-tabler-file-text",
      "i-tabler-layout-grid",
      "i-tabler-list",
      "i-tabler-photo-cog",
      "i-tabler-search",
    ]) {
      expect(nuxtConfig).toContain(`"${icon}"`);
    }
  });
});
