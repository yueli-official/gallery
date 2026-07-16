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
const siblings = computed(() => {
  const index = props.items.findIndex((item) => item.id === props.imageId);
  return {
    previous: index > 0 ? props.items[index - 1] : undefined,
    next:
      index >= 0 && index < props.items.length - 1
        ? props.items[index + 1]
        : undefined,
  };
});

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
        @navigate="emit('navigate', $event)"
      />
    </template>
  </UModal>
</template>
