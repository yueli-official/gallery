<script setup lang="ts">
import { AdminRowActions } from "@yueli/ui/admin";
import type {
  GalleryClassificationCatalogNode,
  GalleryClassificationGovernanceCommand,
} from "~/types/gallery";

type IdentityKind = GalleryClassificationGovernanceCommand["kind"];
type Operation = "status" | "reparent" | "merge" | "delete";

const props = defineProps<{
  items: GalleryClassificationCatalogNode[];
  canGovern: boolean;
}>();
const emit = defineEmits<{
  action: [
    operation: Operation,
    kind: IdentityKind,
    item: GalleryClassificationCatalogNode,
  ];
  edit: [kind: IdentityKind, item: GalleryClassificationCatalogNode];
}>();

const statusLabel: Record<string, string> = {
  draft: "草稿",
  inactive: "已停用",
  replaced: "已合并",
};
function depth(item: GalleryClassificationCatalogNode) {
  let current = item;
  let value = 0;
  while (current.parentId && value < 6) {
    const parent = props.items.find((entry) => entry.id === current.parentId);
    if (!parent) break;
    value += 1;
    current = parent;
  }
  return value;
}
function moreItems(item: GalleryClassificationCatalogNode) {
  const regular =
    item.status === "replaced"
      ? []
      : [
          {
            id: "edit",
            label: "编辑分类",
            icon: "i-tabler-pencil",
            onSelect: () => emit("edit", "category", item),
          },
          {
            id: "status",
            label: item.status === "active" ? "停用分类" : "启用分类",
            icon:
              item.status === "active" ? "i-tabler-eye-off" : "i-tabler-eye",
            onSelect: () => emit("action", "status", "category", item),
          },
          {
            id: "reparent",
            label: "移动到其他分类",
            icon: "i-tabler-arrows-move",
            onSelect: () => emit("action", "reparent", "category", item),
          },
          {
            id: "merge",
            label: "合并到其他分类",
            icon: "i-tabler-git-merge",
            onSelect: () => emit("action", "merge", "category", item),
          },
        ];
  return [
    regular,
    [
      {
        id: "delete",
        label: "删除分类",
        icon: "i-tabler-trash",
        tone: "danger" as const,
        onSelect: () => emit("action", "delete", "category", item),
      },
    ],
  ].filter((group) => group.length);
}
</script>

<template>
  <section aria-label="分类列表">
    <div
      class="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] items-center gap-3 border-y border-default bg-elevated/50 px-3 py-2 text-xs font-medium text-muted"
      aria-hidden="true"
    >
      <span>名称</span>
      <span>标识</span>
      <span class="text-right">操作</span>
    </div>
    <div class="divide-y divide-default border-b border-default">
      <article
        v-for="item in items"
        :key="item.id"
        class="grid min-h-12 grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] items-center gap-3 px-3 py-2"
        data-classification-row
      >
        <div
          class="min-w-0"
          :style="{ paddingLeft: `${depth(item) * 1.25}rem` }"
        >
          <div class="flex flex-wrap items-center gap-2">
            <h3 class="font-medium text-highlighted">{{ item.name }}</h3>
            <UBadge
              v-if="item.status !== 'active'"
              color="warning"
              variant="soft"
              :label="statusLabel[item.status] || item.status"
            />
          </div>
        </div>
        <p class="min-w-0 truncate text-xs text-muted">{{ item.slug }}</p>
        <AdminRowActions
          v-if="canGovern"
          :items="moreItems(item)"
          :label="`更多分类操作：${item.name}`"
          presentation="overflow"
        />
      </article>
    </div>
  </section>
</template>
