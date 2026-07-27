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

function failureDetails(reason: unknown) {
  const error = record(reason);
  const data = record(error?.data);
  const failure = record(error?.failure) ?? record(data?.failure);
  const code =
    stableCode(failure?.code) ||
    stableCode(error?.code) ||
    stableCode(error?.message);
  const violations = [
    ...(Array.isArray(failure?.violations) ? failure.violations : []),
    ...(Array.isArray(error?.violations) ? error.violations : []),
  ];
  const pointers = violations
    .map((violation) => record(violation)?.pointer)
    .filter((pointer): pointer is string => typeof pointer === "string");
  return { code, failure, pointers };
}

export function galleryAdminMutationErrorMessage(reason: unknown): string {
  const { code, failure, pointers } = failureDetails(reason);

  if (code === "common.validation_failed") {
    if (pointers.includes("/expectedUpdatedAt"))
      return "编辑器中的记录版本无效，请重新打开后再试。";
    return "填写的信息不符合要求，请检查后再保存。";
  }
  if (code === "gallery.conflict")
    return "图片状态已经变化，请重新打开编辑器后再保存。";
  if (failure?.kind === "network" || code === "foundation.network.failed")
    return "网络连接失败，请检查网络后重试。";
  if (failure?.kind === "timeout" || code === "foundation.timeout")
    return "保存请求超时，请稍后重试。";
  if (code === "gallery.forbidden")
    return "当前账号没有修改图片的权限。";
  if (code === "gallery.not_found") return "图片记录已不存在，请刷新列表。";
  return "图片信息没有保存，请稍后重试。";
}
