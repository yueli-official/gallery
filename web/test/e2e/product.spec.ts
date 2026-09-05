import { expect, test } from "@playwright/test";

import { productSites } from "./contracts";
import { registerProductSuite } from "./product-suite";
import { loginE2E, settleNuxt } from "./runtime";

registerProductSuite("gallery");

const suite = process.env.GALLERY_E2E_SUITE?.trim() || "all";
if (suite === "all" || suite === "journeys") {
  for (const site of productSites("gallery")) {
    test(`${site.slug} 新访客可以查看空投稿记录`, async ({ browser }) => {
      const context = await browser.newContext();
      const page = await context.newPage();
      try {
        const submissionsResponse = page.waitForResponse(
          (response) =>
            new URL(response.url()).pathname ===
            "/api/gallery/me/submissions",
        );
        await page.goto(new URL("/submissions", site.url).toString(), {
          waitUntil: "domcontentloaded",
        });
        expect((await submissionsResponse).status()).toBe(200);
        await expect(page.getByText("投稿记录加载失败")).toHaveCount(0);
        await expect(
          page.getByText("还没有投稿记录", { exact: true }),
        ).toBeVisible();
      } finally {
        await context.close();
      }
    });

    test(`${site.slug} 已登录用户可以查看投稿记录`, async ({ browser }) => {
      const context = await loginE2E(
        browser,
        {},
        undefined,
        site.url,
      );
      const page = await context.newPage();
      try {
        const submissionsResponse = page.waitForResponse(
          (response) =>
            new URL(response.url()).pathname ===
            "/api/gallery/me/submissions",
        );
        await page.goto(new URL("/submissions", site.url).toString(), {
          waitUntil: "domcontentloaded",
        });
        const response = await submissionsResponse;
        expect(response.status()).toBe(200);
        expect(
          (await response.json() as { total: number }).total,
        ).toBeGreaterThan(0);
        await expect(page.getByText("投稿记录加载失败")).toHaveCount(0);
      } finally {
        await context.close();
      }
    });

    test(`${site.slug} Viewer 可连续浏览并显式关闭`, async ({ page }) => {
      await page.goto(
        new URL("/images/AZsQAAAAcACQAAAAAAAAAQ", site.url).toString(),
        { waitUntil: "domcontentloaded" },
      );
      await expect(page.locator(".gallery-detail.has-image")).toBeVisible();
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
        await expect(page.locator(".gallery-detail.has-image")).toBeVisible();
        visited.add(page.url());
      }
      expect(visited.size).toBeGreaterThanOrEqual(10);

      await page.getByRole("button", { name: "返回图片目录" }).click();
      await expect(page).toHaveURL(new URL("/images", site.url).toString());
    });

    test(`${site.slug} 图册从目录页末尾继续到下一页第一张`, async ({ page }) => {
      await page.goto(new URL("/images", site.url).toString(), {
        waitUntil: "domcontentloaded",
      });
      await settleNuxt(page);

      const firstPageTiles = page.locator("a.gallery-tile-link");
      await expect(firstPageTiles).toHaveCount(24);
      const lastFirstPageHref = await firstPageTiles.last().getAttribute("href");
      expect(lastFirstPageHref).toBeTruthy();

      await page.getByRole("button", { name: "2", exact: true }).click();
      await expect(page).toHaveURL(/page=2/u);
      await expect(page.locator("a.gallery-tile-link")).toHaveCount(24);
      const firstSecondPageHref = await page
        .locator("a.gallery-tile-link")
        .first()
        .getAttribute("href");
      expect(firstSecondPageHref).toBeTruthy();

      await page.goto(new URL("/images", site.url).toString(), {
        waitUntil: "domcontentloaded",
      });
      await settleNuxt(page);
      await page.locator("a.gallery-tile-link").last().click();
      await expect(page.locator('[data-navigation-ready="true"]')).toBeVisible();
      await page.getByRole("button", { name: "进入图册模式" }).click();
      await expect
        .poll(() => page.evaluate(() => Boolean(document.fullscreenElement)))
        .toBe(true);

      await page.getByRole("button", { name: "查看下一张图片" }).click();
      await expect(page).toHaveURL(
        new URL(firstSecondPageHref!, site.url).toString(),
      );
      await expect
        .poll(() => page.evaluate(() => Boolean(document.fullscreenElement)))
        .toBe(true);
      await page.getByRole("button", { name: "查看上一张图片" }).click();
      await expect(page).toHaveURL(
        new URL(lastFirstPageHref!, site.url).toString(),
      );
    });
  }
}
