<script setup lang="ts">
import type { GalleryCollection } from "~/types/gallery";

const { data, error, status, refresh } = await useFetch<{ collections: GalleryCollection[] }>("/api/gallery/collections");
useSeoMeta({ title: "专题集合", description: "由运营方整理的公开图片专题。" });
</script>

<template>
  <div class="gallery-page max-w-6xl">
    <header class="mb-8 max-w-2xl">
      <p class="text-xs font-semibold uppercase tracking-[.18em] text-primary">Editorial collections</p>
      <h1 class="mt-2 text-3xl font-semibold tracking-tight text-highlighted">专题集合</h1>
      <p class="mt-2 text-sm leading-6 text-muted">把多张图片组织成一个主题；集合只负责关系，不改变图片本身。</p>
    </header>
    <div v-if="status === 'pending'" class="grid gap-5 sm:grid-cols-2"><USkeleton v-for="index in 6" :key="index" class="h-56 rounded-lg" /></div>
    <UAlert v-else-if="error" color="error" variant="subtle" title="专题加载失败"><template #actions><UButton label="重试" @click="() => refresh()" /></template></UAlert>
    <div v-else-if="data?.collections.length" class="grid gap-x-6 gap-y-10 sm:grid-cols-2">
      <NuxtLink v-for="collection in data.collections" :key="collection.id" :to="`/collections/${collection.slug}`" class="group rounded-lg focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary">
        <div class="grid aspect-[16/9] place-items-center overflow-hidden rounded-lg bg-elevated">
          <UIcon name="i-tabler-folders" class="size-9 text-dimmed transition-transform group-hover:scale-105" />
        </div>
        <div class="mt-3 flex items-start justify-between gap-4"><div><h2 class="font-semibold text-highlighted">{{ collection.name }}</h2><p class="mt-1 line-clamp-2 text-sm text-muted">{{ collection.description }}</p></div><span class="shrink-0 text-xs tabular-nums text-muted">{{ collection.itemCount }} 张</span></div>
      </NuxtLink>
    </div>
    <div v-else class="border-y border-dashed border-default py-16 text-center text-sm text-muted">还没有公开专题。</div>
  </div>
</template>
