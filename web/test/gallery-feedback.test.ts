import fs from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";

const root = path.resolve(import.meta.dirname, "..");
const actions = fs.readFileSync(
  path.join(root, "app/composables/useGalleryImageActions.ts"),
  "utf8",
);
const app = fs.readFileSync(path.join(root, "app/app.vue"), "utf8");
const feedback = fs.readFileSync(
  path.join(root, "app/utils/feedback.ts"),
  "utf8",
);

describe("gallery feedback", () => {
  it("copies share links on non-secure LAN origins", () => {
    expect(actions).toContain("window.isSecureContext");
    expect(actions).toContain('document.execCommand("copy")');
    expect(actions).toContain('title: "链接已复制"');
  });

  it("recovers a user-only favorite when the BFF fell back to Guest", () => {
    expect(actions).toContain('actionFailureCode(reason) === "gallery.forbidden"');
    expect(actions).toContain("await login()");
  });

  it("uses the shared bounded toast region", () => {
    expect(app).toContain(':toaster="null"');
    expect(app).toContain("FeedbackToastRegion");
    expect(feedback).toContain("createNuxtToastNotifier");
  });
});
