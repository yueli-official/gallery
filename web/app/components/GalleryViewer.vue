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
    previous?: GalleryImageCard;
    next?: GalleryImageCard;
    related?: GalleryRelatedImage[];
    relatedPending?: boolean;
  }>(),
  {
    status: "idle",
    failed: false,
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
const zoom = ref(1);
const fullscreen = ref(false);
const moreDetails = ref<HTMLDetailsElement>();
const pointerStart = ref<{ x: number; y: number }>();
let qualifiedViewTimer: ReturnType<typeof setTimeout> | undefined;
const { favoritePending, toggleFavorite, shareImage, track } =
  useGalleryImageActions(image);

const canZoomOut = computed(() => zoom.value > 1);
const canZoomIn = computed(() => zoom.value < 2.5);
const viewerLabel = computed(() =>
  props.image ? `图片详情：${props.image.title}` : "图片详情",
);
const mediaStyle = computed(() => ({
  "--gallery-image-matte": props.image?.dominantColor || "var(--gallery-panel)",
}));
const filmstripItems = computed<GalleryImageCard[]>(() => {
  const candidates = [
    props.previous,
    props.image,
    props.next,
    ...props.related,
  ].filter((item): item is GalleryImageCard => Boolean(item));
  const unique = new Map<string, GalleryImageCard>();

  for (const item of candidates) {
    if (!unique.has(item.id)) unique.set(item.id, item);
  }

  return [...unique.values()].slice(0, 12);
});

function resetMedia(): void {
  zoom.value = 1;
}

function changeZoom(delta: number): void {
  zoom.value = Math.min(2.5, Math.max(1, zoom.value + delta));
}

async function toggleFullscreen(): Promise<void> {
  if (!import.meta.client) return;
  if (document.fullscreenElement) {
    await document.exitFullscreen();
  } else {
    await document.documentElement.requestFullscreen();
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

function closeMoreDetails(): void {
  if (moreDetails.value) moreDetails.value.open = false;
}

function onDocumentPointerDown(event: PointerEvent): void {
  const details = moreDetails.value;
  const target = event.target;
  if (
    !details?.open ||
    !(target instanceof Node) ||
    details.contains(target)
  )
    return;
  closeMoreDetails();
}

function onKeydown(event: KeyboardEvent): void {
  if (reportOpen.value) return;
  if (event.key === "Escape" && moreDetails.value?.open) {
    event.preventDefault();
    closeMoreDetails();
  } else if (event.key === "ArrowLeft" && props.previous) {
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
    closeMoreDetails();
    beginQualifiedView();
  },
);

onMounted(() => {
  window.addEventListener("keydown", onKeydown);
  document.addEventListener("pointerdown", onDocumentPointerDown);
  document.addEventListener("fullscreenchange", onFullscreenChange);
  onFullscreenChange();
  beginQualifiedView();
});
onBeforeUnmount(() => {
  window.removeEventListener("keydown", onKeydown);
  document.removeEventListener("pointerdown", onDocumentPointerDown);
  document.removeEventListener("fullscreenchange", onFullscreenChange);
  if (qualifiedViewTimer) clearTimeout(qualifiedViewTimer);
});
</script>

<template>
  <!--
    THESIS: 单张图片是一页永久挂展，不是被字段包围的后台详情；拒绝让标题和侧栏压过图片。
    OWN-WORLD: 冷纸白、矿物蓝、近乎满幅的图片展台、低位信息坞和无卡壳图片阵列。
    STORY: 先完整观看，再以低声量标题、统计和图标动作确认信息，随后沿相关图片继续探索。
    FIRST VIEWPORT: 桌面由全宽图片主导，标题、简介和图标动作压缩到图片下方；尺寸与治理动作按需展开。
    FORM: 全宽图片加低位信息坞，采用 .impeccable/mocks/viewer-comp-b-bottom-dock.png 的构图。
    FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md
  -->
  <article
    class="gallery-detail"
    :class="{ 'has-image': image }"
    :aria-label="viewerLabel"
  >
    <div v-if="status === 'pending'" class="gallery-detail-state">
      <USkeleton class="gallery-detail-state-media" />
      <div class="grid gap-3">
        <USkeleton class="h-8 w-2/3" />
        <USkeleton class="h-20 w-full" />
        <USkeleton class="h-11 w-full" />
      </div>
    </div>

    <div v-else-if="failed" class="gallery-detail-error">
      <UIcon name="i-tabler-photo-off" class="size-8" />
      <div>
        <h1>这张图片无法显示</h1>
        <p>它可能尚未发布、已隐藏或不存在。</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UButton label="重新加载" @click="emit('retry')" />
        <UButton
          color="neutral"
          variant="outline"
          label="返回图片目录"
          @click="emit('close')"
        />
      </div>
    </div>

    <template v-else-if="image">
      <nav class="gallery-detail-wayfinding" aria-label="图片详情导航">
        <button type="button" @click="emit('close')">
          <UIcon name="i-tabler-arrow-left" class="size-4" />
          返回图片目录
        </button>
      </nav>

      <div class="gallery-detail-stage">
        <section
          class="gallery-detail-media"
          :class="{ 'is-fullscreen': fullscreen }"
          :style="mediaStyle"
          @pointerdown="onPointerDown"
          @pointerup="onPointerUp"
        >
          <div class="gallery-detail-media-tools" aria-label="图片缩放工具">
            <UButton
              color="neutral"
              variant="soft"
              size="sm"
              icon="i-tabler-minus"
              aria-label="缩小图片"
              :disabled="!canZoomOut"
              @click="changeZoom(-0.5)"
            />
            <UButton
              color="neutral"
              variant="soft"
              size="sm"
              icon="i-tabler-focus-centered"
              aria-label="恢复适合窗口"
              :disabled="zoom === 1"
              @click="resetMedia"
            />
            <UButton
              color="neutral"
              variant="soft"
              size="sm"
              icon="i-tabler-plus"
              aria-label="放大图片"
              :disabled="!canZoomIn"
              @click="changeZoom(0.5)"
            />
            <UButton
              color="neutral"
              variant="soft"
              size="sm"
              :icon="fullscreen ? 'i-tabler-minimize' : 'i-tabler-maximize'"
              :aria-label="fullscreen ? '退出图册模式' : '进入图册模式'"
              @click="toggleFullscreen"
            />
          </div>

          <div class="gallery-detail-canvas">
            <img
              v-bind="galleryImageSources(image.assetId, 'display', true)"
              :alt="image.altText || image.title"
              :width="image.width"
              :height="image.height"
              draggable="false"
              :style="{ transform: `scale(${zoom})` }"
              @dblclick="zoom === 1 ? changeZoom(1) : resetMedia()"
            />
          </div>

          <button
            v-if="previous"
            type="button"
            class="gallery-detail-media-previous"
            aria-label="查看上一张图片"
            @click="navigate(previous)"
          >
            <UIcon name="i-tabler-chevron-left" class="size-5" />
          </button>
          <button
            v-if="next"
            type="button"
            class="gallery-detail-media-next"
            aria-label="查看下一张图片"
            @click="navigate(next)"
          >
            <UIcon name="i-tabler-chevron-right" class="size-5" />
          </button>
          <p class="gallery-detail-media-status" aria-live="polite">
            {{ Math.round(zoom * 100) }}%
          </p>

          <nav class="gallery-detail-filmstrip" aria-label="图册缩略图">
            <button
              v-for="item in filmstripItems"
              :key="item.id"
              type="button"
              :class="{ 'is-current': item.id === image.id }"
              :aria-current="item.id === image.id ? 'true' : undefined"
              :aria-label="
                item.id === image.id
                  ? `当前图片：${item.title}`
                  : `查看图片：${item.title}`
              "
              @click="navigate(item)"
            >
              <img
                v-bind="
                  galleryImageSources(item.assetId, 'thumbnail', false)
                "
                :alt="item.altText || item.title"
                :width="item.width"
                :height="item.height"
                loading="lazy"
              />
            </button>
          </nav>
        </section>

        <aside class="gallery-detail-info">
          <div class="gallery-detail-heading">
            <div class="gallery-detail-titleline">
              <h1>{{ image.title }}</h1>
              <NuxtLink
                v-if="image.primaryCategory"
                :to="{
                  path: '/images',
                  query: { categories: image.primaryCategorySlug },
                }"
                class="gallery-detail-category"
              >
                {{ image.primaryCategory }}
                <UIcon name="i-tabler-arrow-up-right" class="size-3.5" />
              </NuxtLink>
            </div>
            <p
              v-if="image.description"
              class="gallery-detail-description"
            >
              {{ image.description }}
            </p>
            <div class="gallery-detail-stats" aria-label="图片统计与来源">
              <span :aria-label="`浏览 ${compactMetric(image.metrics.views)} 次`">
                <UIcon name="i-tabler-eye" class="size-3.5" />
                <span aria-hidden="true">
                  {{ compactMetric(image.metrics.views) }}
                </span>
              </span>
              <span
                :aria-label="`收藏 ${compactMetric(image.metrics.favorites)} 次`"
              >
                <UIcon name="i-tabler-heart" class="size-3.5" />
                <span aria-hidden="true">
                  {{ compactMetric(image.metrics.favorites) }}
                </span>
              </span>
              <a
                v-if="image.sourceUrl"
                :href="image.sourceUrl"
                target="_blank"
                rel="noopener noreferrer nofollow"
                referrerpolicy="no-referrer"
              >
                原始来源
                <UIcon name="i-tabler-external-link" class="size-3" />
              </a>
              <span v-else>来源待补充</span>
            </div>
          </div>

          <div class="gallery-detail-primary-actions">
            <UButton
              :icon="
                image.favorited ? 'i-tabler-heart-filled' : 'i-tabler-heart'
              "
              color="neutral"
              variant="ghost"
              size="lg"
              class="gallery-detail-action"
              :class="{ 'is-saved': image.favorited }"
              :aria-label="image.favorited ? '取消收藏' : '收藏图片'"
              :title="image.favorited ? '取消收藏' : '收藏图片'"
              :aria-pressed="image.favorited"
              :loading="favoritePending"
              @click="toggleFavorite"
            />
            <UButton
              color="neutral"
              variant="ghost"
              size="lg"
              icon="i-tabler-share-3"
              class="gallery-detail-action"
              aria-label="分享图片"
              title="分享图片"
              @click="share"
            />
          </div>

          <div class="gallery-detail-meta-tail">
            <div v-if="image.tags.length" class="gallery-detail-tags">
              <NuxtLink
                v-for="tag in image.tags"
                :key="tag"
                :to="{ path: '/images', query: { tag } }"
              >
                #{{ tag }}
              </NuxtLink>
            </div>

            <details ref="moreDetails" class="gallery-detail-more">
              <summary aria-label="更多图片信息" title="更多图片信息">
                <UIcon name="i-tabler-dots" class="size-5" />
                <span class="sr-only">更多图片信息</span>
              </summary>
              <div class="gallery-detail-more-panel">
                <dl class="gallery-detail-more-facts">
                  <div>
                    <dt>图片尺寸</dt>
                    <dd>{{ image.width }} × {{ image.height }} px</dd>
                  </div>
                </dl>

                <div class="gallery-detail-secondary-actions">
                  <button
                    type="button"
                    @click="
                      reportKind = 'report';
                      reportOpen = true;
                      closeMoreDetails();
                    "
                  >
                    <UIcon name="i-tabler-flag" class="size-4" />
                    举报图片
                  </button>
                  <button
                    type="button"
                    @click="
                      reportKind = 'source_correction';
                      reportOpen = true;
                      closeMoreDetails();
                    "
                  >
                    <UIcon name="i-tabler-link-plus" class="size-4" />
                    补充来源
                  </button>
                </div>
              </div>
            </details>
          </div>
        </aside>
      </div>

      <section
        class="gallery-detail-related"
        aria-labelledby="related-heading"
      >
        <header>
          <div>
            <h2 id="related-heading">继续看相近的图片</h2>
            <p>根据分类、维度与标签找到的关联内容。</p>
          </div>
          <NuxtLink to="/images">
            浏览全部图片
            <UIcon name="i-tabler-arrow-right" class="size-4" />
          </NuxtLink>
        </header>
        <div v-if="relatedPending" class="gallery-detail-related-grid">
          <USkeleton
            v-for="index in 6"
            :key="index"
            class="aspect-[4/3] rounded-[0.9rem]"
          />
        </div>
        <div v-else-if="related.length" class="gallery-detail-related-grid">
          <button
            v-for="item in related.slice(0, 10)"
            :key="item.id"
            type="button"
            :aria-label="`查看相关图片：${item.title}`"
            @click="emit('navigate', item.id)"
          >
            <span
              class="gallery-detail-related-media"
              :style="{
                aspectRatio: imageAspect(item.width, item.height),
                backgroundColor: item.dominantColor || undefined,
              }"
            >
              <img
                v-bind="galleryImageSources(item.assetId, 'masonry', false)"
                :alt="item.altText || item.title"
                :width="item.width"
                :height="item.height"
              />
            </span>
            <span class="gallery-detail-related-copy">
              <strong>{{ item.title }}</strong>
              <small v-if="item.reasons[0]">{{
                item.reasons[0].label
              }}</small>
            </span>
          </button>
        </div>
        <p v-else class="gallery-detail-related-empty">
          暂时没有足够的关联图片，可以返回完整目录继续浏览。
        </p>
      </section>
    </template>

    <GalleryReportDialog
      v-if="image"
      v-model:open="reportOpen"
      :image-id="image.id"
      :kind="reportKind"
    />
  </article>
</template>
