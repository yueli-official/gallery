export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event);
  const path = getRouterParam(event, "path") || "";
  try {
    return await $fetch(`${config.apiBase}/api/v1/gallery/${path}`, {
      query: getQuery(event),
      headers: getHeader(event, "authorization") ? { authorization: getHeader(event, "authorization")! } : undefined,
    });
  } catch (reason: any) {
    throw createError({
      statusCode: reason?.response?.status || reason?.statusCode || 502,
      statusMessage: reason?.data?.message || reason?.statusMessage || "Gallery API request failed",
    });
  }
});
