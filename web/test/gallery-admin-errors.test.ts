import { describe, expect, it } from "vitest";
import { galleryAdminMutationErrorMessage } from "../app/utils/galleryAdminErrors";
import { remoteFailure } from "./http-failure-fixture";
describe("Gallery admin mutation errors", () => {
  it("distinguishes invalid revision from an optimistic conflict", () => {
    expect(
      galleryAdminMutationErrorMessage(
        remoteFailure("common.validation_failed", 400, {
          violations: [
            {
              pointer: "/expectedUpdatedAt",
              code: "validation.invalid",
              params: {},
            },
          ],
        }),
      ),
    ).toBe("编辑器中的记录版本无效，请重新打开后再试。");
  });
  it("describes a real optimistic conflict without blaming another person", () => {
    expect(
      galleryAdminMutationErrorMessage(remoteFailure("gallery.conflict", 409)),
    ).toBe("图片状态已经变化，请重新打开编辑器后再保存。");
  });
});
