interface Envelope<T> {
  code: string;
  data: T;
  message: string;
  traceId: string;
}

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event);
  const path = getRouterParam(event, "path") || "";
  try {
    const response = await $fetch<Envelope<unknown>>(`${config.apiBase}/api/v1/gallery/${path}`, {
      method: "POST",
      body: await readBody(event),
      headers: getHeader(event, "authorization") ? { authorization: getHeader(event, "authorization")! } : undefined,
    });
    if (response.code !== "ok") throw createError({ statusCode: 502, statusMessage: response.message || "Gallery API request failed" });
    return response.data;
  } catch (reason: any) {
    throw createError({
      statusCode: reason?.response?.status || reason?.statusCode || 502,
      statusMessage: reason?.data?.message || reason?.statusMessage || "Gallery API request failed",
    });
  }
});
