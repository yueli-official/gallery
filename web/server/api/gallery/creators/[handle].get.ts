import type { GalleryArtworkCard, GalleryPublicCreator } from "../../../../app/types/gallery";

interface Envelope<T> { code: string; data: T; message: string; traceId: string }

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event);
  const handle = getRouterParam(event, "handle") || "";
  const response = await $fetch<Envelope<{ creator: GalleryPublicCreator; artworks: GalleryArtworkCard[] }>>(
    `${config.apiBase}/api/v1/gallery/creators/${encodeURIComponent(handle)}`,
  );
  if (response.code !== "ok") throw createError({ statusCode: 404, statusMessage: response.message });
  return response.data;
});
