<script setup lang="ts">
import type { FailureFeedback } from "@yueli/http-runtime";
defineProps<{ feedback?: FailureFeedback | null }>();
</script>
<template>
  <div
    v-if="feedback"
    role="alert"
    class="grid gap-2 text-xs text-error"
    data-gallery-failure
  >
    <p>{{ feedback.message }}</p>
    <p v-for="(message, index) in feedback.summary" :key="index">
      {{ message }}
    </p>
    <details v-if="feedback.technical.traceId" class="text-muted">
      <summary class="cursor-pointer">技术详情</summary>
      <p class="select-all break-all">
        {{ feedback.technical.code }} · {{ feedback.technical.traceId }}
      </p>
    </details>
  </div>
</template>
