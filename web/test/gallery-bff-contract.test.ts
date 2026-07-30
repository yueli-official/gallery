import fs from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";

const root = path.resolve(import.meta.dirname, "..");
const genericRoute = path.join(
  root,
  "server/api/gallery/[...path].ts",
);

function source(relativePath: string) {
  return fs
    .readFileSync(path.join(root, relativePath), "utf8")
    .replace(/\r\n/gu, "\n");
}

function applicationSources(directory: string): string[] {
  return fs.readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const target = path.join(directory, entry.name);
    if (entry.isDirectory()) return applicationSources(target);
    return /\.(?:ts|vue)$/u.test(entry.name) ? [target] : [];
  });
}

describe("Gallery Web BFF contract", () => {
  it("registers Gallery as an explicit runtime target", () => {
    expect(source("nuxt.config.ts")).toContain(
      'gallery: {\n        path: "/api/gallery"',
    );
  });

  it("uses one method-agnostic Gallery BFF route", () => {
    expect(fs.existsSync(genericRoute)).toBe(true);
    expect(
      fs.existsSync(path.join(root, "server/api/gallery/[...path].get.ts")),
    ).toBe(false);
    expect(
      fs.existsSync(path.join(root, "server/api/gallery/[...path].post.ts")),
    ).toBe(false);
  });

  it("routes the reported submissions request through the Gallery client", () => {
    const submissions = source("app/pages/submissions.vue");
    expect(submissions).toContain("useGalleryApi()");
    expect(submissions).toContain('call<SubmissionPage>("/me/submissions"');
  });

  it("does not send Gallery API paths through the Identity client", () => {
    const offenders = applicationSources(path.join(root, "app"))
      .filter((file) => source(path.relative(root, file)).includes("/api/v1/gallery"))
      .map((file) => path.relative(root, file));
    expect(offenders).toEqual([]);
  });
});
