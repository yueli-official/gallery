<script setup lang="ts">
import { createPlatformNotifier } from "@platform/ui/feedback";
import type { GalleryImage } from "~/types/gallery";

const route = useRoute("/images/[imageId]");
const router = useRouter();
const { loggedIn, login } = useAuth();
const { call } = useApi();
const toast = createPlatformNotifier(useToast());
const { data, error, status, refresh } = await useFetch<{ image: GalleryImage }>(() => `/api/gallery/images/${route.params.imageId}`);
const image = computed(() => data.value?.image);
const favoritePending = ref(false);
const reportOpen = ref(false);
const reportKind = ref<"report" | "source_correction">("report");
const reportReason = ref("");
const reportDescription = ref("");
const proposedSourceUrl = ref("");
const reportPending = ref(false);
let qualifiedViewTimer: ReturnType<typeof setTimeout> | undefined;

useSeoMeta({
  title: () => image.value?.title || "图片详情",
  description: () => image.value?.description || image.value?.altText || "查看公开图片详情。",
  ogImage: () => image.value ? galleryRendition(image.value.assetId, "og") : undefined,
});
useHead(() => image.value ? {
  script: [{ type: "application/ld+json", innerHTML: JSON.stringify({
    "@context": "https://schema.org", "@type": "ImageObject", name: image.value.title,
    description: image.value.description || image.value.altText,
    contentUrl: galleryRendition(image.value.assetId, "display"),
    thumbnailUrl: galleryRendition(image.value.assetId, "thumbnail"),
    width: image.value.width, height: image.value.height,
  }).replaceAll("<", "\\u003c") }],
} : {});

async function toggleFavorite() {
  if (!image.value || favoritePending.value) return;
  if (!loggedIn.value) {
    await login();
    return;
  }
  favoritePending.value = true;
  try {
    const next = !image.value.favorited;
    await call(`/api/v1/gallery/me/favorites/${encodeURIComponent(image.value.id)}`, { method: next ? "PUT" : "DELETE", body: next ? { version: 0 } : undefined });
    image.value.favorited = next;
    image.value.metrics.favorites = Math.max(0, image.value.metrics.favorites + (next ? 1 : -1));
  } catch (reason: any) {
    toast.add({ title: "收藏没有保存", description: reason?.data?.message || reason?.message || "请稍后重试", color: "error" });
  } finally {
    favoritePending.value = false;
  }
}

async function shareImage() {
  const url = window.location.href;
  if (navigator.share) await navigator.share({ title: image.value?.title, url });
  else {
    await navigator.clipboard.writeText(url);
    toast.add({ title: "链接已复制", color: "success" });
  }
  if (image.value) void track("share");
}

async function track(type: "qualified_view" | "share") {
  if (!image.value) return;
  try {
    await $fetch(`/api/gallery/images/${encodeURIComponent(image.value.id)}/events`, { method: "POST", body: { type, sessionKey: galleryMetricSession() } });
  } catch { /* Metrics never block viewing. */ }
}

async function submitCase() {
  if (!image.value || reportPending.value) return;
  reportPending.value = true;
  try {
    await $fetch(`/api/gallery/images/${encodeURIComponent(image.value.id)}/cases`, {
      method: "POST",
      body: { kind: reportKind.value, reason: reportReason.value, description: reportDescription.value, proposedSourceUrl: proposedSourceUrl.value },
    });
    reportOpen.value = false;
    reportReason.value = "";
    reportDescription.value = "";
    proposedSourceUrl.value = "";
    toast.add({ title: reportKind.value === "report" ? "举报已提交" : "来源建议已提交", description: "运营人员会独立复核，不会按次数自动下架。", color: "success" });
  } catch (reason: any) {
    toast.add({ title: "提交失败", description: reason?.data?.message || reason?.message || "请稍后重试", color: "error" });
  } finally {
    reportPending.value = false;
  }
}

function closeViewer() {
  if (window.history.length > 1) router.back();
  else void router.push("/images");
}

onMounted(() => { qualifiedViewTimer = setTimeout(() => { void track("qualified_view"); }, 2200); });
onBeforeUnmount(() => { if (qualifiedViewTimer) clearTimeout(qualifiedViewTimer); });
</script>

<template>
  <div class="min-h-[calc(100dvh-4rem)] bg-stone-950 text-white md:grid md:grid-cols-[minmax(0,1fr)_23rem]">
    <section class="relative grid min-h-[62dvh] place-items-center overflow-hidden md:min-h-[calc(100dvh-4rem)]">
      <div class="absolute inset-x-0 top-0 z-10 flex items-center justify-between p-3 sm:p-5">
        <UButton color="neutral" variant="solid" icon="i-tabler-x" aria-label="关闭图片查看器" @click="closeViewer" />
        <div class="flex gap-1">
          <UButton color="neutral" variant="solid" icon="i-tabler-share-3" aria-label="分享图片" @click="shareImage" />
          <UButton color="neutral" variant="solid" :icon="image?.favorited ? 'i-tabler-heart-filled' : 'i-tabler-heart'" :aria-label="image?.favorited ? '取消收藏' : '收藏图片'" :loading="favoritePending" @click="toggleFavorite" />
        </div>
      </div>

      <USkeleton v-if="status === 'pending'" class="h-[70vh] w-[70vw] rounded-lg bg-white/10" />
      <UAlert v-else-if="error" color="error" variant="subtle" icon="i-tabler-photo-off" title="这张图片无法显示" description="它可能尚未发布、已隐藏或不存在。"><template #actions><UButton label="重试" @click="() => refresh()" /></template></UAlert>
      <img
        v-else-if="image"
        :src="galleryRendition(image.assetId, 'display')"
        :alt="image.altText || image.title"
        :width="image.width"
        :height="image.height"
        fetchpriority="high"
        class="max-h-[100dvh] max-w-full select-none object-contain p-0 md:p-8"
      />
    </section>

    <aside v-if="image" class="relative z-20 -mt-4 rounded-t-2xl bg-default px-5 pb-[calc(1.5rem+env(safe-area-inset-bottom))] pt-5 text-default md:mt-0 md:overflow-y-auto md:rounded-none md:border-l md:border-default md:px-6 md:py-8">
      <div class="mx-auto mb-4 h-1 w-10 rounded-full bg-accented md:hidden" aria-hidden="true" />
      <UBadge v-if="image.topic" color="primary" variant="soft" :label="image.topic" />
      <h1 class="mt-3 text-2xl font-semibold tracking-tight text-highlighted">{{ image.title }}</h1>
      <p v-if="image.description" class="mt-4 whitespace-pre-line text-sm leading-7 text-toned">{{ image.description }}</p>

      <div class="mt-6 flex flex-wrap gap-2">
        <UBadge v-for="tag in image.tags" :key="tag" color="neutral" variant="subtle" :label="`#${tag}`" />
      </div>

      <dl class="mt-7 space-y-3 border-t border-default pt-5 text-sm">
        <div class="flex justify-between gap-4"><dt class="text-muted">浏览</dt><dd class="tabular-nums">{{ compactMetric(image.metrics.views) }}</dd></div>
        <div class="flex justify-between gap-4"><dt class="text-muted">收藏</dt><dd class="tabular-nums">{{ compactMetric(image.metrics.favorites) }}</dd></div>
        <div class="flex justify-between gap-4"><dt class="text-muted">尺寸</dt><dd class="tabular-nums">{{ image.width }} × {{ image.height }}</dd></div>
        <div class="flex items-start justify-between gap-4">
          <dt class="text-muted">来源</dt>
          <dd class="max-w-[12rem] text-right">
            <a v-if="image.sourceUrl" :href="image.sourceUrl" target="_blank" rel="noopener noreferrer nofollow" referrerpolicy="no-referrer" class="break-all text-primary hover:underline">打开来源地址</a>
            <span v-else class="text-muted">来源待补充</span>
          </dd>
        </div>
      </dl>

      <div class="mt-7 grid grid-cols-2 gap-2">
        <UButton color="neutral" variant="outline" icon="i-tabler-flag" label="举报" @click="reportKind = 'report'; reportOpen = true" />
        <UButton color="neutral" variant="outline" icon="i-tabler-link-plus" label="补充来源" @click="reportKind = 'source_correction'; reportOpen = true" />
      </div>
    </aside>

    <UModal v-model:open="reportOpen" :title="reportKind === 'report' ? '举报图片' : '建议来源地址'" description="提交后由运营人员独立复核。">
      <template #body>
        <form class="space-y-4" @submit.prevent="submitCase">
          <UFormField v-if="reportKind === 'report'" label="原因" required><UInput v-model="reportReason" placeholder="例如内容不适合公开展示" /></UFormField>
          <UFormField v-else label="来源地址" required><UInput v-model="proposedSourceUrl" type="url" placeholder="https://" /></UFormField>
          <UFormField label="补充说明"><UTextarea v-model="reportDescription" :rows="4" /></UFormField>
          <div class="flex justify-end gap-2"><UButton color="neutral" variant="ghost" label="取消" @click="() => { reportOpen = false; }" /><UButton type="submit" label="提交" :loading="reportPending" /></div>
        </form>
      </template>
    </UModal>
  </div>
</template>
