<script setup lang="ts">
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
        label: item.status === "active" ? "停用维度" : "启用维度",
        icon: item.status === "active" ? "i-tabler-eye-off" : "i-tabler-eye",
        onSelect: () => emit("action", "status", "facet", item),
      },
    ],
    [
      {
        label: "删除维度",
        icon: "i-tabler-trash",
        color: "error" as const,
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
            label: item.status === "active" ? "停用维度值" : "启用维度值",
            icon:
              item.status === "active" ? "i-tabler-eye-off" : "i-tabler-eye",
            onSelect: () => emit("action", "status", "facet_value", item),
          },
          {
            label: "移动到其他父级",
            icon: "i-tabler-arrows-move",
            onSelect: () => emit("action", "reparent", "facet_value", item),
          },
          {
            label: "合并到其他值",
            icon: "i-tabler-git-merge",
            onSelect: () => emit("action", "merge", "facet_value", item),
          },
        ];
  return [
    regular,
    [
      {
        label: "删除维度值",
        icon: "i-tabler-trash",
        color: "error" as const,
        onSelect: () => emit("action", "delete", "facet_value", item),
      },
    ],
  ].filter((group) => group.length);
}
</script>

<template>
  <section aria-labelledby="classification-facet-heading">
    <div class="mb-3">
      <h2
        id="classification-facet-heading"
        class="font-semibold text-highlighted"
      >
        维度
      </h2>
      <p class="mt-1 text-sm text-muted">
        每个维度是一条筛选轴，维度值按行扫描和治理。
      </p>
    </div>
    <div class="overflow-hidden rounded-xl border border-default bg-default">
      <section
        v-for="facet in facets"
        :key="facet.id"
        class="border-b border-default last:border-b-0"
      >
        <header
          class="flex flex-wrap items-center justify-between gap-3 bg-elevated/50 px-4 py-3"
        >
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <h3 class="font-semibold text-highlighted">{{ facet.name }}</h3>
              <UBadge
                v-if="facet.status !== 'active'"
                color="warning"
                variant="soft"
                :label="statusLabel[facet.status] || facet.status"
              />
            </div>
            <p class="mt-0.5 text-xs text-muted">
              {{ facet.slug }} · {{ facet.values.length }} 个值
            </p>
          </div>
          <div class="flex items-center gap-1">
            <UDropdownMenu v-if="canGovern" :items="facetMore(facet)"
              ><UButton
                class="min-h-11"
                color="neutral"
                variant="ghost"
                size="sm"
                icon="i-tabler-dots"
                square
                :aria-label="`更多维度操作：${facet.name}`"
            /></UDropdownMenu>
          </div>
        </header>
        <article
          v-for="value in facet.values"
          :key="value.id"
          class="grid min-h-16 grid-cols-[minmax(0,1fr)_auto] items-center gap-3 border-t border-default px-4 py-2.5 pl-8"
        >
          <div class="min-w-0">
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
            <p class="mt-0.5 text-xs text-muted">{{ value.slug }}</p>
            <details class="mt-1 text-xs text-dimmed">
              <summary class="cursor-pointer">标识信息</summary>
              <p class="mt-1 break-all font-mono">{{ value.id }}</p>
            </details>
          </div>
          <div class="flex items-center gap-1">
            <UDropdownMenu v-if="canGovern" :items="valueMore(value)"
              ><UButton
                class="min-h-11"
                color="neutral"
                variant="ghost"
                size="xs"
                icon="i-tabler-dots"
                square
                :aria-label="`更多维度值操作：${value.name}`"
            /></UDropdownMenu>
          </div>
        </article>
      </section>
    </div>
  </section>
</template>
