<script setup lang="ts">
defineProps<{
  eyebrow?: string;
  title: string;
  to?: string;
  actionLabel?: string;
}>();
</script>

<template>
  <header
    class="gallery-section-header mb-6 flex items-start justify-between gap-8 md:items-end"
  >
    <div>
      <h2
        class="font-display text-[length:var(--gallery-title-section)] font-semibold leading-[1.05] tracking-[-0.04em] text-highlighted"
      >
        {{ title }}
      </h2>
    </div>
    <slot name="action">
      <UButton
        v-if="to"
        :to="to"
        color="neutral"
        variant="ghost"
        trailing-icon="i-tabler-arrow-right"
        :label="actionLabel || '查看全部'"
      />
    </slot>
  </header>
</template>
