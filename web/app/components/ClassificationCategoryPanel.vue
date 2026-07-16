<script setup lang="ts">
import type {
  GalleryClassificationCatalogNode,
  GalleryClassificationGovernanceCommand,
} from "~/types/gallery";

type IdentityKind = GalleryClassificationGovernanceCommand["kind"];
type Operation = "status" | "reparent" | "merge" | "delete";

defineProps<{ items: GalleryClassificationCatalogNode[] }>();
defineEmits<{
  action: [
    operation: Operation,
    kind: IdentityKind,
    item: GalleryClassificationCatalogNode,
  ];
}>();

const statusColor = (status: string) =>
  (status === "active"
    ? "success"
    : status === "replaced"
      ? "warning"
      : "neutral") as any;
</script>

<template>
  <section aria-labelledby="classification-category-heading">
    <div class="mb-3 flex items-end justify-between gap-3">
      <div>
        <h2
          id="classification-category-heading"
          class="font-semibold text-highlighted"
        >
          Category
        </h2>
        <p class="mt-1 text-sm text-muted">
          多归属、单父层级；主分类保存在独立 companion 关系中。
        </p>
      </div>
      <span class="text-xs text-dimmed">{{ items.length }} 项</span>
    </div>
    <div class="divide-y divide-default border-y border-default">
      <article
        v-for="item in items"
        :key="item.id"
        class="grid gap-3 py-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-center"
      >
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h3 class="font-medium text-highlighted">{{ item.name }}</h3>
            <UBadge
              :color="statusColor(item.status)"
              variant="soft"
              :label="item.status"
            />
          </div>
          <p class="mt-1 truncate text-sm text-muted">
            {{ item.slug }} · {{ item.id }}
          </p>
          <p v-if="item.parentId" class="mt-1 text-xs text-dimmed">
            父节点 {{ item.parentId }}
          </p>
        </div>
        <div class="flex flex-wrap gap-2">
          <UButton
            v-if="item.status !== 'replaced'"
            class="min-h-11"
            color="neutral"
            variant="outline"
            size="sm"
            :label="item.status === 'active' ? '停用' : '启用'"
            @click="$emit('action', 'status', 'category', item)"
          />
          <UButton
            v-if="item.status !== 'replaced'"
            class="min-h-11"
            color="neutral"
            variant="ghost"
            size="sm"
            label="移动"
            @click="$emit('action', 'reparent', 'category', item)"
          />
          <UButton
            v-if="item.status !== 'replaced'"
            class="min-h-11"
            color="neutral"
            variant="ghost"
            size="sm"
            label="合并"
            @click="$emit('action', 'merge', 'category', item)"
          />
          <UButton
            class="min-h-11"
            color="error"
            variant="ghost"
            size="sm"
            label="删除"
            @click="$emit('action', 'delete', 'category', item)"
          />
        </div>
      </article>
    </div>
  </section>
</template>
