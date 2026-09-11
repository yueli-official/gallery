import { describe, expect, it } from "vitest";

import type { GallerySubmission } from "../app/types/gallery";
import {
  submissionPreviewURL,
  submissionReviewAction,
  submissionStatus,
} from "../app/utils/submission-review";

function submission(patch: Partial<GallerySubmission> = {}): GallerySubmission {
  return {
    id: "submission id",
    assetId: "asset-1",
    imageId: "",
    title: "测试",
    description: "",
    sourceUrl: "",
    altText: "测试",
    primaryCategoryId: "",
    processingState: "queued",
    reviewState: "pending",
    safetyState: "pending",
    outcome: "pending",
    failureCode: "",
    reviewNote: "",
    ...patch,
  };
}

describe("submission review action", () => {
  it("explains why a queued submission cannot be approved", () => {
    expect(submissionReviewAction(submission())).toEqual({
      canApprove: false,
      label: "等待媒体处理",
      reason: "媒体正在排队；处理完成后这里会自动变为可批准。",
    });
  });

  it("allows ready submissions after a human safety decision", () => {
    expect(
      submissionReviewAction(
        submission({ processingState: "ready", safetyState: "uncertain" }),
      ),
    ).toMatchObject({ canApprove: true, label: "通过" });
  });

  it("uses the operator-gated preview route instead of the public asset route", () => {
    expect(submissionPreviewURL("submission id")).toBe(
      "/admin/submissions/submission%20id/preview",
    );
  });
});


describe("submission workflow status", () => {
  it.each([
    [{ processingState: "queued" }, "处理中"],
    [{ processingState: "processing" }, "处理中"],
    [{ processingState: "failed" }, "处理失败"],
    [{ processingState: "ready", safetyState: "uncertain" }, "等待审核"],
    [{ processingState: "ready", safetyState: "safe" }, "等待审核"],
    [{ outcome: "published", reviewState: "approved", safetyState: "uncertain" }, "审核通过"],
    [{ outcome: "duplicate", reviewState: "approved" }, "审核通过"],
    [{ outcome: "rejected", reviewState: "rejected", processingState: "failed" }, "审核未通过"],
    [{ outcome: "withdrawn" }, "已撤回"],
  ] as const)("maps %j to %s", (patch, expected) => {
    expect(submissionStatus(submission(patch))).toBe(expected);
  });

  it.each(["pending", "blocked", "unavailable"])("retains the approval guard for %s", safetyState => {
    expect(submissionReviewAction(submission({ processingState: "ready", safetyState })).canApprove).toBe(false);
  });
});
