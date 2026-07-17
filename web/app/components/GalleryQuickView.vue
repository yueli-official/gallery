<script setup lang="ts">
import type {
  GalleryImage,
  GalleryImageCard,
  GalleryRelatedImage,
} from "~/types/gallery";

const props = withDefaults(
  defineProps<{ imageId: string; items?: GalleryImageCard[] }>(),
  { items: () => [] },
);
const emit = defineEmits<{ close: []; navigate: [imageId: string] }>();
const requestedId = computed(() => props.imageId);
const open = computed({
  get: () => Boolean(props.imageId),
  set: (value) => {
    if (!value) emit("close");
  },
});

const { data, error, status, refresh } = await useAsyncData(
  "gallery-catalog-quick-view",
  async () => {
    if (!requestedId.value) return null;
    return await $fetch<{ image: GalleryImage }>(
      `/api/gallery/images/${encodeURIComponent(requestedId.value)}`,
    );
  },
  { watch: [requestedId] },
);
const {
  data: relatedData,
  status: relatedStatus,
  refresh: refreshRelated,
} = await useAsyncData(
  "gallery-catalog-quick-view-related",
  async () => {
    if (!requestedId.value) return { items: [] as GalleryRelatedImage[] };
    return await $fetch<{ items: GalleryRelatedImage[] }>(
      `/api/gallery/images/${encodeURIComponent(requestedId.value)}/related`,
      { query: { size: 8 } },
    );
  },
  { watch: [requestedId] },
);
const image = computed(() =>
  data.value?.image.id === props.imageId ? data.value.image : undefined,
);
const sequence = ref(
  createViewerSequence(
    props.imageId,
    props.items.map((item) => item.id),
  ),
);
const continuation = ref<GalleryImageCard[]>([]);
const continuationPending = ref(false);
const pool = computed(() => {
  const values: GalleryImageCard[] = [
    ...props.items,
    ...(relatedData.value?.items || []),
    ...continuation.value,
  ];
  if (image.value) values.push(image.value);
  return new Map(values.map((item) => [item.id, item]));
});
const siblings = computed(() => {
  return {
    previous:
      sequence.value.index > 0
        ? pool.value.get(sequence.value.ids[sequence.value.index - 1]!)
        : undefined,
    next:
      sequence.value.index < sequence.value.ids.length - 1
        ? pool.value.get(sequence.value.ids[sequence.value.index + 1]!)
        : undefined,
  };
});

async function ensureContinuation(): Promise<void> {
  if (!import.meta.client) return;
  if (
    siblings.value.next ||
    continuationPending.value ||
    relatedStatus.value === "pending"
  )
    return;
  continuationPending.value = true;
  try {
    const seed =
      globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random()}`;
    const discovery = await $fetch<{ images: GalleryImageCard[] }>(
      "/api/gallery/discovery",
      { query: { seed } },
    );
    continuation.value = discovery.images;
    sequence.value = extendViewerSequence(
      sequence.value,
      discovery.images.map((item) => item.id),
    );
  } catch {
    // Keep the current related sequence available when discovery is transiently unavailable.
  } finally {
    continuationPending.value = false;
  }
}

watch(
  [requestedId, relatedData],
  ([current, suggestions]) => {
    if (!current) return;
    if (sequence.value.ids.length === 0) {
      sequence.value = createViewerSequence(
        current,
        props.items.map((item) => item.id),
      );
    } else {
      sequence.value = moveViewerSequence(sequence.value, current);
    }
    sequence.value = extendViewerSequence(
      sequence.value,
      (suggestions?.items || []).map((item) => item.id),
    );
    void ensureContinuation();
  },
  { immediate: true },
);

function navigate(imageId: string): void {
  sequence.value = moveViewerSequence(sequence.value, imageId);
  emit("navigate", imageId);
}

function retry(): void {
  void Promise.all([refresh(), refreshRelated()]);
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="image?.title || '图片快速查看'"
    description="在不离开当前目录位置的情况下查看图片信息。"
    :ui="{
      content: 'sm:max-w-[min(94vw,90rem)] overflow-hidden',
      header: 'sr-only',
      body: 'p-0 sm:p-0',
    }"
  >
    <template #body>
      <GalleryViewer
        :image="image"
        :status="status"
        :failed="Boolean(error)"
        mode="modal"
        :previous="siblings.previous"
        :next="siblings.next"
        :related="relatedData?.items || []"
        :related-pending="relatedStatus === 'pending'"
        @close="emit('close')"
        @retry="retry"
        @navigate="navigate"
      />
    </template>
  </UModal>
</template>
