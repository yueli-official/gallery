<script setup lang="ts">
defineProps<{
  icon: string;
  title: string;
  description?: string;
}>();
</script>

<template>
  <div
    class="gallery-compact-empty grid min-h-84 place-items-center rounded-2xl border border-default bg-[color-mix(in_srgb,var(--gallery-panel)_94%,transparent)] px-6 py-16 text-center"
  >
    <div>
      <span
        class="gallery-empty-icon inline-grid size-12 place-items-center rounded-[0.9rem] bg-primary/10 text-primary"
      >
        <UIcon :name="icon" class="size-6" />
      </span>
      <h2 class="mt-4 text-lg font-semibold text-highlighted">{{ title }}</h2>
      <p v-if="description" class="mx-auto mt-2 max-w-sm text-sm leading-6 text-muted">
        {{ description }}
      </p>
      <slot name="actions" />
    </div>
  </div>
</template>
