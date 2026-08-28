<script setup lang="ts">
import { ManageEmpty, SkeletonList } from "~/utils/manageComponents";
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
const createNewValue = "__create__";

const review = (
  item: GalleryClassificationTagProposal,
  decision: "approve" | "reject",
) => {
  const target = targets[item.id] || createNewValue;
  emit("review", item, decision, target === createNewValue ? "" : target);
};
const activeTags = computed(() =>
  props.tags.filter((entry) => entry.status === "active"),
);
const approvalOptions = computed(() => [
  { label: "创建新标签", value: createNewValue },
  ...activeTags.value.map((tag) => ({
    label: `归入「${tag.name}」`,
    value: tag.id,
  })),
]);

function updateTarget(itemID: string, value: unknown) {
  targets[itemID] = String(value || createNewValue);
}
</script>

<template>
  <section
    aria-labelledby="classification-proposal-heading"
    data-classification-proposals
  >
    <div class="mb-3 flex items-center gap-2">
      <h2
        id="classification-proposal-heading"
        class="font-semibold text-highlighted"
      >
        标签提案
      </h2>
      <UBadge
        size="xs"
        color="neutral"
        variant="soft"
        :label="String(total)"
      />
    </div>
    <SkeletonList v-if="!hydrated || pending" :rows="3" />
    <div
      v-else-if="proposals.length"
      class="border-y border-default"
    >
      <div
        class="hidden grid-cols-[minmax(0,1fr)_minmax(14rem,22rem)_auto] items-center gap-3 bg-elevated/50 px-3 py-2 text-xs font-medium text-muted lg:grid"
        aria-hidden="true"
      >
        <span>提案</span>
        <span>处理为</span>
        <span class="text-right">操作</span>
      </div>
      <div class="divide-y divide-default lg:border-t lg:border-default">
      <article
        v-for="item in proposals"
        :key="item.id"
        class="grid gap-3 px-3 py-3 lg:grid-cols-[minmax(0,1fr)_minmax(14rem,22rem)_auto] lg:items-center"
        data-tag-proposal-row
      >
        <p class="min-w-0 truncate font-medium text-highlighted">
          {{ item.inputValue }}
        </p>
        <USelect
          :model-value="targets[item.id] || createNewValue"
          :items="approvalOptions"
          value-key="value"
          size="sm"
          class="w-full"
          :aria-label="`选择 ${item.inputValue} 的标签处理方式`"
          @update:model-value="updateTarget(item.id, $event)"
        />
        <div class="flex justify-end gap-1">
          <UButton
            size="xs"
            color="error"
            variant="ghost"
            label="拒绝"
            :loading="reviewingId === item.id"
            @click="review(item, 'reject')"
          />
          <UButton
            size="xs"
            label="通过"
            :loading="reviewingId === item.id"
            @click="review(item, 'approve')"
          />
        </div>
      </article>
      </div>
    </div>
    <ManageEmpty
      v-else
      icon="i-tabler-tag-off"
      title="没有待审标签"
      description="新的投稿标签会显示在这里。"
    />
  </section>
</template>
