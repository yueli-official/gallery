import { getApiFailure } from "@yueli/http-runtime";
import { galleryFailureCode, galleryFailureMessage } from "./galleryFailure";
export function galleryAdminMutationErrorMessage(reason: unknown): string {
  const failure = getApiFailure(reason);
  if (
    failure?.kind === "remote" &&
    failure.violations.some((item) => item.pointer === "/expectedUpdatedAt")
  )
    return "编辑器中的记录版本无效，请重新打开后再试。";
  if (galleryFailureCode(reason) === "gallery.conflict")
    return "图片状态已经变化，请重新打开编辑器后再保存。";
  return galleryFailureMessage(reason, "图片信息没有保存，请稍后重试。");
}
