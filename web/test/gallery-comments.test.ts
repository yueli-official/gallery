import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

function source(relativePath: string) {
  return readFileSync(
    fileURLToPath(new URL(relativePath, import.meta.url)),
    "utf8",
  );
}

const viewer = source("../app/components/GalleryViewer.vue");
const section = source("../app/components/GalleryCommentSection.vue");
const publicStyles = source("../app/assets/css/main.css");
const management = source("../app/pages/manage/comments.vue");
const layout = source("../app/layouts/manage.vue");
const me = source("../app/composables/useGalleryMe.ts");

describe("Gallery comments", () => {
  it("attaches a two-level public thread to an Image rather than its Submission", () => {
    expect(viewer).toContain("<GalleryCommentSection");
    expect(viewer).toContain('v-show="!fullscreen"');
    expect(section).toContain("PublicCommentThread");
    expect(section).toContain("sortOrder: order.value");
    expect(section).not.toContain("submissionId");
    expect(section).toContain("匿名评论需要审核");
    expect(section).toContain("登录后直接发布");
    expect(section).not.toContain("GalleryCommentForm");
    expect(publicStyles).not.toMatch(
      /\.gallery-detail-comments\s*\{[^}]*border-top/,
    );
    expect(publicStyles).not.toMatch(
      /\.gallery-detail-related\s*\{[^}]*border-top/,
    );
    expect(section).toContain(
      "mt-12 w-full rounded-2xl bg-muted p-3 sm:p-6 lg:mt-16 lg:p-8",
    );
    expect(viewer).toContain("gallery-detail-related mt-20 lg:mt-28");
    expect(publicStyles).not.toMatch(/\.gallery-detail-comments\s*\{/);
    expect(publicStyles).not.toMatch(
      /\.gallery-detail-related\s*\{[^}]*(?:margin|padding|background)/,
    );
  });

  it("uses the shared collection and confirmation language for moderation", () => {
    expect(management).toContain("CommentModerationCollection");
    expect(management).toContain('icon: "i-tabler-photo"');
    expect(management).not.toContain("CollectionPanel");
    expect(management).not.toContain("CollectionSortHeader");
    expect(management).toContain('title="永久删除评论"');
    expect(management).toContain('label: "移入回收站"');
    expect(management).toContain('label: "永久删除"');
    expect(management).toContain("lifecycleChange: changeLifecycle");
    expect(management).toContain("emptyTrash");
    expect(management).not.toContain("window.confirm");
    expect(management).not.toContain("font-mono");
  });

  it("registers the comments page and explicit moderation capabilities", () => {
    expect(layout).toContain('label: "评论"');
    expect(layout).toContain('to: "/manage/comments"');
    for (const capability of [
      "gallery.comment.read",
      "gallery.comment.moderate",
      "gallery.comment.delete",
    ]) {
      expect(me).toContain(`"${capability}"`);
    }
  });
});
