import fs from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";

const appConfig = fs.readFileSync(
  path.resolve(import.meta.dirname, "../app/app.config.ts"),
  "utf8",
);

describe("gallery form control theme", () => {
  it.each(["input", "textarea", "select", "selectMenu"])(
    "makes %s fill its field container",
    (component) => {
      const block = appConfig.match(
        new RegExp(`${component}: \\{[\\s\\S]*?\\n    \\},`),
      )?.[0];
      expect(block).toContain("w-full");
    },
  );

  it("uses the existing inset border for focus instead of an outer outline", () => {
    expect(appConfig.match(/focus-visible:outline-none/g) || []).toHaveLength(4);
    expect(appConfig.match(/focus-visible:ring-inset/g) || []).toHaveLength(4);
    expect(appConfig).not.toContain("focus-visible:outline-3");
  });
});
