<script setup lang="ts">
import { ManageEmpty, SkeletonList } from "@platform/manage/components";
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
}>();
const emit = defineEmits<{
  action: [
    operation: Operation,
    kind: IdentityKind,
    item: GalleryClassificationTag,
  ];
  loadMore: [];
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
            label: item.status === "active" ? "停用标签" : "启用标签",
            icon:
              item.status === "active" ? "i-tabler-eye-off" : "i-tabler-eye",
            onSelect: () => emit("action", "status", "tag", item),
          },
          {
            label: "合并到其他标签",
            icon: "i-tabler-git-merge",
            onSelect: () => emit("action", "merge", "tag", item),
          },
        ];
  return [
    regular,
    [
      {
        label: "删除标签",
        icon: "i-tabler-trash",
        color: "error" as const,
        onSelect: () => emit("action", "delete", "tag", item),
      },
    ],
  ].filter((group) => group.length);
}
</script>

<template>
  <section aria-labelledby="classification-tag-heading">
    <div class="mb-3 flex items-end justify-between gap-3">
      <div>
        <h2
          id="classification-tag-heading"
          class="font-semibold text-highlighted"
        >
          标签
        </h2>
        <p class="mt-1 text-sm text-muted">扁平管理长尾词和同义词关系。</p>
      </div>
      <span class="text-xs text-dimmed">{{ tags.length }} 项</span>
    </div>
    <SkeletonList v-if="!hydrated || pending" :rows="4" />
    <div
      v-else-if="tags.length"
      class="divide-y divide-default border-y border-default"
    >
      <article
        v-for="tag in tags"
        :key="tag.id"
        class="grid gap-3 py-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-center"
      >
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h3 class="font-medium text-highlighted">{{ tag.name }}</h3>
            <UBadge
              v-if="tag.status !== 'active'"
              color="warning"
              variant="soft"
              :label="statusLabel[tag.status] || tag.status"
            />
          </div>
          <p class="mt-1 truncate text-xs text-muted">{{ tag.slug }}</p>
          <p class="mt-1 text-xs text-dimmed">
            {{ tag.assignmentCount }} 个关系 · {{ tag.aliasCount }} 个 Alias
          </p>
          <details class="mt-1 text-xs text-dimmed">
            <summary class="cursor-pointer">标识信息</summary>
            <p class="mt-1 break-all font-mono">{{ tag.id }}</p>
          </details>
        </div>
        <div class="flex flex-wrap gap-2">
          <UDropdownMenu :items="moreItems(tag)"
            ><UButton
              class="min-h-11"
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-tabler-dots"
              square
              :aria-label="`更多标签操作：${tag.name}`"
          /></UDropdownMenu>
        </div>
      </article>
      <div v-if="nextCursor" class="flex justify-center py-4">
        <UButton
          class="min-h-11"
          color="neutral"
          variant="outline"
          label="加载更多 Tag"
          :loading="loadingMore"
          @click="$emit('loadMore')"
        />
      </div>
    </div>
    <ManageEmpty
      v-else
      icon="i-tabler-tags"
      title="还没有 canonical Tag"
      description="批准第一个 Proposal 后会在这里出现。"
    />
  </section>
</template>
