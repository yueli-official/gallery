function record(value: unknown): Record<string, unknown> | undefined {
  return typeof value === "object" && value !== null
    ? (value as Record<string, unknown>)
    : undefined;
}

function stableCode(value: unknown): string {
  return typeof value === "string" && /^[a-z][a-z0-9._-]+$/.test(value)
    ? value
    : "";
}

function gallerySubmissionError(reason: unknown) {
  const error = record(reason);
  const nestedData = record(error?.data);
  const failure = record(error?.failure) ?? record(nestedData?.failure);
  const code =
    stableCode(failure?.code) ||
    stableCode(error?.code) ||
    stableCode(error?.message);

  return { code, error, failure };
}

export function gallerySubmissionErrorMessage(reason: unknown): string {
  const { code, error, failure } = gallerySubmissionError(reason);

  if (failure?.kind === "network" || code === "foundation.network.failed") {
    return "网络连接失败，请检查网络后重试。";
  }
  if (failure?.kind === "timeout" || code === "foundation.timeout") {
    return "请求超时，请稍后重试。";
  }

  const messages: Record<string, string> = {
    "gallery.conflict": "这张图片已经投稿，无需重复提交。",
    "gallery.upstream_failed": "图片处理服务暂时不可用，请稍后重试。",
    "gallery.rate_limited": "投稿过于频繁，请稍后再试。",
    "gallery.challenge_required": "需要先完成安全验证，请按提示操作。",
    "gallery.abuse_unavailable": "投稿验证服务暂时不可用，请稍后重试。",
    "gallery.authorization_unavailable": "权限服务暂时不可用，请稍后重试。",
    "gallery.forbidden": "当前账号没有执行此操作的权限。",
    "common.validation_failed": "投稿信息不完整或格式不正确，请检查后重试。",
  };
  if (messages[code]) return messages[code];

  const message = typeof error?.message === "string" ? error.message.trim() : "";
  return message && !stableCode(message) ? message : "投稿未完成，请稍后重试。";
}

export function gallerySubmissionFailure(reason: unknown): {
  kind: "already-submitted" | "retryable";
  message: string;
} {
  return {
    kind:
      gallerySubmissionError(reason).code === "gallery.conflict"
        ? "already-submitted"
        : "retryable",
    message: gallerySubmissionErrorMessage(reason),
  };
}
