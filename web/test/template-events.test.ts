import { readFileSync, readdirSync } from "node:fs";
import { extname, join } from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

const appRoot = fileURLToPath(new URL("../app", import.meta.url));

function vueFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) {
      return vueFiles(path);
    }
    return extname(entry.name) === ".vue" ? [path] : [];
  });
}

describe("Gallery 模板事件", () => {
  it("不会把事件处理器写成只返回函数的箭头表达式", () => {
    const invalidFiles = vueFiles(appRoot).filter((path) =>
      /@(?:click|submit|change|update:[\w-]+)="\s*\(\)\s*=>/s.test(
        readFileSync(path, "utf8"),
      ),
    );

    expect(invalidFiles).toEqual([]);
  });

  it("client-only 异步页面保持服务端与客户端首帧一致", () => {
    const invalidFiles = vueFiles(appRoot).filter((path) => {
      const source = readFileSync(path, "utf8");
      return (
        /server:\s*false/.test(source) &&
        !source.includes("useClientHydrated()")
      );
    });

    expect(invalidFiles).toEqual([]);
  });
});
