<script setup lang="ts">
import type { GallerySubmission } from "~/types/gallery";

definePageMeta({ middleware: "auth" });
const hydrated = useClientHydrated();
const { call } = useApi();
const { data, error, pending, refresh } = await useAsyncData(
  "gallery-my-submissions",
  () =>
    call<{ submissions: GallerySubmission[]; total: number }>(
      "/api/v1/gallery/me/submissions",
    ),
  { server: false, default: () => ({ submissions: [], total: 0 }) },
);
const withdrawing = ref("");
const stateLabel: Record<string, string> = {
  pending: "等待处理",
  published: "已展示",
  duplicate: "已收录",
  rejected: "未通过",
  withdrawn: "已撤回",
  failed: "处理失败",
};
const stateColor = (value: string) =>
  (({
    published: "success",
    duplicate: "info",
    rejected: "error",
    failed: "error",
    withdrawn: "neutral",
  })[value] || "warning") as any;
async function withdraw(id: string) {
  withdrawing.value = id;
  try {
    await call(
      `/api/v1/gallery/me/submissions/${encodeURIComponent(id)}/withdraw`,
      { method: "POST" },
    );
    await refresh();
  } finally {
    withdrawing.value = "";
  }
}
useSeoMeta({ title: "我的投稿", robots: "noindex,nofollow" });
</script>

<template>
  <div class="gallery-page max-w-5xl">
    <header class="mb-7 flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 class="text-3xl font-semibold tracking-tight text-highlighted">
          我的投稿
        </h1>
        <p class="mt-1.5 text-sm text-muted">
          查看处理结果，或撤回自己提交的图片。
        </p>
      </div>
      <UButton to="/submit" icon="i-tabler-plus" label="继续投稿" />
    </header>
    <div v-if="!hydrated || pending" class="space-y-3">
      <USkeleton v-for="index in 6" :key="index" class="h-28 rounded-lg" />
    </div>
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="投稿记录加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <div
      v-else-if="data.submissions.length"
      class="divide-y divide-default border-y border-default"
    >
      <article
        v-for="submission in data.submissions"
        :key="submission.id"
        class="grid gap-3 py-5 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center"
      >
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h2 class="truncate font-medium text-highlighted">
              {{ submission.title }}
            </h2>
            <UBadge
              :color="stateColor(submission.outcome)"
              variant="soft"
              :label="stateLabel[submission.outcome] || submission.outcome"
            />
          </div>
          <p class="mt-1 text-xs text-muted">
            处理 {{ submission.processingState }} · 安全
            {{ submission.safetyState }} · 审核 {{ submission.reviewState }}
          </p>
          <p v-if="submission.reviewNote" class="mt-2 text-sm text-toned">
            {{ submission.reviewNote }}
          </p>
        </div>
        <div class="flex gap-2">
          <UButton
            v-if="submission.imageId"
            :to="`/images/${submission.imageId}`"
            color="neutral"
            variant="ghost"
            label="查看图片"
          /><UButton
            v-if="['pending', 'published'].includes(submission.outcome)"
            color="error"
            variant="ghost"
            label="撤回"
            :loading="withdrawing === submission.id"
            @click="withdraw(submission.id)"
          />
        </div>
      </article>
    </div>
    <div
      v-else
      class="grid min-h-72 place-items-center border-y border-dashed border-default text-center"
    >
      <div>
        <UIcon name="i-tabler-photo-up" class="mx-auto size-8 text-dimmed" />
        <h2 class="mt-3 font-semibold text-highlighted">还没有投稿记录</h2>
        <UButton to="/submit" class="mt-4" label="投稿一张图片" />
      </div>
    </div>
  </div>
</template>
