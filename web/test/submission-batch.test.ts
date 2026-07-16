import { describe, expect, it } from "vitest";
import {
  applyGallerySubmissionDefaults,
  gallerySubmissionMetadataValid,
} from "../app/utils/gallerySubmissionBatch";

const defaults = {
  title: "城市夜景系列",
  description: "共同说明",
  sourceUrl: "https://example.com/source",
  primaryCategoryId: "photography",
  sceneValueIds: ["city", "night"],
  tags: "夜景, 城市",
};

describe("Gallery 投稿批量默认值", () => {
  it("支持统一、单独修改后恢复共同值", () => {
    const item = {
      ...defaults,
      title: "文件标题",
      sceneValueIds: [],
      status: "ready",
      customized: true,
    };
    expect(applyGallerySubmissionDefaults(item, defaults)).toBe(true);
    expect(item).toMatchObject({
      title: "城市夜景系列",
      customized: false,
      sceneValueIds: ["city", "night"],
    });

    item.title = "单张标题";
    item.primaryCategoryId = "illustration";
    item.customized = true;
    applyGallerySubmissionDefaults(item, defaults);
    expect(item).toMatchObject({
      title: "城市夜景系列",
      primaryCategoryId: "photography",
      customized: false,
    });
  });

  it("完成项锁定且必填元数据按逐项状态验证", () => {
    const completed = {
      ...defaults,
      title: "已完成",
      sceneValueIds: ["city"],
      status: "completed",
      customized: true,
    };
    expect(applyGallerySubmissionDefaults(completed, defaults)).toBe(false);
    expect(completed.title).toBe("已完成");
    expect(gallerySubmissionMetadataValid(completed)).toBe(true);
    expect(
      gallerySubmissionMetadataValid({ ...defaults, primaryCategoryId: "" }),
    ).toBe(false);
  });
});
