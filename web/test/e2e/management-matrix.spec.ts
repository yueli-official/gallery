import AxeBuilder from "@axe-core/playwright";
import { expect, test, type BrowserContext, type Page } from "@playwright/test";
import { productSites } from "./contracts";
import {
  capturePageFailures,
  ensureRegisteredE2EIdentity,
  loginE2E,
  settleNuxt,
} from "./runtime";

const managementRoutes = [
  "/manage",
  "/manage/images",
  "/manage/comments",
  "/manage/submissions",
  "/manage/collections",
  "/manage/classification",
  "/manage/cases",
  "/manage/discovery",
  "/manage/assets",
  "/manage/authorization",
] as const;

async function saveDiscoverySettings(page: Page): Promise<void> {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.request().method() === "PATCH" &&
      candidate.url().includes("/api/gallery/admin/site-settings"),
  );
  await page
    .locator("main")
    .getByRole("button", { name: "保存设置", exact: true })
    .first()
    .click();
  expect((await response).ok()).toBeTruthy();
  await expect(
    page.getByText("设置已保存，前台刷新后生效。", { exact: true }),
  ).toBeVisible();
}

async function restoreDiscoveryName(
  page: Page,
  siteURL: string,
  originalName: string,
): Promise<void> {
  await page.goto(new URL("/manage/discovery", siteURL).toString(), {
    waitUntil: "domcontentloaded",
  });
  await settleNuxt(page);
  const field = page
    .getByRole("textbox", { name: "站点名称", exact: false })
    .first();
  if ((await field.inputValue()) === originalName) return;
  await field.fill(originalName);
  await saveDiscoverySettings(page);
}

async function openFirstImageEditor(page: Page): Promise<void> {
  await page
    .getByRole("button", { name: /^编辑图片/ })
    .first()
    .click();
  await expect(page.getByRole("dialog", { name: "编辑图片" })).toBeVisible();
  await expect(
    page
      .getByRole("dialog", { name: "编辑图片" })
      .getByRole("textbox", { name: "标题", exact: false }),
  ).toBeVisible();
}

async function saveImageEditor(page: Page): Promise<void> {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.request().method() === "PATCH" &&
      /\/api\/gallery\/admin\/images\/[^/?]+$/.test(candidate.url()),
  );
  await page
    .getByRole("dialog", { name: "编辑图片" })
    .getByRole("button", { name: "保存更改", exact: true })
    .click();
  expect((await response).ok()).toBeTruthy();
  await expect(page.getByRole("dialog", { name: "编辑图片" })).toHaveCount(0);
}

async function searchManagedImages(page: Page, query: string): Promise<void> {
  const search = page.getByPlaceholder("搜索标题、说明或替代文本…");
  await search.fill(query);
  await search.press("Enter");
  await expect(page.getByText(query, { exact: true }).first()).toBeVisible();
}

async function restoreImageTitle(
  page: Page,
  siteURL: string,
  temporaryTitle: string,
  originalTitle: string,
): Promise<void> {
  await page.goto(new URL("/manage/images", siteURL).toString(), {
    waitUntil: "domcontentloaded",
  });
  await settleNuxt(page);
  await searchManagedImages(page, temporaryTitle);
  await openFirstImageEditor(page);
  const title = page
    .getByRole("dialog", { name: "编辑图片" })
    .getByRole("textbox", { name: "标题", exact: false });
  if ((await title.inputValue()) === originalTitle) return;
  await title.fill(originalTitle);
  await saveImageEditor(page);
}

async function saveCollectionSettings(page: Page): Promise<void> {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.request().method() === "PATCH" &&
      /\/api\/gallery\/admin\/collections\/[^/?]+$/.test(candidate.url()),
  );
  const save = page
    .locator("[data-gallery-collection-commandbar]")
    .getByRole("button", { name: "保存", exact: true });
  await save.click();
  expect((await response).ok()).toBeTruthy();
  await expect(
    page
      .locator("[data-gallery-collection-commandbar]")
      .getByRole("button", { name: "已保存", exact: true }),
  ).toBeVisible();
}

async function openCollectionSettings(page: Page): Promise<void> {
  const inspector = page.getByRole("heading", {
    name: "专题设置",
    exact: true,
  });
  if (await inspector.isVisible().catch(() => false)) return;
  await page.getByRole("button", { name: "专题设置", exact: true }).click();
  await expect(inspector).toBeVisible();
}

async function restoreCollectionName(
  page: Page,
  collectionURL: string,
  originalName: string,
): Promise<void> {
  await page.goto(collectionURL, { waitUntil: "domcontentloaded" });
  await settleNuxt(page);
  await openCollectionSettings(page);
  const name = page
    .getByRole("textbox", { name: "名称", exact: false })
    .first();
  if ((await name.inputValue()) === originalName) return;
  await name.fill(originalName);
  await saveCollectionSettings(page);
}

async function authenticatedPage(
  context: BrowserContext,
  siteURL: string,
): Promise<Page> {
  const page = await context.newPage();
  await page.goto(new URL("/manage", siteURL).toString(), {
    waitUntil: "domcontentloaded",
  });
  await expect(page).toHaveURL(new URL("/manage", siteURL).toString());
  await settleNuxt(page);
  return page;
}

export function registerManagementSuite(product: string) {
  for (const site of productSites(product)) {
    test.describe(`${site.slug} (${site.product}) management contract`, () => {
      test("匿名调用 Gallery 管理 API 被拒绝", async ({ request }) => {
        for (const path of [
          "/api/gallery/admin/overview",
          "/api/gallery/admin/site-settings",
          "/api/gallery/admin/images?size=1",
          "/api/gallery/admin/comments?size=1",
          "/api/gallery/admin/collections",
        ]) {
          const response = await request.get(
            new URL(path, site.url).toString(),
          );
          expect(
            [401, 403],
            `${path} unexpectedly returned HTTP ${response.status()}`,
          ).toContain(response.status());
        }
      });

      test("普通会员无法进入管理页面或调用管理 API", async ({ browser }) => {
        const context = await ensureRegisteredE2EIdentity(
          browser,
          {
            email: "gallery-member-acceptance@example.test",
            password: "Gallery-member-acceptance-2026!",
          },
          "Gallery Acceptance Member",
          site.url,
        );
        const page = await context.newPage();
        try {
          await page.goto(new URL("/manage", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await expect(page).toHaveURL(new URL("/", site.url).toString());
          await settleNuxt(page);

          for (const path of [
            "/api/gallery/admin/overview",
            "/api/gallery/admin/site-settings",
            "/api/gallery/admin/images?size=1",
            "/api/gallery/admin/collections",
          ]) {
            const response = await context.request.get(
              new URL(path, site.url).toString(),
            );
            expect(
              response.status(),
              `${path} unexpectedly returned HTTP ${response.status()}`,
            ).toBe(403);
          }
        } finally {
          await context.close();
        }
      });

      test("管理员可无 5xx 与浏览器错误遍历全部管理入口", async ({
        browser,
      }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        const failures = capturePageFailures(page);
        try {
          for (const path of managementRoutes) {
            const url = new URL(path, site.url).toString();
            const response = await page.goto(url, {
              waitUntil: "domcontentloaded",
            });
            expect(
              response?.ok(),
              `${path} returned ${response?.status()}`,
            ).toBeTruthy();
            await expect(page).toHaveURL(url);
            await expect(page.locator("main h1").first()).toBeVisible();
            await settleNuxt(page);
          }
          expect(failures).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("图片管理使用统一页头并让搜索与筛选保持同一行", async ({
        browser,
      }) => {
        const context = await loginE2E(browser, {
          viewport: { width: 1440, height: 900 },
        });
        const page = await authenticatedPage(context, site.url);
        const errors = capturePageFailures(page);
        try {
          await page.goto(new URL("/manage/images", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);

          const managePage = page.locator('[data-manage-page][id="images"]');
          await expect(managePage).toBeVisible();
          await expect(
            managePage.locator("[data-manage-page-icon]"),
          ).toBeVisible();
          await expect(
            managePage.getByRole("heading", { name: "图片" }),
          ).toBeVisible();
          await expect(
            managePage.getByRole("link", { name: "投稿图片" }),
          ).toBeVisible();

          const header = managePage.locator("[data-manage-page-header]");
          const collection = managePage
            .locator('section[aria-label="图片列表"]')
            .first();
          const search = collection.locator("[data-collection-table-search]");
          const controls = collection.locator(
            "[data-collection-table-controls]",
          );
          const searchButton = search.getByRole("button", { name: "搜索" });
          const filterButton = controls.getByRole("button", { name: /^筛选/u });
          const [
            headerBox,
            collectionBox,
            searchBox,
            controlsBox,
            searchButtonBox,
            filterButtonBox,
          ] = await Promise.all([
            header.boundingBox(),
            collection.boundingBox(),
            search.boundingBox(),
            controls.boundingBox(),
            searchButton.boundingBox(),
            filterButton.boundingBox(),
          ]);
          expect(headerBox).toBeTruthy();
          expect(collectionBox).toBeTruthy();
          expect(searchBox).toBeTruthy();
          expect(controlsBox).toBeTruthy();
          expect(searchButtonBox).toBeTruthy();
          expect(filterButtonBox).toBeTruthy();
          expect(
            collectionBox!.y - (headerBox!.y + headerBox!.height),
          ).toBeGreaterThanOrEqual(18);
          expect(
            Math.abs(
              searchBox!.y +
                searchBox!.height / 2 -
                (controlsBox!.y + controlsBox!.height / 2),
            ),
          ).toBeLessThanOrEqual(2);
          expect(searchButtonBox!.height).toBe(filterButtonBox!.height);
          const statusBadges = collection.locator("[data-image-status-badge]");
          expect(await statusBadges.count()).toBeGreaterThan(0);
          expect(
            await statusBadges.evaluateAll((elements) =>
              elements.every(
                (element) => (element.textContent?.trim().length || 0) <= 4,
              ),
            ),
          ).toBeTruthy();
          expect(errors).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("评论后台复用集合治理并支持审核与删除确认", async ({ browser }) => {
        const context = await loginE2E(
          browser,
          { viewport: { width: 1440, height: 900 } },
          undefined,
          site.url,
        );
        const page = await context.newPage();
        let status = "pending";
        let patchBody: Record<string, unknown> | undefined;
        await page.route("**/api/gallery/me", (route) =>
          route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              isAdministrator: true,
              roles: ["administrator"],
              capabilities: [
                "gallery.dashboard.read",
                "gallery.comment.read",
                "gallery.comment.moderate",
                "gallery.comment.delete",
              ],
            }),
          }),
        );
        await page.route("**/api/gallery/admin/comments**", async (route) => {
          if (route.request().method() === "PATCH") {
            patchBody = route.request().postDataJSON();
            status = String(patchBody?.status || status);
            await route.fulfill({
              status: 200,
              contentType: "application/json",
              body: JSON.stringify({ comment: { id: "comment-1", status } }),
            });
            return;
          }
          if (route.request().method() === "DELETE") {
            await route.fulfill({
              status: 200,
              contentType: "application/json",
              body: JSON.stringify({ deleted: true }),
            });
            return;
          }
          await route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({
              total: 1,
              page: 1,
              size: 20,
              items: [
                {
                  id: "comment-1",
                  imageId: "AZsQAAAAcACQAAAAAAAAAQ",
                  imageTitle: "海岸晨光 001",
                  authorName: "访客",
                  authorEmail: "visitor@example.test",
                  content: "等待审核的图片评论",
                  status,
                  createdAt: "2026-07-11T08:15:00Z",
                },
              ],
            }),
          });
        });
        try {
          await page.goto(new URL("/manage/comments", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await expect(
            page.getByRole("heading", { name: "评论", exact: true }),
          ).toBeVisible();
          const comments = page.locator("[data-comment-moderation-collection]");
          await expect(
            comments.getByText("来源", { exact: true }),
          ).toBeVisible();
          await expect(comments.getByText("图片", { exact: true })).toHaveCount(
            0,
          );
          await expect(
            page.getByText("等待审核的图片评论", { exact: true }),
          ).toBeVisible();
          await page.getByRole("button", { name: "通过", exact: true }).click();
          expect(patchBody).toEqual({ status: "approved" });

          const actions = page.getByRole("button", { name: "评论操作：访客" });
          await actions.click();
          await page.getByRole("menuitem", { name: "移入回收站" }).click();
          expect(patchBody).toEqual({ status: "trash" });

          await page
            .getByRole("button", { name: "回收站", exact: true })
            .click();
          await actions.click();
          await page.getByRole("menuitem", { name: "永久删除" }).click();
          const dialog = page.getByRole("dialog", { name: "永久删除评论" });
          await expect(dialog).toBeVisible();
          await expect(dialog.getByText(/永久删除.*回复/u)).toBeVisible();
          await dialog.getByRole("button", { name: "取消" }).click();
          await expect(dialog).toHaveCount(0);

          await page.getByRole("button", { name: "清空回收站" }).click();
          const emptyTrash = page.getByRole("dialog", { name: "清空回收站" });
          await expect(emptyTrash).toBeVisible();
          await emptyTrash.getByRole("button", { name: "取消" }).click();
          const accessibility = await new AxeBuilder({ page })
            .exclude("nuxt-devtools-frame")
            .analyze();
          expect(
            accessibility.violations.filter((violation) =>
              ["serious", "critical"].includes(violation.impact || ""),
            ),
          ).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("共享评论治理五列网格样式已交付", async ({ page }) => {
        await page.setViewportSize({ width: 1440, height: 900 });
        await page.goto(site.url, { waitUntil: "domcontentloaded" });
        await settleNuxt(page);
        const columns = await page.evaluate(() => {
          const probe = document.createElement("div");
          probe.className =
            "grid lg:grid-cols-[minmax(16rem,1.4fr)_minmax(10rem,0.8fr)_10rem_8.5rem_7rem]";
          probe.style.position = "fixed";
          probe.style.inset = "0";
          probe.style.visibility = "hidden";
          document.body.append(probe);
          const value = getComputedStyle(probe).gridTemplateColumns;
          probe.remove();
          return value;
        });
        expect(columns).not.toBe("none");
        expect(columns.trim().split(/\s+/u)).toHaveLength(5);
      });

      test("控制台页头与指标卡使用统一页面节奏", async ({ browser }) => {
        const context = await loginE2E(browser, {
          viewport: { width: 1440, height: 900 },
        });
        const page = await authenticatedPage(context, site.url);
        try {
          await page.goto(new URL("/manage", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          const managePage = page.locator('[data-manage-page][id="dashboard"]');
          const header = managePage.locator("[data-manage-page-header]");
          const metric = managePage
            .locator("[data-gallery-dashboard-metric]")
            .first();
          const metricGroup = metric.locator("..");
          const trend = managePage.locator("[data-gallery-dashboard-trend]");
          const [headerBox, metricBox, metricGroupBox, trendBox, metricStyle] =
            await Promise.all([
              header.boundingBox(),
              metric.boundingBox(),
              metricGroup.boundingBox(),
              trend.boundingBox(),
              metric.evaluate((element) => {
                const style = getComputedStyle(element);
                return {
                  paddingInline: style.paddingInline,
                  paddingBlock: style.paddingBlock,
                  columns: style.gridTemplateColumns.split(" ").length,
                };
              }),
            ]);
          expect(headerBox).toBeTruthy();
          expect(metricBox).toBeTruthy();
          expect(metricGroupBox).toBeTruthy();
          expect(trendBox).toBeTruthy();
          expect(Math.abs(metricBox!.x - headerBox!.x)).toBeLessThanOrEqual(1);
          expect(metricBox!.y - (headerBox!.y + headerBox!.height)).toBe(20);
          expect(
            trendBox!.y - (metricGroupBox!.y + metricGroupBox!.height),
          ).toBe(20);
          expect(metricStyle).toEqual({
            paddingInline: "16px",
            paddingBlock: "16px",
            columns: 2,
          });
        } finally {
          await context.close();
        }
      });

      test("后台共享标题与侧栏默认配置完整生效", async ({ browser }) => {
        const context = await loginE2E(browser, {
          viewport: { width: 1280, height: 720 },
        });
        const page = await authenticatedPage(context, site.url);
        try {
          await page.goto(new URL("/manage/submissions", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          const headingStyle = await page
            .locator("[data-manage-page-header] h1")
            .evaluate((element) => {
              const style = getComputedStyle(element);
              return {
                fontSize: style.fontSize,
                lineHeight: style.lineHeight,
                fontFamily: style.fontFamily,
                letterSpacing: style.letterSpacing,
              };
            });
          expect(headingStyle).toMatchObject({
            fontSize: "30px",
            lineHeight: "36px",
            letterSpacing: "-0.75px",
          });
          expect(headingStyle.fontFamily).toContain("Space Grotesk");

          const activeLink = page
            .locator('a[href="/manage/submissions"]')
            .first();
          const navigationStyle = await activeLink.evaluate((element) => {
            const style = getComputedStyle(element);
            const after = getComputedStyle(element, "::after");
            const icon = element.querySelector<HTMLElement>(".iconify");
            return {
              background: style.backgroundColor,
              shadow: style.boxShadow,
              afterWidth: after.width,
              afterHeight: after.height,
              maskSize: icon ? getComputedStyle(icon).maskSize : "",
            };
          });
          expect(navigationStyle.background).not.toBe("rgba(0, 0, 0, 0)");
          expect(navigationStyle.shadow).toContain("inset");
          expect(navigationStyle.afterWidth).toBe("32px");
          expect(navigationStyle.afterHeight).toBe("32px");
          expect(navigationStyle.maskSize).toBe("16px 16px");
        } finally {
          await context.close();
        }
      });

      test("审核与权限任务使用带图标的统一页签并收口专题命名", async ({
        browser,
      }) => {
        const context = await loginE2E(browser, {
          viewport: { width: 1440, height: 900 },
        });
        const page = await authenticatedPage(context, site.url);
        const errors = capturePageFailures(page);
        try {
          const tabbedPages = [
            {
              path: "/manage/submissions",
              surface: "submissions",
              tabs: ["待审核", "处理失败", "安全不确定", "全部投稿"],
            },
            {
              path: "/manage/cases",
              surface: "cases",
              tabs: ["待处理", "待结论", "已解决", "已忽略", "全部"],
            },
            {
              path: "/manage/authorization",
              surface: "authorization",
              tabs: ["申请", "权限", "用户管理"],
            },
          ] as const;

          for (const target of tabbedPages) {
            await page.goto(new URL(target.path, site.url).toString(), {
              waitUntil: "domcontentloaded",
            });
            await settleNuxt(page);

            const surface = page.locator(
              `[data-manage-tabbed-surface][data-manage-surface="${target.surface}"]`,
            );
            await expect(surface).toBeVisible();
            for (const label of target.tabs) {
              const tab = surface.getByRole("tab", {
                name: new RegExp(`^${label}`),
              });
              await expect(tab).toBeVisible();
              const icon = tab.locator('[data-slot="leadingIcon"]');
              await expect(icon).toBeVisible();
              expect(
                await icon.evaluate((element) => {
                  const style = getComputedStyle(element);
                  return (
                    style.maskImage !== "none" ||
                    style.webkitMaskImage !== "none" ||
                    style.backgroundImage !== "none"
                  );
                }),
              ).toBeTruthy();
            }
            const accessibility = await new AxeBuilder({ page })
              .exclude("nuxt-devtools-frame")
              .analyze();
            expect(
              accessibility.violations.filter((violation) =>
                ["serious", "critical"].includes(violation.impact || ""),
              ),
            ).toEqual([]);
          }

          await page.goto(new URL("/manage/cases", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          const takeCase = page
            .getByRole("button", { name: "接手处理" })
            .first();
          await takeCase.hover();
          await expect(
            page.getByText("接手处理", { exact: true }).last(),
          ).toBeVisible();
          await page.mouse.move(0, 0);
          await expect(
            page.getByText("接手处理", { exact: true }).last(),
          ).toBeHidden();
          const relatedImage = page
            .getByRole("link", { name: "查看关联图片" })
            .first();
          await relatedImage.hover();
          await expect(
            page.getByText("查看关联图片", { exact: true }).last(),
          ).toBeVisible();

          await page.goto(
            new URL("/manage/cases?status=reviewing", site.url).toString(),
            { waitUntil: "domcontentloaded" },
          );
          await settleNuxt(page);
          await expect(
            page.getByRole("tab", { name: "待结论", selected: true }),
          ).toBeVisible();
          await expect(
            page.getByRole("button", { name: "填写结论" }).first(),
          ).toBeVisible();

          await page.goto(
            new URL("/manage/authorization", site.url).toString(),
            { waitUntil: "domcontentloaded" },
          );
          await settleNuxt(page);
          const authorization = page.locator(
            '[data-manage-surface="authorization"]',
          );
          await authorization.getByRole("tab", { name: "权限" }).click();
          await expect(page).toHaveURL(/tab=permissions/);
          await expect(
            authorization.getByRole("heading", { name: "角色与能力" }),
          ).toBeVisible();

          await page.goto(new URL("/manage/collections", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await expect(
            page.getByRole("heading", { name: "专题", exact: true }),
          ).toBeVisible();
          await expect(page.getByText("专题策展", { exact: true })).toHaveCount(
            0,
          );

          await page.setViewportSize({ width: 390, height: 844 });
          for (const target of tabbedPages.slice(0, 2)) {
            await page.goto(new URL(target.path, site.url).toString(), {
              waitUntil: "domcontentloaded",
            });
            await settleNuxt(page);
            const navigation = page.locator(
              `[data-manage-surface="${target.surface}"] > nav`,
            );
            const [width, navigationBox, firstTabBox, lastTabBox] =
              await Promise.all([
                navigation.evaluate((element) => ({
                  client: element.clientWidth,
                  scroll: element.scrollWidth,
                })),
                navigation.boundingBox(),
                navigation.getByRole("tab").first().boundingBox(),
                navigation.getByRole("tab").last().boundingBox(),
              ]);
            expect(width.scroll).toBeLessThanOrEqual(width.client + 1);
            expect(navigationBox).toBeTruthy();
            expect(firstTabBox).toBeTruthy();
            expect(lastTabBox).toBeTruthy();
            expect(firstTabBox!.x).toBeGreaterThanOrEqual(navigationBox!.x - 1);
            expect(lastTabBox!.x + lastTabBox!.width).toBeLessThanOrEqual(
              navigationBox!.x + navigationBox!.width + 1,
            );
          }
          expect(errors).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("投稿审核使用简短动作并把拒绝原因收进确认层", async ({
        browser,
      }) => {
        const context = await loginE2E(browser, {
          viewport: { width: 1440, height: 900 },
        });
        const page = await authenticatedPage(context, site.url);
        try {
          await page.goto(new URL("/manage/submissions", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await expect(
            page.getByRole("button", { name: "通过" }).first(),
          ).toBeVisible();
          const reject = page.getByRole("button", { name: "拒绝" }).first();
          await expect(reject).toBeVisible();
          await reject.click();
          const dialog = page.getByRole("dialog", { name: "拒绝投稿" });
          await expect(dialog).toBeVisible();
          await expect(
            dialog.getByRole("textbox", { name: "拒绝原因" }),
          ).toBeVisible();
          await dialog.getByRole("button", { name: "取消" }).click();
          await expect(dialog).toHaveCount(0);

          await page.getByRole("tab", { name: /^处理失败/u }).click();
          await settleNuxt(page);
          await expect(
            page.getByText("derive_failed", { exact: true }),
          ).toHaveCount(0);
          await expect(
            page.getByText("媒体处理失败", { exact: true }),
          ).toHaveCount(0);
          await expect(
            page.locator("[data-submission-decision]").first(),
          ).toBeVisible();
        } finally {
          await context.close();
        }
      });

      test("资源策略可以选择获授权的存储后端", async ({ browser }) => {
        const context = await loginE2E(browser, {
          viewport: { width: 1440, height: 900 },
        });
        const page = await authenticatedPage(context, site.url);
        const errors = capturePageFailures(page);
        try {
          await page.goto(new URL("/manage/assets", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await page.locator("[data-asset-registration-edit]").click();

          const backend = page.locator("[data-asset-storage-backend]");
          await expect(backend).toBeVisible();
          await expect(backend).toBeEnabled();
          await backend.click();
          await expect(
            page.getByRole("option", { name: "腾讯云 COS · blog" }),
          ).toBeVisible();
          await expect(
            page.getByRole("option", { name: "本地存储 · local" }),
          ).toBeVisible();
          await page.keyboard.press("Escape");
          await page.getByRole("button", { name: "取消", exact: true }).click();
          expect(errors).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("站点设置保存后持久化且测试结束恢复原值", async ({ browser }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        let originalName = "";
        try {
          await page.goto(new URL("/manage/discovery", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          const settingsSurface = page.locator(
            '[data-manage-surface="site-settings"]',
          );
          await expect(
            page.getByRole("heading", { name: "站点设置", exact: true }),
          ).toBeVisible();
          await expect(settingsSurface).toBeVisible();
          for (const label of ["站点", "首页", "发现"]) {
            const tab = settingsSurface.getByRole("tab", {
              name: label,
              exact: true,
            });
            await expect(tab).toBeVisible();
            const icon = tab.locator('[data-slot="leadingIcon"]');
            await expect(icon).toBeVisible();
            expect(
              await icon.evaluate((element) => {
                const style = getComputedStyle(element);
                return (
                  style.maskImage !== "none" ||
                  style.webkitMaskImage !== "none" ||
                  style.backgroundImage !== "none"
                );
              }),
            ).toBeTruthy();
          }
          await settingsSurface.getByRole("tab", { name: "首页" }).click();
          await expect(page).toHaveURL(/section=home/);
          await expect(
            settingsSurface.getByRole("heading", { name: "首页信息" }),
          ).toBeVisible();
          await settingsSurface.getByRole("tab", { name: "发现" }).click();
          await expect(page).toHaveURL(/section=discovery/);
          await expect(
            settingsSurface.getByRole("heading", { name: "随机发现" }),
          ).toBeVisible();
          await settingsSurface.getByRole("tab", { name: "站点" }).click();
          await expect(page).not.toHaveURL(/section=/);
          const accessibility = await new AxeBuilder({ page })
            .exclude("nuxt-devtools-frame")
            .analyze();
          expect(
            accessibility.violations.filter((violation) =>
              ["serious", "critical"].includes(violation.impact || ""),
            ),
          ).toEqual([]);
          const field = page
            .getByRole("textbox", { name: "站点名称", exact: false })
            .first();
          originalName = await field.inputValue();
          const temporaryName = `${originalName.slice(0, 65)} · 验收`;
          await field.fill(temporaryName);
          await saveDiscoverySettings(page);

          await page.reload({ waitUntil: "domcontentloaded" });
          await settleNuxt(page);
          await expect(
            page
              .getByRole("textbox", { name: "站点名称", exact: false })
              .first(),
          ).toHaveValue(temporaryName);

          await page
            .getByRole("textbox", { name: "站点名称", exact: false })
            .first()
            .fill(originalName);
          await saveDiscoverySettings(page);
          await page.reload({ waitUntil: "domcontentloaded" });
          await settleNuxt(page);
          await expect(
            page
              .getByRole("textbox", { name: "站点名称", exact: false })
              .first(),
          ).toHaveValue(originalName);
        } finally {
          if (originalName)
            await restoreDiscoveryName(page, site.url, originalName).catch(
              () => undefined,
            );
          await context.close();
        }
      });

      test("新建专题表单使用完整宽度", async ({ browser }) => {
        const context = await loginE2E(browser, {
          viewport: { width: 1440, height: 900 },
        });
        const page = await authenticatedPage(context, site.url);
        try {
          await page.goto(new URL("/manage/collections", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await page.getByRole("button", { name: "新建专题" }).click();
          const dialog = page.getByRole("dialog", { name: "新建专题" });
          const [dialogBox, nameBox, slugBox, descriptionBox] =
            await Promise.all([
              dialog.boundingBox(),
              dialog.getByRole("textbox", { name: "名称" }).boundingBox(),
              dialog.getByRole("textbox", { name: "Slug" }).boundingBox(),
              dialog.getByRole("textbox", { name: "说明" }).boundingBox(),
            ]);
          expect(dialogBox).toBeTruthy();
          for (const field of [nameBox, slugBox, descriptionBox]) {
            expect(field).toBeTruthy();
            expect(field!.width).toBeGreaterThanOrEqual(dialogBox!.width - 64);
          }
          await dialog.getByRole("button", { name: "取消" }).click();
        } finally {
          await context.close();
        }
      });

      test("图片编辑器提供图片预览与删除入口", async ({ browser }) => {
        const context = await loginE2E(browser, {
          viewport: { width: 1440, height: 900 },
        });
        const page = await authenticatedPage(context, site.url);
        let deleteBody: Record<string, unknown> | undefined;
        await page.route("**/api/gallery/admin/images/*", async (route) => {
          if (route.request().method() !== "DELETE") {
            await route.continue();
            return;
          }
          deleteBody = route.request().postDataJSON();
          await route.fulfill({
            status: 200,
            contentType: "application/json",
            body: JSON.stringify({ deleted: true }),
          });
        });
        try {
          await page.goto(new URL("/manage/images", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await openFirstImageEditor(page);
          const dialog = page.getByRole("dialog", { name: "编辑图片" });
          await expect(
            dialog.locator("[data-image-editor-preview]"),
          ).toBeVisible();
          await expect(
            dialog.getByRole("button", { name: "删除图片" }),
          ).toBeVisible();
          await dialog.getByRole("button", { name: "删除图片" }).click();
          await expect(
            dialog.getByRole("button", { name: "确认删除" }),
          ).toBeVisible();
          await dialog.getByRole("button", { name: "确认删除" }).click();
          await expect(dialog).toHaveCount(0);
          expect(deleteBody?.expectedUpdatedAt).toMatch(/^2026-/u);
        } finally {
          await context.close();
        }
      });

      test("图片编辑器使用可读分类标签且选择器有明确名称", async ({
        browser,
      }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        try {
          await page.goto(new URL("/manage/images", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await openFirstImageEditor(page);

          const dialog = page.getByRole("dialog", { name: "编辑图片" });
          const category = dialog.getByRole("button", {
            name: "主分类",
            exact: true,
          });
          const tags = dialog.getByRole("button", {
            name: "标签",
            exact: true,
          });
          await expect(category).toBeVisible();
          await expect(tags).toBeVisible();

          const rawID =
            /(?:^[A-Za-z0-9_-]{22}$)|(?:^[0-9a-f]{8}-[0-9a-f-]{27}$)/iu;
          expect((await category.textContent())?.trim()).not.toMatch(rawID);
          expect((await tags.textContent())?.trim()).not.toMatch(rawID);

          await category.click();
          const categoryLabels = (
            await page.getByRole("option").allTextContents()
          ).map((value) => value.trim());
          expect(categoryLabels).toEqual(
            expect.arrayContaining(["壁纸", "插画", "摄影"]),
          );
        } finally {
          await context.close();
        }
      });

      test("图片标题保存后可检索且测试结束恢复原值", async ({ browser }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        let originalTitle = "";
        let temporaryTitle = "";
        try {
          await page.goto(new URL("/manage/images", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await openFirstImageEditor(page);
          const title = page
            .getByRole("dialog", { name: "编辑图片" })
            .getByRole("textbox", { name: "标题", exact: false });
          originalTitle = await title.inputValue();
          temporaryTitle = `${originalTitle.slice(0, 140)} · 验收`;
          await title.fill(temporaryTitle);
          await saveImageEditor(page);

          await searchManagedImages(page, temporaryTitle);
          await openFirstImageEditor(page);
          await expect(
            page
              .getByRole("dialog", { name: "编辑图片" })
              .getByRole("textbox", { name: "标题", exact: false }),
          ).toHaveValue(temporaryTitle);
          await page
            .getByRole("dialog", { name: "编辑图片" })
            .getByRole("textbox", { name: "标题", exact: false })
            .fill(originalTitle);
          await saveImageEditor(page);

          await searchManagedImages(page, originalTitle);
          await expect(
            page.getByText(originalTitle, { exact: true }).first(),
          ).toBeVisible();
        } finally {
          if (originalTitle && temporaryTitle)
            await restoreImageTitle(
              page,
              site.url,
              temporaryTitle,
              originalTitle,
            ).catch(() => undefined);
          await context.close();
        }
      });

      test("图片专区支持网格列表与批量加入专题", async ({ browser }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        let memberMutation: Record<string, unknown> | undefined;
        await page.route(
          "**/api/gallery/admin/collections/*/members",
          async (route) => {
            memberMutation = route.request().postDataJSON();
            await route.fulfill({
              status: 200,
              contentType: "application/json",
              body: JSON.stringify({ collection: {} }),
            });
          },
        );
        try {
          await page.goto(new URL("/manage/images", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);

          const list = page.getByRole("button", {
            name: "列表视图",
            exact: true,
          });
          const grid = page.getByRole("button", {
            name: "网格视图",
            exact: true,
          });
          await expect(list).toHaveAttribute("aria-pressed", "true");
          await grid.click();
          await expect(page).toHaveURL(/(?:\?|&)view=grid(?:&|$)/u);
          await expect(
            page.locator("[data-gallery-image-grid-item]").first(),
          ).toBeVisible();

          await page
            .getByRole("checkbox", { name: /^选择图片：/u })
            .first()
            .check();
          await page.getByRole("combobox", { name: "批量操作" }).click();
          await page.getByRole("option", { name: "加入专题" }).click();
          await page.getByRole("button", { name: "应用", exact: true }).click();

          const dialog = page.getByRole("dialog", { name: "批量加入专题" });
          await expect(dialog).toBeVisible();
          await dialog
            .getByRole("button", { name: "目标专题", exact: true })
            .click();
          await page.getByRole("option").first().click();
          await dialog
            .getByRole("button", { name: "确认应用", exact: true })
            .click();
          await expect(dialog).toHaveCount(0);
          expect(memberMutation?.version).toEqual(expect.any(Number));
          expect(memberMutation?.add).toHaveLength(1);
          expect(memberMutation?.remove).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("专题编辑器支持网格列表切换与右上角设置", async ({ browser }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        try {
          await page.goto(new URL("/manage/collections", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          const firstCollection = page
            .getByRole("link", { name: /^编辑专题：/u })
            .first();
          await expect(firstCollection).toBeVisible();
          const href = await firstCollection.getAttribute("href");
          expect(href).toBeTruthy();

          await page.goto(new URL(href!, site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          await expect(
            page.locator("[data-gallery-collection-commandbar]"),
          ).toBeVisible();
          await expect(
            page.getByRole("heading", { name: "专题设置", exact: true }),
          ).toHaveCount(0);

          const grid = page.getByRole("button", {
            name: "网格视图",
            exact: true,
          });
          const list = page.getByRole("button", {
            name: "列表视图",
            exact: true,
          });
          await expect(grid).toHaveAttribute("aria-pressed", "true");
          for (const toggle of [grid, list]) {
            const icon = toggle.locator('[data-slot="leadingIcon"]');
            await expect(icon).toBeVisible();
            expect(
              await icon.evaluate((element) => {
                const style = getComputedStyle(element);
                return (
                  style.maskImage !== "none" ||
                  style.webkitMaskImage !== "none" ||
                  style.backgroundImage !== "none"
                );
              }),
            ).toBeTruthy();
          }
          await expect(
            page.locator("[data-gallery-collection-grid-item]").first(),
          ).toBeVisible();
          await list.click();
          await expect(page).toHaveURL(/(?:\?|&)view=list(?:&|$)/u);
          await expect(list).toHaveAttribute("aria-pressed", "true");
          await expect(
            page.locator("[data-gallery-collection-list-item]").first(),
          ).toBeVisible();

          await openCollectionSettings(page);
          await expect(
            page.locator("[data-gallery-collection-inspector]"),
          ).toBeVisible();
          await expect(
            page.locator(
              '[data-gallery-collection-inspector][data-inspector-mode="docked"]',
            ),
          ).toBeVisible();
          for (const name of ["内容", "展示", "SEO"]) {
            const tab = page.getByRole("tab", { name, exact: true });
            await expect(
              tab.locator('[data-slot="leadingIcon"]'),
            ).toBeVisible();
          }
          const [inspectorBox, nameBox, slugBox, descriptionBox] =
            await Promise.all([
              page.locator("[data-gallery-collection-inspector]").boundingBox(),
              page
                .getByRole("textbox", { name: "名称", exact: false })
                .first()
                .boundingBox(),
              page
                .getByRole("textbox", { name: "Slug", exact: false })
                .first()
                .boundingBox(),
              page
                .getByRole("textbox", { name: "说明", exact: false })
                .first()
                .boundingBox(),
            ]);
          expect(inspectorBox).not.toBeNull();
          for (const fieldBox of [nameBox, slugBox, descriptionBox]) {
            expect(fieldBox).not.toBeNull();
            expect(fieldBox!.width).toBeGreaterThan(inspectorBox!.width * 0.95);
          }
          const overflow = await page.evaluate(
            () =>
              document.documentElement.scrollWidth -
              document.documentElement.clientWidth,
          );
          expect(overflow).toBeLessThanOrEqual(1);
          const accessibility = await new AxeBuilder({ page })
            .exclude("nuxt-devtools-frame")
            .analyze();
          expect(
            accessibility.violations.filter((violation) =>
              ["serious", "critical"].includes(violation.impact || ""),
            ),
          ).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("专题编辑器支持拖拽、更多排序与批量移出", async ({ browser }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        try {
          await page.goto(new URL("/manage/collections", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          const firstCollection = page
            .getByRole("link", { name: /^编辑专题：/u })
            .first();
          await expect(firstCollection).toBeVisible();
          const href = await firstCollection.getAttribute("href");
          expect(href).toBeTruthy();
          await page.goto(new URL(href!, site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);

          const handles = page.getByRole("button", { name: /^拖动排序：/u });
          await expect(handles.first()).toBeVisible();
          await expect(handles.nth(2)).toBeVisible();
          await handles
            .first()
            .dragTo(page.locator("[data-gallery-collection-grid-item]").nth(2));
          await expect(
            page.getByRole("button", { name: "保存顺序", exact: true }),
          ).toBeEnabled();

          await expect(
            page.getByRole("button", { name: /^上移：/u }),
          ).toHaveCount(0);
          await page
            .getByRole("button", { name: /^图片操作：/u })
            .first()
            .click();
          for (const name of ["移到顶部", "上移", "下移", "移到底部"]) {
            await expect(
              page.getByRole("menuitem", { name, exact: true }),
            ).toBeVisible();
          }
          await page.keyboard.press("Escape");

          await page
            .getByRole("checkbox", { name: /^选择图片：/u })
            .first()
            .check();
          await page
            .getByRole("button", { name: "批量移出专题", exact: true })
            .click();
          const removeDialog = page.getByRole("dialog", {
            name: "批量移出专题",
          });
          await expect(removeDialog).toBeVisible();
          await removeDialog.getByRole("button", { name: "取消" }).click();

          await openCollectionSettings(page);
          await page.getByRole("tab", { name: "展示", exact: true }).click();
          await expect(
            page.getByRole("combobox", { name: "图片排序" }),
          ).toBeVisible();
          await expect(
            page.getByRole("button", { name: "应用排序", exact: true }),
          ).toBeVisible();
        } finally {
          await context.close();
        }
      });

      test("专题添加图片默认展示可用图片", async ({ browser }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        try {
          await page.goto(new URL("/manage/collections", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          const firstCollection = page
            .getByRole("link", { name: /^编辑专题：/u })
            .first();
          await expect(firstCollection).toBeVisible();
          const href = await firstCollection.getAttribute("href");
          expect(href).toBeTruthy();
          await page.goto(new URL(href!, site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await page.getByRole("button", { name: "添加图片" }).click();
          const dialog = page.getByRole("dialog", { name: "添加图片" });
          await expect(dialog).toBeVisible();
          await expect(
            dialog.getByRole("button", { name: "添加", exact: true }).first(),
          ).toBeVisible();
        } finally {
          await context.close();
        }
      });

      test("专题名称保存后刷新持久化且测试结束恢复原值", async ({
        browser,
      }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        let originalName = "";
        let collectionURL = "";
        try {
          await page.goto(new URL("/manage/collections", site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          const firstCollection = page
            .getByRole("link", { name: /^编辑专题：/u })
            .first();
          await expect(
            firstCollection,
            "验收夹具必须至少包含一个可编辑专题",
          ).toBeVisible();
          const href = await firstCollection.getAttribute("href");
          expect(href).toBeTruthy();
          collectionURL = new URL(href!, site.url).toString();

          await page.goto(collectionURL, { waitUntil: "domcontentloaded" });
          await settleNuxt(page);
          await openCollectionSettings(page);
          const name = page
            .getByRole("textbox", { name: "名称", exact: false })
            .first();
          originalName = await name.inputValue();
          const temporaryName = `${originalName.slice(0, 90)} · 验收`;
          await name.fill(temporaryName);
          await saveCollectionSettings(page);

          await page.reload({ waitUntil: "domcontentloaded" });
          await settleNuxt(page);
          await openCollectionSettings(page);
          await expect(
            page.getByRole("textbox", { name: "名称", exact: false }).first(),
          ).toHaveValue(temporaryName);

          await page
            .getByRole("textbox", { name: "名称", exact: false })
            .first()
            .fill(originalName);
          await saveCollectionSettings(page);
          await page.reload({ waitUntil: "domcontentloaded" });
          await settleNuxt(page);
          await openCollectionSettings(page);
          await expect(
            page.getByRole("textbox", { name: "名称", exact: false }).first(),
          ).toHaveValue(originalName);
        } finally {
          if (collectionURL && originalName)
            await restoreCollectionName(
              page,
              collectionURL,
              originalName,
            ).catch(() => undefined);
          await context.close();
        }
      });

      test("资源策略展示消费者注册且不提供第二套配置入口", async ({
        browser,
      }) => {
        const context = await loginE2E(browser, {}, undefined, site.url);
        const page = await context.newPage();
        try {
          await page.goto(new URL("/manage/assets", site.url).toString(), {
            waitUntil: "networkidle",
          });
          await expect(
            page.getByRole("heading", { name: "资源策略", exact: true }),
          ).toBeVisible();
          await expect(
            page.locator("[data-asset-registration-summary]"),
          ).toBeVisible();
          await expect(
            page.getByText("gallery-submission", { exact: true }),
          ).toBeVisible();
          await expect(page.getByRole("button", { name: "保存" })).toHaveCount(
            0,
          );
          await expect(page.getByLabel("站点名称")).toHaveCount(0);
        } finally {
          await context.close();
        }
      });

      test("分类状态变更必须先生成影响预览且不会直接改写数据", async ({
        browser,
      }) => {
        const context = await loginE2E(browser);
        const page = await authenticatedPage(context, site.url);
        try {
          await page.goto(
            new URL("/manage/classification", site.url).toString(),
            { waitUntil: "domcontentloaded" },
          );
          await settleNuxt(page);
          const firstActions = page
            .getByRole("button", { name: /^更多分类操作：/u })
            .first();
          await expect(
            firstActions,
            "验收夹具必须至少包含一个可治理分类",
          ).toBeVisible();
          await firstActions.click();

          const previewResponse = page.waitForResponse(
            (candidate) =>
              candidate.request().method() === "POST" &&
              /\/api\/gallery\/admin\/classification\/governance\/preview$/.test(
                candidate.url(),
              ),
          );
          await page.getByRole("menuitem", { name: "停用分类" }).click();
          expect((await previewResponse).ok()).toBeTruthy();

          const dialog = page.getByRole("dialog", { name: "治理影响预览" });
          await expect(dialog).toBeVisible();
          await expect(
            dialog.getByText("planned", { exact: true }),
          ).toBeVisible();
          await expect(
            dialog.getByRole("button", { name: "执行计划", exact: true }),
          ).toBeVisible();
          await dialog
            .getByRole("button", { name: "关闭", exact: true })
            .click();
          await expect(dialog).toHaveCount(0);
        } finally {
          await context.close();
        }
      });

      test("分类与维度使用统一表面并提供新增入口", async ({ browser }) => {
        const context = await loginE2E(browser, {
          viewport: { width: 1440, height: 900 },
        });
        const page = await authenticatedPage(context, site.url);
        let createBody: Record<string, unknown> | undefined;
        let updateBody: Record<string, unknown> | undefined;
        let updatePath = "";
        await page.route(
          "**/api/gallery/admin/classification/identities",
          async (route) => {
            createBody = route.request().postDataJSON();
            await route.fulfill({
              status: 200,
              contentType: "application/json",
              body: JSON.stringify({
                identity: {
                  id: "AZsQAAAAcACQAAAAAAAAAQ",
                  catalogRevision: 2,
                },
              }),
            });
          },
        );
        await page.route(
          "**/api/gallery/admin/classification/identities/*",
          async (route) => {
            updatePath = new URL(route.request().url()).pathname;
            updateBody = route.request().postDataJSON();
            await route.fulfill({
              status: 200,
              contentType: "application/json",
              body: JSON.stringify({
                identity: {
                  id: "AZsQAAAAcACQAAAAAAAAAQ",
                  catalogRevision: 3,
                },
              }),
            });
          },
        );
        try {
          await page.goto(
            new URL("/manage/classification", site.url).toString(),
            { waitUntil: "domcontentloaded" },
          );
          await settleNuxt(page);

          const surface = page.locator(
            '[data-manage-surface="classification"]',
          );
          await expect(surface).toBeVisible();
          const searchFrameStyle = await surface
            .locator("[data-classification-search]")
            .evaluate((element) => {
              const style = getComputedStyle(element);
              return {
                borderRadius: style.borderRadius,
                borderTopWidth: style.borderTopWidth,
                boxShadow: style.boxShadow,
              };
            });
          expect(searchFrameStyle).toEqual({
            borderRadius: "0px",
            borderTopWidth: "0px",
            boxShadow: "none",
          });
          const firstClassificationRow = surface
            .locator("[data-classification-row]")
            .first();
          const compactRow = await firstClassificationRow.evaluate(
            (element) => {
              const style = getComputedStyle(element);
              return {
                height: element.getBoundingClientRect().height,
                columns: style.gridTemplateColumns.split(" ").length,
              };
            },
          );
          expect(compactRow.columns).toBe(3);
          expect(compactRow.height).toBeLessThanOrEqual(52);
          await expect(page.getByText("标识信息", { exact: true })).toHaveCount(
            0,
          );
          for (const label of ["分类", "维度", "标签", "标签提案"]) {
            const tab = surface.getByRole("tab", {
              name: new RegExp(`^${label}(?: \\d+)?$`),
            });
            await expect(tab).toBeVisible();
            await expect(
              tab.locator('[data-slot="leadingIcon"]'),
            ).toBeVisible();
          }

          await page.getByRole("button", { name: "新增分类" }).click();
          const categoryDialog = page.getByRole("dialog", { name: "新增分类" });
          await categoryDialog
            .getByRole("textbox", { name: "名称" })
            .fill("验收分类");
          await categoryDialog
            .getByRole("textbox", { name: "标识" })
            .fill("acceptance-category");
          await categoryDialog
            .getByRole("button", { name: "新增分类" })
            .click();
          await expect(categoryDialog).toHaveCount(0);
          expect(createBody).toMatchObject({
            kind: "category",
            name: "验收分类",
            slug: "acceptance-category",
            parentId: "",
          });

          const categoryActions = page
            .getByRole("button", { name: /^更多分类操作：/u })
            .first();
          const categoryActionBox = await categoryActions.boundingBox();
          expect(categoryActionBox).toBeTruthy();
          expect(categoryActionBox!.width).toBe(categoryActionBox!.height);
          const overflowIcon = categoryActions.locator(".iconify");
          await expect(overflowIcon).toBeVisible();
          expect(
            await overflowIcon.evaluate((element) => {
              const style = getComputedStyle(element);
              return (
                style.maskImage !== "none" ||
                style.webkitMaskImage !== "none" ||
                style.backgroundImage !== "none"
              );
            }),
          ).toBeTruthy();
          await categoryActions.click();
          await page.getByRole("menuitem", { name: "编辑分类" }).click();
          const editCategoryDialog = page.getByRole("dialog", {
            name: "编辑分类",
          });
          await expect(
            editCategoryDialog.getByRole("textbox", { name: "名称" }),
          ).toHaveValue("壁纸");
          await editCategoryDialog
            .getByRole("textbox", { name: "名称" })
            .fill("壁纸编辑验收");
          await editCategoryDialog
            .getByRole("button", { name: "编辑分类" })
            .click();
          await expect(editCategoryDialog).toHaveCount(0);
          expect(updatePath).toMatch(
            /\/api\/gallery\/admin\/classification\/identities\/[^/]+$/,
          );
          expect(updateBody).toMatchObject({
            kind: "category",
            name: "壁纸编辑验收",
            slug: "wallpaper",
          });

          await surface.getByRole("tab", { name: /^维度/u }).click();
          await expect(page).toHaveURL(/section=facets/);
          await expect(
            page.getByRole("button", { name: "新增维度" }),
          ).toBeVisible();
          await expect(
            surface.getByRole("button", { name: "新增" }).first(),
          ).toBeVisible();
          await page.getByRole("button", { name: "新增维度" }).click();
          const facetDialog = page.getByRole("dialog", { name: "新增维度" });
          await expect(
            facetDialog.getByRole("textbox", { name: "名称" }),
          ).toBeVisible();
          await facetDialog.getByRole("button", { name: "取消" }).click();
          await expect(facetDialog).toHaveCount(0);

          await page
            .getByRole("button", { name: /^更多维度操作：/u })
            .first()
            .click();
          await page.getByRole("menuitem", { name: "编辑维度" }).click();
          const editFacetDialog = page.getByRole("dialog", {
            name: "编辑维度",
          });
          await expect(
            editFacetDialog.getByRole("textbox", { name: "名称" }),
          ).toHaveValue("场景");
          await editFacetDialog.getByRole("button", { name: "取消" }).click();
          await expect(editFacetDialog).toHaveCount(0);

          await page
            .getByRole("button", { name: /^更多维度值操作：/u })
            .first()
            .click();
          await page.getByRole("menuitem", { name: "编辑维度值" }).click();
          const editValueDialog = page.getByRole("dialog", {
            name: "编辑维度值",
          });
          await expect(
            editValueDialog.getByRole("textbox", { name: "名称" }),
          ).toHaveValue("人物");
          await editValueDialog.getByRole("button", { name: "取消" }).click();
          await expect(editValueDialog).toHaveCount(0);

          await surface
            .getByRole("tab")
            .filter({ hasText: /^标签\d*/u })
            .first()
            .click();
          await expect(
            page.getByRole("button", { name: "新增标签" }),
          ).toBeVisible();
          const firstTagRow = surface
            .locator("[data-classification-row]")
            .first();
          await expect(firstTagRow).not.toContainText("个关系");
          await expect(firstTagRow).not.toContainText("个别名");
          const tagActions = page
            .getByRole("button", { name: /^更多标签操作：/u })
            .first();
          const tagActionBox = await tagActions.boundingBox();
          expect(tagActionBox).toBeTruthy();
          expect(tagActionBox!.width).toBe(tagActionBox!.height);
          await tagActions.click();
          await page.getByRole("menuitem", { name: "编辑标签" }).click();
          const editTagDialog = page.getByRole("dialog", {
            name: "编辑标签",
          });
          await expect(
            editTagDialog.getByRole("textbox", { name: "名称" }),
          ).not.toHaveValue("");
          await editTagDialog.getByRole("button", { name: "取消" }).click();
          await expect(editTagDialog).toHaveCount(0);

          await surface
            .getByRole("tab")
            .filter({ hasText: "标签提案" })
            .click();
          await expect(
            page.getByText("来源与标识", { exact: true }),
          ).toHaveCount(0);
          await expect(surface.locator("select")).toHaveCount(0);
          const proposalRow = surface
            .locator("[data-tag-proposal-row]")
            .first();
          const proposalGrid = await proposalRow.evaluate(
            (element) =>
              getComputedStyle(element).gridTemplateColumns.split(" ").length,
          );
          expect(proposalGrid).toBe(3);
          const proposalButtons = proposalRow.getByRole("button");
          const proposalButtonBoxes = await proposalButtons.evaluateAll(
            (elements) =>
              elements.map((element) => element.getBoundingClientRect().height),
          );
          expect(Math.max(...proposalButtonBoxes)).toBeLessThanOrEqual(30);
          const proposalSelect = surface
            .getByLabel(/选择 .* 的标签处理方式/u)
            .first();
          await proposalSelect.click();
          await expect(
            page.getByRole("option", { name: "创建新标签" }),
          ).toBeVisible();
          await page.keyboard.press("Escape");

          const accessibility = await new AxeBuilder({ page })
            .exclude("nuxt-devtools-frame")
            .analyze();
          expect(
            accessibility.violations.filter((violation) =>
              ["serious", "critical"].includes(violation.impact || ""),
            ),
          ).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("运行中的分类写入接口已加载", async ({ browser }) => {
        const context = await loginE2E(browser, {}, undefined, site.url);
        try {
          const response = await context.request.post(
            new URL(
              "/api/gallery/admin/classification/identities",
              site.url,
            ).toString(),
            { data: {} },
          );
          expect(response.status()).not.toBe(404);
          expect([400, 422]).toContain(response.status());
        } finally {
          await context.close();
        }
      });
    });
  }
}
