import { getApiFailure, resolveFailureFeedback } from "@yueli/http-runtime";
import {
  galleryFailurePresentation,
  type GalleryFailureCode,
} from "../generated/galleryFailure";
const messages: Record<GalleryFailureCode, string> = {
  "gallery.not_found": "内容不存在或已被移除，请刷新后重试。",
  "gallery.gone": "内容已下线或已删除。",
  "gallery.not_initialized": "图库暂未准备完成，请稍后重试。",
  "gallery.forbidden": "当前账号没有执行此操作的权限。",
  "gallery.conflict": "内容状态已变化或已经存在，请刷新后核对。",
  "gallery.invalid_state": "当前状态不允许此操作，请刷新后核对。",
  "gallery.upstream_failed": "图片处理服务暂时不可用，请稍后重试。",
  "gallery.rate_limited": "投稿过于频繁，请稍后再试。",
  "gallery.challenge_required": "需要先完成安全验证，请按提示操作。",
  "gallery.abuse_unavailable": "投稿验证服务暂时不可用，请稍后重试。",
  "gallery.abuse_attempt_replayed": "本次投稿请求已处理，请查看投稿记录。",
  "gallery.authorization_unavailable": "权限服务暂时不可用，请稍后重试。",
  "gallery.gateway.unavailable": "服务暂时不可用，请稍后重试。",
  "gallery.gateway.timeout": "请求超时，请稍后重试。",
  "gallery.request.invalid": "请求格式不正确，请检查后重试。",
  "gallery.request.method_not_allowed":
    "当前请求方式不受支持，请刷新页面后重试。",
  "gallery.request.body_too_large": "提交内容过大，请减少后重试。",
  "gallery.not_authenticated": "请先登录后再继续。",
};
export function galleryFailureCode(reason: unknown) {
  return getApiFailure(reason)?.code || "";
}
export function galleryFailureFeedback(
  reason: unknown,
  fallback: string,
  fields?: Readonly<Record<string, string>>,
) {
  return resolveFailureFeedback(reason, {
    fallback,
    fields,
    resolveText(code) {
      if (Object.hasOwn(galleryFailurePresentation, code))
        return { message: messages[code as GalleryFailureCode] };
      if (code === "common.validation_failed")
        return { message: "填写的信息不符合要求，请检查后重试。" };
      if (code === "common.unauthorized")
        return { message: "请先登录后再继续。" };
      if (code === "common.rate_limited")
        return { message: "操作过于频繁，请稍后重试。" };
      if (
        code === "foundation.network.failed" ||
        code === "foundation.request.network"
      )
        return { message: "网络连接失败，请检查网络后重试。" };
      if (
        code === "foundation.timeout" ||
        code === "foundation.request.timeout"
      )
        return { message: "请求超时，请稍后重试。" };
      if (code.startsWith("validation."))
        return {
          message:
            code === "validation.required"
              ? "请填写此项。"
              : "此项内容不符合要求，请检查后重试。",
        };
      return undefined;
    },
  });
}
export function galleryFailureMessage(
  reason: unknown,
  fallback = "请求未完成，请稍后重试。",
) {
  const feedback = galleryFailureFeedback(reason, fallback);
  return [feedback.message, ...feedback.summary].join(" ");
}

export interface GalleryOperationFailure {
  status: number;
  code: string;
  traceId: string;
  params?: import("@yueli/http-runtime").ProblemParams;
  violations?: readonly import("@yueli/http-runtime").Violation[];
}
export function galleryBatchFailureMessage(
  failure: GalleryOperationFailure | undefined,
  fallback: string,
) {
  if (!failure) return fallback;
  return galleryFailureMessage(
    {
      kind: "remote",
      ...failure,
      params: failure.params ?? {},
      violations: failure.violations ?? [],
      reauth: "not-attempted",
    },
    fallback,
  );
}
