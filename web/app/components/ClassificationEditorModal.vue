<script setup lang="ts">
import type {
  GalleryClassificationCatalogFacet,
  GalleryClassificationCatalogNode,
  GalleryClassificationTag,
} from "~/types/gallery";

type ManagedIdentity =
  | GalleryClassificationCatalogNode
  | GalleryClassificationCatalogFacet
  | GalleryClassificationTag;

const open = defineModel<boolean>("open", { required: true });
const props = defineProps<{
  mode: "reparent" | "merge";
  identity?: ManagedIdentity;
  targets: ManagedIdentity[];
  childCount: number;
}>();
const emit = defineEmits<{ preview: [targetId: string] }>();
const targetId = ref("");

watch(
  () => [open.value, props.identity?.id],
  () => {
    if (open.value) targetId.value = "";
  },
);

function submit() {
  if (!targetId.value) return;
  emit("preview", targetId.value);
}

function close() {
  open.value = false;
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="mode === 'merge' ? '合并标识' : '移动节点'"
    description="系统会根据最新目录生成计划；这里只选择目标，不直接写数据库。"
  >
    <template #body>
      <form class="space-y-4" @submit.prevent="submit">
        <div class="rounded-md bg-elevated p-3 text-sm">
          <p class="font-medium text-highlighted">{{ identity?.name }}</p>
          <p class="mt-1 break-all text-xs text-muted">{{ identity?.id }}</p>
        </div>
        <UFormField
          :label="mode === 'merge' ? '合并目标' : '新父节点'"
          required
        >
          <select
            v-model="targetId"
            class="h-11 w-full rounded-md border border-default bg-default px-3 text-sm text-highlighted"
          >
            <option value="" disabled>请选择</option>
            <option v-if="mode === 'reparent'" value="root">设为根节点</option>
            <option
              v-for="target in targets"
              :key="target.id"
              :value="target.id"
            >
              {{ target.name }} · {{ target.slug }}
            </option>
          </select>
        </UFormField>
        <p v-if="mode === 'merge' && childCount" class="text-xs text-muted">
          其直接子节点将显式移动到目标；预览会再次验证完整性和环。
        </p>
        <div class="flex justify-end gap-2">
          <UButton
            class="min-h-11"
            color="neutral"
            variant="ghost"
            label="取消"
            @click="close"
          />
          <UButton
            class="min-h-11"
            type="submit"
            label="生成预览"
            :disabled="!targetId"
          />
        </div>
      </form>
    </template>
  </UModal>
</template>
