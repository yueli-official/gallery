<script setup lang="ts">
import type {
  GalleryClassificationCatalogNode,
  GalleryClassificationGovernanceCommand,
} from "~/types/gallery";

type IdentityKind = GalleryClassificationGovernanceCommand["kind"];
type Operation = "status" | "reparent" | "merge" | "delete";

const props = defineProps<{ items: GalleryClassificationCatalogNode[] }>();
const emit = defineEmits<{
  action: [
    operation: Operation,
    kind: IdentityKind,
    item: GalleryClassificationCatalogNode,
  ];
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
            label: "移动到其他分类",
            icon: "i-tabler-arrows-move",
            onSelect: () => emit("action", "reparent", "category", item),
          },
          {
            label: "合并到其他分类",
            icon: "i-tabler-git-merge",
            onSelect: () => emit("action", "merge", "category", item),
          },
        ];
  return [
    regular,
    [
      {
        label: "删除分类",
        icon: "i-tabler-trash",
        color: "error" as const,
        onSelect: () => emit("action", "delete", "category", item),
      },
    ],
  ].filter((group) => group.length);
}
</script>

<template>
  <section aria-labelledby="classification-category-heading">
    <div class="mb-3 flex items-end justify-between gap-3">
      <div>
        <h2
          id="classification-category-heading"
          class="font-semibold text-highlighted"
        >
          分类
        </h2>
        <p class="mt-1 text-sm text-muted">用树形层级组织主要浏览入口。</p>
      </div>
      <span class="text-xs text-dimmed">{{ items.length }} 项</span>
    </div>
    <div class="divide-y divide-default border-y border-default">
      <article
        v-for="item in items"
        :key="item.id"
        class="grid gap-3 py-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-center"
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
          <p class="mt-1 text-xs text-muted">{{ item.slug }}</p>
          <details class="mt-1 text-xs text-dimmed">
            <summary class="cursor-pointer">标识信息</summary>
            <p class="mt-1 break-all font-mono">{{ item.id }}</p>
          </details>
        </div>
        <div class="flex items-center justify-end gap-1">
          <UButton
            v-if="item.status !== 'replaced'"
            class="min-h-11"
            color="primary"
            variant="soft"
            size="sm"
            :label="item.status === 'active' ? '停用' : '启用'"
            @click="$emit('action', 'status', 'category', item)"
          />
          <UDropdownMenu :items="moreItems(item)"
            ><UButton
              class="min-h-11"
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-tabler-dots"
              square
              :aria-label="`更多分类操作：${item.name}`"
          /></UDropdownMenu>
        </div>
      </article>
    </div>
  </section>
</template>
