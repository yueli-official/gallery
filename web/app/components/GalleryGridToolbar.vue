<script setup lang="ts" generic="T extends string">
defineProps<{
  sortBy: T;
  sortOrder: "asc" | "desc";
  items: { label: string; value: T }[];
}>();
const emit = defineEmits<{
  "update:sortBy": [value: T];
  "update:sortOrder": [value: "asc" | "desc"];
}>();
</script>

<template>
  <div class="flex min-w-0 flex-wrap items-center justify-between gap-2" data-gallery-grid-toolbar>
    <span class="text-xs text-muted">选择本页</span>
    <div class="flex items-center gap-1">
      <USelect :model-value="String(sortBy)" :items="items.map(item => ({ label: item.label, value: String(item.value) }))" aria-label="排序依据" size="xs" class="w-28"
        @update:model-value="emit('update:sortBy', $event as T)" />
      <UButton :icon="sortOrder === 'asc' ? 'i-tabler-sort-ascending' : 'i-tabler-sort-descending'"
        :aria-label="sortOrder === 'asc' ? '当前升序，切换降序' : '当前降序，切换升序'"
        color="neutral" variant="ghost" size="xs" square
        @click="emit('update:sortOrder', sortOrder === 'asc' ? 'desc' : 'asc')" />
    </div>
  </div>
</template>
