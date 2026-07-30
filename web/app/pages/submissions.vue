<script setup lang="ts">
import { createGalleryNotifier } from "~/utils/feedback";
import type { GallerySubmission } from "~/types/gallery";

interface SubmissionPage {
  submissions: GallerySubmission[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

const { loggedIn, login } = useAuth();
const route = useRoute();
const router = useRouter();
const hydrated = useClientHydrated();
const { call } = useGalleryApi();
const toast = createGalleryNotifier(useToast());
const page = computed(() => Math.max(1, Number(route.query.page) || 1));
const outcome = computed(() => String(route.query.outcome || ""));
const processingState = computed(() =>
  String(route.query.processingState || ""),
);
const reviewState = computed(() => String(route.query.reviewState || ""));
const { data, error, pending, refresh } = await useAsyncData(
  "gallery-my-submissions",
  () =>
    call<SubmissionPage>("/me/submissions", {
      query: {
        page: page.value,
        size: 20,
        outcome: outcome.value || undefined,
        processingState: processingState.value || undefined,
        reviewState: reviewState.value || undefined,
      },
    }),
  {
    server: false,
    watch: [page, outcome, processingState, reviewState],
    default: () => ({
      submissions: [],
      total: 0,
      page: 1,
      pageSize: 20,
      totalPages: 0,
    }),
  },
);
const withdrawing = ref("");
const outcomeItems = [
  { label: "全部结果", value: "all" },
  { label: "等待处理", value: "pending" },
  { label: "已展示", value: "published" },
  { label: "已收录", value: "duplicate" },
  { label: "未通过", value: "rejected" },
  { label: "处理失败", value: "failed" },
  { label: "已撤回", value: "withdrawn" },
];
const processingItems = [
  { label: "全部处理状态", value: "all" },
  { label: "排队中", value: "queued" },
  { label: "处理中", value: "processing" },
  { label: "处理完成", value: "ready" },
  { label: "处理失败", value: "failed" },
];
const reviewItems = [
  { label: "全部审核状态", value: "all" },
  { label: "无需人工审核", value: "not_required" },
  { label: "等待审核", value: "pending" },
  { label: "审核通过", value: "approved" },
  { label: "审核未通过", value: "rejected" },
];
const stateLabel: Record<string, string> = {
  pending: "等待处理",
  published: "已展示",
  duplicate: "已收录",
  rejected: "未通过",
  withdrawn: "已撤回",
  failed: "处理失败",
};
const processingLabel: Record<string, string> = {
  queued: "排队中",
  processing: "处理中",
  ready: "处理完成",
  failed: "处理失败",
};
const reviewLabel: Record<string, string> = {
  not_required: "无需人工审核",
  pending: "等待审核",
  approved: "审核通过",
  rejected: "审核未通过",
};
const safetyLabel: Record<string, string> = {
  pending: "等待安全检查",
  safe: "安全检查通过",
  uncertain: "需要复核",
  blocked: "安全检查未通过",
  unavailable: "安全检查不可用",
};
const failureLabel: Record<string, string> = {
  unsupported_format: "文件格式不受支持",
  animated_image: "检测到动画图片",
  processing_failed: "图片处理失败",
  safety_unavailable: "安全检查暂时不可用",
};
const stateColor = (value: string) =>
  (({
    published: "success",
    duplicate: "info",
    rejected: "error",
    failed: "error",
    withdrawn: "neutral",
  })[value] || "warning") as any;

function submissionSummary(submission: GallerySubmission) {
  if (submission.failureCode)
    return failureLabel[submission.failureCode] || submission.failureCode;
  if (submission.outcome !== "pending")
    return stateLabel[submission.outcome] || submission.outcome;
  if (submission.processingState !== "ready")
    return processingLabel[submission.processingState] || "等待媒体处理";
  if (submission.safetyState !== "safe")
    return safetyLabel[submission.safetyState] || "等待安全判断";
  if (submission.reviewState === "pending") return "等待人工审核";
  return reviewLabel[submission.reviewState] || "等待发布";
}

function setQuery(
  key: "outcome" | "processingState" | "reviewState",
  value: string,
) {
  const query = Object.fromEntries(
    Object.entries(route.query).filter(
      ([entry]) => entry !== key && entry !== "page",
    ),
  );
  if (value !== "all") query[key] = value;
  void router.push({ query });
}
function setPage(value: number) {
  const query = { ...route.query };
  if (value <= 1) delete query.page;
  else query.page = String(value);
  void router.push({ query });
}
function clearFilters() {
  void router.push({ query: {} });
}

async function withdraw(id: string) {
  withdrawing.value = id;
  try {
    await call(
      `/me/submissions/${encodeURIComponent(id)}/withdraw`,
      { method: "POST" },
    );
    await refresh();
  } catch (reason: any) {
    toast.add({
      title: "无法撤回这条投稿",
      description: reason?.data?.message || "状态可能已经变化，请刷新后重试",
      color: "error",
    });
    await refresh();
  } finally {
    withdrawing.value = "";
  }
}
useSeoMeta({ title: "我的投稿", robots: "noindex,nofollow" });
</script>

<template>
  <div class="gallery-page max-w-6xl">
    <header class="gallery-page-header">
      <div>
        <h1 class="gallery-page-title">我的投稿</h1>
        <p class="gallery-page-copy">
          跟踪每张图片的处理、审核与最终结果，失败原因会保留在对应记录中。
        </p>
      </div>
      <UButton to="/submit" icon="i-tabler-library-plus" label="批量投稿" />
    </header>

    <UAlert
      v-if="!loggedIn"
      class="mb-6"
      color="neutral"
      variant="subtle"
      icon="i-tabler-clock-shield"
      title="正在查看此浏览器的临时投稿"
      description="临时身份保留 30 天。登录后会自动把这些投稿转入你的账号。"
    >
      <template #actions>
        <UButton label="登录并长期保留" @click="void login()" />
      </template>
    </UAlert>

    <section
      class="mb-6 rounded-xl border border-default bg-default/75 p-4"
      aria-label="筛选投稿记录"
    >
      <div class="grid gap-3 sm:grid-cols-3">
        <USelect
          :model-value="outcome || 'all'"
          :items="outcomeItems"
          value-key="value"
          aria-label="按最终结果筛选"
          @update:model-value="setQuery('outcome', String($event))"
        />
        <USelect
          :model-value="processingState || 'all'"
          :items="processingItems"
          value-key="value"
          aria-label="按处理状态筛选"
          @update:model-value="setQuery('processingState', String($event))"
        />
        <USelect
          :model-value="reviewState || 'all'"
          :items="reviewItems"
          value-key="value"
          aria-label="按审核状态筛选"
          @update:model-value="setQuery('reviewState', String($event))"
        />
      </div>
      <div
        class="mt-3 flex items-center justify-between gap-3 text-xs text-muted"
      >
        <span>共 {{ data.total }} 条记录</span>
        <UButton
          v-if="outcome || processingState || reviewState"
          color="neutral"
          variant="link"
          size="xs"
          label="清除筛选"
          @click="clearFilters"
        />
      </div>
    </section>

    <div v-if="!hydrated || pending" class="space-y-3">
      <USkeleton v-for="index in 6" :key="index" class="h-32 rounded-lg" />
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
      class="overflow-hidden rounded-2xl border border-default bg-default/70 px-5 shadow-sm"
    >
      <article
        v-for="submission in data.submissions"
        :key="submission.id"
        class="grid gap-4 border-b border-default py-5 last:border-b-0 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-start"
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
          <p class="mt-2 flex items-center gap-1.5 text-xs text-muted">
            <UIcon name="i-tabler-progress-check" class="size-4 shrink-0" />
            {{ submissionSummary(submission) }}
          </p>
          <div
            v-if="submission.failureCode || submission.reviewNote"
            class="mt-3 rounded-lg bg-elevated px-3 py-2 text-sm leading-6 text-toned"
          >
            <p v-if="submission.failureCode">
              <span class="font-medium text-error">失败原因：</span
              >{{
                failureLabel[submission.failureCode] || submission.failureCode
              }}
            </p>
            <p v-if="submission.reviewNote">
              <span class="font-medium">审核说明：</span
              >{{ submission.reviewNote }}
            </p>
          </div>
          <p v-if="submission.createdAt" class="mt-3 text-xs text-dimmed">
            提交于 {{ new Date(submission.createdAt).toLocaleString("zh-CN") }}
          </p>
        </div>
        <div class="flex gap-2">
          <UButton
            v-if="submission.imageId"
            :to="`/images/${submission.imageId}`"
            color="neutral"
            variant="outline"
            size="sm"
            label="查看图片"
          />
          <UButton
            v-if="['pending', 'published'].includes(submission.outcome)"
            color="error"
            variant="ghost"
            size="sm"
            label="撤回"
            :loading="withdrawing === submission.id"
            @click="withdraw(submission.id)"
          />
        </div>
      </article>
    </div>
    <div v-else class="gallery-compact-empty grid place-items-center">
      <div>
        <span class="gallery-empty-icon"
          ><UIcon
            :name="
              outcome || processingState || reviewState
                ? 'i-tabler-filter-off'
                : 'i-tabler-photo-up'
            "
            class="size-6"
        /></span>
        <h2 class="mt-3 font-semibold text-highlighted">
          {{
            outcome || processingState || reviewState
              ? "没有符合筛选条件的记录"
              : "还没有投稿记录"
          }}
        </h2>
        <UButton
          v-if="outcome || processingState || reviewState"
          class="mt-4"
          color="neutral"
          variant="outline"
          label="清除筛选"
          @click="clearFilters"
        /><UButton v-else to="/submit" class="mt-4" label="投稿图片" />
      </div>
    </div>

    <nav
      v-if="data.totalPages > 1"
      class="mt-8 flex items-center justify-center gap-3"
      aria-label="投稿记录分页"
    >
      <UButton
        color="neutral"
        variant="outline"
        icon="i-tabler-arrow-left"
        label="上一页"
        :disabled="page <= 1"
        @click="setPage(page - 1)"
      />
      <span class="text-sm tabular-nums text-muted"
        >{{ page }} / {{ data.totalPages }}</span
      >
      <UButton
        color="neutral"
        variant="outline"
        trailing-icon="i-tabler-arrow-right"
        label="下一页"
        :disabled="page >= data.totalPages"
        @click="setPage(page + 1)"
      />
    </nav>
  </div>
</template>
