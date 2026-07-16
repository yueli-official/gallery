<script setup lang="ts">
import { ManageEmpty, SkeletonList } from "@platform/manage/components";
import type {
  GalleryClassificationTag,
  GalleryClassificationTagProposal,
} from "~/types/gallery";

const props = defineProps<{
  proposals: GalleryClassificationTagProposal[];
  total: number;
  tags: GalleryClassificationTag[];
  pending: boolean;
  hydrated: boolean;
  reviewingId: string;
}>();
const emit = defineEmits<{
  review: [
    item: GalleryClassificationTagProposal,
    decision: "approve" | "reject",
    targetTagId: string,
  ];
}>();
const targets = reactive<Record<string, string>>({});

const review = (
  item: GalleryClassificationTagProposal,
  decision: "approve" | "reject",
) => emit("review", item, decision, targets[item.id] || "");
const activeTags = computed(() =>
  props.tags.filter((entry) => entry.status === "active"),
);
</script>

<template>
  <section aria-labelledby="classification-proposal-heading">
    <div class="mb-3 flex items-end justify-between gap-3">
      <div>
        <h2
          id="classification-proposal-heading"
          class="font-semibold text-highlighted"
        >
          Tag Proposal
        </h2>
        <p class="mt-1 text-sm text-muted">
          批准时重新解析最新 Lookup Registry；创建 canonical
          Tag，或显式归并为已有 Tag 的 Alias。
        </p>
      </div>
      <span class="text-xs text-dimmed">{{ total }} 条待审</span>
    </div>
    <SkeletonList v-if="!hydrated || pending" :rows="3" />
    <div
      v-else-if="proposals.length"
      class="divide-y divide-default border-y border-default"
    >
      <article
        v-for="item in proposals"
        :key="item.id"
        class="grid gap-3 py-4 lg:grid-cols-[minmax(0,1fr)_minmax(14rem,22rem)_auto] lg:items-center"
      >
        <div class="min-w-0">
          <p class="font-medium text-highlighted">{{ item.inputValue }}</p>
          <p class="mt-1 truncate text-xs text-muted">
            lookup: {{ item.lookupKey }} · submission {{ item.submissionId }}
          </p>
        </div>
        <UFormField label="批准方式">
          <select
            v-model="targets[item.id]"
            class="h-11 w-full rounded-md border border-default bg-default px-3 text-sm text-highlighted"
            :aria-label="`选择 ${item.inputValue} 的 Tag 处理方式`"
          >
            <option value="">批准并创建新 canonical Tag</option>
            <option v-for="tag in activeTags" :key="tag.id" :value="tag.id">
              作为 {{ tag.name }} 的 Alias
            </option>
          </select>
        </UFormField>
        <div class="flex gap-2">
          <UButton
            class="min-h-11"
            color="neutral"
            variant="outline"
            label="拒绝"
            :loading="reviewingId === item.id"
            @click="review(item, 'reject')"
          />
          <UButton
            class="min-h-11"
            label="批准"
            :loading="reviewingId === item.id"
            @click="review(item, 'approve')"
          />
        </div>
      </article>
    </div>
    <ManageEmpty
      v-else
      icon="i-tabler-tag-off"
      title="没有待审 Tag"
      description="未知投稿词会进入独立 Proposal，不会提前污染公开目录。"
    />
  </section>
</template>
