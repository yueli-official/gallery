<script setup lang="ts">
defineProps<{
  title: string;
  description?: string;
}>();
</script>

<template>
  <header
    class="gallery-page-header mb-4 flex items-start justify-between gap-2 border-b border-default pb-3 pt-1.5 md:mb-6 md:items-end md:gap-8 md:pb-4"
  >
    <div>
      <h1
        class="gallery-page-title font-display text-[1.75rem] font-bold leading-[1.15] tracking-[-0.04em] text-highlighted md:text-[length:var(--gallery-title-page)]"
      >
        {{ title }}
      </h1>
      <p
        v-if="description"
        class="gallery-page-copy mt-1 max-w-[38rem] text-[0.9rem] leading-[1.7] text-muted"
      >
        {{ description }}
      </p>
    </div>
    <slot name="actions" />
  </header>
</template>
