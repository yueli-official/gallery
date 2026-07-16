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
defineEmits<{
  action: [
    operation: Operation,
    kind: IdentityKind,
    item: GalleryClassificationTag,
  ];
  loadMore: [];
}>();

const statusColor = (status: string) =>
  (status === "active"
    ? "success"
    : status === "replaced"
      ? "warning"
      : "neutral") as any;
</script>

<template>
  <section aria-labelledby="classification-tag-heading">
    <div class="mb-3 flex items-end justify-between gap-3">
      <div>
        <h2
          id="classification-tag-heading"
          class="font-semibold text-highlighted"
        >
          Tag
        </h2>
        <p class="mt-1 text-sm text-muted">
          扁平长尾词；Alias 和 replacement 直接解析到 canonical Tag，不允许链。
        </p>
      </div>
      <span class="text-xs text-dimmed">keyset cursor</span>
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
              :color="statusColor(tag.status)"
              variant="soft"
              :label="tag.status"
            />
          </div>
          <p class="mt-1 truncate text-xs text-muted">
            {{ tag.slug }} · {{ tag.id }}
          </p>
          <p class="mt-1 text-xs text-dimmed">
            {{ tag.assignmentCount }} 个关系 · {{ tag.aliasCount }} 个 Alias
          </p>
        </div>
        <div class="flex flex-wrap gap-2">
          <UButton
            v-if="tag.status !== 'replaced'"
            class="min-h-11"
            color="neutral"
            variant="ghost"
            size="sm"
            :label="tag.status === 'active' ? '停用' : '启用'"
            @click="$emit('action', 'status', 'tag', tag)"
          />
          <UButton
            v-if="tag.status !== 'replaced'"
            class="min-h-11"
            color="neutral"
            variant="ghost"
            size="sm"
            label="合并"
            @click="$emit('action', 'merge', 'tag', tag)"
          />
          <UButton
            class="min-h-11"
            color="error"
            variant="ghost"
            size="sm"
            label="删除"
            @click="$emit('action', 'delete', 'tag', tag)"
          />
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
