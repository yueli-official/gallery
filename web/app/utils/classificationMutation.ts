import { getApiFailure } from "@yueli/http-runtime";
import { galleryFailureMessage } from "./galleryFailure";
export function classificationMutationErrorMessage(
  cause: unknown,
  fallback: string,
) {
  const failure = getApiFailure(cause);
  if (
    failure?.kind === "remote" &&
    failure.code === "gallery.conflict" &&
    failure.params.resource === "classification_slug"
  )
    return "这个标识已经存在，请换一个标识。";
  return galleryFailureMessage(cause, fallback);
}
