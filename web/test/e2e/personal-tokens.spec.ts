import { expect, request, test, type APIRequestContext } from "@playwright/test";

import { expectNoHorizontalOverflow, loginE2E } from "./runtime";

test.use({ trace: "off", video: "off", screenshot: "off" });

async function json(
  client: APIRequestContext,
  method: string,
  path: string,
  status: number,
  data?: unknown,
  headers?: Record<string, string>,
) {
  const response = await client.fetch(path, { method, data, headers });
  expect(
    response.status(),
    `${method} ${path}: ${response.status() === status ? "" : await response.text()}`,
  ).toBe(status);
  if (status === 204) {
    expect(await response.text()).toBe("");
    return;
  }
  const body = await response.json();
  if (status >= 400) {
    expect(response.headers()["content-type"]).toContain(
      "application/problem+json",
    );
    expect(body.status).toBe(status);
    expect(body.traceId).toBeTruthy();
  }
  return body;
}

test("developer token catalog, media upload, Gallery ownership, and revocation", async ({
  browser,
}, testInfo) => {
  test.setTimeout(240_000);
  const gallery =
    process.env.GALLERY_E2E_URL || "http://gallery.dev.yuelili.test:3007";
  const account =
    process.env.GALLERY_E2E_ACCOUNT_URL ||
    "http://account-gallery.dev.yuelili.test:3000";
  const api = process.env.GALLERY_E2E_API_URL || "http://127.0.0.1:8091";
  const asset = process.env.GALLERY_E2E_ASSET_URL || "http://127.0.0.1:8082";
  const marker = `${Date.now().toString(36)}-${testInfo.project.name || "desktop"}`;
  const scope = (key: string) =>
    `site:${Buffer.from("gallery-main-web").toString("base64url")}:${key}`;
  const capabilityKeys = [
    "media.upload",
    "gallery.submission.create",
    "gallery.submission.read_own",
    "gallery.submission.withdraw",
    "gallery.favorite.read",
    "gallery.favorite.manage",
    "gallery.comment.create",
    "gallery.dashboard.read",
    "gallery.image.read",
    "gallery.image.update",
    "gallery.image.hide",
    "gallery.submission.read",
    "gallery.submission.review",
    "gallery.collection.read",
    "gallery.collection.manage",
    "gallery.classification.read",
    "gallery.classification.proposal_review",
    "gallery.classification.govern",
    "gallery.case.read",
    "gallery.case.resolve",
    "gallery.discovery.read",
    "gallery.discovery.manage",
    "gallery.comment.read",
    "gallery.comment.moderate",
    "gallery.comment.delete",
  ];
  const clients: APIRequestContext[] = [];
  let tokenID = 0;
  let token = "";
  let submissionID = "";
  let commentID = "";

  const desktop = await loginE2E(
    browser,
    { viewport: { width: 1440, height: 1000 } },
    undefined,
    gallery,
  );
  const page = await desktop.newPage();

  try {
    const catalog = await json(
      desktop.request,
      "GET",
      `${account}/api/v1/pat/scopes`,
      200,
    );
    const galleryScopes = catalog.items.filter(
      (item: { site: string }) => item.site === "gallery-main-web",
    );
    expect(galleryScopes.map((item: { key: string }) => item.key).sort()).toEqual(
      capabilityKeys.map(scope).sort(),
    );
    expect(catalog.unavailableSites).not.toContain("Gallery");

    await page.goto(`${account}/developer-tokens`);
    const create = page.getByRole("button", { name: "创建令牌", exact: true });
    await expect(create).toBeVisible();
    await expect
      .poll(() =>
        create.evaluate((node) =>
          Boolean(
            (node as HTMLElement & { __vueParentComponent?: unknown })
              .__vueParentComponent,
          ),
        ),
      )
      .toBe(true);
    await create.click();
    const dialog = page.getByRole("dialog");
    await dialog
      .getByPlaceholder("例如：本地脚本")
      .fill(`Gallery 完整令牌 ${marker}`);
    for (const permission of galleryScopes) {
      await dialog
        .getByRole("checkbox", { name: permission.label, exact: true })
        .check();
    }
    await expect(
      dialog.getByRole("checkbox", { name: "上传投稿图片", exact: true }),
    ).toBeVisible();
    await expectNoHorizontalOverflow(page, "Gallery PAT 桌面权限目录");
    await page.screenshot({
      path: testInfo.outputPath("developer-token-permissions-desktop.png"),
      fullPage: false,
    });
    const createdResponse = page.waitForResponse(
      (response) =>
        response.url().endsWith("/api/v1/pat") &&
        response.request().method() === "POST",
    );
    await dialog.getByRole("button", { name: "创建令牌", exact: true }).click();
    const response = await createdResponse;
    expect(response.status()).toBe(201);
    const created = await response.json();
    tokenID = created.id;
    token = created.token;

    const galleryClient = await request.newContext({
      baseURL: api,
      extraHTTPHeaders: { authorization: `Bearer ${token}` },
    });
    const assetClient = await request.newContext({
      baseURL: asset,
      extraHTTPHeaders: {
        authorization: `Bearer ${token}`,
        "x-yueli-token-site": "gallery-main-web",
      },
    });
    clients.push(galleryClient, assetClient);

    await json(galleryClient, "GET", "/api/v1/gallery/discovery", 200);
    await json(galleryClient, "GET", "/api/v1/gallery/me", 403);
    await json(
      galleryClient,
      "GET",
      "/api/v1/authorization/manage/console",
      403,
    );
    await json(
      galleryClient,
      "GET",
      "/api/v1/internal/personal-token/permissions?userKey=TestA123",
      403,
    );
    const png = Buffer.from(
      "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAIAAAAlC+aJAAAAWElEQVR4nO3PQQ0AIBDAsAP/nuGNAvZoFSzZOjNnyNi1dwfgUQCeBeBZAB4F4FkAHgXgWQAeBeBZAB4F4FkAHgXgWQAeBeBZAB4F4FkAHgXgWQAeBeBZAB4F4FkAHgXgWQAeBeBZAB4F4FkA3gCJYQJ/r4tQWAAAAABJRU5ErkJggg==",
      "base64",
    );
    const upload = await json(
      assetClient,
      "POST",
      "/api/v1/assets/upload-init",
      201,
      {
        filename: `gallery-pat-${marker}.png`,
        mime: "image/png",
        size: png.length,
        siteKey: "gallery",
        profileKey: "gallery-submission",
        category: "gallery-submission",
        visibility: "private",
        multipart: false,
      },
    );
    const uploadResponse = await assetClient.put(upload.uploadUrl, {
      data: png,
      headers: upload.uploadHeaders || {},
    });
    expect(uploadResponse.ok(), await uploadResponse.text()).toBe(true);
    const finalized = await json(
      assetClient,
      "POST",
      "/api/v1/assets/finalize",
      201,
      { uploadToken: upload.uploadToken },
    );

    const options = await json(
      galleryClient,
      "GET",
      "/api/v1/gallery/submission-options",
      200,
    );
    const submission = await json(
      galleryClient,
      "POST",
      "/api/v1/gallery/submissions",
      201,
      {
        assetId: finalized.asset.id,
        title: `PAT 投稿 ${marker}`,
        description: "Gallery PAT Playwright acceptance",
        categoryIds: [options.categories[0].id],
        primaryCategoryId: options.categories[0].id,
        facets: [
          {
            facetId: options.facets[0].id,
            valueIds: [options.facets[0].values[0].id],
          },
        ],
        tags: [],
      },
    );
    submissionID = submission.submission.id;
    const mine = await json(
      galleryClient,
      "GET",
      `/api/v1/gallery/me/submissions?size=100`,
      200,
    );
    expect(mine.items.map((item: { id: string }) => item.id)).toContain(
      submissionID,
    );
    const images = await json(
      galleryClient,
      "GET",
      "/api/v1/gallery/images?size=1",
      200,
    );
    const imageID = images.items[0].id;
    const comment = await json(
      galleryClient,
      "POST",
      `/api/v1/gallery/images/${imageID}/comments`,
      201,
      { content: `PAT 评论 ${marker}` },
    );
    commentID = comment.comment.id;
    await json(
      galleryClient,
      "GET",
      `/api/v1/gallery/admin/images?size=1`,
      200,
    );
    await json(
      galleryClient,
      "DELETE",
      `/api/v1/gallery/admin/comments/${commentID}`,
      204,
    );
    commentID = "";
    await json(
      galleryClient,
      "POST",
      `/api/v1/gallery/me/submissions/${submissionID}/withdraw`,
      200,
    );
    submissionID = "";

    const mobile = await loginE2E(
      browser,
      { viewport: { width: 390, height: 844 }, isMobile: true },
      undefined,
      gallery,
    );
    try {
      const mobilePage = await mobile.newPage();
      await mobilePage.goto(`${account}/developer-tokens`);
      const mobileCreate = mobilePage.getByRole("button", {
        name: "创建令牌",
        exact: true,
      });
      await expect
        .poll(() =>
          mobileCreate.evaluate((node) =>
            Boolean(
              (node as HTMLElement & { __vueParentComponent?: unknown })
                .__vueParentComponent,
            ),
          ),
        )
        .toBe(true);
      await mobileCreate.click();
      const mobileDialog = mobilePage.getByRole("dialog");
      await expect(mobileDialog).toBeVisible();
      await expect(
        mobileDialog.getByRole("checkbox", {
          name: "上传投稿图片",
          exact: true,
        }),
      ).toBeVisible();
      await mobileDialog.getByLabel("搜索分类或操作").fill("投稿");
      const mobileSubmissionPermission = mobileDialog.getByRole("checkbox", {
        name: "读取自己的投稿",
        exact: true,
      });
      await expect(mobileSubmissionPermission).toBeVisible();
      await mobileSubmissionPermission.scrollIntoViewIfNeeded();
      await expectNoHorizontalOverflow(mobilePage, "Gallery PAT 手机权限目录");
      await mobilePage.screenshot({
        path: testInfo.outputPath("developer-token-permissions-mobile.png"),
        fullPage: false,
      });
    } finally {
      await mobile.close();
    }

    await json(desktop.request, "DELETE", `${account}/api/v1/pat/${tokenID}`, 204);
    tokenID = 0;
    await json(galleryClient, "GET", "/api/v1/gallery/me/submissions", 401);
    await json(assetClient, "GET", "/api/v1/assets/upload-policy", 403);
  } finally {
    if (commentID && token) {
      const cleanup = await request.newContext({
        baseURL: api,
        extraHTTPHeaders: { authorization: `Bearer ${token}` },
      });
      await cleanup.delete(`/api/v1/gallery/admin/comments/${commentID}`);
      await cleanup.dispose();
    }
    if (submissionID && token) {
      const cleanup = await request.newContext({
        baseURL: api,
        extraHTTPHeaders: { authorization: `Bearer ${token}` },
      });
      await cleanup.post(
        `/api/v1/gallery/me/submissions/${submissionID}/withdraw`,
      );
      await cleanup.dispose();
    }
    if (tokenID) {
      await desktop.request.delete(`${account}/api/v1/pat/${tokenID}`);
    }
    for (const client of clients) await client.dispose();
    await desktop.close();
  }
});
