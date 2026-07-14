import type { GalleryArtwork } from "../../../../app/types/gallery";

interface Envelope<T> {
  code: string;
  data: T;
  message: string;
  traceId: string;
}

export default defineEventHandler(async (event) => {
  const artworkId = getRouterParam(event, "artworkId");
  if (!artworkId) {
    throw createError({ statusCode: 400, statusMessage: "Artwork ID is required" });
  }
  const config = useRuntimeConfig(event);
  const response = await $fetch<Envelope<{ artwork: GalleryArtwork }>>(
    `${config.apiBase}/api/v1/gallery/artworks/${encodeURIComponent(artworkId)}`,
  );
  if (response.code !== "ok") {
    throw createError({
      statusCode: 502,
      statusMessage: response.message || "Gallery API request failed",
    });
  }
  return response.data.artwork;
});

