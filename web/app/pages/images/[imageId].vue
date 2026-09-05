<script setup lang="ts">
import type {
  GalleryDiscovery,
  GalleryImage,
  GalleryImageCard,
  GalleryImagePage,
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
const navigationSession = useGalleryViewerNavigation(imageId.value);
const sequence = computed({
  get: () => navigationSession.value.sequence,
  set: (value) => {
    navigationSession.value.sequence = value;
  },
});
const candidatePool = computed({
  get: () => navigationSession.value.candidates,
  set: (value) => {
    navigationSession.value.candidates = value;
  },
});
const continuationPending = ref(false);
const previousContinuationPending = ref(false);
const closeTarget = ref("/images");
const navigationReady = ref(false);

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
    const catalog = navigationSession.value.catalog;
    if (catalog) {
      if (!catalog.nextPage) return;
      const requestedPage = catalog.nextPage;
      const response = await $fetch<GalleryImagePage>("/api/gallery/images", {
        query: { ...catalog.request, page: requestedPage },
      });
      mergeCandidates(response.items);
      sequence.value = extendViewerSequence(
        sequence.value,
        response.items.map((item) => item.id),
      );
      navigationSession.value = {
        ...navigationSession.value,
        catalog: {
          ...catalog,
          nextPage:
            requestedPage < response.totalPages
              ? requestedPage + 1
              : undefined,
        },
      };
      return;
    }
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

async function ensurePreviousContinuation(): Promise<void> {
  const catalog = navigationSession.value.catalog;
  if (
    !import.meta.client ||
    previous.value ||
    !catalog?.previousPage ||
    previousContinuationPending.value
  )
    return;
  previousContinuationPending.value = true;
  try {
    const requestedPage = catalog.previousPage;
    const response = await $fetch<GalleryImagePage>("/api/gallery/images", {
      query: { ...catalog.request, page: requestedPage },
    });
    mergeCandidates(response.items);
    sequence.value = prependViewerSequence(
      sequence.value,
      response.items.map((item) => item.id),
    );
    navigationSession.value = {
      ...navigationSession.value,
      catalog: {
        ...catalog,
        previousPage: requestedPage > 1 ? requestedPage - 1 : undefined,
      },
    };
  } catch {
    // Keep the catalog boundary available for a later retry.
  } finally {
    previousContinuationPending.value = false;
  }
}

watch(
  [image, related],
  ([current, suggestions]) => {
    if (!current) return;
    mergeCandidates([current, ...suggestions]);
    sequence.value = moveViewerSequence(sequence.value, current.id);
    if (!navigationSession.value.catalog) {
      sequence.value = extendViewerSequence(
        sequence.value,
        suggestions.map((item) => item.id),
      );
    }
    void ensurePreviousContinuation();
    void ensureContinuation();
  },
  { immediate: true },
);

onMounted(() => {
  navigationReady.value = true;
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
  navigationSession.value = createViewerNavigationSession(imageId.value);
  void router.replace(closeTarget.value);
}

function navigate(targetId: string): void {
  sequence.value = moveViewerSequence(sequence.value, targetId);
  void router.replace(`/images/${encodeURIComponent(targetId)}`);
}

function retry(): void {
  void Promise.all([refresh(), refreshRelated()]);
}

onBeforeRouteLeave((to) => {
  if (/^\/images\/[^/]+/.test(to.path)) return;
  navigationSession.value = createViewerNavigationSession(imageId.value);
});
</script>

<template>
  <GalleryViewer
    :data-navigation-ready="navigationReady ? 'true' : 'false'"
    :image="image"
    :status="status"
    :failed="Boolean(error)"
    :previous="previous"
    :next="next"
    :related="related"
    :related-pending="relatedStatus === 'pending'"
    @close="closeViewer"
    @retry="retry"
    @navigate="navigate"
  />
</template>
