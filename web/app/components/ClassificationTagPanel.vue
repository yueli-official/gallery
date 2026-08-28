<script setup lang="ts">
import { AdminRowActions } from "@yueli/ui/admin";
import { ManageEmpty, SkeletonList } from "~/utils/manageComponents";
import type {
  GalleryClassificationGovernanceCommand,
  GalleryClassificationTag,
} from "~/types/gallery";

type IdentityKind = GalleryClassificationGovernanceCommand["kind"];
type Operation = "status" | "merge" | "delete";

defineProps<{
  tags: GalleryClassificationTag[];
  pending: boolean;
  hydrated: boolean;
  nextCursor: string;
  loadingMore: boolean;
  canGovern: boolean;
}>();
const emit = defineEmits<{
  action: [
    operation: Operation,
    kind: IdentityKind,
    item: GalleryClassificationTag,
  ];
  loadMore: [];
  edit: [kind: IdentityKind, item: GalleryClassificationTag];
}>();

const statusLabel: Record<string, string> = {
  inactive: "已停用",
  replaced: "已合并",
};
function moreItems(item: GalleryClassificationTag) {
  const regular =
    item.status === "replaced"
      ? []
      : [
          {
            id: "edit",
            label: "编辑标签",
            icon: "i-tabler-pencil",
            onSelect: () => emit("edit", "tag", item),
          },
          {
            id: "status",
            label: item.status === "active" ? "停用标签" : "启用标签",
            icon:
              item.status === "active" ? "i-tabler-eye-off" : "i-tabler-eye",
            onSelect: () => emit("action", "status", "tag", item),
          },
          {
            id: "merge",
            label: "合并到其他标签",
            icon: "i-tabler-git-merge",
            onSelect: () => emit("action", "merge", "tag", item),
          },
        ];
  return [
    regular,
    [
      {
        id: "delete",
        label: "删除标签",
        icon: "i-tabler-trash",
        tone: "danger" as const,
        onSelect: () => emit("action", "delete", "tag", item),
      },
    ],
  ].filter((group) => group.length);
}
</script>

<template>
  <section aria-label="标签列表" data-classification-tags>
    <SkeletonList v-if="!hydrated || pending" :rows="4" />
    <div
      v-else-if="tags.length"
      class="border-y border-default"
    >
      <div
        class="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] items-center gap-3 bg-elevated/50 px-3 py-2 text-xs font-medium text-muted"
        aria-hidden="true"
      >
        <span>名称</span>
        <span>标识</span>
        <span class="text-right">操作</span>
      </div>
      <div class="divide-y divide-default border-t border-default">
      <article
        v-for="tag in tags"
        :key="tag.id"
        class="grid min-h-12 grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] items-center gap-3 px-3 py-2"
        data-classification-row
      >
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h3 class="font-medium text-highlighted">{{ tag.name }}</h3>
            <UTooltip
              v-if="tag.assignmentCount"
              :text="`${tag.assignmentCount} 个关系`"
            >
              <UBadge
                size="xs"
                color="neutral"
                variant="soft"
                :label="String(tag.assignmentCount)"
              />
            </UTooltip>
            <UTooltip
              v-if="tag.aliasCount"
              :text="`${tag.aliasCount} 个别名`"
            >
              <UBadge
                size="xs"
                color="neutral"
                variant="soft"
                :label="`${tag.aliasCount} 别名`"
              />
            </UTooltip>
            <UBadge
              v-if="tag.status !== 'active'"
              color="warning"
              variant="soft"
              :label="statusLabel[tag.status] || tag.status"
            />
          </div>
        </div>
        <p class="min-w-0 truncate text-xs text-muted">{{ tag.slug }}</p>
        <AdminRowActions
          v-if="canGovern"
          :items="moreItems(tag)"
          :label="`更多标签操作：${tag.name}`"
          presentation="overflow"
        />
      </article>
      </div>
      <div v-if="nextCursor" class="flex justify-center py-4">
        <UButton
          class="min-h-11"
          color="neutral"
          variant="outline"
          label="加载更多标签"
          :loading="loadingMore"
          @click="$emit('loadMore')"
        />
      </div>
    </div>
    <ManageEmpty
      v-else
      icon="i-tabler-tags"
      title="还没有标签"
      description="通过标签提案或新增标签后会显示在这里。"
    />
  </section>
</template>
