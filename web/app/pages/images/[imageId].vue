<script setup lang="ts">
import type {
  GalleryDiscovery,
  GalleryImage,
  GalleryImageCard,
  GalleryRelatedImage,
} from "~/types/gallery";

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
const sequence = ref(createViewerSequence(imageId.value));
const candidatePool = ref<GalleryImageCard[]>([]);
const continuationPending = ref(false);
const closeTarget = ref("/images");

function mergeCandidates(items: GalleryImageCard[]): void {
  const byId = new Map(candidatePool.value.map((item) => [item.id, item]));
  items.forEach((item) => byId.set(item.id, item));
  candidatePool.value = [...byId.values()];
}

const candidateMap = computed(
  () => new Map(candidatePool.value.map((item) => [item.id, item])),
);
const previous = computed(() =>
  sequence.value.index > 0
    ? candidateMap.value.get(sequence.value.ids[sequence.value.index - 1]!)
    : undefined,
);
const next = computed(() =>
  sequence.value.index < sequence.value.ids.length - 1
    ? candidateMap.value.get(sequence.value.ids[sequence.value.index + 1]!)
    : undefined,
);

async function ensureContinuation(): Promise<void> {
  if (!import.meta.client) return;
  if (
    next.value ||
    continuationPending.value ||
    relatedStatus.value === "pending"
  )
    return;
  continuationPending.value = true;
  try {
    const seed =
      globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random()}`;
    const discovery = await $fetch<GalleryDiscovery>("/api/gallery/discovery", {
      query: { seed },
    });
    mergeCandidates(discovery.images);
    sequence.value = extendViewerSequence(
      sequence.value,
      discovery.images.map((item) => item.id),
    );
  } catch {
    // Related results remain usable; the next image can retry continuation.
  } finally {
    continuationPending.value = false;
  }
}

watch(
  [image, related],
  ([current, suggestions]) => {
    if (!current) return;
    mergeCandidates([current, ...suggestions]);
    sequence.value = moveViewerSequence(sequence.value, current.id);
    sequence.value = extendViewerSequence(
      sequence.value,
      suggestions.map((item) => item.id),
    );
    void ensureContinuation();
  },
  { immediate: true },
);

onMounted(() => {
  const back = window.history.state?.back;
  closeTarget.value = viewerCloseTarget(
    typeof back === "string" ? back : undefined,
  );
});

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
  void router.replace(closeTarget.value);
}

function navigate(targetId: string): void {
  sequence.value = moveViewerSequence(sequence.value, targetId);
  void router.replace(`/images/${encodeURIComponent(targetId)}`);
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
