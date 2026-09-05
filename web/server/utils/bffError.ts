import { randomUUID } from "node:crypto";
import type { GalleryFailureCode } from "../../app/generated/galleryFailure";
const statuses: Record<number, GalleryFailureCode> = {
  400: "gallery.request.invalid",
  401: "gallery.not_authenticated",
  403: "gallery.forbidden",
  404: "gallery.not_found",
  405: "gallery.request.method_not_allowed",
  409: "gallery.conflict",
  413: "gallery.request.body_too_large",
  502: "gallery.gateway.unavailable",
  504: "gallery.gateway.timeout",
};
export function sanitizeBffError(error: unknown, traceId = randomUUID()) {
  const raw =
    error && typeof error === "object"
      ? Number((error as { statusCode?: unknown }).statusCode)
      : 500;
  const status = statuses[raw] ? raw : raw === 429 ? 429 : 500;
  const code =
    statuses[status] ||
    (status === 429 ? "common.rate_limited" : "common.internal");
  return {
    type: `https://errors.yueli.dev/problems/${code}`,
    status,
    code,
    traceId,
  };
}
