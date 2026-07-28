<script setup lang="ts">
import { createGalleryNotifier } from "~/utils/feedback";

const props = defineProps<{
  imageId: string;
  open: boolean;
  kind: "report" | "source_correction";
}>();
const emit = defineEmits<{ "update:open": [open: boolean] }>();
const toast = createGalleryNotifier(useToast());
const reportReason = ref("");
const reportDescription = ref("");
const proposedSourceUrl = ref("");
const pending = ref(false);
const openModel = computed({
  get: () => props.open,
  set: (value) => emit("update:open", value),
});

watch(openModel, (open) => {
  if (open) return;
  reportReason.value = "";
  reportDescription.value = "";
  proposedSourceUrl.value = "";
});

async function submit(): Promise<void> {
  if (!props.imageId || pending.value) return;
  pending.value = true;
  try {
    await $fetch(
      `/api/gallery/images/${encodeURIComponent(props.imageId)}/cases`,
      {
        method: "POST",
        body: {
          kind: props.kind,
          reason: reportReason.value,
          description: reportDescription.value,
          proposedSourceUrl: proposedSourceUrl.value,
        },
      },
    );
    openModel.value = false;
    // feedback-contract: the dialog closes after submission, so no stable inline surface remains.
    toast.add({
      title: props.kind === "report" ? "举报已提交" : "来源建议已提交",
      description: "运营人员会独立复核，不会按次数自动下架。",
      color: "success",
    });
  } catch (reason: any) {
    toast.add({
      title: "提交失败",
      description: reason?.data?.message || reason?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    pending.value = false;
  }
}

function close(): void {
  openModel.value = false;
}
</script>

<template>
  <UModal
    v-model:open="openModel"
    :title="kind === 'report' ? '举报图片' : '建议来源地址'"
    description="提交后由运营人员独立复核。"
  >
    <template #body>
      <form class="space-y-4" @submit.prevent="submit">
        <UFormField v-if="kind === 'report'" label="原因" required>
          <UInput v-model="reportReason" placeholder="例如内容不适合公开展示" />
        </UFormField>
        <UFormField v-else label="来源地址" required>
          <UInput
            v-model="proposedSourceUrl"
            type="url"
            placeholder="https://"
          />
        </UFormField>
        <UFormField label="补充说明">
          <UTextarea v-model="reportDescription" :rows="4" />
        </UFormField>
        <div class="flex justify-end gap-2">
          <UButton
            color="neutral"
            variant="ghost"
            label="取消"
            @click="close"
          />
          <UButton type="submit" label="提交" :loading="pending" />
        </div>
      </form>
    </template>
  </UModal>
</template>
