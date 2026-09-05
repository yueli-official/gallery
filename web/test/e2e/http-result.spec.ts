import { expect, test } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";
import { loginE2E, settleNuxt } from "./runtime";

const site = process.env.GALLERY_E2E_URL || "http://192.168.5.7:3007";

test("HTTP Result projects pages, 201 Location, 204 and per-item Problems", async ({
  browser,
}) => {
  const context = await loginE2E(browser, {}, undefined, site);
  try {
    const request = context.request;
    const pageResponse = await request.get(
      `${site}/api/gallery/images?page=1&size=12`,
    );
    const page = await pageResponse.json();
    expect(pageResponse.status()).toBe(200);
    expect(page).toMatchObject({ page: 1, size: 12 });
    expect(page.items.length).toBeGreaterThan(0);
    expect(page).not.toHaveProperty("pageSize");
    expect(page).not.toHaveProperty("totalPages");
    const imageId = page.items[0].id;
    const name = `http-result-${Date.now()}`;
    const created = await request.post(
      `${site}/api/gallery/admin/collections`,
      { data: { name, slug: name, visibility: "private" } },
    );
    expect(created.status()).toBe(201);
    const collection = (await created.json()).collection;
    expect(created.headers().location).toBe(
      `/api/gallery/admin/collections/${collection.id}`,
    );
    const comment = await request.post(
      `${site}/api/gallery/images/${imageId}/comments`,
      { data: { content: name } },
    );
    expect(comment.status()).toBe(201);
    const commentId = (await comment.json()).comment.id;
    const deleted = await request.delete(
      `${site}/api/gallery/admin/comments/${commentId}`,
    );
    expect(deleted.status()).toBe(204);
    expect(await deleted.text()).toBe("");
    const event = await request.post(
      `${site}/api/gallery/images/${imageId}/events`,
      { data: { type: "qualified_view", sessionKey: name } },
    );
    expect(event.status()).toBe(204);
    const imageResponse = await request.get(
      `${site}/api/gallery/admin/images/${imageId}`,
    );
    const image = (await imageResponse.json()).image;
    const batch = await request.post(`${site}/api/gallery/admin/images/bulk`, {
      data: {
        action: "set_primary_category",
        primaryCategoryId: image.primaryCategoryId,
        imageIds: [imageId, "not-an-id"],
      },
    });
    expect(batch.status()).toBe(200);
    const results = (await batch.json()).results;
    expect(results[0]).toMatchObject({ success: true });
    expect(results[1]).toMatchObject({
      success: false,
      failure: {
        status: 404,
        code: "gallery.not_found",
        traceId: batch.headers()["x-trace-id"],
      },
    });
    expect(results[1]).not.toHaveProperty("error");
    const cursor = await request.get(
      `${site}/api/gallery/admin/classification/tags?size=1`,
    );
    expect(await cursor.json()).toHaveProperty("items");
    expect(await cursor.json()).not.toHaveProperty("page");
    const browserPage = await context.newPage();
    const favoritesBefore = (await (await request.get(`${site}/api/gallery/me/favorites?size=60`)).json()).collection;
    expect(favoritesBefore.total).toBeLessThan(60);
    const alreadyFavorite = favoritesBefore.items.some((item: { id: string }) => item.id === imageId);
    const favorite = await request.put(`${site}/api/gallery/me/favorites/${imageId}`, { data: { version: favoritesBefore.version } });
    expect(favorite.status()).toBe(200);
    const favoriteVersion = (await favorite.json()).collection.version;
    try {
      await browserPage.goto(`${site}/favorites`); await settleNuxt(browserPage);
      await expect(browserPage.locator(`a[href="/images/${imageId}"]`).first()).toBeVisible();
    } finally {
      if (!alreadyFavorite) {
        const removed = await request.delete(`${site}/api/gallery/me/favorites/${imageId}?version=${favoriteVersion}`);
        expect(removed.status()).toBe(200);
      }
    }
    for (let i = 0; i < 3; i++) {
      const response = await browserPage.goto(`${site}/manage/images`);
      expect(await response!.text()).toContain("data-admin-shell");
      await settleNuxt(browserPage);
    }
  } finally {
    await context.close();
  }
});

for (const width of [390, 1440]) {
  test(`image editor retains its draft and shows safe field feedback at ${width}px`, async ({
    browser,
  }, info) => {
    const context = await loginE2E(
      browser,
      { viewport: { width, height: 900 } },
      undefined,
      site,
    );
    const page = await context.newPage();
    try {
      await page.goto(`${site}/manage/images`);
      await settleNuxt(page);
      await page
        .getByRole("button", { name: /^编辑图片/ })
        .first()
        .click();
      const dialog = page.getByRole("dialog", { name: "编辑图片" });
      const title = dialog.getByRole("textbox", { name: "标题", exact: false });
      await expect(title).toBeVisible();
      await title.fill("保留这份编辑草稿");
      await page.route("**/api/gallery/admin/images/*", (route) => {
        if (route.request().method() !== "PATCH") return route.continue();
        return route.fulfill({
          status: 400,
          contentType: "application/problem+json",
          headers: { "x-trace-id": "gallery-field-trace" },
          json: {
            type: "https://errors.yueli.dev/problems/common.validation_failed",
            status: 400,
            code: "common.validation_failed",
            traceId: "gallery-field-trace",
            detail: "SQL password=secret",
            violations: [
              { pointer: "/title", code: "validation.invalid" },
              { pointer: "/unknown", code: "validation.required" },
            ],
          },
        });
      });
      await dialog
        .getByRole("button", { name: "保存更改", exact: true })
        .click();
      await expect(title).toHaveValue("保留这份编辑草稿");
      await expect(title).toHaveAttribute("aria-invalid", "true");
      await expect(dialog.locator("[data-gallery-failure]")).toContainText(
        "请填写此项",
      );
      await expect(page.getByText("SQL password=secret")).toHaveCount(0);
      await dialog.getByText("技术详情", { exact: true }).click();
      await expect(dialog.locator("[data-gallery-failure]")).toContainText(
        "gallery-field-trace",
      );
      await page.screenshot({ path: info.outputPath(`field-${width}.png`) });
      const axe = await new AxeBuilder({ page })
        .include('[role="dialog"]')
        .analyze();
      expect(
        axe.violations.filter((v) =>
          ["serious", "critical"].includes(v.impact || ""),
        ),
      ).toEqual([]);
    } finally {
      await context.close();
    }
  });
}

test("real upload completes through the same-origin Asset proxy and creates a submission", async ({
  browser,
}, info) => {
  const context = await loginE2E(
    browser,
    { viewport: { width: 1440, height: 1000 } },
    undefined,
    site,
  );
  const page = await context.newPage();
  try {
    await page.goto(`${site}/submit`);
    await settleNuxt(page);
    await page
      .getByRole("combobox", { name: /主分类/ })
      .first()
      .click();
    await page.getByRole("option").first().click();
    await page.getByRole("combobox", { name: /场景/ }).first().click();
    await page.getByRole("option").first().click();
    await page.keyboard.press("Escape");
    const dataURL = await page.evaluate(() => {
      const canvas = document.createElement("canvas");
      canvas.width = 64;
      canvas.height = 64;
      const drawing = canvas.getContext("2d")!;
      drawing.fillStyle = `#${(Date.now() % 16777216).toString(16).padStart(6, "0")}`;
      drawing.fillRect(0, 0, 64, 64);
      return canvas.toDataURL("image/png");
    });
    await page
      .locator('input[type="file"]')
      .first()
      .setInputFiles({
        name: `http-result-${Date.now()}.png`,
        mimeType: "image/png",
        buffer: Buffer.from(dataURL.split(",")[1]!, "base64"),
      });
    const submitted = page.waitForResponse(
      (response) =>
        new URL(response.url()).pathname === "/api/gallery/submissions" &&
        response.request().method() === "POST",
    );
    await page.getByRole("button", { name: "开始投稿", exact: true }).click();
    const response = await submitted;
    expect(response.status()).toBe(201);
    await expect(
      page.getByText("已完成", { exact: true }).first(),
    ).toBeVisible();
    await page.screenshot({ path: info.outputPath("real-submission.png") });
  } finally {
    await context.close();
  }
});
