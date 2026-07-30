import { sanitizeBffError } from "./bffError";

export function withSafeBffErrors(
  handler: ReturnType<typeof defineEventHandler>,
) {
  return defineEventHandler(async (event) => {
    try {
      return await handler(event);
    } catch (error) {
      const safe = sanitizeBffError(error);
      setResponseStatus(event, safe.statusCode, safe.statusMessage);
      return safe;
    }
  });
}
