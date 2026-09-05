import { describe, expect, it } from "vitest";
import { classificationMutationErrorMessage } from "../app/utils/classificationMutation";
import { remoteFailure } from "./http-failure-fixture";
describe("classification mutation errors", () => {
  it("uses the action fallback for malformed transport failures", () => {
    expect(
      classificationMutationErrorMessage(
        { statusCode: 404, message: "restart /server/path" },
        "创建失败，请刷新后重试。",
      ),
    ).toBe("创建失败，请刷新后重试。");
  });
  it("explains duplicate slugs without exposing internal resource keys", () => {
    expect(
      classificationMutationErrorMessage(
        remoteFailure("gallery.conflict", 409, {
          params: { resource: "classification_slug" },
        }),
        "创建失败",
      ),
    ).toBe("这个标识已经存在，请换一个标识。");
  });
  it("localizes validation codes without using server detail", () => {
    const message = classificationMutationErrorMessage(
      remoteFailure("common.validation_failed", 400, {
        violations: [
          {
            pointer: "/slug",
            code: "validation.invalid",
            params: { detail: "SQL password=secret" },
          },
        ],
      }),
      "创建失败",
    );
    expect(message).toContain("此项内容不符合要求");
    expect(message).not.toContain("SQL");
  });
});
