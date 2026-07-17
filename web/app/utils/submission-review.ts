import type { GallerySubmission } from "~/types/gallery";

export interface SubmissionReviewAction {
  canApprove: boolean;
  label: string;
  reason: string;
}

export function submissionReviewAction(
  item: GallerySubmission,
): SubmissionReviewAction {
  if (item.reviewState !== "pending" || item.outcome !== "pending")
    return { canApprove: false, label: "审核已结束", reason: "" };
  if (item.processingState === "queued")
    return {
      canApprove: false,
      label: "等待媒体处理",
      reason: "媒体正在排队；处理完成后这里会自动变为可批准。",
    };
  if (item.processingState === "processing")
    return {
      canApprove: false,
      label: "正在检查媒体",
      reason: "正在读取尺寸、内容指纹并生成审核缩略图。",
    };
  if (item.processingState === "failed")
    return {
      canApprove: false,
      label: "媒体处理失败",
      reason: "媒体处理失败，修复或重新投稿后才能批准。",
    };
  if (item.safetyState === "pending")
    return {
      canApprove: false,
      label: "等待安全判断",
      reason: "媒体已就绪，仍在等待安全判断。",
    };
  if (item.safetyState === "blocked")
    return {
      canApprove: false,
      label: "内容已阻止",
      reason: "安全判断已阻止这条投稿，不能进入公开目录。",
    };
  if (item.safetyState === "unavailable")
    return {
      canApprove: false,
      label: "安全判断不可用",
      reason: "当前没有可用的安全判断，需要重新处理。",
    };
  return { canApprove: true, label: "批准进入目录", reason: "" };
}

export function submissionPreviewURL(submissionId: string): string {
  return `/api/v1/gallery/admin/submissions/${encodeURIComponent(submissionId)}/preview`;
}
