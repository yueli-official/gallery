<script setup lang="ts">
import type { GalleryPresentationRule } from "~/types/presentation";
defineProps<{ visual?: GalleryPresentationRule }>();
</script>

<template>
  <span class="grid size-9 shrink-0 place-items-center overflow-hidden rounded-md border border-default bg-muted text-primary" aria-hidden="true">
    <img v-if="visual?.mode === 'image' && visual.assetId" :src="galleryRendition(visual, 'thumbnail')" alt="" class="size-full object-cover" loading="lazy" />
    <UIcon v-else :name="`i-tabler-${visual?.icon || 'category'}`" class="size-5" />
  </span>
</template>
