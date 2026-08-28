import { expect, test } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";
import { productSites } from "./contracts";
import {
  capturePageFailures,
  loginE2E,
  requiredEnv,
  settleNuxt,
} from "./runtime";

const accountURL = requiredEnv("GALLERY_E2E_ACCOUNT_URL");
const missing = "__gallery_e2e_missing__";

export function registerJourneySuite(product: string) {
  for (const site of productSites(product)) {
    const contract = site.contract;
    test.describe(`${site.slug} (${site.product})`, () => {
      test("公开入口完成渲染且没有浏览器错误", async ({ page }) => {
        const errors = capturePageFailures(page);
        const response = await page.goto(
          new URL(contract.public.path, site.url).toString(),
          { waitUntil: "domcontentloaded" },
        );
        expect(response?.ok()).toBeTruthy();
        await expect(
          page.locator(contract.public.readySelector).first(),
        ).toBeVisible();
        await settleNuxt(page);
        expect(errors).toEqual([]);
      });

      test("公开首页板块只保留标题与动作", async ({ page }) => {
        await page.goto(new URL("/?seed=section-copy", site.url).toString(), {
          waitUntil: "domcontentloaded",
        });
        await settleNuxt(page);
        const collections = page.getByRole("heading", {
          name: "从专题进入",
          exact: true,
        });
        await collections.scrollIntoViewIfNeeded();
        await expect(collections).toBeVisible();
        for (const copy of [
          "沿着一个清晰主题，查看经过整理的图片集合。",
          "最近完成处理和审核的公开图片。",
          "近期获得更多有效浏览的图片。",
        ]) {
          await expect(page.getByText(copy, { exact: true })).toHaveCount(0);
        }
      });

      test("公开图片目录支持分类、维度与标签组合筛选", async ({ page }) => {
        await page.setViewportSize({ width: 1440, height: 1000 });
        await page.goto(new URL("/images", site.url).toString(), {
          waitUntil: "domcontentloaded",
        });
        await settleNuxt(page);

        const filters = page.getByRole("complementary", {
          name: "图片过滤器",
        });
        for (const name of ["分类筛选", "维度筛选", "标签筛选"]) {
          await expect(
            filters.getByRole("tab", { name, exact: true }),
          ).toBeVisible();
        }
        await expect(
          filters.getByRole("heading", { name: "分类", exact: true }),
        ).toBeVisible();
        await filters.getByRole("checkbox", { name: "壁纸" }).check();
        await filters.getByRole("tab", { name: "维度筛选" }).click();
        for (const name of ["场景", "方向", "风格"]) {
          await expect(
            filters.getByRole("heading", { name, exact: true }),
          ).toBeVisible();
        }
        await filters.getByRole("checkbox", { name: "场景：人物" }).check();
        await filters.getByRole("tab", { name: "标签筛选" }).click();
        await expect(
          filters.getByRole("heading", { name: "热门标签", exact: true }),
        ).toBeVisible();
        await filters.getByRole("button", { name: /^标签：晨光，/u }).click();
        await filters
          .getByRole("button", { name: "应用", exact: true })
          .click();

        await expect(page).toHaveURL(/categories=wallpaper/u);
        await expect(page).toHaveURL(/facets=scene:people/u);
        await expect(page).toHaveURL(/tag=demo-tag-1/u);
        await expect(
          page.getByRole("button", { name: "移除筛选：壁纸" }),
        ).toBeVisible();
        await expect(
          page.getByRole("button", { name: "移除筛选：场景：人物" }),
        ).toBeVisible();
        await expect(
          page.getByRole("button", { name: "移除筛选：标签：晨光" }),
        ).toBeVisible();
        const accessibility = await new AxeBuilder({ page })
          .exclude("nuxt-devtools-frame")
          .analyze();
        expect(
          accessibility.violations.filter((violation) =>
            ["serious", "critical"].includes(violation.impact || ""),
          ),
        ).toEqual([]);
      });

      test("图片详情提供两级评论与匿名审核反馈", async ({ page }) => {
        let postedBody: Record<string, unknown> | undefined;
        await page.route(
          "**/api/gallery/images/*/comments**",
          async (route) => {
            if (route.request().method() === "POST") {
              postedBody = route.request().postDataJSON();
              await route.fulfill({
                status: 200,
                contentType: "application/json",
                body: JSON.stringify({
                  pending: true,
                  comment: {
                    id: "comment-pending",
                    authorName: "访客",
                    isAnonymous: true,
                    content: postedBody?.content,
                    createdAt: "2026-07-11T08:30:00Z",
                  },
                }),
              });
              return;
            }
            await route.fulfill({
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
                    content: "这张图片的颜色很舒服。",
                    createdAt: "2026-07-11T08:00:00Z",
                    replies: [
                      {
                        id: "comment-2",
                        parentId: "comment-1",
                        authorName: "路过的访客",
                        isAnonymous: true,
                        content: "我也很喜欢。",
                        createdAt: "2026-07-11T08:05:00Z",
                      },
                    ],
                  },
                ],
              }),
            });
          },
        );

        await page.goto(
          new URL("/images/AZsQAAAAcACQAAAAAAAAAQ", site.url).toString(),
          { waitUntil: "domcontentloaded" },
        );
        await settleNuxt(page);
        const comments = page.locator("[data-gallery-comments]");
        await comments.scrollIntoViewIfNeeded();
        await expect(
          comments.getByRole("heading", { name: /评论/ }),
        ).toBeVisible();
        await expect(
          comments.getByText("这张图片的颜色很舒服。", { exact: true }),
        ).toBeVisible();
        await expect(
          comments.getByText("我也很喜欢。", { exact: true }),
        ).toBeVisible();
        await comments.getByRole("button", { name: "回复" }).first().click();
        await expect(
          comments.getByRole("textbox", { name: "写下回复…" }),
        ).toBeVisible();

        await comments.getByPlaceholder("昵称 *").first().fill("访客");
        await comments
          .getByRole("textbox", { name: "写下你的评论…" })
          .fill("一条待审核评论");
        await comments.getByRole("button", { name: "发表评论" }).click();
        await expect(
          comments.getByText("评论已提交，待审核后显示", { exact: true }),
        ).toBeVisible();
        expect(postedBody).toMatchObject({
          content: "一条待审核评论",
          authorName: "访客",
        });
        const accessibility = await new AxeBuilder({ page })
          .exclude("nuxt-devtools-frame")
          .analyze();
        expect(
          accessibility.violations.filter((violation) =>
            ["serious", "critical"].includes(violation.impact || ""),
          ),
        ).toEqual([]);
      });

      test("匿名访问管理入口进入账户登录流程", async ({ page }) => {
        const errors = capturePageFailures(page);
        await page.goto(new URL(contract.manage.path, site.url).toString(), {
          waitUntil: "domcontentloaded",
        });
        await expect(page).toHaveURL(
          new RegExp(
            `^${accountURL.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}/login(?:\\?|$)`,
          ),
        );
        await expect(
          page.getByRole("heading", { name: "欢迎回来" }),
        ).toBeVisible();
        await page.waitForLoadState("load");
        expect(errors).toEqual([]);
      });

      test("已登录运营者可以进入管理界面", async ({ browser }) => {
        const context = await loginE2E(browser);
        const page = await context.newPage();
        const errors = capturePageFailures(page);
        try {
          const manageURL = new URL(contract.manage.path, site.url).toString();
          await page.goto(manageURL, { waitUntil: "domcontentloaded" });
          await expect(page).toHaveURL(manageURL);
          await expect(
            page.locator(contract.manage.readySelector).first(),
          ).toBeVisible();
          await expect(
            page
              .getByRole("heading", {
                name: new RegExp(contract.manage.heading),
              })
              .first(),
          ).toBeVisible();
          await settleNuxt(page);
          expect(errors).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("设置页脏数据保护阻止意外离开", async ({ browser }) => {
        const settings = contract.manage.settings;
        test.skip(!settings, "该产品没有可编辑的通用设置页");
        if (!settings) return;
        const context = await loginE2E(browser);
        const page = await context.newPage();
        const errors = capturePageFailures(page);
        try {
          await page.goto(new URL(settings.path, site.url).toString(), {
            waitUntil: "domcontentloaded",
          });
          await settleNuxt(page);
          const field = page
            .getByRole("textbox", { name: settings.fieldLabel, exact: false })
            .first();
          await expect(field).toBeVisible();
          await expect(
            page
              .locator("main")
              .getByRole("button", { name: "保存设置", exact: true })
              .first(),
          ).toBeVisible();
          await field.fill(`${await field.inputValue()} · 未保存`);

          const dialogHandled = new Promise<void>((resolve) => {
            page.once("dialog", async (dialog) => {
              expect(dialog.type()).toBe("confirm");
              expect(dialog.message()).toContain("未保存");
              await dialog.dismiss();
              resolve();
            });
          });
          await page
            .locator(`a[href="${contract.manage.path}"]`)
            .first()
            .click();
          await dialogHandled;
          await expect(page).toHaveURL(
            new RegExp(
              `${settings.path.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}(?:\\?|$)`,
            ),
          );
          expect(errors).toEqual([]);
        } finally {
          await context.close();
        }
      });

      test("空结果状态有明确反馈", async ({ browser, page }) => {
        const context = contract.empty.authenticated
          ? await loginE2E(browser)
          : undefined;
        const targetPage = context ? await context.newPage() : page;
        const errors = capturePageFailures(targetPage);
        try {
          await targetPage.goto(
            new URL(contract.empty.path, site.url).toString(),
            { waitUntil: "domcontentloaded" },
          );
          const emptyState = targetPage
            .getByText(contract.empty.text, { exact: true })
            .first();
          if (contract.empty.inputPlaceholder) {
            const input = targetPage.getByPlaceholder(
              contract.empty.inputPlaceholder,
            );
            await expect(async () => {
              await input.fill(missing);
              await expect(input).toHaveValue(missing);
              await expect(emptyState).toBeVisible({ timeout: 2_000 });
            }).toPass({ timeout: 15_000, intervals: [250, 500, 1_000] });
          }
          await expect(emptyState).toBeVisible();
          await settleNuxt(targetPage);
          expect(errors).toEqual([]);
        } finally {
          await context?.close();
        }
      });

      test("缺失实体返回产品错误状态", async ({ page }) => {
        const errors = capturePageFailures(page);
        const errorURL = new URL(contract.error.path, site.url).toString();
        const response = await page.goto(errorURL, {
          waitUntil: "domcontentloaded",
        });
        expect(response?.status()).toBe(contract.error.status);
        await expect(
          page.getByText(contract.error.text, { exact: false }).first(),
        ).toBeVisible();
        await settleNuxt(page);
        expect(
          errors.filter(
            (error) =>
              error !== `http ${contract.error.status}: ${errorURL}` &&
              !error.includes(`status of ${contract.error.status}`),
          ),
        ).toEqual([]);
      });
    });
  }
}
