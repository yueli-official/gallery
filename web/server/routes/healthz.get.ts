export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event);
  const imageId = String(config.healthMediaImageId || "");
  if (!imageId) return { status: "up" };

  // Local readiness checks the same signed delivery path used by the browser.
  // Return status only: this endpoint must never become an unsigned image alias.
  const result = await $fetch<{
    image: { delivery?: { renditions: Record<string, string> } };
  }>(`${String(config.apiBase).replace(/\/$/u, "")}/api/v1/gallery/images/${encodeURIComponent(imageId)}`, {
    timeout: 10_000,
  });
  const mediaURL = result.image.delivery?.renditions.thumbnail;
  if (!mediaURL) throw createError({ statusCode: 503, statusMessage: "Media is not ready" });
  const target = new URL(mediaURL);
  if (!target.pathname.startsWith("/asset-api/api/v1/protected-media/")) {
    throw createError({ statusCode: 503, statusMessage: "Media delivery is not configured" });
  }
  const media = await event.$fetch<ArrayBuffer>(`${target.pathname}${target.search}`, {
    responseType: "arrayBuffer",
    timeout: 10_000,
  });
  const header = new TextDecoder().decode(new Uint8Array(media).subarray(0, 12));
  if (!header.startsWith("RIFF") || header.slice(8, 12) !== "WEBP") {
    throw createError({ statusCode: 503, statusMessage: "Media is not ready" });
  }
  return { status: "up", media: "ready" };
});
