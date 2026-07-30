import { describe, expect, it } from "vitest";

import type { GallerySubmission } from "../app/types/gallery";
import {
  submissionPreviewURL,
  submissionReviewAction,
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
      ).canApprove,
    ).toBe(true);
  });

  it("uses the operator-gated preview route instead of the public asset route", () => {
    expect(submissionPreviewURL("submission id")).toBe(
      "/admin/submissions/submission%20id/preview",
    );
  });
});
