import { describe, expect, it } from "vitest";

import { classificationMutationErrorMessage } from "../app/utils/classificationMutation";

describe("classification mutation errors", () => {
  it("replaces a protocol-level 404 with the actionable API restart message", () => {
    expect(
      classificationMutationErrorMessage(
        {
          failure: {
            code: "foundation.problem.invalid_body",
          },
        },
        "创建失败",
      ),
    ).toContain("重启 Gallery API");
  });

  it("explains duplicate slugs without exposing the resource key", () => {
    expect(
      classificationMutationErrorMessage(
        {
          failure: {
            code: "gallery.conflict",
            params: { resource: "classification_slug" },
          },
        },
        "创建失败",
      ),
    ).toBe("这个标识已经存在，请换一个标识。");
  });

  it("keeps server validation details next to the form", () => {
    expect(
      classificationMutationErrorMessage(
        {
          failure: {
            violations: [{ params: { detail: "标识只能使用小写字母" } }],
          },
        },
        "创建失败",
      ),
    ).toBe("标识只能使用小写字母");
  });
});
