<script setup lang="ts">
import type { GalleryImage } from "~/types/gallery";

const props = defineProps<{ imageId: string }>();
const emit = defineEmits<{ close: [] }>();
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
const image = computed(() =>
  data.value?.image.id === props.imageId ? data.value.image : undefined,
);
const reportOpen = ref(false);
const reportKind = ref<"report" | "source_correction">("report");
const { favoritePending, toggleFavorite, shareImage } =
  useGalleryImageActions(image);

function sharePreview(): void {
  if (!image.value) return;
  const target = new URL(
    `/images/${encodeURIComponent(image.value.id)}`,
    window.location.origin,
  ).href;
  void shareImage(target);
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="image?.title || '图片快速查看'"
    description="在不离开当前目录位置的情况下查看图片信息。"
    :ui="{ content: 'sm:max-w-6xl' }"
  >
    <template #body>
      <div class="gallery-quick-view">
        <div class="gallery-quick-view-media">
          <USkeleton
            v-if="status === 'pending'"
            class="h-[55vh] w-full rounded-lg"
          />
          <UAlert
            v-else-if="error"
            color="error"
            variant="subtle"
            icon="i-tabler-photo-off"
            title="图片暂时无法显示"
          >
            <template #actions>
              <UButton label="重试" @click="refresh()" />
            </template>
          </UAlert>
          <img
            v-else-if="image"
            v-bind="galleryImageSources(image.assetId, 'preview', true)"
            :alt="image.altText || image.title"
            :width="image.width"
            :height="image.height"
            class="max-h-[68vh] max-w-full object-contain"
          />
        </div>

        <aside v-if="image" class="gallery-quick-view-details">
          <div>
            <UBadge
              v-if="image.primaryCategory"
              color="neutral"
              variant="soft"
              :label="image.primaryCategory"
            />
            <h2>{{ image.title }}</h2>
            <p v-if="image.description">{{ image.description }}</p>
          </div>

          <div class="flex flex-wrap gap-2">
            <UBadge
              v-for="tag in image.tags.slice(0, 8)"
              :key="tag"
              color="neutral"
              variant="subtle"
              :label="`#${tag}`"
            />
          </div>

          <dl class="gallery-quick-view-facts">
            <div>
              <dt>尺寸</dt>
              <dd>{{ image.width }} × {{ image.height }}</dd>
            </div>
            <div>
              <dt>浏览</dt>
              <dd>{{ compactMetric(image.metrics.views) }}</dd>
            </div>
            <div>
              <dt>收藏</dt>
              <dd>{{ compactMetric(image.metrics.favorites) }}</dd>
            </div>
          </dl>

          <div class="mt-auto grid grid-cols-2 gap-2">
            <UButton
              color="neutral"
              variant="outline"
              icon="i-tabler-share-3"
              label="分享"
              @click="sharePreview"
            />
            <UButton
              color="neutral"
              variant="outline"
              :icon="
                image.favorited ? 'i-tabler-heart-filled' : 'i-tabler-heart'
              "
              :label="image.favorited ? '取消收藏' : '收藏'"
              :loading="favoritePending"
              @click="toggleFavorite"
            />
            <UButton
              class="col-span-2"
              :to="`/images/${encodeURIComponent(image.id)}`"
              trailing-icon="i-tabler-arrow-right"
              label="打开完整详情"
              block
            />
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-tabler-flag"
              label="举报"
              @click="
                reportKind = 'report';
                reportOpen = true;
              "
            />
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-tabler-link-plus"
              label="补充来源"
              @click="
                reportKind = 'source_correction';
                reportOpen = true;
              "
            />
          </div>
        </aside>
      </div>
    </template>
  </UModal>
  <GalleryReportDialog
    v-if="image"
    v-model:open="reportOpen"
    :image-id="image.id"
    :kind="reportKind"
  />
</template>
