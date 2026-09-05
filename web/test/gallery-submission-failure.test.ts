import { describe, expect, it } from "vitest";
import {
  gallerySubmissionFailure,
  gallerySubmissionErrorMessage,
} from "../app/utils/gallerySubmissionFailure";
import { remoteFailure } from "./http-failure-fixture";
const network = {
  failure: {
    kind: "network",
    code: "foundation.request.network",
    reauth: "not-attempted",
  },
};
describe("Gallery submission failure", () => {
  it("uses structured failures and never interprets raw thrown text", () => {
    expect(
      gallerySubmissionErrorMessage(
        remoteFailure("gallery.upstream_failed", 502),
      ),
    ).toBe("图片处理服务暂时不可用，请稍后重试。");
    expect(gallerySubmissionErrorMessage(network)).toBe(
      "网络连接失败，请检查网络后重试。",
    );
    expect(
      gallerySubmissionErrorMessage(new Error("SQL password=secret")),
    ).toBe("投稿未完成，请稍后重试。");
    expect(
      gallerySubmissionErrorMessage(new Error("gallery.upstream_failed")),
    ).toBe("投稿未完成，请稍后重试。");
  });
  it("treats an explicit duplicate as already submitted and leaves uncertain requests retryable", () => {
    expect(
      gallerySubmissionFailure(remoteFailure("gallery.conflict", 409)),
    ).toEqual({
      kind: "already-submitted",
      message: "这张图片已经投稿，无需重复提交。",
    });
    expect(gallerySubmissionFailure(network)).toEqual({
      kind: "retryable",
      message: "网络连接失败，请检查网络后重试。",
    });
  });
});
