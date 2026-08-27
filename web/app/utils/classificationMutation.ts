export interface ClassificationMutationFailure {
  code?: string;
  message?: string;
  statusCode?: number;
  failure?: {
    code?: string;
    status?: number;
    params?: { resource?: string };
    violations?: Array<{ params?: { detail?: string } }>;
  };
  data?: {
    code?: string;
    message?: string;
    params?: { resource?: string };
    violations?: Array<{ params?: { detail?: string } }>;
    failure?: ClassificationMutationFailure["failure"];
  };
}

export function classificationMutationErrorMessage(
  cause: unknown,
  fallback: string,
) {
  const failure = cause as ClassificationMutationFailure;
  const apiFailure = failure.failure || failure.data?.failure;
  const code =
    apiFailure?.code || failure.code || failure.data?.code || failure.message || "";
  if (
    failure.statusCode === 404 ||
    apiFailure?.status === 404 ||
    code === "foundation.problem.invalid_body" ||
    code === "foundation.problem.invalid_content_type"
  ) {
    return "当前运行的 Gallery API 尚未加载分类写入接口，请从原启动终端重启 Gallery API 后再试。";
  }
  if (
    code === "gallery.conflict" &&
    (apiFailure?.params?.resource || failure.data?.params?.resource) ===
      "classification_slug"
  ) {
    return "这个标识已经存在，请换一个标识。";
  }
  const detail =
    apiFailure?.violations?.[0]?.params?.detail ||
    failure.data?.violations?.[0]?.params?.detail;
  if (detail) return detail;
  return failure.data?.message || failure.message || fallback;
}
