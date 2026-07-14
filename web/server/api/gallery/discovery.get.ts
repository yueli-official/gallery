import type { GalleryDiscovery } from "../../../app/types/gallery";

interface Envelope<T> {
  code: string;
  data: T;
  message: string;
  traceId: string;
}

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event);
  const response = await $fetch<Envelope<GalleryDiscovery>>(
    `${config.apiBase}/api/v1/gallery/discovery`,
  );
  if (response.code !== "ok") {
    throw createError({
      statusCode: 502,
      statusMessage: response.message || "Gallery API request failed",
    });
  }
  return response.data;
});

