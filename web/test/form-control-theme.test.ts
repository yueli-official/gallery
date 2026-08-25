import fs from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";

const appConfig = fs.readFileSync(
  path.resolve(import.meta.dirname, "../app/app.config.ts"),
  "utf8",
);

describe("gallery form control theme", () => {
  it("delegates form controls to the shared single-border preset", () => {
    expect(appConfig).toContain("createUiPreset");
    expect(appConfig).not.toMatch(/\b(?:input|textarea|select|selectMenu):\s*\{/);
    expect(appConfig).not.toContain("focus-visible:ring-inset");
  });
});
