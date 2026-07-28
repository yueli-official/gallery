<script setup lang="ts">
defineProps<{
  items: { key: string; label: string; count?: number }[];
}>();
const model = defineModel<string>({ required: true });
</script>

<template>
  <div
    class="flex items-center gap-1 overflow-x-auto overflow-y-hidden border-b border-default [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
    role="tablist"
  >
    <button
      v-for="item in items"
      :key="item.key"
      type="button"
      role="tab"
      class="-mb-px flex min-h-11 shrink-0 items-center gap-1.5 border-b-2 px-3 py-2 text-sm font-medium transition"
      :class="
        model === item.key
          ? 'border-primary text-primary'
          : 'border-transparent text-muted hover:text-default'
      "
      :aria-selected="model === item.key"
      @click="model = item.key"
    >
      {{ item.label }}
      <span
        v-if="item.count != null"
        class="rounded-full px-1.5 text-xs tabular-nums"
        :class="
          model === item.key
            ? 'bg-primary/10 text-primary'
            : 'bg-elevated text-dimmed'
        "
        >{{ item.count }}</span
      >
    </button>
  </div>
</template>
