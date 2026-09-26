<script setup lang="ts">

import { ManagePage, TabbedSurface } from "@yueli/ui/admin";
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
import {
  CollectionPanel,
  CollectionSortHeader,
  CollectionViewToggle,
} from "@yueli/ui/collection/pattern";
import type {
  GalleryAdminSubmissionPage,
  GallerySubmission,
} from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "投稿与审核 · 图库管理" });
const router = useRouter();
const { call } = useGalleryApi();
const { can } = useGalleryMe();
const hydrated = useClientHydrated();
const canReviewSubmissions = computed(() => can("gallery.submission.review"));
type SubmissionSortBy = "createdAt" | "updatedAt" | "title";
type SubmissionSortOrder = "asc" | "desc";
interface SubmissionCollectionQuery {
  view: "list" | "grid";
  q: string;
  sortBy: SubmissionSortBy;
  sortOrder: SubmissionSortOrder;
  page: number;
  size: number;
  processingState: string;
  reviewState: string;
  safetyState: string;
  outcome: string;
}
const sortByValues = ["createdAt", "updatedAt", "title"] as const;
const sortOrderValues = ["asc", "desc"] as const;
const pageSizes = [20, 40, 60] as const;
const defaultQuery: SubmissionCollectionQuery = {
  view: "list",
  q: "",
  sortBy: "createdAt",
  sortOrder: "asc",
  page: 1,
  size: 20,
  processingState: "ready",
  reviewState: "pending",
  safetyState: "",
  outcome: "pending",
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
      "/admin/submissions",
      {
        query: {
          q: nextQuery.q || undefined,
          sortBy: nextQuery.sortBy,
          sortOrder: nextQuery.sortOrder,
          page: nextQuery.page,
          size: nextQuery.size,
          processingState: nextQuery.processingState === "all" ? undefined : nextQuery.processingState || undefined,
          reviewState: nextQuery.reviewState === "all" ? undefined : nextQuery.reviewState || undefined,
          safetyState: nextQuery.safetyState || undefined,
          outcome: nextQuery.outcome === "all" ? undefined : nextQuery.outcome || undefined,
        },
      },
    );
    const lastPage = Math.max(
      1,
      galleryPageCount(data) || Math.ceil(data.total / nextQuery.size),
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
    view: { kind: "enum", values: ["list", "grid"] as const, default: defaultQuery.view },
    q: { kind: "string", default: defaultQuery.q, maxLength: 200 },
    sortBy: {
      kind: "enum",
      values: sortByValues,
      default: defaultQuery.sortBy,
    },
    sortOrder: {
      kind: "enum",
      values: sortOrderValues,
      default: defaultQuery.sortOrder,
    },
    page: { kind: "positive-integer", default: defaultQuery.page },
    size: {
      kind: "positive-integer",
      values: pageSizes,
      default: defaultQuery.size,
    },
    processingState: { kind: "string", default: defaultQuery.processingState, maxLength: 100 },
    reviewState: { kind: "string", default: defaultQuery.reviewState, maxLength: 100 },
    safetyState: { kind: "string", default: "", maxLength: 100 },
    outcome: { kind: "string", default: defaultQuery.outcome, maxLength: 100 },
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
    submissionReviewAction(item).canApprove,
  querySync,
  dataQueryKey: ({ view: _view, ...query }) => JSON.stringify(query),
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
const viewMode = computed({
  get: () => collectionQuery.value.view,
  set: (value: "list" | "grid") => updateQuery({ view: value }, false),
});
const page = computed({
  get: () => collectionQuery.value.page,
  set: (value: number) => updateQuery({ page: value }, false),
});
const size = computed({
  get: () => collectionQuery.value.size,
  set: (value: number) => updateQuery({ size: value }),
});
const q = computed(() => collectionQuery.value.q);
const sortBy = computed(() => collectionQuery.value.sortBy);
const sortOrder = computed(() => collectionQuery.value.sortOrder);
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
const rejecting = shallowRef<GallerySubmission>();

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
  { label: "审核通过", value: "approved" },
  { label: "审核未通过", value: "rejected" },
];
const outcomeItems = [
  { label: "全部结果", value: "all" },
  { label: "进行中", value: "pending" },
  { label: "已发布", value: "published" },
  { label: "重复内容", value: "duplicate" },
  { label: "审核未通过", value: "rejected" },
  { label: "已撤回", value: "withdrawn" },
  { label: "失败", value: "failed" },
];
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
    ].filter(value => Boolean(value) && value !== "all").length,
);
type SubmissionPreset = "review" | "failed" | "rejected" | "approved" | "all";
const presetQueries = {
  review: { processingState: "ready", reviewState: "pending", safetyState: "", outcome: "pending" },
  failed: { processingState: "failed", reviewState: "all", safetyState: "", outcome: "all" },
  rejected: { processingState: "all", reviewState: "rejected", safetyState: "", outcome: "all" },
  approved: { processingState: "all", reviewState: "approved", safetyState: "", outcome: "all" },
  all: { processingState: "all", reviewState: "all", safetyState: "", outcome: "all" },
} satisfies Record<SubmissionPreset, Partial<SubmissionCollectionQuery>>;
const activePreset = computed(() => {
  return (Object.keys(presetQueries) as SubmissionPreset[]).find(key =>
    Object.entries(presetQueries[key]).every(([field, value]) => collectionQuery.value[field as keyof SubmissionCollectionQuery] === value),
  ) || "custom";
});
const presetModel = computed({
  get: () => activePreset.value,
  set: (value: string) => { if (value in presetQueries) preset(value as SubmissionPreset); },
});
const presetItems = [
  { value: "review", label: "等待审核", icon: "i-tabler-inbox" },
  { value: "failed", label: "处理失败", icon: "i-tabler-alert-triangle" },
  { value: "rejected", label: "审核未通过", icon: "i-tabler-x" },
  { value: "approved", label: "审核通过", icon: "i-tabler-check" },
  { value: "all", label: "全部投稿", icon: "i-tabler-photo" },
].map(item => ({ ...item, ui: { trigger: "shrink-0 gap-1 px-2 text-xs sm:px-3 sm:text-sm", leadingIcon: "size-4" } }));

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
    processingState: "all",
    reviewState: "all",
    safetyState: "",
    outcome: "all",
    sortBy: "createdAt",
    sortOrder: "asc",
  });
}
function preset(kind: SubmissionPreset) {
  updateQuery(presetQueries[kind]);
}
function openReject(item: GallerySubmission) {
  rejecting.value = item;
}
function closeReject() {
  rejecting.value = undefined;
}
function setRejectOpen(open: boolean) {
  if (!open) closeReject();
}
async function review(
  item: GallerySubmission,
  decision: "approve" | "reject",
): Promise<boolean> {
  if (!canReviewSubmissions.value) return false;
  acting.value = item.id;
  actionErrors.value = Object.fromEntries(
    Object.entries(actionErrors.value).filter(([id]) => id !== item.id),
  );
  try {
    await call(
      `/admin/submissions/${encodeURIComponent(item.id)}/review`,
      {
        method: "POST",
        body: { decision, note: note.value[item.id] || "" },
      },
    );
    await refresh();
    return true;
  } catch (reason: any) {
    actionErrors.value = {
      ...actionErrors.value,
      [item.id]:
        galleryFailureMessage(reason, "操作没有完成；状态可能已变化，请刷新后重试。"),
    };
    return false;
  } finally {
    acting.value = "";
  }
}

async function rejectSubmission() {
  const item = rejecting.value;
  if (!item || !note.value[item.id]?.trim()) return;
  if (await review(item, "reject")) closeReject();
}

async function approveSubmission(item: GallerySubmission) {
  await review(item, "approve");
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
        failure?: import("~/utils/galleryFailure").GalleryOperationFailure;
      }>;
    }>("/admin/submissions/bulk-review", {
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
      message: failed.length ? galleryBatchFailureMessage(failed[0]?.failure, "部分项目未完成，请刷新后重试。") : undefined,
    };
    await refresh();
  } catch (reason: any) {
    bulkResult.value = {
      approved: 0,
      failed: selectedIds.value.length,
      message: galleryFailureMessage(reason, "批量请求中断，当前选择已保留。"),
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
    id: "outcome",
    label: "处理结果",
    value: outcome.value || "all",
    options: outcomeItems,
    class: "w-32",
  },
]);
function changeControl(id: string, value: CollectionControlValue) {
  if (typeof value !== "string") return;
  const normalized = value;
  if (id === "processingState") updateQuery({ processingState: normalized });
  if (id === "reviewState") updateQuery({ reviewState: normalized });
  if (id === "outcome") updateQuery({ outcome: normalized });
}

function changeColumnSort(nextSortBy: SubmissionSortBy) {
  if (sortBy.value === nextSortBy) {
    updateQuery({ sortOrder: sortOrder.value === "asc" ? "desc" : "asc" });
    return;
  }
  updateQuery({
    sortBy: nextSortBy,
    sortOrder: nextSortBy === "title" ? "asc" : "desc",
  });
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
  emptyDescription: "切换投稿状态可以查看处理进度和审核结果。",
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
  if (item.reviewNote) return item.reviewNote;
  if (item.failureCode === "derive_failed") return "公开图片生成失败";
  if (item.failureCode || item.processingState === "failed") return "图片处理失败，请重新投稿";
  if (item.outcome === "duplicate") return "重复内容已收录到现有图片";
  if (item.outcome === "published") return "已发布到图片目录";
  return submissionReviewAction(item).canApprove ? "请查看图片后通过或拒绝" : submissionReviewAction(item).reason;
}

const dateTimeFormatter = new Intl.DateTimeFormat("zh-CN", {
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
  hour12: false,
});
function formatDateTime(value?: string) {
  return value ? dateTimeFormatter.format(new Date(value)) : "—";
}
</script>

<template>
  <ManagePage id="submissions" title="投稿审核" icon="i-tabler-photo-check">
    <TabbedSurface
      v-model="presetModel"
      :items="presetItems"
      navigation-label="投稿审核队列"
      data-manage-surface="submissions"
    >
      <CollectionPanel
        compact-pagination
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
        :selectable="canReviewSubmissions && panelState === 'ready'"
        class="gallery-compact-collection rounded-none border-0 shadow-none"
        :layout="viewMode === 'grid' ? 'grid' : 'rows'"
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

        <template #view>
          <CollectionViewToggle v-model="viewMode" :items="[
            {key: 'list', label: '列表视图', icon: 'i-tabler-list'},
            {key: 'grid', label: '网格视图', icon: 'i-tabler-layout-grid'},
          ]" />
        </template>
        <template #columns>
          <GalleryGridToolbar v-if="viewMode === 'grid'"
            :sort-by="sortBy" :sort-order="sortOrder"
            :items="[{label: '提交时间', value: 'createdAt'}, {label: '更新时间', value: 'updatedAt'}, {label: '标题', value: 'title'}]"
            @update:sort-by="updateQuery({sortBy: $event})"
            @update:sort-order="updateQuery({sortOrder: $event})" />
          <div v-else
            class="grid grid-cols-[minmax(0,1fr)_7rem] items-center gap-3 md:grid-cols-[minmax(0,1fr)_9rem_7rem]"
          >
            <CollectionSortHeader
              label="投稿"
              :active="sortBy === 'title'"
              :sort-order="sortOrder"
              @sort="changeColumnSort('title')"
            />
            <CollectionSortHeader
              class="hidden md:inline-flex"
              label="提交时间"
              :active="sortBy === 'createdAt'"
              :sort-order="sortOrder"
              @sort="changeColumnSort('createdAt')"
            />
            <span class="text-right">操作</span>
          </div>
        </template>
        <template #bulk-actions
          ><UButton
            v-if="canReviewSubmissions"
            size="xs"
            icon="i-tabler-checks"
            label="批量通过"
            :loading="bulkPending"
            @click="bulkApprove"
        /></template>
        <template #item="{ item }">
          <article :class="viewMode === 'grid' ? '-m-4 overflow-hidden rounded-lg' : 'space-y-3'" data-gallery-submission-item>
            <div
              :class="viewMode === 'grid' ? 'space-y-2' : 'grid grid-cols-[minmax(0,1fr)_7rem] items-center gap-3 md:grid-cols-[minmax(0,1fr)_9rem_7rem]'"
            >
              <div :class="viewMode === 'grid' ? 'min-w-0' : 'flex min-w-0 items-center gap-3'">
                <div
                  :class="viewMode === 'grid' ? 'relative aspect-[4/3] overflow-hidden bg-elevated' : 'relative hidden aspect-[4/3] w-20 shrink-0 overflow-hidden rounded-lg bg-elevated sm:block'"
                >
                  <GallerySubmissionPreview
                    :submission-id="item.id"
                    :alt="item.altText"
                  />
                </div>
                <div :class="viewMode === 'grid' ? 'min-w-0 px-2.5 pt-2.5' : 'min-w-0'">
                  <div class="flex min-w-0 flex-wrap items-center gap-1">
                    <h2 class="truncate text-sm font-medium text-highlighted">
                      {{ item.title }}
                    </h2>
                    <UBadge
                      :color="submissionStatusColor(item)"
                      variant="soft"
                      :label="submissionStatus(item)"
                    />
                  </div>
                  <p class="mt-1 truncate text-xs text-muted">
                    {{ item.description || item.altText }}
                  </p>
                  <p
                    class="mt-1 flex items-center gap-1.5 text-xs text-dimmed"
                    data-submission-decision
                  >
                    <UIcon
                      :name="
                        submissionReviewAction(item).canApprove
                          ? 'i-tabler-circle-check'
                          : item.failureCode || item.safetyState === 'blocked'
                            ? 'i-tabler-alert-triangle'
                            : 'i-tabler-progress'
                      "
                      class="size-3.5 shrink-0"
                    />
                    <span class="truncate">{{ decisionSummary(item) }}</span>
                  </p>
                </div>
              </div>
              <time
                :class="viewMode === 'grid' ? 'block px-2.5 text-xs text-muted' : 'hidden text-xs text-muted md:block'"
                :datetime="item.createdAt"
              >
                {{ formatDateTime(item.createdAt) }}
              </time>
              <div :class="viewMode === 'grid' ? 'flex justify-end gap-1 px-2.5 pb-2.5' : 'flex justify-end gap-1'">
                <UTooltip v-if="item.imageId" text="查看公开图片">
                  <UButton
                    :to="`/images/${item.imageId}`"
                    target="_blank"
                    rel="noopener"
                    color="neutral"
                    variant="ghost"
                    size="xs"
                    icon="i-tabler-external-link"
                    square
                    :aria-label="`查看公开图片：${item.title}`"
                  />
                </UTooltip>
                <UButton
                  v-if="
                    canReviewSubmissions &&
                    item.reviewState === 'pending' &&
                    item.outcome === 'pending'
                  "
                  size="xs"
                  label="通过"
                  :loading="acting === item.id"
                  :disabled="!submissionReviewAction(item).canApprove"
                  @click="approveSubmission(item)"
                />
                <UButton
                  v-if="
                    canReviewSubmissions &&
                    item.reviewState === 'pending' &&
                    item.outcome === 'pending'
                  "
                  size="xs"
                  color="error"
                  variant="ghost"
                  label="拒绝"
                  :disabled="acting === item.id"
                  @click="openReject(item)"
                />
              </div>
            </div>

            <UAlert
              v-if="actionErrors[item.id]"
              color="error"
              variant="subtle"
              title="本项操作失败"
              :description="actionErrors[item.id]"
            />
          </article>
        </template>
      </CollectionPanel>
    </TabbedSurface>
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
    <UModal
      :open="Boolean(rejecting)"
      title="拒绝投稿"
      description="填写原因后，这条投稿将不会进入图片目录。"
      @update:open="setRejectOpen"
    >
      <template #body>
        <UFormField label="拒绝原因" required>
          <UTextarea
            v-if="rejecting"
            v-model="note[rejecting.id]"
            :rows="4"
            class="w-full"
            placeholder="说明不通过的原因"
            autofocus
          />
        </UFormField>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton
            color="neutral"
            variant="outline"
            label="取消"
            @click="closeReject"
          />
          <UButton
            color="error"
            label="确认拒绝"
            :loading="Boolean(rejecting && acting === rejecting.id)"
            :disabled="!rejecting || !note[rejecting.id]?.trim()"
            @click="rejectSubmission"
          />
        </div>
      </template>
    </UModal>
  </ManagePage>
</template>
