<script setup lang="ts">
import type { GalleryArtwork } from "~/types/gallery";

const route = useRoute("/artworks/[artworkId]");
const { data, error, status } = await useFetch<GalleryArtwork>(
  () => `/api/gallery/artworks/${route.params.artworkId}`,
  { key: `gallery-artwork-${route.params.artworkId}` },
);

useSeoMeta({
  title: () => data.value?.title ?? "作品详情",
  description: () => data.value?.description ?? "查看图片作品详情。",
});
</script>

<template>
  <div>
    <div v-if="status === 'pending'" class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]">
      <USkeleton class="aspect-[4/5] rounded-xl" />
      <div class="space-y-4">
        <USkeleton class="h-8 w-3/4" />
        <USkeleton class="h-24 w-full" />
      </div>
    </div>
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      icon="i-tabler-photo-off"
      title="无法读取这件作品"
      description="作品不存在、尚未发布，或已被限制展示。"
    />
    <article v-else-if="data" class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]">
      <div class="space-y-4">
        <figure
          v-for="(asset, index) in data.assets"
          :key="asset.id"
          class="overflow-hidden rounded-xl bg-elevated"
        >
          <img
            :src="asset.detailUrl || asset.cardUrl"
            :alt="asset.altText || `${data.title} · 第 ${index + 1} 张`"
            :width="asset.width || undefined"
            :height="asset.height || undefined"
            :loading="index === 0 ? 'eager' : 'lazy'"
            :fetchpriority="index === 0 ? 'high' : 'auto'"
            class="mx-auto block h-auto max-w-full"
          />
        </figure>
      </div>
      <aside class="h-fit space-y-6 lg:sticky lg:top-24">
        <div>
          <p class="text-sm text-primary">{{ data.creator.displayName }}</p>
          <h1 class="mt-2 text-3xl font-semibold tracking-tight text-highlighted">
            {{ data.title }}
          </h1>
          <p v-if="data.description" class="mt-4 whitespace-pre-line leading-7 text-toned">
            {{ data.description }}
          </p>
        </div>
        <div class="flex flex-wrap gap-2">
          <UBadge v-for="tag in data.tags" :key="tag" color="neutral" variant="subtle">
            #{{ tag }}
          </UBadge>
        </div>
        <dl class="grid gap-3 border-t border-default pt-5 text-sm">
          <div class="flex justify-between gap-4"><dt class="text-muted">内容分级</dt><dd>{{ data.contentRating }}</dd></div>
          <div class="flex justify-between gap-4"><dt class="text-muted">AI 使用</dt><dd>{{ data.aiUsage }}</dd></div>
          <div class="flex justify-between gap-4"><dt class="text-muted">训练授权</dt><dd>{{ data.aiTrainingPermission }}</dd></div>
          <div class="flex justify-between gap-4"><dt class="text-muted">许可</dt><dd>{{ data.license }}</dd></div>
        </dl>
      </aside>
    </article>
  </div>
</template>

