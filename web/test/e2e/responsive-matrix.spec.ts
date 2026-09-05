import { expect, test, type Page } from "@playwright/test";
import { productSites, type BrowserContract } from "./contracts";
import { expectNoHorizontalOverflow, loginE2E, settleNuxt } from "./runtime";

function scenario(contract: BrowserContract, key: string) {
  const value = contract.visual?.scenarios?.find((item) => item.key === key);
  if (!value) throw new Error(`缺少响应式验收场景：${key}`);
  return value;
}

async function openReady(
  page: Page,
  siteURL: string,
  route: { path: string; readySelector: string; status?: number },
) {
  const response = await page.goto(new URL(route.path, siteURL).toString(), {
    waitUntil: "domcontentloaded",
  });
  expect(response?.status()).toBe(route.status || 200);
  await expect(page.locator(route.readySelector).first()).toBeVisible();
  await settleNuxt(page);
  await expectNoHorizontalOverflow(page, `${route.path} `);
  await expect(page.locator("main").first()).toBeVisible();
}

export function registerResponsiveSuite(product: string) {
  for (const site of productSites(product)) {
    const contract = site.contract;
    test.describe(`${site.slug} (${site.product}) 极端响应式合同`, () => {
      test("320px 窄屏关键消费路径完整且无横向溢出", async ({ page }) => {
        await page.setViewportSize({ width: 320, height: 720 });
        for (const key of ["catalog", "collections", "viewer"]) {
          await openReady(page, site.url, scenario(contract, key));
        }
        await expect(
          page.getByRole("button", { name: "查看下一张图片" }),
        ).toBeVisible();
      });

      test("1920px 宽屏公共壳统一为 1200px 并保持居中", async ({ browser }) => {
        const viewport = { width: 1920, height: 1080 };
        const publicContext = await browser.newContext({ viewport });
        try {
          const page = await publicContext.newPage();
          await openReady(page, site.url, {
            path: contract.public.path,
            readySelector: contract.public.readySelector,
          });
          for (const selector of [
            ".gallery-header-inner",
            ".gallery-main",
            ".gallery-page",
            ".gallery-footer-inner",
          ]) {
            const bounds = await page.locator(selector).first().boundingBox();
            expect(bounds, selector).not.toBeNull();
            expect(bounds!.width, selector).toBeCloseTo(1200, 0);
            expect(
              Math.abs(bounds!.x - (viewport.width - bounds!.width) / 2),
              selector,
            ).toBeLessThanOrEqual(2);
          }

          await openReady(page, site.url, scenario(contract, "viewer"));
          const detailBounds = await page
            .locator(".gallery-detail")
            .boundingBox();
          const commentsBounds = await page
            .locator(".gallery-detail-comments")
            .boundingBox();
          const relatedBounds = await page
            .locator(".gallery-detail-related")
            .boundingBox();
          expect(detailBounds).not.toBeNull();
          expect(detailBounds!.width).toBeCloseTo(1200, 0);
          expect(commentsBounds).not.toBeNull();
          expect(relatedBounds).not.toBeNull();
          expect(commentsBounds!.width).toBeCloseTo(relatedBounds!.width, 0);
          expect(commentsBounds!.x).toBeCloseTo(relatedBounds!.x, 0);
          const sectionStyles = await page.evaluate(() => {
            const comments = getComputedStyle(
              document.querySelector<HTMLElement>(
                ".gallery-detail-comments",
              )!,
            );
            const related = getComputedStyle(
              document.querySelector<HTMLElement>(
                ".gallery-detail-related",
              )!,
            );
            return {
              commentBackground: comments.backgroundColor,
              relatedBackground: related.backgroundColor,
              commentBorder: comments.borderTopWidth,
              commentShadow: comments.boxShadow,
              commentMargin: Number.parseFloat(comments.marginTop),
              relatedMargin: Number.parseFloat(related.marginTop),
            };
          });
          expect(sectionStyles.commentBackground).not.toBe(
            sectionStyles.relatedBackground,
          );
          expect(sectionStyles.commentBorder).toBe("0px");
          expect(sectionStyles.commentShadow).toBe("none");
          expect(sectionStyles.relatedMargin).toBeGreaterThan(
            sectionStyles.commentMargin,
          );
        } finally {
          await publicContext.close();
        }
      });

      test("1920px 管理后台仍使用可用工作画布", async ({ browser }) => {
        const viewport = { width: 1920, height: 1080 };
        const context = await loginE2E(browser, { viewport });
        try {
          const managePage = await context.newPage();
          await openReady(
            managePage,
            site.url,
            scenario(contract, "manage-authorization"),
          );
          const manageBounds = await managePage
            .locator("#authorization")
            .boundingBox();
          expect(manageBounds).not.toBeNull();
          const panel = await managePage.locator("[data-admin-console-panel]").boundingBox();
          expect(panel).not.toBeNull();
          expect(manageBounds!.x).toBeGreaterThanOrEqual(panel!.x - 1);
          expect(manageBounds!.x + manageBounds!.width).toBeLessThanOrEqual(panel!.x + panel!.width + 1);
          expect(manageBounds!.width).toBeGreaterThanOrEqual(Math.min(1200, panel!.width));
        } finally {
          await context.close();
        }
      });

      test("200% 有效缩放下目录可重排且关键控件可操作", async ({ browser }) => {
        const context = await browser.newContext({
          viewport: { width: 640, height: 450 },
          screen: { width: 1280, height: 900 },
          deviceScaleFactor: 2,
        });
        try {
          const page = await context.newPage();
          await openReady(page, site.url, scenario(contract, "catalog"));
          expect(
            await page.evaluate(() => ({
              width: window.innerWidth,
              pixelRatio: window.devicePixelRatio,
            })),
          ).toEqual({ width: 640, pixelRatio: 2 });
          await expect(
            page.getByRole("button", { name: "筛选", exact: true }),
          ).toBeVisible();
          await page.getByRole("button", { name: "筛选", exact: true }).click();
          await expect(page.getByRole("dialog")).toBeVisible();
        } finally {
          await context.close();
        }
      });

      test("目录筛选在桌面与抽屉断点之间保持单一入口", async ({ page }) => {
        await page.setViewportSize({ width: 1023, height: 800 });
        await openReady(page, site.url, scenario(contract, "catalog"));
        const filterButton = page.getByRole("button", {
          name: "筛选",
          exact: true,
        });
        const desktopFilters = page.getByRole("complementary", {
          name: "图片过滤器",
        });
        await expect(filterButton).toBeVisible();
        await expect(desktopFilters).toBeHidden();
        await filterButton.click();
        await expect(page.getByRole("dialog")).toBeVisible();

        await page.setViewportSize({ width: 1024, height: 800 });
        await expect(desktopFilters).toBeVisible();
        await expect(filterButton).toBeHidden();
        await expect(page.getByRole("dialog")).toHaveCount(0);
        await expectNoHorizontalOverflow(page, "/images 1024px ");
      });
    });
  }
}
