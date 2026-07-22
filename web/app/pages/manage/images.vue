<script setup lang="ts">
import { PageHeader } from "@yueli/ui/dashboard/pattern";
import { createPlatformNotifier } from "@platform/ui/feedback";
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
  CollectionViewToggle,
} from "@yueli/ui/collection/pattern";
import { ManageTaxonomyChips } from "@platform/manage/components";
import type {
  GalleryAdminImage,
  GalleryAdminImagePage,
  GalleryClassificationTagPage,
  GallerySubmissionOptions,
} from "~/types/gallery";

interface BulkResult {
  changed: number;
  failed: number;
  interrupted?: boolean;
  message?: string;
}

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "图片 · 图库管理" });

const ALL = "__all__" as const;
const router = useRouter();
const { call } = useApi();
const hydrated = useClientHydrated();
const toast = createPlatformNotifier(useToast());

type ImageStatus = "" | "published" | "draft" | "hidden" | "deleted";
type ImageSort = "created" | "updated" | "title";
type ImageDirection = "asc" | "desc";
type ImageView = "list" | "grid";
interface ImageCollectionQuery {
  q: string;
  status: ImageStatus;
  page: number;
  size: number;
  sort: ImageSort;
  direction: ImageDirection;
  view: ImageView;
  category: string;
  facet: string;
}
const statuses = ["", "published", "draft", "hidden", "deleted"] as const;
const sorts = ["created", "updated", "title"] as const;
const directions = ["asc", "desc"] as const;
const views = ["list", "grid"] as const;
const pageSizes = [12, 24, 48, 60] as const;
const defaultQuery: ImageCollectionQuery = {
  q: "",
  status: "",
  page: 1,
  size: 24,
  sort: "created",
  direction: "desc",
  view: "list",
  category: ALL,
  facet: ALL,
};
const counts = ref<Record<string, number>>({});
function apiSort(query: Readonly<ImageCollectionQuery>) {
  if (query.sort === "title")
    return query.direction === "asc" ? "title_asc" : "title_desc";
  if (query.sort === "updated")
    return query.direction === "asc" ? "updated_asc" : "updated";
  return query.direction === "asc" ? "oldest" : "newest";
}
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
    const data = await call<GalleryAdminImagePage>(
      "/api/v1/gallery/admin/images",
      {
        query: {
          q: nextQuery.q || undefined,
          sort: apiSort(nextQuery),
          page: nextQuery.page,
          size: nextQuery.size,
          publicationState: nextQuery.status || undefined,
          categoryId:
            nextQuery.category === ALL ? undefined : nextQuery.category,
          facetValueId: nextQuery.facet === ALL ? undefined : nextQuery.facet,
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
    sort: { kind: "enum", values: sorts, default: defaultQuery.sort },
    direction: {
      kind: "enum",
      values: directions,
      default: defaultQuery.direction,
    },
    view: { kind: "enum", values: views, default: defaultQuery.view },
    category: {
      kind: "string",
      default: defaultQuery.category,
      maxLength: 200,
    },
    facet: { kind: "string", default: defaultQuery.facet, maxLength: 200 },
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
  querySync,
  dataQueryKey: (query) => JSON.stringify({ ...query, view: undefined }),
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
const sort = computed({
  get: () => collectionQuery.value.sort,
  set: (value: ImageSort) => updateCollectionQuery({ sort: value }),
});
const direction = computed({
  get: () => collectionQuery.value.direction,
  set: (value: ImageDirection) => updateCollectionQuery({ direction: value }),
});
const view = computed({
  get: () => collectionQuery.value.view,
  set: (value: ImageView) => updateCollectionQuery({ view: value }, false),
});
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
      "/api/v1/gallery/admin/classification/tags?size=100",
    ),
  { server: false, default: () => ({ page: { items: [], nextCursor: "" } }) },
);
const tagOptions = computed(() =>
  tagData.value.page.items
    .filter((item) => item.status === "active")
    .map((item) => ({ label: item.name, value: item.id })),
);
const sortOptions = [
  { label: "创建时间", value: "created" },
  { label: "更新时间", value: "updated" },
  { label: "标题", value: "title" },
];
const statusOptions = computed(() => [
  { value: "", label: `全部 · ${counts.value.all || 0}` },
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
    value: status.value,
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
  {
    kind: "select",
    id: "sort",
    label: "图片排序",
    value: sort.value,
    options: sortOptions,
    icon: "i-tabler-arrows-sort",
    class: "w-32",
  },
  {
    kind: "direction",
    id: "direction",
    label: "排序方向",
    value: direction.value,
    ascendingLabel: "切换为倒序",
    descendingLabel: "切换为正序",
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
  if (id === "status" && statuses.includes(value as ImageStatus))
    status.value = value as ImageStatus;
  if (id === "category") category.value = value;
  if (id === "facet") facet.value = value;
  if (id === "sort" && sorts.includes(value as ImageSort))
    sort.value = value as ImageSort;
  if (id === "direction" && directions.includes(value as ImageDirection))
    direction.value = value as ImageDirection;
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
const editPending = ref(false);
const editLoading = ref(false);
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
async function openEdit(item: GalleryAdminImage) {
  hydrateEditForm(item);
  editLoading.value = true;
  try {
    const response = await call<{ image: GalleryAdminImage }>(
      `/api/v1/gallery/admin/images/${encodeURIComponent(item.id)}`,
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
    !editing.value?.updatedAt ||
    !editForm.title.trim() ||
    !editForm.altText.trim() ||
    editForm.primaryCategoryId === ALL
  )
    return;
  editPending.value = true;
  try {
    await call(
      `/api/v1/gallery/admin/images/${encodeURIComponent(editing.value.id)}`,
      {
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
      },
    );
    editing.value = undefined;
    await refresh();
  } catch (reason: any) {
    toast.add({
      title: "图片信息没有保存",
      description: reason?.data?.message || "记录可能已被其他人更新",
      color: "error",
    });
    await refresh();
  } finally {
    editPending.value = false;
  }
}

const batchAction = ref<string>();
const batchOpen = ref(false);
const batchPending = ref(false);
const batchCategory = ref(ALL);
const batchReason = ref("");
const batchResult = ref<BulkResult>();
const batchOptions = [
  { label: "设置主分类", value: "set_primary_category" },
  { label: "下架", value: "hide" },
];
function prepareBatch() {
  if (batchAction.value && selectionCount.value) batchOpen.value = true;
}
function cancelBatch() {
  batchOpen.value = false;
}
async function runBatch() {
  if (!batchAction.value || !selectedIds.value.length) return;
  batchPending.value = true;
  try {
    const response = await call<{
      results: Array<{ imageId: string; success: boolean; error?: string }>;
    }>("/api/v1/gallery/admin/images/bulk", {
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

function anomalyBadges(image: GalleryAdminImage) {
  const badges: Array<{
    label: string;
    color: "warning" | "error" | "neutral";
  }> = [];
  if (!image.publicRenditionReady)
    badges.push({ label: "公开版本未就绪", color: "warning" });
  return badges;
}
function moreItems(image: GalleryAdminImage) {
  return image.publicationState === "published"
    ? [
        [
          {
            label: "打开公开页",
            icon: "i-tabler-external-link",
            to: `/images/${image.id}`,
            target: "_blank",
          },
        ],
      ]
    : [];
}
</script>

<template>
  <div>
    <PageHeader title="图片">
      <template #subtitle>管理已审核图片的公开状态、分类、维度与标签</template>
      <template #actions
        ><UButton to="/submit" icon="i-tabler-upload" label="投稿图片"
      /></template>
    </PageHeader>

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
      :layout="view === 'grid' ? 'grid' : 'rows'"
      :selection-count="selectionCount"
      :page-selected="isPageSelected"
      :page-indeterminate="isPageIndeterminate"
      :is-selected="imageWorkflow.isSelected"
      label="图片列表"
      selectable
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
          v-model="view"
          :items="[
            { key: 'list', label: '列表视图', icon: 'i-tabler-list' },
            { key: 'grid', label: '网格视图', icon: 'i-tabler-layout-grid' },
          ]"
        />
      </template>

      <template #columns>
        <div class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3">
          <span>图片、分类与浏览数据</span>
          <span class="hidden w-32 text-right md:block">状态与操作</span>
        </div>
      </template>

      <template #bulk-actions>
        <USelect
          v-model="batchAction"
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
          v-if="view === 'list'"
          class="grid min-w-0 gap-3 sm:grid-cols-[5rem_minmax(0,1fr)] md:grid-cols-[5rem_minmax(0,1fr)_auto] md:items-center"
        >
          <img
            :src="galleryRendition(image.assetId, 'thumbnail')"
            :alt="image.altText"
            class="hidden aspect-[4/3] w-20 rounded-lg bg-elevated object-cover sm:block"
          />
          <div class="min-w-0">
            <div class="flex min-w-0 flex-wrap items-center gap-2">
              <p class="truncate text-sm font-medium text-highlighted">
                {{ image.title }}
              </p>
              <UBadge
                v-for="badge in anomalyBadges(image)"
                :key="badge.label"
                :color="badge.color"
                variant="subtle"
                :label="badge.label"
              />
            </div>
            <div
              class="mt-1 flex min-w-0 items-center gap-2 text-xs text-muted"
            >
              <span class="truncate font-mono">{{ image.id }}</span
              ><span class="text-dimmed">·</span
              ><span class="shrink-0"
                >{{ compactMetric(image.metrics.views) }} 次浏览</span
              >
            </div>
            <ManageTaxonomyChips
              class="mt-1.5"
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
          <div class="flex justify-end gap-1">
            <UButton
              color="primary"
              variant="soft"
              size="xs"
              icon="i-tabler-pencil"
              label="编辑图片"
              @click="openEdit(image)"
            />
            <UDropdownMenu
              v-if="moreItems(image).length"
              :items="moreItems(image)"
              ><UButton
                color="neutral"
                variant="ghost"
                size="xs"
                icon="i-tabler-dots"
                square
                :aria-label="`更多操作：${image.title}`"
            /></UDropdownMenu>
          </div>
        </div>

        <div v-else class="-m-4 overflow-hidden rounded-lg">
          <div
            class="group relative aspect-[4/3] overflow-hidden border-b border-default bg-elevated"
          >
            <img
              :src="galleryRendition(image.assetId, 'grid-sm')"
              :alt="image.altText"
              class="size-full object-cover transition-transform duration-300 group-hover:scale-[1.02]"
            />
            <UButton
              class="absolute right-2 top-2"
              color="primary"
              variant="solid"
              size="xs"
              icon="i-tabler-pencil"
              square
              aria-label="编辑图片"
              @click="openEdit(image)"
            />
          </div>
          <div class="p-3">
            <h2 class="truncate text-sm font-medium text-highlighted">
              {{ image.title }}
            </h2>
            <p class="mt-1 truncate text-xs text-muted">
              {{ image.primaryCategory || "未分类" }} ·
              {{ compactMetric(image.metrics.views) }} 次浏览
            </p>
            <div
              v-if="anomalyBadges(image).length"
              class="mt-2 flex flex-wrap gap-1"
            >
              <UBadge
                v-for="badge in anomalyBadges(image)"
                :key="badge.label"
                :color="badge.color"
                variant="subtle"
                :label="badge.label"
              />
            </div>
          </div>
        </div>
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
      :open="Boolean(editing)"
      title="编辑图片"
      description="统一维护目录文案、主分类、维度和标签。"
      @update:open="
        (open) => {
          if (!open) editing = undefined;
        }
      "
    >
      <template #body>
        <div v-if="editLoading" class="space-y-3 py-2">
          <USkeleton class="h-10 rounded-lg" />
          <USkeleton class="h-24 rounded-lg" />
          <USkeleton class="h-32 rounded-lg" />
        </div>
        <form v-else class="space-y-5" @submit.prevent="saveEdit">
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
              :items="categoryOptions.filter((item) => item.value !== ALL)"
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
              :items="tagOptions"
              value-key="value"
              multiple
              class="w-full"
              :search-input="{ placeholder: '搜索标签…' }"
              placeholder="选择标签"
            />
          </UFormField>
          <div class="flex justify-end gap-2">
            <UButton
              color="neutral"
              variant="outline"
              label="取消"
              @click="editing = undefined"
            /><UButton
              type="submit"
              label="保存更改"
              :loading="editPending"
              :disabled="editForm.primaryCategoryId === ALL"
            />
          </div>
        </form>
      </template>
    </UModal>

    <UModal
      v-model:open="batchOpen"
      :title="batchAction === 'hide' ? '批量下架' : '批量设置主分类'"
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
                  : batchCategory === ALL
              "
            />
          </div></form
      ></template>
    </UModal>
  </div>
</template>
