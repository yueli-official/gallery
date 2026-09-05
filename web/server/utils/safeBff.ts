import { sanitizeBffError } from "./bffError";

export function withSafeBffErrors(
  handler: ReturnType<typeof defineEventHandler>,
) {
  return defineEventHandler(async (event) => {
    try {
      return await handler(event);
    } catch (error) {
      const safe = sanitizeBffError(error);
      setResponseStatus(event, safe.status);
      setResponseHeader(event, "content-type", "application/problem+json");
      setResponseHeader(event, "x-trace-id", safe.traceId);
      return safe;
    }
  });
}
