<script setup lang="ts">
import { SkeletonList } from "@platform/manage/components";
import type { GalleryClassificationGovernancePreview } from "~/types/gallery";

const open = defineModel<boolean>("open", { required: true });
const props = defineProps<{
  previewing: boolean;
  executing: boolean;
  error: string;
  preview?: GalleryClassificationGovernancePreview;
  destructive: boolean;
}>();
defineEmits<{ execute: []; deleteAll: [] }>();

const needsDeleteConfirmation = computed(() =>
  Boolean(
    props.preview?.diagnostics.some(
      (item) => item.code === "govern.delete_confirmation_required",
    ),
  ),
);
const stepLabel = (kind: string) =>
  ({
    change_status: "切换状态",
    move_child: "移动子节点",
    migrate_assignments: "迁移对象归属",
    migrate_primary: "迁移主分类",
    migrate_aliases: "迁移别名",
    set_replacement: "建立直接替代",
    clear_primary_assignments: "清除主分类",
    delete_assignments: "删除对象归属",
    delete_aliases: "删除别名",
    delete_references: "删除历史引用",
    delete_identity: "删除标识",
  })[kind] || kind;

function close() {
  open.value = false;
}
</script>

<template>
  <UModal
    v-model:open="open"
    title="治理影响预览"
    description="执行时会重算 revision 和 impact token；任何变化都会拒绝旧计划。"
  >
    <template #body>
      <div class="space-y-4">
        <div v-if="previewing" class="py-8"><SkeletonList :rows="4" /></div>
        <UAlert
          v-else-if="error"
          color="error"
          variant="subtle"
          title="操作失败"
          :description="error"
        />
        <template v-else-if="preview">
          <div class="flex items-center justify-between gap-3">
            <UBadge
              :color="preview.outcome === 'planned' ? 'success' : 'warning'"
              variant="soft"
              :label="preview.outcome"
            />
            <span class="text-xs text-muted"
              >revision {{ preview.catalogRevision }}</span
            >
          </div>
          <div v-if="preview.diagnostics.length" class="space-y-2">
            <UAlert
              v-for="item in preview.diagnostics"
              :key="`${item.code}:${item.reference}`"
              color="warning"
              variant="subtle"
              :title="item.code"
              :description="item.reference || item.path.join('.')"
            />
          </div>
          <ol
            v-if="preview.plan.steps.length"
            class="divide-y divide-default rounded-lg border border-default px-4"
          >
            <li
              v-for="(step, index) in preview.plan.steps"
              :key="`${index}:${step.kind}:${step.sourceId}`"
              class="flex items-start justify-between gap-3 py-3 text-sm"
            >
              <div>
                <p class="font-medium text-highlighted">
                  {{ index + 1 }}. {{ stepLabel(step.kind) }}
                </p>
                <p class="mt-1 break-all text-xs text-muted">
                  {{ step.sourceId
                  }}<template v-if="step.targetId">
                    → {{ step.targetId }}</template
                  >
                </p>
              </div>
              <UBadge
                v-if="step.affectedCount"
                color="neutral"
                variant="soft"
                :label="`${step.affectedCount} 项`"
              />
            </li>
          </ol>
        </template>
        <div class="flex flex-wrap justify-end gap-2">
          <UButton
            class="min-h-11"
            color="neutral"
            variant="ghost"
            label="关闭"
            @click="close"
          />
          <UButton
            v-if="needsDeleteConfirmation"
            class="min-h-11"
            color="error"
            variant="outline"
            label="生成全部删除计划"
            :loading="previewing"
            @click="$emit('deleteAll')"
          />
          <UButton
            v-if="preview?.outcome === 'planned'"
            class="min-h-11"
            :color="destructive ? 'error' : 'primary'"
            :label="destructive ? '确认并执行' : '执行计划'"
            :loading="executing"
            @click="$emit('execute')"
          />
        </div>
      </div>
    </template>
  </UModal>
</template>
