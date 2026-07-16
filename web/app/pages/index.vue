<script setup lang="ts">
import type { GalleryDiscovery } from "~/types/gallery";

const route = useRoute();
const router = useRouter();
const seed = computed(() => String(route.query.seed || ""));
const { data, error, status, refresh } = await useFetch<GalleryDiscovery>(
  "/api/gallery/discovery",
  {
    query: computed(() => ({ seed: seed.value || undefined })),
    watch: [seed],
  },
);

useSeoMeta({
  title: () => data.value?.site.name || "月离图库",
  description: () =>
    data.value?.site.description || "随机发现、收藏和投稿公开图片。",
  robots: () => (seed.value ? "noindex,follow" : "index,follow"),
});

function nextBatch() {
  const next =
    globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random()}`;
  void router.replace({ path: "/", query: { seed: next } });
}
</script>

<template>
  <div class="gallery-page">
    <header class="mb-5 flex flex-wrap items-end justify-between gap-4 sm:mb-7">
      <div>
        <p
          class="text-xs font-semibold uppercase tracking-[.18em] text-primary"
        >
          Random archive
        </p>
        <h1
          class="mt-1 text-2xl font-semibold tracking-tight text-highlighted sm:text-3xl"
        >
          随机看看
        </h1>
        <p class="mt-1.5 text-sm text-muted">
          不按热度，把不同主题的图片重新洗牌。
        </p>
      </div>
      <UButton
        color="neutral"
        variant="outline"
        icon="i-tabler-refresh"
        label="换一批"
        :loading="status === 'pending'"
        @click="nextBatch"
      />
    </header>

    <div
      v-if="status === 'pending'"
      class="gallery-masonry"
      aria-label="正在加载随机图片"
    >
      <USkeleton
        v-for="index in 20"
        :key="index"
        class="mb-3 h-64 break-inside-avoid rounded-lg"
        :style="{ height: `${180 + (index % 4) * 46}px` }"
      />
    </div>

    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      icon="i-tabler-alert-circle"
      title="随机图片暂时没有加载出来"
      description="请确认 Gallery API 与 Asset 服务可用。"
    >
      <template #actions
        ><UButton color="error" variant="soft" label="重试" @click="refresh()"
      /></template>
    </UAlert>

    <GalleryMasonry
      v-else-if="data?.images.length"
      :items="data.images"
      priority
    />

    <div
      v-else
      class="grid min-h-[45vh] place-items-center border-y border-dashed border-default py-16 text-center"
    >
      <div class="max-w-sm">
        <UIcon name="i-tabler-photo" class="mx-auto size-9 text-dimmed" />
        <h2 class="mt-4 text-lg font-semibold text-highlighted">
          图库正在等待第一批图片
        </h2>
        <p class="mt-2 text-sm leading-6 text-muted">
          图片完成处理和审核后，会直接出现在随机流里。
        </p>
        <UButton
          to="/submit"
          class="mt-5"
          icon="i-tabler-photo-up"
          label="投稿一张图片"
        />
      </div>
    </div>
  </div>
</template>
