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
    <div class="mb-3 flex items-end justify-between gap-3">
      <div>
        <h2
          id="classification-proposal-heading"
          class="font-semibold text-highlighted"
        >
          标签提案
        </h2>
        <p class="mt-1 text-sm text-muted">
          投稿中未收录的标签会进入这里；通过时创建新标签或归入现有标签。
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
          <details class="mt-1 text-xs text-dimmed">
            <summary class="cursor-pointer">来源与标识</summary>
            <p class="mt-1 break-all font-mono">
              {{ item.lookupKey }} · {{ item.submissionId }}
            </p>
          </details>
        </div>
        <UFormField label="处理为">
          <USelect
            :model-value="targets[item.id] || createNewValue"
            :items="approvalOptions"
            value-key="value"
            class="w-full"
            :aria-label="`选择 ${item.inputValue} 的标签处理方式`"
            @update:model-value="updateTarget(item.id, $event)"
          />
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
            label="通过"
            :loading="reviewingId === item.id"
            @click="review(item, 'approve')"
          />
        </div>
      </article>
    </div>
    <ManageEmpty
      v-else
      icon="i-tabler-tag-off"
      title="没有待审标签"
      description="投稿中的新词会先进入这里，批准后才加入公开目录。"
    />
  </section>
</template>
