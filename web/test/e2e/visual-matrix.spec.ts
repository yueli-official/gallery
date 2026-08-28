import { expect, test, type Page } from "@playwright/test";
import { productSites } from "./contracts";
import {
  expectNoHorizontalOverflow,
  loginE2E,
  settleNuxt,
} from "./runtime";

const viewports = [
  { name: "mobile", width: 390, height: 844 },
  { name: "tablet", width: 768, height: 1024 },
  { name: "medium", width: 1100, height: 900 },
  { name: "desktop", width: 1440, height: 900 },
] as const;

const themes = ["light", "dark"] as const;

async function setTheme(page: Page, theme: (typeof themes)[number]) {
  await page.emulateMedia({ colorScheme: theme });
  await page.addInitScript(
    (value) => window.localStorage.setItem("nuxt-color-mode", value),
    theme,
  );
}

async function settle(page: Page) {
  await settleNuxt(page);
  await page.waitForFunction(
    () =>
      [
        ...document.querySelectorAll(
          '[data-submission-preview-state="loading"]',
        ),
      ].every((element) => {
        const bounds = element.getBoundingClientRect();
        return bounds.bottom <= 0 || bounds.top >= window.innerHeight;
      }),
    undefined,
    { timeout: 30_000 },
  );
  await page.locator("img").evaluateAll((images) => {
    for (const image of images as HTMLImageElement[]) {
      const bounds = image.getBoundingClientRect();
      if (bounds.bottom > 0 && bounds.top < window.innerHeight)
        image.loading = "eager";
    }
  });
  await page.waitForFunction(
    () =>
      [...document.images].every((image) => {
        const bounds = image.getBoundingClientRect();
        const visible = bounds.bottom > 0 && bounds.top < window.innerHeight;
        return !visible || (image.complete && image.naturalWidth > 0);
      }),
    undefined,
    { timeout: 30_000 },
  );
  await page.addStyleTag({
    content: `
      *, *::before, *::after {
        animation-delay: 0s !important;
        animation-duration: 0s !important;
        caret-color: transparent !important;
        transition-delay: 0s !important;
        transition-duration: 0s !important;
      }
      nuxt-devtools-frame, nuxt-devtools-client, vue-devtools,
      #nuxt-devtools-container, [data-nuxt-devtools] { display: none !important; }
    `,
  });
}

export function registerVisualSuite(product: string) {
  for (const site of productSites(product)) {
    const contract = site.contract;
    test.describe(`${site.slug} (${site.product}) 视觉合同`, () => {
      for (const theme of themes) {
        for (const viewport of viewports) {
          test(`公开入口 ${theme} ${viewport.name}`, async ({ page }) => {
            await page.setViewportSize(viewport);
            await setTheme(page, theme);
            const response = await page.goto(
              new URL(contract.public.path, site.url).toString(),
              { waitUntil: "domcontentloaded" },
            );
            expect(response?.ok()).toBeTruthy();
            await settle(page);
            await expect(
              page.locator(contract.public.readySelector).first(),
            ).toBeVisible();
            const rootClass =
              (await page.locator("html").getAttribute("class")) || "";
            expect(rootClass.split(/\s+/).includes("dark")).toBe(
              theme === "dark",
            );
            await expectNoHorizontalOverflow(page);
            await expect(page).toHaveScreenshot([
              "screenshots",
              site.slug,
              "public",
              theme,
              `${viewport.name}.png`,
            ]);
          });
        }
      }

      const views = contract.manage.views?.length
        ? contract.manage.views
        : [{ key: "default", label: "" }];
      const managePath = contract.visual?.managePath || contract.manage.path;
      for (const theme of themes) {
        for (const viewport of viewports) {
          for (const view of views) {
            test(`管理入口 ${view.key} ${theme} ${viewport.name}`, async ({
              browser,
            }) => {
              const context = await loginE2E(
                browser,
                { viewport, colorScheme: theme },
                theme,
              );
              const page = await context.newPage();
              try {
                const url = new URL(managePath, site.url);
                if (view.key !== "default")
                  url.searchParams.set("view", view.key);
                const response = await page.goto(url.toString(), {
                  waitUntil: "domcontentloaded",
                });
                expect(response?.ok()).toBeTruthy();
                await settle(page);
                await expect(
                  page.locator(contract.manage.readySelector).first(),
                ).toBeVisible();
                if (view.label) {
                  await expect(
                    page
                      .getByRole("button", { name: view.label, exact: true })
                      .first(),
                  ).toHaveAttribute("aria-pressed", "true");
                }
                await expect(
                  page.locator("[data-manage-skeleton]"),
                ).toHaveCount(0);
                const rootClass =
                  (await page.locator("html").getAttribute("class")) || "";
                expect(rootClass.split(/\s+/).includes("dark")).toBe(
                  theme === "dark",
                );
                await expectNoHorizontalOverflow(page);
                await expect(page).toHaveScreenshot([
                  "screenshots",
                  site.slug,
                  `manage-${view.key}`,
                  theme,
                  `${viewport.name}.png`,
                ]);
              } finally {
                await context.close();
              }
            });
          }
        }
      }

      for (const scenario of contract.visual?.scenarios || []) {
        const scenarioViewports = scenario.viewports?.length
          ? viewports.filter((viewport) =>
              scenario.viewports?.includes(viewport.name),
            )
          : viewports;
        for (const theme of themes) {
          for (const viewport of scenarioViewports) {
            test(`领域场景 ${scenario.key} ${theme} ${viewport.name}`, async ({
              browser,
              page,
            }) => {
              const context = scenario.authenticated
                ? await loginE2E(
                    browser,
                    { viewport, colorScheme: theme },
                    theme,
                    site.url,
                  )
                : undefined;
              const targetPage = context ? await context.newPage() : page;
              try {
                if (!context) {
                  await targetPage.setViewportSize(viewport);
                  await setTheme(targetPage, theme);
                }
                const response = await targetPage.goto(
                  new URL(scenario.path, site.url).toString(),
                  { waitUntil: "domcontentloaded" },
                );
                expect(response?.status()).toBe(scenario.status || 200);
                if (scenario.actions?.length) {
                  const controls = scenario.actions.map((action) =>
                    targetPage
                      .getByRole(action.role, {
                        name: action.name,
                        exact: true,
                      })
                      .first(),
                  );
                  await expect(controls[0]).toBeVisible();
                  await settle(targetPage);
                  for (const [index, action] of scenario.actions.entries()) {
                    const control = controls[index];
                    await control.click();
                    if (action.role === "checkbox")
                      await expect(control).toBeChecked();
                  }
                  await settle(targetPage);
                } else {
                  await expect(
                    targetPage.locator(scenario.readySelector).first(),
                  ).toBeVisible();
                  await settle(targetPage);
                }
                await expect(
                  targetPage.locator(scenario.readySelector).first(),
                ).toBeVisible();
                await expectNoHorizontalOverflow(targetPage);
                await expect(targetPage).toHaveScreenshot([
                  "screenshots",
                  site.slug,
                  scenario.key,
                  theme,
                  `${viewport.name}.png`,
                ]);
              } finally {
                await context?.close();
              }
            });
          }
        }
      }

      for (const theme of themes) {
        for (const viewport of viewports.filter((item) =>
          ["mobile", "desktop"].includes(item.name),
        )) {
          test(`评论场景 public ${theme} ${viewport.name}`, async ({ page }) => {
            await page.setViewportSize(viewport);
            await setTheme(page, theme);
            await page.route("**/api/gallery/images/*/comments**", (route) =>
              route.fulfill({
                status: 200,
                contentType: "application/json",
                body: JSON.stringify({
                  total: 2,
                  page: 1,
                  size: 20,
                  items: [
                    {
                      id: "comment-1",
                      authorName: "测试用户",
                      isAnonymous: false,
                      content: "这张海岸晨光的颜色很舒服。",
                      createdAt: "2026-07-11T08:00:00Z",
                      replies: [
                        {
                          id: "comment-2",
                          parentId: "comment-1",
                          authorName: "路过的访客",
                          isAnonymous: true,
                          content: "我也很喜欢这组蓝色。",
                          createdAt: "2026-07-11T08:05:00Z",
                        },
                      ],
                    },
                  ],
                }),
              }),
            );
            await page.goto(
              new URL("/images/AZsQAAAAcACQAAAAAAAAAQ", site.url).toString(),
              { waitUntil: "domcontentloaded" },
            );
            await settle(page);
            const comments = page.locator("[data-gallery-comments]");
            await expect(comments.getByText("这张海岸晨光的颜色很舒服。", { exact: true })).toBeVisible();
            await comments.scrollIntoViewIfNeeded();
            await expectNoHorizontalOverflow(page);
            await expect(page).toHaveScreenshot([
              "screenshots",
              site.slug,
              "comments-public",
              theme,
              `${viewport.name}.png`,
            ]);
          });

          test(`评论场景 admin ${theme} ${viewport.name}`, async ({ browser }) => {
            const context = await loginE2E(
              browser,
              { viewport, colorScheme: theme },
              theme,
              site.url,
            );
            const page = await context.newPage();
            try {
              await page.route("**/api/gallery/admin/comments**", (route) =>
                route.fulfill({
                  status: 200,
                  contentType: "application/json",
                  body: JSON.stringify({
                    total: 2,
                    page: 1,
                    size: 20,
                    items: [
                      {
                        id: "comment-1",
                        imageId: "AZsQAAAAcACQAAAAAAAAAQ",
                        imageTitle: "海岸晨光 001",
                        authorName: "等待审核的访客",
                        authorEmail: "visitor@example.test",
                        content: "这是一条等待运营审核的评论。",
                        status: "pending",
                        createdAt: "2026-07-11T08:15:00Z",
                      },
                      {
                        id: "comment-2",
                        imageId: "AZsQAAAAcACQAAAAAAAAAg",
                        imageTitle: "山间薄雾 002",
                        authorName: "测试用户",
                        userKey: "TestA123",
                        content: "登录用户已经发布的评论。",
                        status: "approved",
                        createdAt: "2026-07-11T08:20:00Z",
                      },
                    ],
                  }),
                }),
              );
              await page.goto(new URL("/manage/comments", site.url).toString(), {
                waitUntil: "domcontentloaded",
              });
              await expect(page.getByText("这是一条等待运营审核的评论。", { exact: true })).toBeVisible();
              await settle(page);
              await expectNoHorizontalOverflow(page);
              await expect(page).toHaveScreenshot([
                "screenshots",
                site.slug,
                "comments-admin",
                theme,
                `${viewport.name}.png`,
              ]);
            } finally {
              await context.close();
            }
          });
        }
      }
    });
  }
}
