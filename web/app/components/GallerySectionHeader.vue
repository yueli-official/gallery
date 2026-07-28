<script setup lang="ts">
defineProps<{
  eyebrow?: string;
  title: string;
  description?: string;
  to?: string;
  actionLabel?: string;
}>();
</script>

<template>
  <header class="gallery-section-header">
    <div>
      <h2>{{ title }}</h2>
      <p v-if="description">{{ description }}</p>
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
