<script setup lang="ts">
import { PageHeader } from "@yueli/ui/dashboard/pattern";
import { ManageTabs } from "@platform/manage/components";
import {
  createCollectionRouteQueryCodec,
  createJsonCollectionQueryPolicy,
  type CollectionControl,
  type CollectionControlValue,
  type CollectionPanelMessages,
  type CollectionPanelState,
  type CollectionWorkflow,
} from "@yueli/ui/collection";
import { useVueCollectionWorkflow } from "@yueli/ui/collection/vue";
import { createVueRouterCollectionQuerySync } from "@yueli/ui/collection/vue-router";
import { CollectionPanel } from "@yueli/ui/collection/pattern";
import type {
  GalleryAdminSubmissionPage,
  GallerySubmission,
} from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "投稿与审核 · 图库管理" });
const router = useRouter();
const { call } = useApi();
const { can } = useGalleryMe();
const hydrated = useClientHydrated();
const canReviewSubmissions = computed(() => can("gallery.submission.review"));
type SubmissionSort = "oldest" | "newest" | "updated";
interface SubmissionCollectionQuery {
  q: string;
  sort: SubmissionSort;
  page: number;
  size: number;
  processingState: string;
  reviewState: string;
  safetyState: string;
  outcome: string;
}
const sorts = ["oldest", "newest", "updated"] as const;
const pageSizes = [20, 40, 60] as const;
const defaultQuery: SubmissionCollectionQuery = {
  q: "",
  sort: "oldest",
  page: 1,
  size: 20,
  processingState: "",
  reviewState: "",
  safetyState: "",
  outcome: "",
};
async function loadSubmissions(
  nextQuery: Readonly<SubmissionCollectionQuery>,
  activeWorkflow: CollectionWorkflow<
    GallerySubmission,
    string,
    SubmissionCollectionQuery
  >,
) {
  const token = activeWorkflow.beginLoad();
  try {
    const data = await call<GalleryAdminSubmissionPage>(
      "/api/v1/gallery/admin/submissions",
      {
        query: {
          q: nextQuery.q || undefined,
          sort: nextQuery.sort,
          page: nextQuery.page,
          size: nextQuery.size,
          processingState: nextQuery.processingState || undefined,
          reviewState: nextQuery.reviewState || undefined,
          safetyState: nextQuery.safetyState || undefined,
          outcome: nextQuery.outcome || undefined,
        },
      },
    );
    const lastPage = Math.max(
      1,
      data.totalPages || Math.ceil(data.total / nextQuery.size),
    );
    if (nextQuery.page > lastPage) {
      activeWorkflow.setQuery({ ...nextQuery, page: lastPage });
      return;
    }
    activeWorkflow.resolveLoad(token, {
      items: data.items ?? [],
      total: data.total ?? 0,
    });
  } catch {
    activeWorkflow.rejectLoad(token, {
      key: "gallery.submissions.collection.load_failed",
    });
  }
}
const querySync = createVueRouterCollectionQuerySync({
  router,
  codec: createCollectionRouteQueryCodec({
    q: { kind: "string", default: defaultQuery.q, maxLength: 200 },
    sort: { kind: "enum", values: sorts, default: defaultQuery.sort },
    page: { kind: "positive-integer", default: defaultQuery.page },
    size: {
      kind: "positive-integer",
      values: pageSizes,
      default: defaultQuery.size,
    },
    processingState: { kind: "string", default: "", maxLength: 100 },
    reviewState: { kind: "string", default: "", maxLength: 100 },
    safetyState: { kind: "string", default: "", maxLength: 100 },
    outcome: { kind: "string", default: "", maxLength: 100 },
  }),
});
const {
  snapshot: submissionCollection,
  workflow: submissionWorkflow,
  reload: refresh,
} = useVueCollectionWorkflow({
  initialQuery: defaultQuery,
  queryPolicy: createJsonCollectionQueryPolicy<SubmissionCollectionQuery>(),
  keyOf: (item: GallerySubmission) => item.id,
  isSelectable: (item: GallerySubmission) =>
    canReviewSubmissions.value && submissionReviewAction(item).canApprove,
  querySync,
  dataQueryKey: (query) => JSON.stringify(query),
  load: loadSubmissions,
});
const collectionQuery = computed(() => submissionCollection.value.query);
function updateQuery(
  patch: Partial<SubmissionCollectionQuery>,
  resetPage = true,
) {
  submissionWorkflow.setQuery({
    ...collectionQuery.value,
    ...patch,
    ...(resetPage ? { page: 1 } : {}),
  });
}
const page = computed({
  get: () => collectionQuery.value.page,
  set: (value: number) => updateQuery({ page: value }, false),
});
const size = computed({
  get: () => collectionQuery.value.size,
  set: (value: number) => updateQuery({ size: value }),
});
const q = computed(() => collectionQuery.value.q);
const sort = computed(() => collectionQuery.value.sort);
const processingState = computed(() => collectionQuery.value.processingState);
const reviewState = computed(() => collectionQuery.value.reviewState);
const safetyState = computed(() => collectionQuery.value.safetyState);
const outcome = computed(() => collectionQuery.value.outcome);
const qDraft = ref(q.value);
watch(q, (value) => {
  if (qDraft.value !== value) qDraft.value = value;
});
let searchTimer: ReturnType<typeof setTimeout> | undefined;
watch(qDraft, (value) => {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(() => updateQuery({ q: value.trim() }), 300);
});
onScopeDispose(() => {
  if (searchTimer) clearTimeout(searchTimer);
});

const note = ref<Record<string, string>>({});
const acting = ref("");
const actionErrors = ref<Record<string, string>>({});

const sortItems = [
  { label: "等待最久", value: "oldest" },
  { label: "最新投稿", value: "newest" },
  { label: "最近变化", value: "updated" },
];
const processingItems = [
  { label: "全部处理状态", value: "all" },
  { label: "排队中", value: "queued" },
  { label: "处理中", value: "processing" },
  { label: "已就绪", value: "ready" },
  { label: "处理失败", value: "failed" },
];
const reviewItems = [
  { label: "全部审核状态", value: "all" },
  { label: "无需审核", value: "not_required" },
  { label: "等待审核", value: "pending" },
  { label: "已批准", value: "approved" },
  { label: "已拒绝", value: "rejected" },
];
const safetyItems = [
  { label: "全部安全状态", value: "all" },
  { label: "等待检查", value: "pending" },
  { label: "安全", value: "safe" },
  { label: "不确定", value: "uncertain" },
  { label: "已阻止", value: "blocked" },
  { label: "不可用", value: "unavailable" },
];
const outcomeItems = [
  { label: "全部结果", value: "all" },
  { label: "进行中", value: "pending" },
  { label: "已发布", value: "published" },
  { label: "重复内容", value: "duplicate" },
  { label: "已拒绝", value: "rejected" },
  { label: "已撤回", value: "withdrawn" },
  { label: "失败", value: "failed" },
];
const safetyLabel = Object.fromEntries(
  safetyItems.slice(1).map((item) => [item.value, item.label]),
);
const outcomeLabel = Object.fromEntries(
  outcomeItems.slice(1).map((item) => [item.value, item.label]),
);
const filterCount = computed(
  () =>
    [
      processingState.value,
      reviewState.value,
      safetyState.value,
      outcome.value,
    ].filter(Boolean).length,
);
const activePreset = computed(() => {
  if (
    reviewState.value === "pending" &&
    outcome.value === "pending" &&
    !processingState.value &&
    !safetyState.value
  )
    return "review";
  if (
    processingState.value === "failed" &&
    !reviewState.value &&
    !safetyState.value &&
    !outcome.value
  )
    return "failed";
  if (
    safetyState.value === "uncertain" &&
    outcome.value === "pending" &&
    !processingState.value &&
    !reviewState.value
  )
    return "uncertain";
  if (!filterCount.value) return "all";
  return "custom";
});
const presetModel = computed({
  get: () => activePreset.value,
  set: (value: string) => {
    if (["review", "failed", "uncertain", "all"].includes(value))
      preset(value as "review" | "failed" | "uncertain" | "all");
  },
});
const presetItems = [
  { key: "review", label: "待审核" },
  { key: "failed", label: "处理失败" },
  { key: "uncertain", label: "安全不确定" },
  { key: "all", label: "全部投稿" },
];

const selectedIds = computed<readonly string[]>(() =>
  submissionCollection.value.selection.mode === "keys"
    ? submissionCollection.value.selection.keys
    : [],
);
const selectionCount = computed(
  () => submissionCollection.value.selection.count,
);
const isPageSelected = computed(
  () => submissionCollection.value.isPageSelected,
);
const isPageIndeterminate = computed(
  () => submissionCollection.value.isPageIndeterminate,
);
function toggleOne(id: string) {
  submissionWorkflow.toggleKey(id);
}
function togglePage(selected: boolean) {
  submissionWorkflow.togglePage(selected);
}
function clearSelection() {
  submissionWorkflow.clearSelection();
}
function replaceSelection(ids: readonly string[]) {
  submissionWorkflow.clearSelection();
  for (const id of ids) submissionWorkflow.toggleKey(id);
}
const bulkPending = ref(false);
const bulkResult = ref<{
  approved: number;
  failed: number;
  message?: string;
}>();

function search() {
  if (searchTimer) clearTimeout(searchTimer);
  updateQuery({ q: qDraft.value.trim() });
}
function clearFilters() {
  qDraft.value = "";
  updateQuery({
    q: "",
    processingState: "",
    reviewState: "",
    safetyState: "",
    outcome: "",
    sort: "oldest",
  });
}
function preset(kind: "review" | "failed" | "uncertain" | "all") {
  if (kind === "review")
    updateQuery({
      processingState: "",
      reviewState: "pending",
      safetyState: "",
      outcome: "pending",
    });
  else if (kind === "failed")
    updateQuery({
      processingState: "failed",
      reviewState: "",
      safetyState: "",
      outcome: "",
    });
  else if (kind === "uncertain")
    updateQuery({
      processingState: "",
      reviewState: "",
      safetyState: "uncertain",
      outcome: "pending",
    });
  else
    updateQuery({
      processingState: "",
      reviewState: "",
      safetyState: "",
      outcome: "",
    });
}
async function review(item: GallerySubmission, decision: "approve" | "reject") {
  if (!canReviewSubmissions.value) return;
  acting.value = item.id;
  actionErrors.value = Object.fromEntries(
    Object.entries(actionErrors.value).filter(([id]) => id !== item.id),
  );
  try {
    await call(
      `/api/v1/gallery/admin/submissions/${encodeURIComponent(item.id)}/review`,
      {
        method: "POST",
        body: { decision, note: note.value[item.id] || "" },
      },
    );
    await refresh();
  } catch (reason: any) {
    actionErrors.value = {
      ...actionErrors.value,
      [item.id]:
        reason?.data?.message || "操作没有完成；状态可能已变化，请刷新后重试。",
    };
  } finally {
    acting.value = "";
  }
}

async function bulkApprove() {
  if (!canReviewSubmissions.value || !selectedIds.value.length) return;
  bulkPending.value = true;
  bulkResult.value = undefined;
  try {
    const response = await call<{
      results: Array<{
        submissionId: string;
        success: boolean;
        error?: string;
      }>;
    }>("/api/v1/gallery/admin/submissions/bulk-review", {
      method: "POST",
      body: {
        submissionIds: selectedIds.value,
        decision: "approve",
        note: "",
      },
    });
    const failed = response.results.filter((item) => !item.success);
    replaceSelection(failed.map((item) => item.submissionId));
    bulkResult.value = {
      approved: response.results.length - failed.length,
      failed: failed.length,
    };
    await refresh();
  } catch (reason: any) {
    bulkResult.value = {
      approved: 0,
      failed: selectedIds.value.length,
      message: reason?.data?.message || "批量请求中断，当前选择已保留。",
    };
  } finally {
    bulkPending.value = false;
  }
}

const controls = computed<CollectionControl[]>(() => [
  {
    kind: "select",
    id: "processingState",
    label: "处理状态",
    value: processingState.value || "all",
    options: processingItems,
    class: "w-32",
  },
  {
    kind: "select",
    id: "reviewState",
    label: "审核状态",
    value: reviewState.value || "all",
    options: reviewItems,
    class: "w-32",
  },
  {
    kind: "select",
    id: "safetyState",
    label: "安全状态",
    value: safetyState.value || "all",
    options: safetyItems,
    class: "w-32",
  },
  {
    kind: "select",
    id: "outcome",
    label: "处理结果",
    value: outcome.value || "all",
    options: outcomeItems,
    class: "w-32",
  },
  {
    kind: "select",
    id: "sort",
    label: "投稿排序",
    value: sort.value,
    options: sortItems,
    class: "w-28",
  },
]);
function changeControl(id: string, value: CollectionControlValue) {
  if (typeof value !== "string") return;
  const normalized = value === "all" ? "" : value;
  if (id === "processingState") updateQuery({ processingState: normalized });
  if (id === "reviewState") updateQuery({ reviewState: normalized });
  if (id === "safetyState") updateQuery({ safetyState: normalized });
  if (id === "outcome") updateQuery({ outcome: normalized });
  if (id === "sort" && sorts.includes(value as SubmissionSort))
    updateQuery({ sort: value as SubmissionSort });
}
const messages: CollectionPanelMessages = {
  searchPlaceholder: "搜索标题、说明或来源…",
  searchAction: "搜索",
  filtersAction: "状态筛选",
  activeFilters: (count) => `状态筛选（${count}）`,
  clearFilters: "清除筛选",
  selectPage: "选择本页可批准投稿",
  selectItem: (label) => `选择投稿：${label}`,
  bulkRegion: "投稿批量审核",
  selected: (count) => `已选择 ${count} 条投稿`,
  selectAllResults: "选择全部结果",
  clearSelection: "取消选择",
  emptyTitle: "当前筛选没有投稿",
  emptyDescription: "切换处理、安全或结果状态可以查看历史记录。",
  errorTitle: "投稿队列加载失败",
  retry: "重新加载",
  showing: (first, last, count) => `显示 ${first}–${last}，共 ${count} 条`,
  pageSize: "每页",
  pageSizeControl: "每页投稿数量",
  pageSizeOption: (value) => `${value} 条`,
};
const panelState = computed<CollectionPanelState>(() =>
  submissionCollection.value.issue
    ? "error"
    : !hydrated.value ||
        ["idle", "loading", "refreshing"].includes(
          submissionCollection.value.loadState,
        )
      ? "loading"
      : "ready",
);
const submissionKey = (item: GallerySubmission) => item.id;
const submissionLabel = (item: GallerySubmission) => item.title;

function decisionSummary(item: GallerySubmission) {
  if (item.failureCode || item.processingState === "failed")
    return "媒体处理失败";
  if (item.safetyState === "blocked") return "安全判断已阻止";
  if (item.safetyState === "uncertain") return "需要复核安全判断";
  if (submissionReviewAction(item).canApprove)
    return "已完成检查，可以批准进入目录";
  if (item.outcome === "published") return "已发布到图片目录";
  if (item.outcome === "duplicate") return "重复内容已合并到现有图片";
  return outcomeLabel[item.outcome] || "等待处理";
}
</script>

<template>
  <div>
    <PageHeader title="投稿审核">
      <template #subtitle>
        先处理能进入目录的投稿；媒体失败和安全不确定保留为独立队列。
      </template>
    </PageHeader>

    <ManageTabs v-model="presetModel" :items="presetItems" class="mb-4" />
    <CollectionPanel
      v-model:search="qDraft"
      :items="submissionCollection.items"
      :item-key="submissionKey"
      :item-label="submissionLabel"
      :controls="controls"
      :messages="messages"
      :state="panelState"
      error-message="请确认 Gallery API、数据库和登录状态正常。"
      :total="submissionCollection.total"
      :page="page"
      :page-size="size"
      :page-sizes="pageSizes"
      :active-filter-count="filterCount"
      :selection-count="selectionCount"
      :page-selected="isPageSelected"
      :page-indeterminate="isPageIndeterminate"
      :is-selected="submissionWorkflow.isSelected"
      :is-item-selectable="(item) => submissionReviewAction(item).canApprove"
      label="投稿审核队列"
      :selectable="canReviewSubmissions"
      @search="search"
      @control-change="changeControl"
      @clear-filters="clearFilters"
      @retry="refresh"
      @toggle-page="togglePage"
      @toggle-item="toggleOne"
      @clear-selection="clearSelection"
      @page-change="page = $event"
      @page-size-change="size = $event"
    >
      <template #columns
        ><div class="grid grid-cols-[minmax(0,1fr)_auto] gap-3">
          <span>投稿、检查与审核状态</span
          ><span class="hidden w-48 text-right xl:block">审核操作</span>
        </div></template
      >
      <template #bulk-actions
        ><UButton
          v-if="canReviewSubmissions"
          size="xs"
          icon="i-tabler-checks"
          label="批量批准"
          :loading="bulkPending"
          @click="bulkApprove"
      /></template>
      <template #item="{ item }">
        <article
          class="grid grid-cols-[5rem_minmax(0,1fr)] gap-3 sm:grid-cols-[6.5rem_minmax(0,1fr)] xl:grid-cols-[6.5rem_minmax(0,1fr)_19rem] xl:items-start"
        >
          <div
            class="relative aspect-[4/3] self-start overflow-hidden rounded-lg bg-elevated"
          >
            <GallerySubmissionPreview
              :submission-id="item.id"
              :alt="item.altText"
            />
            <span
              class="absolute bottom-1.5 left-1.5 rounded bg-default/90 px-1.5 py-0.5 text-[11px] font-medium text-default backdrop-blur"
            >
              {{ outcomeLabel[item.outcome] }}
            </span>
          </div>
          <div class="min-w-0">
            <div class="flex min-w-0 items-center gap-2">
              <h2 class="truncate font-semibold text-highlighted">
                {{ item.title }}
              </h2>
              <UBadge
                v-if="item.safetyState !== 'safe'"
                :color="item.safetyState === 'blocked' ? 'error' : 'warning'"
                variant="soft"
                :label="safetyLabel[item.safetyState]"
              />
            </div>
            <p
              v-if="item.description"
              class="mt-1 line-clamp-2 text-sm leading-5 text-muted"
            >
              {{ item.description }}
            </p>
            <p class="mt-2 flex items-center gap-1.5 text-xs text-muted">
              <UIcon
                :name="
                  submissionReviewAction(item).canApprove
                    ? 'i-tabler-circle-check'
                    : item.failureCode || item.safetyState === 'blocked'
                      ? 'i-tabler-alert-triangle'
                      : 'i-tabler-progress'
                "
                class="size-4 shrink-0"
              />
              {{ decisionSummary(item) }}
            </p>
            <UAlert
              v-if="item.failureCode"
              class="mt-3"
              color="error"
              variant="subtle"
              icon="i-tabler-alert-triangle"
              title="媒体处理失败"
              :description="item.failureCode"
            />
            <p
              v-if="item.reviewNote"
              class="mt-3 border-l-2 border-default pl-3 text-sm text-muted"
            >
              审核记录：{{ item.reviewNote }}
            </p>
            <UAlert
              v-if="actionErrors[item.id]"
              class="mt-3"
              color="error"
              variant="subtle"
              title="本项操作失败"
              :description="actionErrors[item.id]"
            />
          </div>
          <div class="col-span-2 sm:col-span-1 sm:col-start-2 xl:col-start-3">
            <div class="flex flex-wrap items-center gap-2 xl:justify-end">
              <UButton
                v-if="
                  canReviewSubmissions &&
                  item.reviewState === 'pending' &&
                  item.outcome === 'pending'
                "
                :label="submissionReviewAction(item).label"
                :loading="acting === item.id"
                :disabled="!submissionReviewAction(item).canApprove"
                @click="review(item, 'approve')"
              />
              <UButton
                v-if="item.imageId"
                :to="`/images/${item.imageId}`"
                target="_blank"
                color="neutral"
                variant="ghost"
                icon="i-tabler-external-link"
                aria-label="查看公开图片"
              />
            </div>
            <p
              v-if="submissionReviewAction(item).reason"
              class="mt-2 text-xs leading-5 text-muted xl:text-right"
            >
              {{ submissionReviewAction(item).reason }}
            </p>
            <details
              v-if="
                canReviewSubmissions &&
                item.reviewState === 'pending' &&
                item.outcome === 'pending'
              "
              class="group mt-3 rounded-lg border border-default bg-elevated/35 px-3 py-2"
            >
              <summary
                class="flex min-h-8 cursor-pointer list-none items-center justify-between gap-2 text-sm font-medium text-default"
              >
                备注或拒绝
                <UIcon
                  name="i-tabler-chevron-down"
                  class="size-4 text-muted transition group-open:rotate-180"
                />
              </summary>
              <div class="space-y-2 border-t border-default pt-3">
                <UTextarea
                  v-model="note[item.id]"
                  :rows="3"
                  placeholder="记录判断；拒绝时必须填写原因"
                />
                <UButton
                  color="error"
                  variant="outline"
                  label="拒绝"
                  :loading="acting === item.id"
                  :disabled="!note[item.id]?.trim()"
                  @click="review(item, 'reject')"
                />
              </div>
            </details>
          </div>
        </article>
      </template>
    </CollectionPanel>
    <div
      v-if="bulkResult"
      class="mt-3 flex items-center justify-between gap-2 rounded-lg border border-default bg-elevated px-3 py-2.5 text-xs"
      :class="bulkResult.failed ? 'text-warning' : 'text-success'"
      role="status"
    >
      <span>{{
        bulkResult.message ||
        `已批准 ${bulkResult.approved} 条${bulkResult.failed ? `，${bulkResult.failed} 条失败并保留选择` : ""}`
      }}</span>
      <UButton
        icon="i-tabler-x"
        color="neutral"
        variant="ghost"
        size="xs"
        square
        aria-label="关闭批量结果"
        @click="bulkResult = undefined"
      />
    </div>
  </div>
</template>
