<script setup lang="ts">
import { galleryAdminMutationErrorMessage } from "~/utils/galleryAdminErrors";
import { ManagePage } from "@yueli/ui/admin";
import { createGalleryNotifier } from "~/utils/feedback";
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
import { ManageTaxonomyChips } from "~/utils/manageComponents";
import type {
  GalleryAdminImage,
  GalleryAdminImagePage,
  GalleryClassificationTagPage,
  GalleryCollection,
  GallerySubmissionOptions,
} from "~/types/gallery";

interface BulkResult {
  changed: number;
  failed: number;
  interrupted?: boolean;
  message?: string;
}

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "图片 · 图库管理" });

const ALL = "__all__" as const;
const router = useRouter();
const { call } = useGalleryApi();
const { can } = useGalleryMe();
const hydrated = useClientHydrated();
const toast = createGalleryNotifier(useToast());
const canImageUpdate = computed(() => can("gallery.image.update"));
const canImageHide = computed(() => can("gallery.image.hide"));
const canCollectionManage = computed(() => can("gallery.collection.manage"));
const canBulkManage = computed(
  () => canImageUpdate.value || canImageHide.value || canCollectionManage.value,
);

type ImageStatus = "" | "published" | "draft" | "hidden" | "deleted";
type ImageSortBy = "createdAt" | "updatedAt" | "title" | "views";
type ImageSortOrder = "asc" | "desc";
type ImageView = "list" | "grid";
interface ImageCollectionQuery {
  q: string;
  status: ImageStatus;
  page: number;
  size: number;
  sortBy: ImageSortBy;
  sortOrder: ImageSortOrder;
  category: string;
  facet: string;
  view: ImageView;
}
const statuses = ["", "published", "draft", "hidden", "deleted"] as const;
const sortByValues = ["createdAt", "updatedAt", "title", "views"] as const;
const sortOrderValues = ["asc", "desc"] as const;
const viewValues = ["list", "grid"] as const;
const pageSizes = [12, 24, 48, 60] as const;
const defaultQuery: ImageCollectionQuery = {
  q: "",
  status: "",
  page: 1,
  size: 24,
  sortBy: "createdAt",
  sortOrder: "desc",
  category: ALL,
  facet: ALL,
  view: "list",
};
const counts = ref<Record<string, number>>({});
async function loadImages(
  nextQuery: Readonly<ImageCollectionQuery>,
  activeWorkflow: CollectionWorkflow<
    GalleryAdminImage,
    string,
    ImageCollectionQuery
  >,
) {
  const token = activeWorkflow.beginLoad();
  try {
    const data = await call<GalleryAdminImagePage>("/admin/images", {
      query: {
        q: nextQuery.q || undefined,
        sortBy: nextQuery.sortBy,
        sortOrder: nextQuery.sortOrder,
        page: nextQuery.page,
        size: nextQuery.size,
        publicationState: nextQuery.status || undefined,
        categoryId: nextQuery.category === ALL ? undefined : nextQuery.category,
        facetValueId: nextQuery.facet === ALL ? undefined : nextQuery.facet,
      },
    });
    const lastPage = Math.max(
      1,
      data.totalPages || Math.ceil(data.total / nextQuery.size),
    );
    if (nextQuery.page > lastPage) {
      activeWorkflow.setQuery({ ...nextQuery, page: lastPage });
      return;
    }
    counts.value = data.counts || {};
    activeWorkflow.resolveLoad(token, {
      items: data.items || [],
      total: data.total || 0,
    });
  } catch {
    activeWorkflow.rejectLoad(token, {
      key: "gallery.images.collection.load_failed",
    });
  }
}
const queryPolicy = createJsonCollectionQueryPolicy<ImageCollectionQuery>();
const querySync = createVueRouterCollectionQuerySync({
  router,
  codec: createCollectionRouteQueryCodec({
    q: { kind: "string", default: defaultQuery.q, maxLength: 200 },
    status: { kind: "enum", values: statuses, default: defaultQuery.status },
    page: { kind: "positive-integer", default: defaultQuery.page },
    size: {
      kind: "positive-integer",
      values: pageSizes,
      default: defaultQuery.size,
    },
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
    category: {
      kind: "string",
      default: defaultQuery.category,
      maxLength: 200,
    },
    facet: { kind: "string", default: defaultQuery.facet, maxLength: 200 },
    view: { kind: "enum", values: viewValues, default: defaultQuery.view },
  }),
});
const {
  snapshot: imageCollection,
  workflow: imageWorkflow,
  reload: refresh,
} = useVueCollectionWorkflow({
  initialQuery: defaultQuery,
  queryPolicy,
  keyOf: (image: GalleryAdminImage) => image.id,
  isSelectable: () => canBulkManage.value,
  querySync,
  dataQueryKey: ({ view: _view, ...query }) => JSON.stringify(query),
  load: loadImages,
});
const collectionQuery = computed(() => imageCollection.value.query);
function updateCollectionQuery(
  patch: Partial<ImageCollectionQuery>,
  resetPage = true,
) {
  imageWorkflow.setQuery({
    ...collectionQuery.value,
    ...patch,
    ...(resetPage ? { page: 1 } : {}),
  });
}
const status = computed({
  get: () => collectionQuery.value.status,
  set: (value: ImageStatus) => updateCollectionQuery({ status: value }),
});
const page = computed({
  get: () => collectionQuery.value.page,
  set: (value: number) => updateCollectionQuery({ page: value }, false),
});
const size = computed({
  get: () => collectionQuery.value.size,
  set: (value: number) => updateCollectionQuery({ size: value }),
});
const sortBy = computed({
  get: () => collectionQuery.value.sortBy,
  set: (value: ImageSortBy) => updateCollectionQuery({ sortBy: value }),
});
const sortOrder = computed({
  get: () => collectionQuery.value.sortOrder,
  set: (value: ImageSortOrder) => updateCollectionQuery({ sortOrder: value }),
});
const viewMode = computed({
  get: () => collectionQuery.value.view,
  set: (value: ImageView) => updateCollectionQuery({ view: value }, false),
});

function changeColumnSort(nextSortBy: ImageSortBy) {
  if (sortBy.value === nextSortBy) {
    sortOrder.value = sortOrder.value === "asc" ? "desc" : "asc";
    return;
  }
  updateCollectionQuery({
    sortBy: nextSortBy,
    sortOrder: nextSortBy === "title" ? "asc" : "desc",
  });
}
const category = computed({
  get: () => collectionQuery.value.category,
  set: (value: string) => updateCollectionQuery({ category: value }),
});
const facet = computed({
  get: () => collectionQuery.value.facet,
  set: (value: string) => updateCollectionQuery({ facet: value }),
});
const searchInput = ref(collectionQuery.value.q);
let searchTimer: ReturnType<typeof setTimeout> | undefined;
watch(searchInput, (value) => {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(
    () => updateCollectionQuery({ q: value.trim() }),
    300,
  );
});
watch(
  () => collectionQuery.value.q,
  (value) => {
    if (searchInput.value !== value) searchInput.value = value;
  },
);
onScopeDispose(() => {
  if (searchTimer) clearTimeout(searchTimer);
});
function submitSearch(value: string) {
  if (searchTimer) clearTimeout(searchTimer);
  searchInput.value = value;
  updateCollectionQuery({ q: value.trim() });
}

const { data: options } = await useFetch<GallerySubmissionOptions>(
  "/api/gallery/submission-options",
  {
    server: false,
    default: () => ({ categories: [], facets: [] }),
  },
);
const categoryOptions = computed(() => [
  { label: "全部分类", value: ALL },
  ...(options.value?.categories || []).map((item) => ({
    label: item.name,
    value: item.id,
  })),
]);
const facetOptions = computed(() => [
  { label: "全部维度", value: ALL },
  ...(options.value?.facets || []).flatMap((group) =>
    group.values.map((item) => ({
      label: `${group.name} · ${item.name}`,
      value: item.id,
    })),
  ),
]);
const { data: tagData } = await useAsyncData(
  "gallery-manage-image-tags",
  () =>
    call<{ page: GalleryClassificationTagPage }>(
      "/admin/classification/tags?size=100",
    ),
  { server: false, default: () => ({ page: { items: [], nextCursor: "" } }) },
);
const tagOptions = computed(() =>
  tagData.value.page.items
    .filter((item) => item.status === "active")
    .map((item) => ({ label: item.name, value: item.id })),
);
const { data: collectionData, refresh: refreshCollections } =
  await useAsyncData(
    "gallery-manage-image-collection-options",
    () =>
      canCollectionManage.value
        ? call<{ collections: GalleryCollection[] }>("/admin/collections")
        : Promise.resolve({ collections: [] }),
    { server: false, default: () => ({ collections: [] }) },
  );
const collectionOptions = computed(() =>
  collectionData.value.collections.map((item) => ({
    label: `${item.name} · ${item.itemCount} 张`,
    value: item.id,
  })),
);
const statusOptions = computed(() => [
  { value: ALL, label: `全部 · ${counts.value.all || 0}` },
  { value: "published", label: `已公开 · ${counts.value.published || 0}` },
  { value: "draft", label: `草稿 · ${counts.value.draft || 0}` },
  { value: "hidden", label: `已隐藏 · ${counts.value.hidden || 0}` },
  { value: "deleted", label: `已删除 · ${counts.value.deleted || 0}` },
]);
const controls = computed<CollectionControl[]>(() => [
  {
    kind: "select",
    id: "status",
    label: "公开状态",
    value: status.value || ALL,
    options: statusOptions.value,
    class: "w-32",
  },
  {
    kind: "select",
    id: "category",
    label: "分类",
    value: category.value,
    options: categoryOptions.value,
    searchPlaceholder: "搜索分类…",
    icon: "i-tabler-folders",
    class: "w-36",
  },
  {
    kind: "select",
    id: "facet",
    label: "维度",
    value: facet.value,
    options: facetOptions.value,
    searchPlaceholder: "搜索维度…",
    icon: "i-tabler-adjustments",
    class: "w-40",
  },
]);
const activeFilterCount = computed(
  () =>
    [status.value !== "", category.value !== ALL, facet.value !== ALL].filter(
      Boolean,
    ).length,
);
function clearFilters() {
  updateCollectionQuery({ status: "", category: ALL, facet: ALL });
}
function changeControl(id: string, value: CollectionControlValue) {
  if (typeof value !== "string") return;
  if (id === "status") {
    const nextStatus = value === ALL ? "" : value;
    if (statuses.includes(nextStatus as ImageStatus))
      status.value = nextStatus as ImageStatus;
  }
  if (id === "category") category.value = value;
  if (id === "facet") facet.value = value;
}
const messages: CollectionPanelMessages = {
  searchPlaceholder: "搜索标题、说明或替代文本…",
  searchAction: "搜索",
  filtersAction: "筛选",
  activeFilters: (count) => `筛选（${count}）`,
  clearFilters: "清除筛选",
  selectPage: "选择当前页图片",
  selectItem: (label) => `选择图片：${label}`,
  bulkRegion: "图片批量操作",
  selected: (count) => `已选择 ${count} 张图片`,
  selectAllResults: "选择全部结果",
  clearSelection: "取消选择",
  emptyTitle: "没有符合条件的图片",
  emptyDescription: "调整筛选或搜索条件后重试。",
  errorTitle: "图片加载失败",
  retry: "重新加载",
  showing: (first, last, total) => `显示 ${first}–${last}，共 ${total} 张`,
  pageSize: "每页",
  pageSizeControl: "每页图片数量",
  pageSizeOption: (value) => `${value} 张`,
};
const panelState = computed<CollectionPanelState>(() => {
  if (imageCollection.value.issue) return "error";
  if (
    !hydrated.value ||
    ["idle", "loading", "refreshing"].includes(imageCollection.value.loadState)
  )
    return "loading";
  return "ready";
});
const items = computed(() => imageCollection.value.items);
const selectedIds = computed<readonly string[]>(() =>
  imageCollection.value.selection.mode === "keys"
    ? imageCollection.value.selection.keys
    : [],
);
const selectionCount = computed(() => imageCollection.value.selection.count);
const isPageSelected = computed(() => imageCollection.value.isPageSelected);
const isPageIndeterminate = computed(
  () => imageCollection.value.isPageIndeterminate,
);
function toggleOne(id: string) {
  imageWorkflow.toggleKey(id);
}
function togglePage(selected: boolean) {
  imageWorkflow.togglePage(selected);
}
function replaceSelection(ids: readonly string[]) {
  imageWorkflow.clearSelection();
  for (const id of ids) imageWorkflow.toggleKey(id);
}
function clearSelection() {
  imageWorkflow.clearSelection();
}
const imageKey = (image: GalleryAdminImage) => image.id;
const imageLabel = (image: GalleryAdminImage) => image.title;

const editing = shallowRef<GalleryAdminImage>();
const editCategoryOptions = computed(() => {
  const items = categoryOptions.value.filter((item) => item.value !== ALL);
  const current = editing.value;
  if (
    current?.primaryCategoryId &&
    current.primaryCategory &&
    !items.some((item) => item.value === current.primaryCategoryId)
  ) {
    return [
      ...items,
      { label: current.primaryCategory, value: current.primaryCategoryId },
    ];
  }
  return items;
});
const editTagOptions = computed(() => {
  const items = new Map(
    tagOptions.value.map((item) => [item.value, item] as const),
  );
  for (const tag of editing.value?.tags || []) {
    items.set(tag.id, { label: tag.name, value: tag.id });
  }
  return [...items.values()];
});
const editPending = ref(false);
const editLoading = ref(false);
const deleteConfirming = ref(false);
const deletePending = ref(false);
const deleteError = ref("");
const editForm = reactive({
  title: "",
  description: "",
  altText: "",
  sourceUrl: "",
  primaryCategoryId: ALL,
  facetValueIds: {} as Record<string, string>,
  tagIds: [] as string[],
});
function hydrateEditForm(item: GalleryAdminImage) {
  editing.value = item;
  deleteConfirming.value = false;
  deleteError.value = "";
  Object.assign(editForm, {
    title: item.title,
    description: item.description,
    altText: item.altText,
    sourceUrl: item.sourceUrl,
    primaryCategoryId: item.primaryCategoryId || ALL,
    facetValueIds: Object.fromEntries(
      (item.facets || []).map((assignment) => [
        assignment.facetId,
        assignment.valueId,
      ]),
    ),
    tagIds: (item.tags || []).map((tag) => tag.id),
  });
}

function closeEdit() {
  editing.value = undefined;
  deleteConfirming.value = false;
  deleteError.value = "";
}

function setEditOpen(open: boolean) {
  if (!open) closeEdit();
}

function beginDelete() {
  deleteConfirming.value = true;
}

function cancelDelete() {
  deleteConfirming.value = false;
  deleteError.value = "";
}
async function openEdit(item: GalleryAdminImage) {
  if (!canImageUpdate.value) return;
  hydrateEditForm(item);
  editLoading.value = true;
  try {
    const response = await call<{ image: GalleryAdminImage }>(
      `/admin/images/${encodeURIComponent(item.id)}`,
    );
    hydrateEditForm(response.image);
  } catch (reason: any) {
    editing.value = undefined;
    toast.add({
      title: "无法打开图片编辑器",
      description:
        reason?.data?.message || "图片记录可能已变化，请刷新后重试。",
      color: "error",
    });
  } finally {
    editLoading.value = false;
  }
}
async function saveEdit() {
  if (
    !canImageUpdate.value ||
    !editing.value?.updatedAt ||
    !editForm.title.trim() ||
    !editForm.altText.trim() ||
    editForm.primaryCategoryId === ALL
  )
    return;
  editPending.value = true;
  try {
    await call(`/admin/images/${encodeURIComponent(editing.value.id)}`, {
      method: "PATCH",
      body: {
        expectedUpdatedAt: editing.value.updatedAt,
        title: editForm.title,
        description: editForm.description,
        altText: editForm.altText,
        sourceUrl: editForm.sourceUrl,
        classification: {
          primaryCategoryId: editForm.primaryCategoryId,
          facetValueIds: Object.values(editForm.facetValueIds).filter(
            (value) => value && value !== ALL,
          ),
          tagIds: editForm.tagIds,
        },
      },
    });
    editing.value = undefined;
    await refresh();
  } catch (reason: any) {
    toast.add({
      title: "图片信息没有保存",
      description: galleryAdminMutationErrorMessage(reason),
      color: "error",
    });
    await refresh();
  } finally {
    editPending.value = false;
  }
}

async function deleteEditingImage() {
  if (!canImageHide.value || !editing.value?.updatedAt || deletePending.value)
    return;
  deletePending.value = true;
  deleteError.value = "";
  try {
    await call(`/admin/images/${encodeURIComponent(editing.value.id)}`, {
      method: "DELETE",
      body: { expectedUpdatedAt: editing.value.updatedAt },
    });
    closeEdit();
    await refresh();
  } catch (reason: any) {
    deleteError.value = galleryAdminMutationErrorMessage(reason);
  } finally {
    deletePending.value = false;
  }
}

const batchAction = ref<string>();
const batchOpen = ref(false);
const batchPending = ref(false);
const batchCategory = ref(ALL);
const batchCollectionId = ref("");
const batchReason = ref("");
const batchResult = ref<BulkResult>();
const batchOptions = computed(() => [
  ...(canImageUpdate.value
    ? [{ label: "设置主分类", value: "set_primary_category" }]
    : []),
  ...(canImageHide.value ? [{ label: "下架", value: "hide" }] : []),
  ...(canCollectionManage.value
    ? [{ label: "加入专题", value: "add_to_collection" }]
    : []),
]);
function prepareBatch() {
  if (
    batchAction.value &&
    batchOptions.value.some((item) => item.value === batchAction.value) &&
    selectionCount.value
  )
    batchOpen.value = true;
}
function cancelBatch() {
  batchOpen.value = false;
}
async function runBatch() {
  if (
    !batchAction.value ||
    !batchOptions.value.some((item) => item.value === batchAction.value) ||
    !selectedIds.value.length
  )
    return;
  batchPending.value = true;
  try {
    if (batchAction.value === "add_to_collection") {
      const target = collectionData.value.collections.find(
        (item) => item.id === batchCollectionId.value,
      );
      if (!target) return;
      const selectedCount = selectedIds.value.length;
      await call<{ collection: GalleryCollection }>(
        `/admin/collections/${encodeURIComponent(target.id)}/members`,
        {
          method: "POST",
          body: {
            version: target.version,
            add: selectedIds.value,
            remove: [],
          },
        },
      );
      clearSelection();
      batchResult.value = { changed: selectedCount, failed: 0 };
      batchOpen.value = false;
      batchAction.value = undefined;
      batchCollectionId.value = "";
      await refreshCollections();
      return;
    }
    const response = await call<{
      results: Array<{ imageId: string; success: boolean; error?: string }>;
    }>("/admin/images/bulk", {
      method: "POST",
      body: {
        imageIds: selectedIds.value,
        action: batchAction.value,
        primaryCategoryId:
          batchCategory.value === ALL ? "" : batchCategory.value,
        reason: batchReason.value.trim(),
      },
    });
    const failed = response.results.filter((item) => !item.success);
    replaceSelection(failed.map((item) => item.imageId));
    batchResult.value = {
      changed: response.results.length - failed.length,
      failed: failed.length,
    };
    batchOpen.value = false;
    batchAction.value = undefined;
    batchCategory.value = ALL;
    batchCollectionId.value = "";
    batchReason.value = "";
    await refresh();
  } catch (reason: any) {
    batchResult.value = {
      changed: 0,
      failed: selectedIds.value.length,
      interrupted: true,
      message: reason?.data?.message || "批量请求中断，当前选择已保留。",
    };
  } finally {
    batchPending.value = false;
  }
}

function imageStatusBadge(image: GalleryAdminImage) {
  if (!image.publicRenditionReady)
    return {
      label: "待生成",
      color: "warning" as const,
      help: "公开版本尚未就绪",
    };
  return {
    label: publicationLabel(image.publicationState),
    color:
      image.publicationState === "published"
        ? ("success" as const)
        : image.publicationState === "deleted"
          ? ("error" as const)
          : image.publicationState === "hidden"
            ? ("warning" as const)
            : ("neutral" as const),
    help: "",
  };
}
const timestampFormatter = new Intl.DateTimeFormat("zh-CN", {
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
  hour12: false,
});
function formatTimestamp(value?: string) {
  return value ? timestampFormatter.format(new Date(value)) : "—";
}
function publicationLabel(value: GalleryAdminImage["publicationState"]) {
  return {
    published: "已公开",
    draft: "草稿",
    hidden: "已隐藏",
    deleted: "已删除",
  }[value];
}
</script>

<template>
  <ManagePage
    id="images"
    title="图片"
    icon="i-tabler-photo"
    :data-manage-images-state="panelState"
  >
    <template #actions>
      <UButton to="/submit" icon="i-tabler-upload" label="投稿图片" />
    </template>

    <CollectionPanel
      v-model:search="searchInput"
      :items="items"
      :item-key="imageKey"
      :item-label="imageLabel"
      :controls="controls"
      :messages="messages"
      :state="panelState"
      error-message="暂时无法读取图片，请稍后重试。"
      :total="imageCollection.total"
      :page="page"
      :page-size="size"
      :page-sizes="pageSizes"
      :active-filter-count="activeFilterCount"
      :layout="viewMode === 'grid' ? 'grid' : 'rows'"
      :selection-count="selectionCount"
      :page-selected="isPageSelected"
      :page-indeterminate="isPageIndeterminate"
      :is-selected="imageWorkflow.isSelected"
      label="图片管理"
      :selectable="canBulkManage"
      @search="submitSearch"
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
        <CollectionViewToggle
          v-model="viewMode"
          :items="[
            { key: 'list', label: '列表视图', icon: 'i-tabler-list' },
            {
              key: 'grid',
              label: '网格视图',
              icon: 'i-tabler-layout-grid',
            },
          ]"
        />
      </template>

      <template v-if="viewMode === 'list'" #columns>
        <div
          class="grid grid-cols-[minmax(0,1fr)_5rem_5rem] items-center gap-3 md:grid-cols-[minmax(0,1fr)_8rem_5rem_8rem_5rem] lg:grid-cols-[minmax(0,1fr)_8rem_5rem_8rem_8rem_5rem]"
        >
          <CollectionSortHeader
            label="标题"
            :active="sortBy === 'title'"
            :sort-order="sortOrder"
            @sort="changeColumnSort('title')"
          />
          <span class="hidden md:inline">分类</span>
          <CollectionSortHeader
            label="浏览"
            :active="sortBy === 'views'"
            :sort-order="sortOrder"
            @sort="changeColumnSort('views')"
          />
          <CollectionSortHeader
            class="hidden md:inline-flex"
            label="更新"
            :active="sortBy === 'updatedAt'"
            :sort-order="sortOrder"
            @sort="changeColumnSort('updatedAt')"
          />
          <span class="hidden lg:inline">状态</span>
          <span class="text-right">操作</span>
        </div>
      </template>

      <template #bulk-actions>
        <USelect
          v-if="canBulkManage"
          v-model="batchAction"
          aria-label="批量操作"
          :items="batchOptions"
          placeholder="批量操作"
          size="xs"
          class="w-28"
        />
        <UButton
          size="xs"
          color="primary"
          variant="soft"
          :disabled="!batchAction"
          @click="prepareBatch"
          >应用</UButton
        >
      </template>

      <template #item="{ item: image }">
        <div
          v-if="viewMode === 'list'"
          class="grid min-w-0 grid-cols-[minmax(0,1fr)_5rem_5rem] items-center gap-3 md:grid-cols-[minmax(0,1fr)_8rem_5rem_8rem_5rem] lg:grid-cols-[minmax(0,1fr)_8rem_5rem_8rem_8rem_5rem]"
          data-gallery-image-list-item
        >
          <div class="flex min-w-0 items-center gap-3">
            <img
              :src="galleryRendition(image.assetId, 'thumbnail')"
              :alt="image.altText"
              class="hidden aspect-[4/3] w-16 shrink-0 rounded-lg bg-elevated object-cover sm:block"
            />
            <div class="min-w-0">
              <button
                v-if="canImageUpdate"
                type="button"
                class="block max-w-full truncate text-left text-sm font-medium text-highlighted hover:text-primary"
                @click="openEdit(image)"
              >
                {{ image.title }}
              </button>
              <p v-else class="truncate text-sm font-medium text-highlighted">
                {{ image.title }}
              </p>
              <p class="mt-1 truncate text-xs text-muted">
                {{ image.description || image.altText }}
              </p>
              <p class="mt-1 truncate font-mono text-xs text-dimmed md:hidden">
                {{ image.primaryCategory || "未分类" }} ·
                {{ formatTimestamp(image.updatedAt) }}
              </p>
            </div>
          </div>
          <div class="hidden min-w-0 md:block">
            <ManageTaxonomyChips
              :items="
                image.primaryCategory
                  ? [
                      {
                        key: image.primaryCategoryId || image.primaryCategory,
                        label: image.primaryCategory,
                        kind: 'category',
                      },
                    ]
                  : []
              "
            />
          </div>
          <div class="text-right text-xs tabular-nums text-muted md:text-left">
            {{ compactMetric(image.metrics.views) }}
          </div>
          <time
            class="hidden text-xs text-muted md:block"
            :datetime="image.updatedAt"
          >
            {{ formatTimestamp(image.updatedAt) }}
          </time>
          <div class="hidden lg:block">
            <span data-image-status-badge>
              <UTooltip
                :text="
                  imageStatusBadge(image).help || imageStatusBadge(image).label
                "
              >
                <UBadge
                  :color="imageStatusBadge(image).color"
                  variant="soft"
                  :label="imageStatusBadge(image).label"
                />
              </UTooltip>
            </span>
          </div>
          <div class="flex justify-end gap-1">
            <UTooltip
              v-if="image.publicationState === 'published'"
              text="查看公开图片"
            >
              <UButton
                :to="`/images/${image.id}`"
                target="_blank"
                rel="noopener"
                color="neutral"
                variant="ghost"
                size="xs"
                icon="i-tabler-external-link"
                square
                :aria-label="`查看公开图片：${image.title}`"
              />
            </UTooltip>
            <UTooltip v-if="canImageUpdate" text="编辑图片">
              <UButton
                color="neutral"
                variant="ghost"
                size="xs"
                icon="i-tabler-pencil"
                square
                :aria-label="`编辑图片：${image.title}`"
                @click="openEdit(image)"
              />
            </UTooltip>
          </div>
        </div>

        <article
          v-else
          class="group -m-4 overflow-hidden rounded-lg"
          data-gallery-image-grid-item
        >
          <div class="relative aspect-[4/3] overflow-hidden bg-elevated">
            <img
              v-bind="galleryImageSources(image.assetId, 'grid', false)"
              :alt="image.altText"
              class="size-full object-cover transition-transform duration-200 group-hover:scale-[1.02]"
            />
            <span class="absolute right-2 top-2" data-image-status-badge>
              <UTooltip
                :text="
                  imageStatusBadge(image).help || imageStatusBadge(image).label
                "
              >
                <UBadge
                  :color="imageStatusBadge(image).color"
                  variant="solid"
                  :label="imageStatusBadge(image).label"
                  size="xs"
                />
              </UTooltip>
            </span>
          </div>
          <div class="space-y-3 p-4">
            <div class="min-w-0">
              <button
                v-if="canImageUpdate"
                type="button"
                class="block max-w-full truncate text-left text-sm font-semibold text-highlighted hover:text-primary"
                @click="openEdit(image)"
              >
                {{ image.title }}
              </button>
              <p v-else class="truncate text-sm font-semibold text-highlighted">
                {{ image.title }}
              </p>
              <p class="mt-1 line-clamp-2 min-h-8 text-xs leading-4 text-muted">
                {{ image.description || image.altText }}
              </p>
            </div>
            <div class="flex min-w-0 items-center justify-between gap-3">
              <ManageTaxonomyChips
                class="min-w-0"
                :items="
                  image.primaryCategory
                    ? [
                        {
                          key: image.primaryCategoryId || image.primaryCategory,
                          label: image.primaryCategory,
                          kind: 'category',
                        },
                      ]
                    : []
                "
              />
              <span class="shrink-0 text-xs tabular-nums text-muted">
                {{ compactMetric(image.metrics.views) }} 次浏览
              </span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <time class="text-xs text-dimmed" :datetime="image.updatedAt">
                更新于 {{ formatTimestamp(image.updatedAt) }}
              </time>
              <div class="flex shrink-0 justify-end gap-1">
                <UTooltip
                  v-if="image.publicationState === 'published'"
                  text="查看公开图片"
                >
                  <UButton
                    :to="`/images/${image.id}`"
                    target="_blank"
                    rel="noopener"
                    color="neutral"
                    variant="ghost"
                    size="xs"
                    icon="i-tabler-external-link"
                    square
                    :aria-label="`查看公开图片：${image.title}`"
                  />
                </UTooltip>
                <UTooltip v-if="canImageUpdate" text="编辑图片">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    size="xs"
                    icon="i-tabler-pencil"
                    square
                    :aria-label="`编辑图片：${image.title}`"
                    @click="openEdit(image)"
                  />
                </UTooltip>
              </div>
            </div>
          </div>
        </article>
      </template>
    </CollectionPanel>

    <div
      v-if="batchResult"
      class="mt-3 flex items-center gap-2 rounded-lg border border-default bg-elevated px-3 py-2.5 text-xs"
      role="status"
    >
      <UIcon
        :name="
          batchResult.interrupted || batchResult.failed
            ? 'i-tabler-alert-triangle'
            : 'i-tabler-circle-check'
        "
        :class="
          batchResult.interrupted || batchResult.failed
            ? 'text-warning'
            : 'text-success'
        "
      />
      <span class="min-w-0 flex-1">{{
        batchResult.interrupted
          ? batchResult.message
          : `已处理 ${batchResult.changed} 张${batchResult.failed ? `，${batchResult.failed} 张待处理` : ""}`
      }}</span>
      <UButton
        icon="i-tabler-x"
        color="neutral"
        variant="ghost"
        size="xs"
        square
        aria-label="关闭批量结果"
        @click="batchResult = undefined"
      />
    </div>

    <UModal
      :open="canImageUpdate && Boolean(editing)"
      title="编辑图片"
      description="统一维护目录文案、主分类、维度和标签。"
      :ui="{ content: 'sm:max-w-3xl' }"
      @update:open="setEditOpen"
    >
      <template #body>
        <div v-if="editLoading" class="space-y-3 py-2">
          <USkeleton class="h-10 rounded-lg" />
          <USkeleton class="h-24 rounded-lg" />
          <USkeleton class="h-32 rounded-lg" />
        </div>
        <form
          v-else
          id="gallery-image-edit-form"
          class="space-y-5"
          @submit.prevent="saveEdit"
        >
          <div
            v-if="editing"
            data-image-editor-preview
            class="grid max-h-72 min-h-48 place-items-center overflow-hidden rounded-xl bg-elevated p-2"
          >
            <img
              :src="galleryRendition(editing.assetId, 'preview')"
              :alt="editForm.altText || editing.title"
              class="max-h-68 max-w-full rounded-lg object-contain"
            />
          </div>
          <UAlert
            v-if="deleteError"
            color="error"
            variant="subtle"
            title="图片没有删除"
            :description="deleteError"
          />
          <UFormField label="标题" required
            ><UInput v-model="editForm.title" maxlength="160" class="w-full"
          /></UFormField>
          <UFormField label="替代文本" required
            ><UInput v-model="editForm.altText" class="w-full"
          /></UFormField>
          <UFormField label="说明"
            ><UTextarea v-model="editForm.description" :rows="4" class="w-full"
          /></UFormField>
          <UFormField label="来源地址"
            ><UInput
              v-model="editForm.sourceUrl"
              type="url"
              placeholder="https://"
              class="w-full"
          /></UFormField>
          <div class="border-t border-default pt-5">
            <h3 class="text-sm font-semibold text-highlighted">目录信息</h3>
            <p class="mt-1 text-xs text-muted">
              主分类用于导航；每个维度最多选择一个值。
            </p>
          </div>
          <UFormField label="主分类" required>
            <USelectMenu
              v-model="editForm.primaryCategoryId"
              aria-label="主分类"
              :items="editCategoryOptions"
              value-key="value"
              class="w-full"
              :search-input="{ placeholder: '搜索分类…' }"
            />
          </UFormField>
          <div class="grid gap-4 sm:grid-cols-2">
            <UFormField
              v-for="group in options?.facets || []"
              :key="group.id"
              :label="group.name"
            >
              <USelect
                v-model="editForm.facetValueIds[group.id]"
                :items="[
                  { label: '未设置', value: ALL },
                  ...group.values.map((item) => ({
                    label: item.name,
                    value: item.id,
                  })),
                ]"
                value-key="value"
                class="w-full"
              />
            </UFormField>
          </div>
          <UFormField label="标签" hint="可多选">
            <USelectMenu
              v-model="editForm.tagIds"
              aria-label="标签"
              :items="editTagOptions"
              value-key="value"
              multiple
              class="w-full"
              :search-input="{ placeholder: '搜索标签…' }"
              placeholder="选择标签"
            />
          </UFormField>
        </form>
      </template>
      <template v-if="!editLoading" #footer>
        <div
          v-if="deleteConfirming"
          class="flex w-full flex-wrap items-center justify-between gap-3"
        >
          <p class="min-w-0 text-sm font-medium text-error">
            确定删除「{{ editing?.title }}」吗？
          </p>
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              label="取消删除"
              @click="cancelDelete"
            />
            <UButton
              color="error"
              icon="i-tabler-trash"
              label="确认删除"
              :loading="deletePending"
              @click="deleteEditingImage"
            />
          </div>
        </div>
        <div v-else class="flex w-full items-center justify-between gap-3">
          <UButton
            v-if="canImageHide"
            color="error"
            variant="ghost"
            icon="i-tabler-trash"
            label="删除图片"
            @click="beginDelete"
          />
          <span v-else />
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="outline"
              label="取消"
              @click="closeEdit"
            />
            <UButton
              form="gallery-image-edit-form"
              type="submit"
              label="保存更改"
              :loading="editPending"
              :disabled="editForm.primaryCategoryId === ALL"
            />
          </div>
        </div>
      </template>
    </UModal>

    <UModal
      v-if="canBulkManage"
      v-model:open="batchOpen"
      :title="
        batchAction === 'hide'
          ? '批量下架'
          : batchAction === 'add_to_collection'
            ? '批量加入专题'
            : '批量设置主分类'
      "
      :description="`将处理选中的 ${selectionCount} 张图片；失败项目会保留选择。`"
    >
      <template #body
        ><form class="space-y-4" @submit.prevent="runBatch">
          <UFormField
            v-if="batchAction === 'set_primary_category'"
            label="主分类"
            required
            ><USelectMenu
              v-model="batchCategory"
              :items="categoryOptions.filter((item) => item.value !== ALL)"
              value-key="value"
              class="w-full"
              :search-input="{ placeholder: '搜索分类…' }"
          /></UFormField>
          <UFormField
            v-else-if="batchAction === 'add_to_collection'"
            label="目标专题"
            required
          >
            <USelectMenu
              v-model="batchCollectionId"
              aria-label="目标专题"
              :items="collectionOptions"
              value-key="value"
              class="w-full"
              placeholder="选择专题"
              :search-input="{ placeholder: '搜索专题…' }"
            />
            <p v-if="!collectionOptions.length" class="mt-2 text-xs text-muted">
              暂无可用专题，请先创建专题。
            </p>
          </UFormField>
          <UFormField v-else label="下架原因" required
            ><UTextarea
              v-model="batchReason"
              :rows="4"
              placeholder="说明政策、版权或内容原因"
              class="w-full"
          /></UFormField>
          <div class="flex justify-end gap-2">
            <UButton
              color="neutral"
              variant="outline"
              label="取消"
              @click="cancelBatch"
            /><UButton
              type="submit"
              :color="batchAction === 'hide' ? 'error' : 'primary'"
              label="确认应用"
              :loading="batchPending"
              :disabled="
                batchAction === 'hide'
                  ? !batchReason.trim()
                  : batchAction === 'add_to_collection'
                    ? !batchCollectionId
                    : batchCategory === ALL
              "
            />
          </div></form
      ></template>
    </UModal>
  </ManagePage>
</template>
