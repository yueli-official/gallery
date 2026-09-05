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
    class="gallery-detail w-full px-[clamp(1rem,2.4vw,2rem)] pb-[clamp(4rem,8vw,7rem)] pt-[0.85rem] text-default max-[35rem]:pt-2"
    :class="{ 'has-image': image }"
    :aria-label="viewerLabel"
  >
    <div
      v-if="status === 'pending'"
      class="gallery-detail-state grid min-h-[70dvh] grid-cols-1 items-center gap-6 max-[35rem]:min-h-[60dvh]"
    >
      <USkeleton
        class="gallery-detail-state-media min-h-[64dvh] rounded-[0.9rem] max-[64rem]:min-h-[56dvh] max-[35rem]:-mx-4 max-[35rem]:min-h-[48dvh] max-[35rem]:rounded-none"
      />
      <div class="grid gap-3">
        <USkeleton class="h-8 w-2/3" />
        <USkeleton class="h-20 w-full" />
        <USkeleton class="h-11 w-full" />
      </div>
    </div>

    <div
      v-else-if="failed"
      class="gallery-detail-error mx-auto grid min-h-[65dvh] max-w-xl place-content-center justify-items-start gap-5"
    >
      <UIcon name="i-tabler-photo-off" class="size-8" />
      <div>
        <h1 class="text-[1.65rem] font-semibold text-highlighted">
          这张图片无法显示
        </h1>
        <p class="mt-1.5 text-muted">它可能尚未发布、已隐藏或不存在。</p>
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
      <nav
        class="gallery-detail-wayfinding mb-2 flex min-h-10 items-center gap-4 text-[0.8rem] text-muted"
        aria-label="图片详情导航"
      >
        <button
          type="button"
          class="inline-flex min-h-9 items-center gap-1.5 rounded-lg px-2 transition-colors hover:bg-muted hover:text-highlighted"
          @click="emit('close')"
        >
          <UIcon name="i-tabler-arrow-left" class="size-4" />
          返回图片目录
        </button>
      </nav>

      <div class="gallery-detail-stage grid grid-cols-1 items-start">
        <section
          class="gallery-detail-media relative grid min-h-[clamp(28rem,calc(100dvh-22rem),50rem)] min-w-0 overflow-hidden rounded-[0.9rem] max-[64rem]:min-h-[min(70dvh,46rem)] max-[35rem]:-mx-4 max-[35rem]:min-h-[54dvh] max-[35rem]:rounded-none"
          :class="{ 'is-fullscreen': fullscreen }"
          :style="mediaStyle"
          @pointerdown="onPointerDown"
          @pointerup="onPointerUp"
        >
          <div
            class="gallery-detail-media-tools absolute left-4 top-4 z-3 flex gap-1 max-[35rem]:left-2.5 max-[35rem]:top-2.5"
            aria-label="图片缩放工具"
          >
            <UButton
              color="neutral"
              variant="soft"
              size="sm"
              icon="i-tabler-minus"
              class="max-[35rem]:hidden"
              aria-label="缩小图片"
              :disabled="!canZoomOut"
              @click="changeZoom(-0.5)"
            />
            <UButton
              color="neutral"
              variant="soft"
              size="sm"
              icon="i-tabler-focus-centered"
              class="max-[35rem]:hidden"
              aria-label="恢复适合窗口"
              :disabled="zoom === 1"
              @click="resetMedia"
            />
            <UButton
              color="neutral"
              variant="soft"
              size="sm"
              icon="i-tabler-plus"
              class="max-[35rem]:hidden"
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

          <div
            class="gallery-detail-canvas grid size-full min-h-[inherit] place-items-center overflow-auto px-[clamp(1rem,2.5vw,2.5rem)] py-[clamp(2.75rem,4vw,4rem)] max-[64rem]:min-h-[min(70dvh,46rem)] max-[35rem]:min-h-[54dvh] max-[35rem]:px-4 max-[35rem]:pb-14 max-[35rem]:pt-18"
          >
            <img
              v-bind="galleryImageSources(image.assetId, 'display', true)"
              :alt="image.altText || image.title"
              :width="image.width"
              :height="image.height"
              draggable="false"
              class="block max-h-[min(calc(100dvh-25rem),48rem)] max-w-full select-none object-contain [filter:drop-shadow(0_0.9rem_1.8rem_rgb(13_24_44_/_0.14))] transition-transform duration-200 ease-out max-[35rem]:max-h-[66dvh]"
              :style="{ transform: `scale(${zoom})`, transformOrigin: 'center' }"
              @dblclick="zoom === 1 ? changeZoom(1) : resetMedia()"
            />
          </div>

          <button
            v-if="previous"
            type="button"
            class="gallery-detail-media-previous absolute left-4 top-1/2 z-3 grid size-11 -translate-y-1/2 place-items-center rounded-lg border border-transparent bg-[color-mix(in_srgb,var(--gallery-viewer-overlay)_76%,transparent)] text-highlighted opacity-65 transition-[background-color,opacity,transform] hover:scale-105 hover:bg-default hover:opacity-100 max-[35rem]:left-2.5"
            aria-label="查看上一张图片"
            @click="navigate(previous)"
          >
            <UIcon name="i-tabler-chevron-left" class="size-5" />
          </button>
          <button
            v-if="next"
            type="button"
            class="gallery-detail-media-next absolute right-4 top-1/2 z-3 grid size-11 -translate-y-1/2 place-items-center rounded-lg border border-transparent bg-[color-mix(in_srgb,var(--gallery-viewer-overlay)_76%,transparent)] text-highlighted opacity-65 transition-[background-color,opacity,transform] hover:scale-105 hover:bg-default hover:opacity-100 max-[35rem]:right-2.5"
            aria-label="查看下一张图片"
            @click="navigate(next)"
          >
            <UIcon name="i-tabler-chevron-right" class="size-5" />
          </button>
          <p
            class="gallery-detail-media-status absolute bottom-4 right-4 flex gap-1.5 rounded-md bg-[var(--gallery-viewer-overlay)] px-2 py-1 text-[0.7rem] tabular-nums text-muted max-[35rem]:bottom-2.5 max-[35rem]:right-2.5"
            aria-live="polite"
          >
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

        <aside
          class="gallery-detail-info grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-start gap-x-4 gap-y-3 pt-4 max-[35rem]:gap-x-2 max-[35rem]:gap-y-2.5 max-[35rem]:pt-3.5"
        >
          <div class="gallery-detail-heading">
            <div
              class="gallery-detail-titleline flex min-w-0 items-baseline gap-3 max-[35rem]:items-start max-[35rem]:flex-col max-[35rem]:gap-1.5"
            >
              <h1
                class="max-w-[48ch] min-w-0 text-[clamp(1.05rem,1.35vw,1.25rem)] font-bold leading-[1.3] tracking-[-0.015em] text-highlighted [overflow-wrap:anywhere] [text-wrap:pretty] max-[35rem]:max-w-none max-[35rem]:text-[clamp(1.05rem,4.8vw,1.25rem)]"
              >
                {{ image.title }}
              </h1>
              <NuxtLink
                v-if="image.primaryCategory"
                :to="{
                  path: '/images',
                  query: { categories: image.primaryCategorySlug },
                }"
                class="gallery-detail-category inline-flex shrink-0 items-center gap-1 whitespace-nowrap text-xs font-bold text-primary"
              >
                {{ image.primaryCategory }}
                <UIcon name="i-tabler-arrow-up-right" class="size-3.5" />
              </NuxtLink>
            </div>
            <p
              v-if="image.description"
              class="gallery-detail-description mt-1 max-w-[52rem] whitespace-pre-line text-[0.78rem] leading-[1.55] text-muted [overflow-wrap:anywhere]"
            >
              {{ image.description }}
            </p>
            <div
              class="gallery-detail-stats mt-1.5 flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1.5 text-[0.72rem] tabular-nums text-dimmed [&>a]:inline-flex [&>a]:items-center [&>a]:gap-1 [&>a]:text-inherit [&>a]:underline [&>a]:decoration-transparent [&>a]:underline-offset-[0.2rem] [&>a]:transition-colors [&>a:hover]:text-primary [&>span]:inline-flex [&>span]:items-center [&>span]:gap-1"
              aria-label="图片统计与来源"
            >
              <span>
                <UIcon name="i-tabler-eye" class="size-3.5" aria-hidden="true" />
                <span class="sr-only">浏览</span>
                <span>
                  {{ compactMetric(image.metrics.views) }}
                </span>
                <span class="sr-only">次</span>
              </span>
              <span>
                <UIcon name="i-tabler-heart" class="size-3.5" aria-hidden="true" />
                <span class="sr-only">收藏</span>
                <span>
                  {{ compactMetric(image.metrics.favorites) }}
                </span>
                <span class="sr-only">次</span>
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

          <div
            class="gallery-detail-primary-actions flex items-center justify-end gap-0.5"
          >
            <UButton
              :icon="
                image.favorited ? 'i-tabler-heart-filled' : 'i-tabler-heart'
              "
              color="neutral"
              variant="ghost"
              size="lg"
              class="gallery-detail-action w-10 justify-center"
              :class="{ 'text-primary': image.favorited }"
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
              class="gallery-detail-action w-10 justify-center"
              aria-label="分享图片"
              title="分享图片"
              @click="share"
            />
          </div>

          <div
            class="gallery-detail-meta-tail col-span-full flex min-w-0 items-center justify-between gap-x-6 gap-y-2 max-[35rem]:gap-1.5"
          >
            <div
              v-if="image.tags.length"
              class="gallery-detail-tags flex flex-wrap gap-x-3 gap-y-1.5"
            >
              <NuxtLink
                v-for="tag in image.tags"
                :key="tag"
                :to="{ path: '/images', query: { tag } }"
                class="text-xs text-dimmed underline decoration-transparent underline-offset-4 transition-colors hover:text-primary"
              >
                #{{ tag }}
              </NuxtLink>
            </div>

            <details ref="moreDetails" class="gallery-detail-more relative ml-auto">
              <summary
                class="grid size-10 cursor-pointer list-none place-items-center rounded-lg text-muted transition-colors hover:bg-muted hover:text-highlighted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary [&::-webkit-details-marker]:hidden"
                aria-label="更多图片信息"
                title="更多图片信息"
              >
                <UIcon name="i-tabler-dots" class="size-5" />
                <span class="sr-only">更多图片信息</span>
              </summary>
              <div
                class="gallery-detail-more-panel absolute right-0 top-[calc(100%+0.45rem)] z-8 w-[min(24rem,calc(100vw-2rem))] rounded-xl bg-default p-4 shadow-[0_1rem_3rem_rgb(13_24_44_/_0.18)] max-[35rem]:fixed max-[35rem]:bottom-4 max-[35rem]:left-4 max-[35rem]:right-4 max-[35rem]:top-auto max-[35rem]:w-auto"
              >
                <dl class="gallery-detail-more-facts m-0">
                  <div
                    class="flex items-baseline justify-between gap-4 text-[0.72rem] text-dimmed"
                  >
                    <dt>图片尺寸</dt>
                    <dd class="tabular-nums text-muted">
                      {{ image.width }} × {{ image.height }} px
                    </dd>
                  </div>
                </dl>

                <div
                  class="gallery-detail-secondary-actions mt-2 flex flex-wrap gap-x-4 gap-y-2 [&_button]:inline-flex [&_button]:min-h-10 [&_button]:items-center [&_button]:gap-1.5 [&_button]:text-[0.78rem] [&_button]:text-muted [&_button:hover]:text-highlighted"
                >
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

      <GalleryCommentSection v-show="!fullscreen" :image-id="image.id" />

      <section
        class="gallery-detail-related mt-20 lg:mt-28"
        aria-labelledby="related-heading"
      >
        <header class="mb-6 flex items-start justify-between gap-6 sm:items-end">
          <div>
            <h2
              id="related-heading"
              class="font-display text-[clamp(1.45rem,2.4vw,2rem)] font-semibold leading-[1.1] tracking-[-0.025em] text-highlighted"
            >
              继续看相近的图片
            </h2>
            <p class="mt-1.5 max-w-72 text-[0.82rem] text-muted sm:max-w-none">
              根据分类、维度与标签找到的关联内容。
            </p>
          </div>
          <NuxtLink
            to="/images"
            class="inline-flex min-h-10 w-10 shrink-0 items-center justify-center gap-1.5 overflow-hidden text-[0.8rem] text-transparent sm:w-auto sm:justify-start sm:overflow-visible sm:text-muted"
          >
            浏览全部图片
            <UIcon name="i-tabler-arrow-right" class="size-4 shrink-0 text-muted" />
          </NuxtLink>
        </header>
        <div
          v-if="relatedPending"
          class="gallery-detail-related-grid grid grid-cols-2 items-start gap-x-3.5 gap-y-6 sm:grid-cols-3 lg:grid-cols-5"
        >
          <USkeleton
            v-for="index in 6"
            :key="index"
            class="aspect-[4/3] rounded-[0.9rem]"
          />
        </div>
        <div
          v-else-if="related.length"
          class="gallery-detail-related-grid grid grid-cols-2 items-start gap-x-3.5 gap-y-6 sm:grid-cols-3 lg:grid-cols-5"
        >
          <button
            v-for="item in related.slice(0, 10)"
            :key="item.id"
            type="button"
            class="group min-w-0 rounded-[0.9rem] text-left"
            :aria-label="`查看相关图片：${item.title}`"
            @click="emit('navigate', item.id)"
          >
            <span
              class="gallery-detail-related-media block overflow-hidden rounded-[0.9rem] bg-muted"
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
                class="size-full object-cover transition-[filter,transform] duration-200 ease-out group-hover:scale-[1.02] group-hover:brightness-[0.96] group-hover:saturate-[0.94] group-focus-visible:scale-[1.02]"
              />
            </span>
            <span class="gallery-detail-related-copy grid min-w-0 gap-1 px-0.5 pt-2.5">
              <strong class="truncate text-[0.8rem] font-semibold text-highlighted">{{ item.title }}</strong>
              <small v-if="item.reasons[0]" class="truncate text-xs text-dimmed">{{
                item.reasons[0].label
              }}</small>
            </span>
          </button>
        </div>
        <p v-else class="gallery-detail-related-empty py-8 text-xs text-dimmed">
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
