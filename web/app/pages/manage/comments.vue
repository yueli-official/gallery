<script setup lang="ts">
import { ManagePage } from "@yueli/ui/admin";
import {
  createCollectionRouteQueryCodec,
  createJsonCollectionQueryPolicy,
  type CollectionWorkflow,
} from "@yueli/ui/collection";
import { useVueCollectionWorkflow } from "@yueli/ui/collection/vue";
import { createVueRouterCollectionQuerySync } from "@yueli/ui/collection/vue-router";
import { CommentModerationCollection } from "@yueli/ui/comments/admin";
import type {
  CommentModerationCollectionActions,
  CommentModerationCollectionModel,
  CommentModerationItem,
  CommentModerationLifecycle,
} from "@yueli/ui/comments/admin";
import type {
  GalleryAdminComment,
  GalleryAdminCommentPage,
  GalleryCommentStatus,
} from "~/types/gallery";
import { createGalleryNotifier } from "~/utils/feedback";

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "评论 · 图库管理" });

type CommentSortOrder = "asc" | "desc";
interface CommentQuery {
  q: string;
  status: "" | GalleryCommentStatus;
  sortOrder: CommentSortOrder;
  page: number;
  size: number;
}

const router = useRouter();
const { call } = useGalleryApi();
const toast = createGalleryNotifier(useToast());
const { can, isAdministrator } = useGalleryMe();
const canModerate = computed(
  () => can("gallery.comment.moderate") || isAdministrator.value,
);
const canDelete = computed(
  () => can("gallery.comment.delete") || isAdministrator.value,
);
const pageSizes = [20, 40, 60] as const;
const statuses = ["", "pending", "approved", "spam", "trash"] as const;
const defaultQuery: CommentQuery = {
  q: "",
  status: "",
  sortOrder: "desc",
  page: 1,
  size: 20,
};

async function loadComments(
  query: Readonly<CommentQuery>,
  activeWorkflow: CollectionWorkflow<
    GalleryAdminComment,
    string,
    CommentQuery
  >,
) {
  const token = activeWorkflow.beginLoad();
  try {
    const response = await call<GalleryAdminCommentPage>("/admin/comments", {
      query: {
        q: query.q || undefined,
        status: query.status || undefined,
        sortBy: "createdAt",
        sortOrder: query.sortOrder,
        page: query.page,
        size: query.size,
      },
    });
    const lastPage = Math.max(
      1,
      Math.ceil(response.total / Math.max(1, query.size)),
    );
    if (query.page > lastPage) {
      activeWorkflow.setQuery({ ...query, page: lastPage });
      return;
    }
    activeWorkflow.resolveLoad(token, {
      items: response.items || [],
      total: response.total || 0,
    });
  } catch {
    activeWorkflow.rejectLoad(token, {
      key: "gallery.comments.collection.load_failed",
    });
  }
}

const querySync = createVueRouterCollectionQuerySync({
  router,
  codec: createCollectionRouteQueryCodec({
    q: { kind: "string", default: "", maxLength: 200 },
    status: { kind: "enum", values: statuses, default: "" },
    sortOrder: {
      kind: "enum",
      values: ["asc", "desc"] as const,
      default: "desc",
    },
    page: { kind: "positive-integer", default: 1 },
    size: { kind: "positive-integer", values: pageSizes, default: 20 },
  }),
});
const {
  snapshot: collection,
  workflow,
  reload,
} = useVueCollectionWorkflow({
  initialQuery: defaultQuery,
  queryPolicy: createJsonCollectionQueryPolicy<CommentQuery>(),
  keyOf: (comment: GalleryAdminComment) => comment.id,
  isSelectable: () => canModerate.value,
  querySync,
  dataQueryKey: (query) => JSON.stringify(query),
  load: loadComments,
});

const query = computed(() => collection.value.query);
const lifecycle = computed<CommentModerationLifecycle>(
  () => query.value.status || "all",
);
function updateQuery(patch: Partial<CommentQuery>, resetPage = true) {
  workflow.setQuery({
    ...query.value,
    ...patch,
    ...(resetPage ? { page: 1 } : {}),
  });
}
const page = computed({
  get: () => query.value.page,
  set: (value: number) => updateQuery({ page: value }, false),
});
const size = computed({
  get: () => query.value.size,
  set: (value: number) => updateQuery({ size: value }),
});
const sortOrder = computed(() => query.value.sortOrder);
const searchInput = ref(query.value.q);
watch(
  () => query.value.q,
  (value) => {
    if (searchInput.value !== value) searchInput.value = value;
  },
);
let searchTimer: ReturnType<typeof setTimeout> | undefined;
watch(searchInput, (value) => {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(() => updateQuery({ q: value.trim() }), 300);
});
onScopeDispose(() => {
  if (searchTimer) clearTimeout(searchTimer);
});

const busy = ref("");
const emptyingTrash = ref(false);
const batchAction = ref<"" | GalleryCommentStatus>("");
const batchItems = computed(() =>
  lifecycle.value === "trash"
    ? [{ label: "恢复", value: "approved" }]
    : lifecycle.value === "spam"
      ? [
          { label: "恢复", value: "approved" },
          { label: "移入回收站", value: "trash" },
        ]
      : [
          { label: "通过", value: "approved" },
          { label: "标记垃圾", value: "spam" },
          { label: "移入回收站", value: "trash" },
        ],
);
const batchRunning = ref(false);
const batchResult = ref<{ success: number; failed: number }>();
const deleteTarget = shallowRef<GalleryAdminComment>();
const deleteOpen = ref(false);

const selectedIds = computed<readonly string[]>(() =>
  collection.value.selection.mode === "keys"
    ? collection.value.selection.keys
    : [],
);
const selectionCount = computed(() => collection.value.selection.count);
const isPageSelected = computed(() => collection.value.isPageSelected);
const isPageIndeterminate = computed(
  () => collection.value.isPageIndeterminate,
);
function toggleOne(id: string) {
  workflow.toggleKey(id);
}
function togglePage(selected: boolean) {
  workflow.togglePage(selected);
}
function clearSelection() {
  workflow.clearSelection();
}
function replaceSelection(ids: readonly string[]) {
  workflow.clearSelection();
  for (const id of ids) workflow.toggleKey(id);
}

function changeLifecycle(value: CommentModerationLifecycle) {
  updateQuery({
    status: value === "all" ? "" : value,
  });
}
function submitSearch(value: string) {
  if (searchTimer) clearTimeout(searchTimer);
  searchInput.value = value;
  updateQuery({ q: value.trim() });
}
function changeSort() {
  updateQuery({ sortOrder: sortOrder.value === "asc" ? "desc" : "asc" });
}

async function setStatus(
  id: string,
  status: GalleryCommentStatus,
  refreshAfter = true,
) {
  busy.value = id;
  try {
    await call(`/admin/comments/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: { status },
    });
    if (workflow.isSelected(id)) workflow.toggleKey(id);
    if (refreshAfter) await reload();
    return true;
  } catch {
    return false;
  } finally {
    busy.value = "";
  }
}

async function approveComment(id: string) {
  await setStatus(id, "approved");
}

async function emptyTrash() {
  if (emptyingTrash.value) return false;
  emptyingTrash.value = true;
  try {
    for (let batch = 0; batch < 100; batch += 1) {
      const result = await call<GalleryAdminCommentPage>("/admin/comments", {
        query: {
          status: "trash",
          sortBy: "createdAt",
          sortOrder: "asc",
          page: 1,
          size: 100,
        },
      });
      if (!result.items.length) break;
      const outcomes = await Promise.allSettled(
        result.items.map((comment) =>
          call(`/admin/comments/${encodeURIComponent(comment.id)}`, {
            method: "DELETE",
          }),
        ),
      );
      if (outcomes.some((outcome) => outcome.status === "rejected")) {
        throw new Error("部分评论未能永久删除");
      }
      if (result.items.length < 100) break;
    }
    clearSelection();
    await reload();
    return true;
  } catch (error: any) {
    toast.add({
      title: "回收站未清空",
      description: galleryFailureMessage(error, "请稍后重试。"),
      color: "error",
    });
    return false;
  } finally {
    emptyingTrash.value = false;
  }
}

async function applyBatch() {
  if (!batchAction.value || !selectedIds.value.length || batchRunning.value)
    return;
  const ids = [...selectedIds.value];
  batchRunning.value = true;
  const results = await Promise.all(
    ids.map(async (id) => ({
      id,
      success: await setStatus(
        id,
        batchAction.value as GalleryCommentStatus,
        false,
      ),
    })),
  );
  const failed = results.filter((result) => !result.success).map((result) => result.id);
  batchResult.value = { success: results.length - failed.length, failed: failed.length };
  replaceSelection(failed);
  batchAction.value = "";
  await reload();
  batchRunning.value = false;
}

function askDelete(comment: GalleryAdminComment) {
  deleteTarget.value = comment;
  deleteOpen.value = true;
}
function closeDelete() {
  deleteOpen.value = false;
}
async function confirmDelete() {
  const comment = deleteTarget.value;
  if (!comment) return;
  busy.value = comment.id;
  try {
    await call(`/admin/comments/${encodeURIComponent(comment.id)}`, {
      method: "DELETE",
    });
    deleteOpen.value = false;
    deleteTarget.value = undefined;
    await reload();
  } finally {
    busy.value = "";
  }
}

const statusMeta: Record<GalleryCommentStatus, { label: string; color: "warning" | "success" | "error" | "neutral" }> = {
  pending: { label: "待审核", color: "warning" },
  approved: { label: "已通过", color: "success" },
  spam: { label: "垃圾评论", color: "error" },
  trash: { label: "回收站", color: "neutral" },
};
function rowActions(comment: GalleryAdminComment) {
  const actions = [];
  if (comment.status === "trash") {
    if (canModerate.value) {
      actions.push({
        id: "restore",
        label: "恢复评论",
        icon: "i-tabler-restore",
        loading: busy.value === comment.id,
        onSelect: () => void setStatus(comment.id, "approved"),
      });
    }
    if (canDelete.value) {
      actions.push({
        id: "delete-permanently",
        label: "永久删除",
        icon: "i-tabler-trash-x",
        tone: "danger" as const,
        loading: busy.value === comment.id,
        onSelect: () => askDelete(comment),
      });
    }
    return actions;
  }
  if (canModerate.value) {
    actions.push(
      comment.status === "spam"
        ? {
            id: "restore",
            label: "恢复评论",
            icon: "i-tabler-restore",
            loading: busy.value === comment.id,
            onSelect: () => void setStatus(comment.id, "approved"),
          }
        : {
            id: "spam",
            label: "标记为垃圾",
            icon: "i-tabler-alert-triangle",
            loading: busy.value === comment.id,
            onSelect: () => void setStatus(comment.id, "spam"),
          },
    );
    actions.push({
      id: "trash",
      label: "移入回收站",
      icon: "i-tabler-trash",
      loading: busy.value === comment.id,
      onSelect: () => void setStatus(comment.id, "trash"),
    });
  }
  return actions;
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
function moderationItem(comment: GalleryAdminComment): CommentModerationItem {
  return {
    id: comment.id,
    content: comment.content,
    createdAt: comment.createdAt,
    authorName: comment.authorName || "匿名用户",
    avatarUrl: comment.avatarUrl,
    authorEmail: comment.authorEmail,
    anonymous: !comment.userKey,
    reply: Boolean(comment.parentId),
    approve: canModerate.value && comment.status === "pending",
    approving: busy.value === comment.id,
    actions: rowActions(comment),
    ...(comment.status === "approved"
      ? {}
      : { status: statusMeta[comment.status] }),
    source: {
      label: comment.imageTitle || "图片已删除",
      ...(comment.imageId ? { to: `/images/${comment.imageId}` } : {}),
      icon: "i-tabler-photo",
    },
  };
}
const moderationModel = computed<CommentModerationCollectionModel>(() => ({
  search: searchInput.value,
  searchPlaceholder: "搜索评论、评论者或图片…",
  items: collection.value.items.map(moderationItem),
  state: collection.value.issue
    ? "error"
    : ["idle", "loading", "refreshing"].includes(collection.value.loadState)
      ? "loading"
      : "ready",
  errorMessage: collection.value.issue
    ? "暂时无法读取评论，请稍后重试。"
    : undefined,
  total: collection.value.total,
  page: page.value,
  pageSize: size.value,
  pageSizes,
  activeFilterCount: 0,
  controls: [],
  sortOrder: sortOrder.value,
  lifecycle: lifecycle.value,
  emptyingTrash: emptyingTrash.value,
  selection: {
    enabled: canModerate.value,
    count: selectionCount.value,
    pageSelected: isPageSelected.value,
    pageIndeterminate: isPageIndeterminate.value,
    isSelected: workflow.isSelected,
  },
}));
const moderationActions: CommentModerationCollectionActions = {
  updateSearch: (value) => {
    searchInput.value = value;
  },
  search: submitSearch,
  controlChange: () => undefined,
  clearFilters: () => changeLifecycle("all"),
  retry: reload,
  sort: changeSort,
  lifecycleChange: changeLifecycle,
  emptyTrash,
  approve: approveComment,
  pageChange: (value) => {
    page.value = value;
  },
  pageSizeChange: (value) => {
    size.value = value;
  },
  togglePage,
  toggleItem: (id) => toggleOne(id),
  clearSelection,
};
</script>

<template>
  <ManagePage id="comments" title="评论" icon="i-tabler-messages">
    <CommentModerationCollection
      :model="moderationModel"
      :actions="moderationActions"
      :format-date="formatDate"
    >

      <template #bulk-actions>
        <USelect
          v-model="batchAction"
          :items="batchItems"
          value-key="value"
          placeholder="批量操作"
          size="xs"
          class="w-28"
        />
        <UButton
          label="应用"
          size="xs"
          variant="soft"
          :loading="batchRunning"
          :disabled="!batchAction"
          @click="applyBatch"
        />
      </template>

    </CommentModerationCollection>

    <div
      v-if="batchResult"
      class="flex items-center justify-between gap-2 rounded-lg border border-default bg-elevated px-3 py-2.5 text-xs"
      :class="batchResult.failed ? 'text-warning' : 'text-success'"
      role="status"
    >
      <span>已处理 {{ batchResult.success }} 条评论<span v-if="batchResult.failed">，{{ batchResult.failed }} 条失败</span></span>
      <UButton icon="i-tabler-x" color="neutral" variant="ghost" size="xs" square aria-label="关闭批量结果" @click="batchResult = undefined" />
    </div>

    <UModal
      v-model:open="deleteOpen"
      title="永久删除评论"
      :description="`永久删除「${deleteTarget?.authorName || '匿名用户'}」的这条评论及其回复？此操作不可撤销。`"
    >
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton label="取消" color="neutral" variant="outline" @click="closeDelete" />
          <UButton label="永久删除" icon="i-tabler-trash-x" color="error" :loading="busy === deleteTarget?.id" @click="confirmDelete" />
        </div>
      </template>
    </UModal>
  </ManagePage>
</template>
