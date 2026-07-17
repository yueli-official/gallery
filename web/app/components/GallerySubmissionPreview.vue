<script setup lang="ts">
const props = defineProps<{ submissionId: string; alt: string }>();
const { call } = useApi();
const source = ref("");
const failed = ref(false);

async function load() {
  source.value = "";
  failed.value = false;
  try {
    const response = await call<{ url: string }>(
      submissionPreviewURL(props.submissionId),
    );
    source.value = response.url;
  } catch {
    failed.value = true;
  }
}

watch(() => props.submissionId, load, { immediate: true });
</script>

<template>
  <img
    v-if="source"
    :src="source"
    :alt="alt"
    loading="lazy"
    class="aspect-[4/3] size-full object-cover"
  />
  <div
    v-else-if="failed"
    class="flex size-full items-center justify-center text-muted"
    role="img"
    :aria-label="`${alt}（预览加载失败）`"
  >
    <UIcon name="i-tabler-photo-off" class="size-5" />
  </div>
  <USkeleton v-else class="size-full rounded-none" />
</template>
