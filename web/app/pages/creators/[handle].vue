<script setup lang="ts">
import type { GalleryArtworkCard, GalleryPublicCreator } from "~/types/gallery";

const route = useRoute();
const { data, error } = await useFetch<{ creator: GalleryPublicCreator; artworks: GalleryArtworkCard[] }>(
  `/api/gallery/creators/${encodeURIComponent(String(route.params.handle))}`,
);
useSeoMeta({
  title: () => data.value?.creator.displayName || "创作者",
  description: () => data.value?.creator.bio || `查看 @${data.value?.creator.handle || route.params.handle} 的公开作品。`,
});
</script>

<template>
  <div v-if="data" class="space-y-10 py-4 sm:py-8">
    <header class="flex flex-col gap-6 border-b border-default pb-8 sm:flex-row sm:items-center">
      <div class="grid size-20 shrink-0 place-items-center rounded-2xl bg-primary/10 text-primary"><UIcon name="i-tabler-user" class="size-9" /></div>
      <div class="min-w-0">
        <p class="text-sm text-muted">@{{ data.creator.handle }}</p>
        <h1 class="mt-1 font-display text-3xl font-semibold text-highlighted">{{ data.creator.displayName }}</h1>
        <p class="mt-3 max-w-2xl text-sm leading-6 text-muted">{{ data.creator.bio || "这位创作者还没有填写个人介绍。" }}</p>
      </div>
    </header>
    <section>
      <div class="mb-5 flex items-end justify-between"><h2 class="text-xl font-semibold text-highlighted">公开作品</h2><span class="text-sm text-muted">{{ data.artworks.length }} 件</span></div>
      <GalleryArtworkGrid v-if="data.artworks.length" :items="data.artworks" />
      <div v-else class="rounded-xl border border-dashed border-default py-16 text-center text-sm text-muted">暂时没有公开作品。</div>
    </section>
  </div>
  <UAlert v-else-if="error" color="error" variant="subtle" title="没有找到这个创作者" />
</template>
