import { expect, test } from "@playwright/test";

import { productSites } from "../../../../../tests/e2e/contracts";
import { registerProductSuite } from "../../../../../tests/e2e/product-suite";

registerProductSuite("gallery");

const suite = process.env.PLATFORMCTL_E2E_SUITE?.trim() || "all";
if (suite === "all" || suite === "journeys") {
  for (const site of productSites("gallery")) {
    test(`${site.slug} Viewer 可连续浏览并显式关闭`, async ({ page }) => {
      await page.goto(
        new URL("/images/AZsQAAAAcACQAAAAAAAAAQ", site.url).toString(),
        { waitUntil: "domcontentloaded" },
      );
      await expect(page.locator(".gallery-viewer.has-image")).toBeVisible();
      await expect(
        page.locator('[data-navigation-ready="true"]'),
      ).toBeVisible();

      const visited = new Set([page.url()]);
      for (let index = 0; index < 12; index += 1) {
        const next = page.getByRole("button", { name: "查看下一张图片" });
        await expect(next).toBeVisible();
        const before = page.url();
        await next.click();
        await expect.poll(() => page.url()).not.toBe(before);
        await expect(page.locator(".gallery-viewer.has-image")).toBeVisible();
        visited.add(page.url());
      }
      expect(visited.size).toBeGreaterThanOrEqual(10);

      await page.getByRole("button", { name: "返回图片目录" }).click();
      await expect(page).toHaveURL(new URL("/images", site.url).toString());
    });
  }
}
