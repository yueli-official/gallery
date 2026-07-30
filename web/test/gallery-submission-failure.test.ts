import { describe, expect, it } from "vitest";
import {
  gallerySubmissionFailure,
  gallerySubmissionErrorMessage,
} from "../app/utils/gallerySubmissionFailure";

describe("Gallery submission failure", () => {
  it("does not leak Gallery error codes to the queue", () => {
    expect(
      gallerySubmissionErrorMessage(new Error("gallery.upstream_failed")),
    ).toBe("图片处理服务暂时不可用，请稍后重试。");
    expect(
      gallerySubmissionErrorMessage({
        failure: { kind: "network", code: "foundation.network.failed" },
      }),
    ).toBe("网络连接失败，请检查网络后重试。");
  });

  it("treats duplicate submissions as already submitted instead of retryable", () => {
    expect(
      gallerySubmissionFailure({
        data: {
          failure: {
            kind: "conflict",
            code: "gallery.conflict",
          },
        },
      }),
    ).toEqual({
      kind: "already-submitted",
      message: "这张图片已经投稿，无需重复提交。",
    });

    expect(
      gallerySubmissionFailure({
        failure: {
          kind: "network",
          code: "foundation.network.failed",
        },
      }),
    ).toEqual({
      kind: "retryable",
      message: "网络连接失败，请检查网络后重试。",
    });
  });
});
