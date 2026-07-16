<script setup lang="ts">
import type { GalleryImage, GalleryRelatedImage } from "~/types/gallery";

const route = useRoute("/images/[imageId]");
const router = useRouter();
const imageId = computed(() => String(route.params.imageId));
const { data, error, status, refresh } = await useFetch<{
  image: GalleryImage;
}>(() => `/api/gallery/images/${encodeURIComponent(imageId.value)}`);
const {
  data: relatedData,
  status: relatedStatus,
  refresh: refreshRelated,
} = await useAsyncData(
  "gallery-image-related",
  async () => {
    if (!data.value?.image) return { items: [] as GalleryRelatedImage[] };
    return await $fetch<{ items: GalleryRelatedImage[] }>(
      `/api/gallery/images/${encodeURIComponent(imageId.value)}/related`,
      { query: { size: 10 } },
    );
  },
  { watch: [imageId], immediate: Boolean(data.value?.image) },
);

if (import.meta.server && error.value) {
  const responseStatus =
    error.value.statusCode === 410
      ? 410
      : error.value.statusCode === 404
        ? 404
        : 502;
  setResponseStatus(responseStatus);
}

const image = computed(() => data.value?.image);
const related = computed(() => relatedData.value?.items || []);
const previous = computed(() => related.value[1]);
const next = computed(() => related.value[0]);

useSeoMeta({
  title: () => image.value?.title || "图片详情",
  description: () =>
    image.value?.description || image.value?.altText || "查看公开图片详情。",
  robots: () => (error.value ? "noindex, nofollow" : "index, follow"),
  ogType: "website",
  ogTitle: () => image.value?.title || "图片详情",
  ogDescription: () =>
    image.value?.description || image.value?.altText || "查看公开图片详情。",
  ogImage: () =>
    image.value ? galleryRendition(image.value.assetId, "og") : undefined,
  twitterCard: "summary_large_image",
});
useHead(() =>
  image.value
    ? {
        script: [
          {
            type: "application/ld+json",
            innerHTML: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "ImageObject",
              name: image.value.title,
              caption: image.value.altText,
              description: image.value.description || image.value.altText,
              contentUrl: galleryRendition(image.value.assetId, "display"),
              thumbnailUrl: galleryRendition(image.value.assetId, "thumbnail"),
              width: image.value.width,
              height: image.value.height,
              uploadDate: image.value.publishedAt,
              representativeOfPage: true,
            }).replaceAll("<", "\\u003c"),
          },
        ],
      }
    : {},
);

function closeViewer(): void {
  if (import.meta.client && window.history.length > 1) router.back();
  else void router.push("/images");
}

function navigate(targetId: string): void {
  void router.push(`/images/${encodeURIComponent(targetId)}`);
}

function retry(): void {
  void Promise.all([refresh(), refreshRelated()]);
}
</script>

<template>
  <GalleryViewer
    :image="image"
    :status="status"
    :failed="Boolean(error)"
    mode="page"
    :previous="previous"
    :next="next"
    :related="related"
    :related-pending="relatedStatus === 'pending'"
    @close="closeViewer"
    @retry="retry"
    @navigate="navigate"
  />
</template>
