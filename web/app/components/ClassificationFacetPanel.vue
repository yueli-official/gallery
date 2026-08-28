<script setup lang="ts">
import { AdminRowActions } from "@yueli/ui/admin";
import type {
  GalleryClassificationCatalogFacet,
  GalleryClassificationCatalogNode,
  GalleryClassificationGovernanceCommand,
} from "~/types/gallery";

type IdentityKind = GalleryClassificationGovernanceCommand["kind"];
type ManagedIdentity =
  GalleryClassificationCatalogFacet | GalleryClassificationCatalogNode;
type Operation = "status" | "reparent" | "merge" | "delete";

defineProps<{
  facets: GalleryClassificationCatalogFacet[];
  canGovern: boolean;
}>();
const emit = defineEmits<{
  action: [operation: Operation, kind: IdentityKind, item: ManagedIdentity];
  createValue: [facet: GalleryClassificationCatalogFacet];
  edit: [kind: IdentityKind, item: ManagedIdentity];
}>();

const statusLabel: Record<string, string> = {
  draft: "草稿",
  inactive: "已停用",
  replaced: "已合并",
};
function facetMore(item: GalleryClassificationCatalogFacet) {
  return [
    [
      {
        id: "edit",
        label: "编辑维度",
        icon: "i-tabler-pencil",
        onSelect: () => emit("edit", "facet", item),
      },
      {
        id: "status",
        label: item.status === "active" ? "停用维度" : "启用维度",
        icon: item.status === "active" ? "i-tabler-eye-off" : "i-tabler-eye",
        onSelect: () => emit("action", "status", "facet", item),
      },
    ],
    [
      {
        id: "delete",
        label: "删除维度",
        icon: "i-tabler-trash",
        tone: "danger" as const,
        onSelect: () => emit("action", "delete", "facet", item),
      },
    ],
  ];
}
function valueMore(item: GalleryClassificationCatalogNode) {
  const regular =
    item.status === "replaced"
      ? []
      : [
          {
            id: "edit",
            label: "编辑维度值",
            icon: "i-tabler-pencil",
            onSelect: () => emit("edit", "facet_value", item),
          },
          {
            id: "status",
            label: item.status === "active" ? "停用维度值" : "启用维度值",
            icon:
              item.status === "active" ? "i-tabler-eye-off" : "i-tabler-eye",
            onSelect: () => emit("action", "status", "facet_value", item),
          },
          {
            id: "reparent",
            label: "移动到其他父级",
            icon: "i-tabler-arrows-move",
            onSelect: () => emit("action", "reparent", "facet_value", item),
          },
          {
            id: "merge",
            label: "合并到其他值",
            icon: "i-tabler-git-merge",
            onSelect: () => emit("action", "merge", "facet_value", item),
          },
        ];
  return [
    regular,
    [
      {
        id: "delete",
        label: "删除维度值",
        icon: "i-tabler-trash",
        tone: "danger" as const,
        onSelect: () => emit("action", "delete", "facet_value", item),
      },
    ],
  ].filter((group) => group.length);
}
</script>

<template>
  <section aria-label="维度列表">
    <div
      class="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] items-center gap-3 border-y border-default bg-elevated/50 px-3 py-2 text-xs font-medium text-muted"
      aria-hidden="true"
    >
      <span>名称</span>
      <span>标识</span>
      <span class="text-right">操作</span>
    </div>
    <div class="divide-y divide-default border-b border-default">
      <section v-for="facet in facets" :key="facet.id">
        <header
          class="grid min-h-12 grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] items-center gap-3 bg-elevated/35 px-3 py-2"
        >
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <h3 class="font-semibold text-highlighted">{{ facet.name }}</h3>
              <UTooltip :text="`${facet.values.length} 个值`">
                <UBadge
                  size="xs"
                  color="neutral"
                  variant="soft"
                  :label="String(facet.values.length)"
                />
              </UTooltip>
              <UBadge
                v-if="facet.status !== 'active'"
                color="warning"
                variant="soft"
                :label="statusLabel[facet.status] || facet.status"
              />
            </div>
          </div>
          <p class="min-w-0 truncate text-xs text-muted">
            {{ facet.slug }}
          </p>
          <div class="flex items-center gap-1">
            <UButton
              v-if="canGovern"
              label="新增"
              icon="i-tabler-plus"
              color="neutral"
              variant="ghost"
              size="xs"
              @click="emit('createValue', facet)"
            />
            <AdminRowActions
              v-if="canGovern"
              :items="facetMore(facet)"
              :label="`更多维度操作：${facet.name}`"
              presentation="overflow"
            />
          </div>
        </header>
        <article
          v-for="value in facet.values"
          :key="value.id"
          class="grid min-h-12 grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] items-center gap-3 border-t border-default px-3 py-2"
          data-classification-row
        >
          <div class="min-w-0 pl-4">
            <div class="flex items-center gap-2">
              <span class="text-sm font-medium text-highlighted">{{
                value.name
              }}</span
              ><UBadge
                v-if="value.status !== 'active'"
                size="xs"
                color="warning"
                variant="soft"
                :label="statusLabel[value.status] || value.status"
              />
            </div>
          </div>
          <p class="min-w-0 truncate text-xs text-muted">{{ value.slug }}</p>
          <AdminRowActions
            v-if="canGovern"
            :items="valueMore(value)"
            :label="`更多维度值操作：${value.name}`"
            presentation="overflow"
          />
        </article>
      </section>
    </div>
  </section>
</template>
