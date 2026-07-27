import { describe, expect, it } from "vitest";
import { galleryAdminMutationErrorMessage } from "../app/utils/galleryAdminErrors";

describe("Gallery admin mutation errors", () => {
  it("does not describe validation failures as another operator's update", () => {
    expect(
      galleryAdminMutationErrorMessage({
        kind: "remote",
        status: 400,
        code: "common.validation_failed",
        violations: [
          {
            pointer: "/expectedUpdatedAt",
            code: "validation.invalid",
          },
        ],
      }),
    ).toBe("编辑器中的记录版本无效，请重新打开后再试。");
  });

  it("describes a real optimistic conflict without blaming another person", () => {
    expect(
      galleryAdminMutationErrorMessage({
        kind: "remote",
        status: 409,
        code: "gallery.conflict",
      }),
    ).toBe("图片状态已经变化，请重新打开编辑器后再保存。");
  });
});
