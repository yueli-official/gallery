<script setup lang="ts">
import type { GallerySearchSuggestion } from "~/types/gallery";

defineProps<{ items: GallerySearchSuggestion[] }>();
const emit = defineEmits<{ select: [key: string] }>();
</script>

<template>
  <div v-if="items.length" class="gallery-search-suggestions">
    <p>相关分类与维度</p>
    <UButton
      v-for="item in items"
      :key="item.key"
      color="neutral"
      variant="ghost"
      size="xs"
      @click="emit('select', item.key)"
    >
      <span class="flex w-full items-center justify-between gap-3">
        <span>{{ item.label }}</span>
        <span class="text-[0.68rem] text-dimmed">{{ item.context }}</span>
      </span>
    </UButton>
  </div>
</template>
