import { describe, expect, it, vi } from "vitest";
import {
  gallerySubmissionFailure,
  gallerySubmissionErrorMessage,
  waitForGalleryAssetReady,
} from "../app/utils/galleryAssetReadiness";

describe("Gallery asset readiness", () => {
  it("waits for a quarantined upload before creating the submission", async () => {
    const load = vi
      .fn()
      .mockResolvedValueOnce({
        id: "asset-1",
        securityState: "quarantined",
        scanStatus: "pending",
      })
      .mockResolvedValueOnce({
        id: "asset-1",
        securityState: "ready",
        scanStatus: "clean",
      });
    const sleep = vi.fn().mockResolvedValue(undefined);

    const asset = await waitForGalleryAssetReady("asset-1", load, {
      attempts: 3,
      intervalMs: 1,
      sleep,
    });

    expect(asset.securityState).toBe("ready");
    expect(load).toHaveBeenCalledTimes(2);
    expect(sleep).toHaveBeenCalledTimes(1);
  });

  it("reports rejected media in user-facing language", async () => {
    await expect(
      waitForGalleryAssetReady(
        "asset-1",
        async () => ({
          id: "asset-1",
          securityState: "rejected",
          scanStatus: "malicious",
        }),
        { attempts: 1, intervalMs: 1 },
      ),
    ).rejects.toThrow("图片未通过安全检查，请更换文件后重试");
  });

  it("does not leak Gallery or Asset error codes to the queue", () => {
    expect(
      gallerySubmissionErrorMessage({
        failure: {
          kind: "remote",
          code: "gallery.upstream_failed",
          params: { upstreamCode: "asset.security.not_ready" },
        },
      }),
    ).toBe("图片仍在进行安全检查，请稍后重试。");
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
