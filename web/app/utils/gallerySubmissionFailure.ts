import { galleryFailureCode, galleryFailureMessage } from "./galleryFailure";
export function gallerySubmissionErrorMessage(reason: unknown): string {
  if (galleryFailureCode(reason) === "gallery.conflict")
    return "这张图片已经投稿，无需重复提交。";
  return galleryFailureMessage(reason, "投稿未完成，请稍后重试。");
}
export function gallerySubmissionFailure(reason: unknown): {
  kind: "already-submitted" | "retryable";
  message: string;
} {
  return {
    kind:
      galleryFailureCode(reason) === "gallery.conflict"
        ? "already-submitted"
        : "retryable",
    message: gallerySubmissionErrorMessage(reason),
  };
}
