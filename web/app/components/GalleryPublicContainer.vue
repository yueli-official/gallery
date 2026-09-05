<script setup lang="ts">
withDefaults(defineProps<{ as?: string }>(), { as: "div" });
</script>

<template>
  <component :is="as" class="mx-auto w-full max-w-[75rem]">
    <slot />
  </component>
</template>
