<script setup lang="ts">
import type {
  GalleryImage,
  GalleryImageCard,
  GalleryRelatedImage,
} from "~/types/gallery";

const props = withDefaults(
  defineProps<{
    image?: GalleryImage;
    status?: "idle" | "pending" | "success" | "error";
    failed?: boolean;
    mode?: "modal" | "page";
    previous?: GalleryImageCard;
    next?: GalleryImageCard;
    related?: GalleryRelatedImage[];
    relatedPending?: boolean;
  }>(),
  {
    status: "idle",
    failed: false,
    mode: "page",
    related: () => [],
    relatedPending: false,
  },
);
const emit = defineEmits<{
  close: [];
  retry: [];
  navigate: [imageId: string];
}>();

const image = computed(() => props.image);
const reportOpen = ref(false);
const reportKind = ref<"report" | "source_correction">("report");
const viewer = useTemplateRef<HTMLElement>("viewer");
const zoom = ref(1);
const fullscreen = ref(false);
const pointerStart = ref<{ x: number; y: number }>();
let qualifiedViewTimer: ReturnType<typeof setTimeout> | undefined;
const { favoritePending, toggleFavorite, shareImage, track } =
  useGalleryImageActions(image);

const canZoomOut = computed(() => zoom.value > 1);
const canZoomIn = computed(() => zoom.value < 2.5);
const viewerLabel = computed(() =>
  props.image ? `正在查看：${props.image.title}` : "图片查看器",
);

function resetMedia(): void {
  zoom.value = 1;
}

function changeZoom(delta: number): void {
  zoom.value = Math.min(2.5, Math.max(1, zoom.value + delta));
}

async function toggleFullscreen(): Promise<void> {
  if (!import.meta.client || !viewer.value) return;
  if (document.fullscreenElement) {
    await document.exitFullscreen();
  } else {
    await viewer.value.requestFullscreen();
  }
}

function onFullscreenChange(): void {
  fullscreen.value = Boolean(document.fullscreenElement);
}

function navigate(target?: GalleryImageCard): void {
  if (!target) return;
  resetMedia();
  emit("navigate", target.id);
}

function share(): void {
  if (!props.image || !import.meta.client) return;
  const url = new URL(
    `/images/${encodeURIComponent(props.image.id)}`,
    window.location.origin,
  ).href;
  void shareImage(url);
}

function onKeydown(event: KeyboardEvent): void {
  if (reportOpen.value) return;
  if (event.key === "ArrowLeft" && props.previous) {
    event.preventDefault();
    navigate(props.previous);
  } else if (event.key === "ArrowRight" && props.next) {
    event.preventDefault();
    navigate(props.next);
  } else if ((event.key === "+" || event.key === "=") && canZoomIn.value) {
    event.preventDefault();
    changeZoom(0.5);
  } else if (event.key === "-" && canZoomOut.value) {
    event.preventDefault();
    changeZoom(-0.5);
  } else if (event.key === "0") {
    event.preventDefault();
    resetMedia();
  } else if (event.key === "Escape" && props.mode === "modal") {
    emit("close");
  }
}

function onPointerDown(event: PointerEvent): void {
  if (zoom.value !== 1 || event.pointerType === "mouse") return;
  pointerStart.value = { x: event.clientX, y: event.clientY };
}

function onPointerUp(event: PointerEvent): void {
  const start = pointerStart.value;
  pointerStart.value = undefined;
  if (!start || zoom.value !== 1) return;
  const deltaX = event.clientX - start.x;
  const deltaY = event.clientY - start.y;
  if (Math.abs(deltaX) < 70 || Math.abs(deltaX) < Math.abs(deltaY) * 1.25)
    return;
  navigate(deltaX > 0 ? props.previous : props.next);
}

function beginQualifiedView(): void {
  if (qualifiedViewTimer) clearTimeout(qualifiedViewTimer);
  if (!props.image) return;
  qualifiedViewTimer = setTimeout(() => void track("qualified_view"), 2200);
}

watch(
  () => props.image?.id,
  () => {
    resetMedia();
    beginQualifiedView();
  },
);

onMounted(() => {
  window.addEventListener("keydown", onKeydown);
  document.addEventListener("fullscreenchange", onFullscreenChange);
  beginQualifiedView();
});
onBeforeUnmount(() => {
  window.removeEventListener("keydown", onKeydown);
  document.removeEventListener("fullscreenchange", onFullscreenChange);
  if (qualifiedViewTimer) clearTimeout(qualifiedViewTimer);
});
</script>

<template>
  <article
    ref="viewer"
    class="gallery-viewer"
    :class="[
      `gallery-viewer-${mode}`,
      { 'has-image': image, 'is-fullscreen': fullscreen },
    ]"
    :aria-label="viewerLabel"
  >
    <section
      class="gallery-viewer-media"
      @pointerdown="onPointerDown"
      @pointerup="onPointerUp"
    >
      <div class="gallery-viewer-toolbar">
        <UButton
          color="neutral"
          variant="soft"
          icon="i-tabler-x"
          :aria-label="mode === 'modal' ? '关闭快速查看' : '返回图片目录'"
          @click="emit('close')"
        />
        <div class="flex items-center gap-1">
          <UButton
            color="neutral"
            variant="soft"
            icon="i-tabler-minus"
            aria-label="缩小图片"
            :disabled="!canZoomOut"
            @click="changeZoom(-0.5)"
          />
          <UButton
            color="neutral"
            variant="soft"
            icon="i-tabler-arrows-maximize"
            aria-label="恢复适合窗口"
            :disabled="zoom === 1"
            @click="resetMedia"
          />
          <UButton
            color="neutral"
            variant="soft"
            icon="i-tabler-plus"
            aria-label="放大图片"
            :disabled="!canZoomIn"
            @click="changeZoom(0.5)"
          />
          <UButton
            color="neutral"
            variant="soft"
            :icon="fullscreen ? 'i-tabler-minimize' : 'i-tabler-maximize'"
            :aria-label="fullscreen ? '退出全屏' : '进入全屏'"
            @click="toggleFullscreen"
          />
        </div>
      </div>

      <USkeleton
        v-if="status === 'pending'"
        class="h-[62vh] w-[70%] rounded-xl bg-white/10"
      />
      <UAlert
        v-else-if="failed"
        class="m-6 max-w-md"
        color="error"
        variant="subtle"
        icon="i-tabler-photo-off"
        title="这张图片无法显示"
        description="它可能尚未发布、已隐藏或不存在。"
      >
        <template #actions>
          <UButton label="重试" @click="emit('retry')" />
        </template>
      </UAlert>
      <div v-else-if="image" class="gallery-viewer-canvas">
        <img
          v-bind="galleryImageSources(image.assetId, 'preview', true)"
          :alt="image.altText || image.title"
          :width="image.width"
          :height="image.height"
          draggable="false"
          class="gallery-viewer-image"
          :style="{ transform: `scale(${zoom})` }"
          @dblclick="zoom === 1 ? changeZoom(1) : resetMedia()"
        />
      </div>

      <UButton
        v-if="previous"
        class="gallery-viewer-previous"
        color="neutral"
        variant="soft"
        icon="i-tabler-chevron-left"
        aria-label="查看上一张图片"
        @click="navigate(previous)"
      />
      <UButton
        v-if="next"
        class="gallery-viewer-next"
        color="neutral"
        variant="soft"
        icon="i-tabler-chevron-right"
        aria-label="查看下一张图片"
        @click="navigate(next)"
      />
      <p v-if="image" class="gallery-viewer-zoom" aria-live="polite">
        {{ Math.round(zoom * 100) }}%
      </p>
    </section>

    <aside v-if="image" class="gallery-viewer-panel">
      <div class="gallery-viewer-summary">
        <UBadge
          v-if="image.primaryCategory"
          color="neutral"
          variant="soft"
          :label="image.primaryCategory"
        />
        <h1>{{ image.title }}</h1>
        <p v-if="image.description">{{ image.description }}</p>
      </div>

      <div v-if="image.tags.length" class="flex flex-wrap gap-2">
        <UBadge
          v-for="tag in image.tags"
          :key="tag"
          color="neutral"
          variant="subtle"
          :label="`#${tag}`"
        />
      </div>

      <dl class="gallery-viewer-facts">
        <div>
          <dt>浏览</dt>
          <dd>{{ compactMetric(image.metrics.views) }}</dd>
        </div>
        <div>
          <dt>收藏</dt>
          <dd>{{ compactMetric(image.metrics.favorites) }}</dd>
        </div>
        <div>
          <dt>尺寸</dt>
          <dd>{{ image.width }} × {{ image.height }}</dd>
        </div>
        <div>
          <dt>来源</dt>
          <dd>
            <a
              v-if="image.sourceUrl"
              :href="image.sourceUrl"
              target="_blank"
              rel="noopener noreferrer nofollow"
              referrerpolicy="no-referrer"
              >打开来源地址</a
            >
            <span v-else>来源待补充</span>
          </dd>
        </div>
      </dl>

      <div class="gallery-viewer-actions">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-tabler-share-3"
          label="分享"
          @click="share"
        />
        <UButton
          color="neutral"
          variant="outline"
          :icon="image.favorited ? 'i-tabler-heart-filled' : 'i-tabler-heart'"
          :label="image.favorited ? '取消收藏' : '收藏'"
          :loading="favoritePending"
          @click="toggleFavorite"
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

      <UButton
        v-if="mode === 'modal'"
        :to="`/images/${encodeURIComponent(image.id)}`"
        trailing-icon="i-tabler-arrow-right"
        label="打开完整详情"
        block
      />

      <section class="gallery-viewer-related" aria-labelledby="related-heading">
        <div class="flex items-end justify-between gap-4">
          <div>
            <h2 id="related-heading">继续探索</h2>
            <p>基于分类、属性和标签推荐</p>
          </div>
          <UIcon name="i-tabler-route" class="size-5 text-white/45" />
        </div>
        <div v-if="relatedPending" class="grid grid-cols-2 gap-2">
          <USkeleton
            v-for="index in 4"
            :key="index"
            class="aspect-[4/3] rounded-lg bg-white/10"
          />
        </div>
        <div v-else-if="related.length" class="gallery-viewer-related-grid">
          <button
            v-for="item in related.slice(0, 6)"
            :key="item.id"
            type="button"
            :aria-label="`查看相关图片：${item.title}`"
            @click="emit('navigate', item.id)"
          >
            <img
              v-bind="galleryImageSources(item.assetId, 'thumbnail', false)"
              :alt="item.altText || item.title"
              :width="item.width"
              :height="item.height"
            />
            <span>{{ item.title }}</span>
            <small v-if="item.reasons[0]">{{ item.reasons[0].label }}</small>
          </button>
        </div>
        <p v-else class="gallery-viewer-related-empty">暂无足够的关联图片</p>
      </section>
    </aside>

    <GalleryReportDialog
      v-if="image"
      v-model:open="reportOpen"
      :image-id="image.id"
      :kind="reportKind"
    />
  </article>
</template>
