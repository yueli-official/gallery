<script setup lang="ts">
import type { GalleryImage } from "~/types/gallery";

const route = useRoute("/images/[imageId]");
const router = useRouter();
const { data, error, status, refresh } = await useFetch<{
  image: GalleryImage;
}>(() => `/api/gallery/images/${route.params.imageId}`);
if (import.meta.server && error.value)
  setResponseStatus(error.value.statusCode === 404 ? 404 : 502);
const image = computed(() => data.value?.image);
const { favoritePending, toggleFavorite, shareImage, track } =
  useGalleryImageActions(image);
const reportOpen = ref(false);
const reportKind = ref<"report" | "source_correction">("report");
let qualifiedViewTimer: ReturnType<typeof setTimeout> | undefined;

useSeoMeta({
  title: () => image.value?.title || "图片详情",
  description: () =>
    image.value?.description || image.value?.altText || "查看公开图片详情。",
  ogImage: () =>
    image.value ? galleryRendition(image.value.assetId, "og") : undefined,
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
              description: image.value.description || image.value.altText,
              contentUrl: galleryRendition(image.value.assetId, "display"),
              thumbnailUrl: galleryRendition(image.value.assetId, "thumbnail"),
              width: image.value.width,
              height: image.value.height,
            }).replaceAll("<", "\\u003c"),
          },
        ],
      }
    : {},
);

function closeViewer() {
  if (window.history.length > 1) router.back();
  else void router.push("/images");
}

onMounted(() => {
  qualifiedViewTimer = setTimeout(() => {
    void track("qualified_view");
  }, 2200);
});
onBeforeUnmount(() => {
  if (qualifiedViewTimer) clearTimeout(qualifiedViewTimer);
});
</script>

<template>
  <div
    class="gallery-viewer min-h-[calc(100dvh-4rem)] text-white md:grid md:grid-cols-[minmax(0,1fr)_24rem]"
  >
    <section
      class="relative grid min-h-[62dvh] place-items-center overflow-hidden md:min-h-[calc(100dvh-4rem)]"
    >
      <div
        class="absolute inset-x-0 top-0 z-10 flex items-center justify-between p-3 sm:p-5"
      >
        <UButton
          color="neutral"
          variant="soft"
          icon="i-tabler-x"
          aria-label="关闭图片查看器"
          @click="closeViewer"
        />
        <div class="flex gap-1">
          <UButton
            color="neutral"
            variant="soft"
            icon="i-tabler-share-3"
            aria-label="分享图片"
            @click="shareImage()"
          />
          <UButton
            color="neutral"
            variant="soft"
            :icon="
              image?.favorited ? 'i-tabler-heart-filled' : 'i-tabler-heart'
            "
            :aria-label="image?.favorited ? '取消收藏' : '收藏图片'"
            :loading="favoritePending"
            @click="toggleFavorite"
          />
        </div>
      </div>

      <USkeleton
        v-if="status === 'pending'"
        class="h-[70vh] w-[70vw] rounded-lg bg-white/10"
      />
      <UAlert
        v-else-if="error"
        color="error"
        variant="subtle"
        icon="i-tabler-photo-off"
        title="这张图片无法显示"
        description="它可能尚未发布、已隐藏或不存在。"
        ><template #actions
          ><UButton label="重试" @click="refresh()" /></template
      ></UAlert>
      <img
        v-else-if="image"
        :src="galleryRendition(image.assetId, 'display')"
        :alt="image.altText || image.title"
        :width="image.width"
        :height="image.height"
        fetchpriority="high"
        class="gallery-viewer-image max-h-[100dvh] max-w-full select-none object-contain p-0 md:p-8"
      />
    </section>

    <aside
      v-if="image"
      class="gallery-viewer-panel relative z-20 -mt-4 rounded-t-2xl px-5 pb-[calc(1.5rem+env(safe-area-inset-bottom))] pt-5 md:mt-0 md:overflow-y-auto md:rounded-none md:border-l md:px-7 md:py-9"
    >
      <div
        class="mx-auto mb-4 h-1 w-10 rounded-full bg-white/20 md:hidden"
        aria-hidden="true"
      />
      <UBadge
        v-if="image.primaryCategory"
        color="neutral"
        variant="soft"
        :label="image.primaryCategory"
      />
      <h1 class="mt-3 text-2xl font-semibold tracking-tight text-white">
        {{ image.title }}
      </h1>
      <p
        v-if="image.description"
        class="mt-4 whitespace-pre-line text-sm leading-7 text-white/68"
      >
        {{ image.description }}
      </p>

      <div class="mt-6 flex flex-wrap gap-2">
        <UBadge
          v-for="tag in image.tags"
          :key="tag"
          color="neutral"
          variant="subtle"
          :label="`#${tag}`"
        />
      </div>

      <dl class="gallery-viewer-facts mt-7 space-y-3 border-t pt-5 text-sm">
        <div class="flex justify-between gap-4">
          <dt class="text-white/52">浏览</dt>
          <dd class="tabular-nums">{{ compactMetric(image.metrics.views) }}</dd>
        </div>
        <div class="flex justify-between gap-4">
          <dt class="text-white/52">收藏</dt>
          <dd class="tabular-nums">
            {{ compactMetric(image.metrics.favorites) }}
          </dd>
        </div>
        <div class="flex justify-between gap-4">
          <dt class="text-white/52">尺寸</dt>
          <dd class="tabular-nums">{{ image.width }} × {{ image.height }}</dd>
        </div>
        <div class="flex items-start justify-between gap-4">
          <dt class="text-white/52">来源</dt>
          <dd class="max-w-[12rem] text-right">
            <a
              v-if="image.sourceUrl"
              :href="image.sourceUrl"
              target="_blank"
              rel="noopener noreferrer nofollow"
              referrerpolicy="no-referrer"
              class="break-all text-primary hover:underline"
              >打开来源地址</a
            >
            <span v-else class="text-white/52">来源待补充</span>
          </dd>
        </div>
      </dl>

      <div class="mt-7 grid grid-cols-2 gap-2">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-tabler-flag"
          label="举报"
          @click="
            reportKind = 'report';
            reportOpen = true;
          "
        />
        <UButton
          color="neutral"
          variant="outline"
          icon="i-tabler-link-plus"
          label="补充来源"
          @click="
            reportKind = 'source_correction';
            reportOpen = true;
          "
        />
      </div>
    </aside>

    <GalleryReportDialog
      v-if="image"
      v-model:open="reportOpen"
      :image-id="image.id"
      :kind="reportKind"
    />
  </div>
</template>
