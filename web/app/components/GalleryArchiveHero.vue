<script setup lang="ts">
withDefaults(defineProps<{ title: string; count?: number | string; countLabel?: string }>(), { countLabel: "张图片" });
</script>
<template>
  <header class="gallery-page-heading">
    <div class="gallery-page-intro"><h1>{{ title }}<span v-if="count !== undefined" class="gallery-page-count">{{ count }} {{ countLabel }}</span></h1></div>
    <div v-if="$slots.actions" class="gallery-page-actions"><slot name="actions" /></div>
  </header>
</template>
<style scoped>
.gallery-page-heading{display:flex;align-items:center;justify-content:space-between;gap:1rem;margin-bottom:1.25rem;flex-wrap:wrap}.gallery-page-intro{min-width:0}.gallery-page-heading h1{display:flex;align-items:baseline;flex-wrap:wrap;gap:.7rem;font-size:1.35rem;font-weight:650;letter-spacing:-.025em}.gallery-page-count{font-size:.7rem;font-weight:400;letter-spacing:0;color:var(--gallery-muted);white-space:nowrap}.gallery-page-actions{display:flex;align-items:center;gap:.5rem;flex-shrink:0}
</style>
