<script setup lang="ts">
import { createPlatformNotifier } from "@platform/ui/feedback";
import {
  ManageActiveFilters,
  ManageCollectionFooter,
  ManageCollectionToolbar,
  ManageEmpty,
  ManageHeader,
  ManageLifecycleTabs,
  ManagePageSelection,
  ManageRowShell,
  ManageSortDirectionButton,
  ManageTaxonomyChips,
  ManageViewToggle,
  SkeletonList,
} from "@platform/manage/components";
import {
  manageCollectionQueryFingerprint,
  serializeManageCollectionQuery,
  type ManageCollectionDefinition,
} from "@platform/manage/collection";
import { useManageCollectionState } from "@platform/manage/use-manage-collection-state";
import { useManageSelection } from "@platform/manage/use-manage-selection";
import type {
  GalleryAdminImage,
  GalleryAdminImagePage,
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

const ALL = "__all__";
const route = useRoute();
const router = useRouter();
const { call } = useApi();
const hydrated = useClientHydrated();
const toast = createPlatformNotifier(useToast());

const collectionDefinition = {
  resourceKind: "gallery-image",
  statuses: ["", "published", "draft", "hidden", "deleted"],
  views: ["list", "grid"],
  sortKeys: ["created", "updated", "title"],
  pageSizes: [12, 24, 48, 60],
  defaultStatus: "",
  defaultView: "list",
  defaultSort: "created",
  defaultDirection: "desc",
  defaultPageSize: 24,
  pagination: "server",
  selection: "page",
  filters: ["category", "facet", "processing", "review", "safety"],
  quickEditFields: ["title", "description", "altText", "sourceUrl"],
  bulkActions: ["set_primary_category", "hide"],
} as const satisfies ManageCollectionDefinition;

const {
  status,
  searchInput,
  q,
  page,
  size,
  sort,
  direction,
  view,
  state: collectionState,
  filterModel,
} = useManageCollectionState({
  definition: collectionDefinition,
  routeQuery: () => route.query,
  replaceQuery: (query) => router.replace({ query }),
});
const category = filterModel("category", ALL);
const facet = filterModel("facet", ALL);
const processing = filterModel("processing", ALL);
const review = filterModel("review", ALL);
const safety = filterModel("safety", ALL);

const apiSort = computed(() => {
  if (sort.value === "title")
    return direction.value === "asc" ? "title_asc" : "title_desc";
  if (sort.value === "updated")
    return direction.value === "asc" ? "updated_asc" : "updated";
  return direction.value === "asc" ? "oldest" : "newest";
});

const { data, pending, error, refresh } = await useAsyncData(
  "gallery-manage-images",
  () =>
    call<GalleryAdminImagePage>("/api/v1/gallery/admin/images", {
      query: {
        q: q.value || undefined,
        sort: apiSort.value,
        page: page.value,
        size: size.value,
        publicationState: status.value || undefined,
        categoryId: category.value === ALL ? undefined : category.value,
        facetValueId: facet.value === ALL ? undefined : facet.value,
        processingState:
          processing.value === ALL ? undefined : processing.value,
        reviewState: review.value === ALL ? undefined : review.value,
        safetyState: safety.value === ALL ? undefined : safety.value,
      },
    }),
  {
    server: false,
    watch: [
      q,
      apiSort,
      page,
      size,
      status,
      category,
      facet,
      processing,
      review,
      safety,
    ],
    default: () => ({
      items: [],
      page: 1,
      pageSize: 24,
      total: 0,
      totalPages: 0,
      counts: {},
    }),
  },
);

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
const processingOptions = [
  { label: "全部处理状态", value: ALL },
  { label: "排队中", value: "queued" },
  { label: "处理中", value: "processing" },
  { label: "已就绪", value: "ready" },
  { label: "处理失败", value: "failed" },
];
const reviewOptions = [
  { label: "全部审核状态", value: ALL },
  { label: "无需审核", value: "not_required" },
  { label: "等待审核", value: "pending" },
  { label: "已批准", value: "approved" },
  { label: "已拒绝", value: "rejected" },
];
const safetyOptions = [
  { label: "全部安全状态", value: ALL },
  { label: "等待检查", value: "pending" },
  { label: "安全", value: "safe" },
  { label: "不确定", value: "uncertain" },
  { label: "已阻止", value: "blocked" },
  { label: "不可用", value: "unavailable" },
];
const sortOptions = [
  { label: "创建时间", value: "created" },
  { label: "更新时间", value: "updated" },
  { label: "标题", value: "title" },
];
const counts = computed<Record<string, number>>(() => data.value.counts || {});
const tabs = computed(() => [
  { key: "", label: "全部", count: counts.value.all || 0 },
  { key: "published", label: "已公开", count: counts.value.published || 0 },
  { key: "draft", label: "草稿", count: counts.value.draft || 0 },
  { key: "hidden", label: "已隐藏", count: counts.value.hidden || 0 },
  { key: "deleted", label: "已删除", count: counts.value.deleted || 0 },
]);
const activeFilters = computed(() => [
  ...(category.value !== ALL
    ? [
        {
          key: "category",
          label: `分类：${categoryOptions.value.find((item) => item.value === category.value)?.label || "未知"}`,
        },
      ]
    : []),
  ...(facet.value !== ALL
    ? [
        {
          key: "facet",
          label: `维度：${facetOptions.value.find((item) => item.value === facet.value)?.label || "未知"}`,
        },
      ]
    : []),
  ...(processing.value !== ALL
    ? [
        {
          key: "processing",
          label:
            processingOptions.find((item) => item.value === processing.value)
              ?.label || processing.value,
        },
      ]
    : []),
  ...(review.value !== ALL
    ? [
        {
          key: "review",
          label:
            reviewOptions.find((item) => item.value === review.value)?.label ||
            review.value,
        },
      ]
    : []),
  ...(safety.value !== ALL
    ? [
        {
          key: "safety",
          label:
            safetyOptions.find((item) => item.value === safety.value)?.label ||
            safety.value,
        },
      ]
    : []),
]);
function removeFilter(key: string) {
  if (key === "category") category.value = ALL;
  if (key === "facet") facet.value = ALL;
  if (key === "processing") processing.value = ALL;
  if (key === "review") review.value = ALL;
  if (key === "safety") safety.value = ALL;
}
function clearFilters() {
  category.value =
    facet.value =
    processing.value =
    review.value =
    safety.value =
      ALL;
}

const items = computed(() => data.value.items || []);
const selectionResetKey = computed(() =>
  manageCollectionQueryFingerprint(
    serializeManageCollectionQuery(collectionState.value, collectionDefinition),
  ),
);
const {
  selectedIds,
  selectionCount,
  isPageSelected,
  isPageIndeterminate,
  isSelected,
  toggleOne,
  togglePage,
  replace: replaceSelection,
  clear: clearSelection,
} = useManageSelection({
  visibleIds: computed(() => items.value.map((item) => item.id)),
  filteredTotal: computed(() => data.value.total || 0),
  resetKey: selectionResetKey,
});

const editing = shallowRef<GalleryAdminImage>();
const editPending = ref(false);
const editForm = reactive({
  title: "",
  description: "",
  altText: "",
  sourceUrl: "",
});
function openEdit(item: GalleryAdminImage) {
  editing.value = item;
  Object.assign(editForm, {
    title: item.title,
    description: item.description,
    altText: item.altText,
    sourceUrl: item.sourceUrl,
  });
}
async function saveEdit() {
  if (
    !editing.value?.updatedAt ||
    !editForm.title.trim() ||
    !editForm.altText.trim()
  )
    return;
  editPending.value = true;
  try {
    await call(
      `/api/v1/gallery/admin/images/${encodeURIComponent(editing.value.id)}`,
      {
        method: "PATCH",
        body: { expectedUpdatedAt: editing.value.updatedAt, ...editForm },
      },
    );
    editing.value = undefined;
    toast.add({ title: "图片信息已保存", color: "success" });
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
  if (image.processingState === "failed")
    badges.push({ label: "处理失败", color: "error" });
  else if (["queued", "processing"].includes(image.processingState))
    badges.push({
      label: image.processingState === "queued" ? "排队中" : "处理中",
      color: "warning",
    });
  if (image.reviewState === "pending")
    badges.push({ label: "待审核", color: "warning" });
  if (image.reviewState === "rejected")
    badges.push({ label: "审核拒绝", color: "error" });
  if (!["safe"].includes(image.safetyState))
    badges.push({
      label:
        safetyOptions.find((item) => item.value === image.safetyState)?.label ||
        image.safetyState,
      color: image.safetyState === "blocked" ? "error" : "warning",
    });
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
    <ManageHeader title="图片">
      <template #subtitle>管理图片生命周期、分类和质量异常</template>
      <template #actions
        ><UButton to="/submit" icon="i-tabler-upload" label="投稿图片"
      /></template>
    </ManageHeader>

    <ManageLifecycleTabs v-model="status" :items="tabs" class="mb-4" />
    <ManageCollectionToolbar
      v-model:search="searchInput"
      search-placeholder="搜索标题、说明或替代文本…"
      :filter-count="activeFilters.length"
      class="mb-3"
    >
      <template #filters>
        <USelectMenu
          v-model="category"
          :items="categoryOptions"
          value-key="value"
          icon="i-tabler-folders"
          size="sm"
          class="w-full sm:w-36"
          :search-input="{ placeholder: '搜索分类…' }"
        />
        <USelectMenu
          v-model="facet"
          :items="facetOptions"
          value-key="value"
          icon="i-tabler-adjustments"
          size="sm"
          class="w-full sm:w-40"
          :search-input="{ placeholder: '搜索维度…' }"
        />
        <USelect
          v-model="processing"
          :items="processingOptions"
          value-key="value"
          icon="i-tabler-progress"
          size="sm"
          class="w-full sm:w-36"
        />
        <USelect
          v-model="review"
          :items="reviewOptions"
          value-key="value"
          icon="i-tabler-clipboard-check"
          size="sm"
          class="w-full sm:w-36"
        />
        <USelect
          v-model="safety"
          :items="safetyOptions"
          value-key="value"
          icon="i-tabler-shield-check"
          size="sm"
          class="w-full sm:w-36"
        />
        <USelect
          v-model="sort"
          :items="sortOptions"
          value-key="value"
          icon="i-tabler-arrows-sort"
          size="sm"
          class="w-full sm:w-32"
        />
        <ManageSortDirectionButton v-model="direction" />
      </template>
      <template #actions>
        <ManageViewToggle
          v-model="view"
          :items="[
            { key: 'list', label: '列表视图', icon: 'i-tabler-list' },
            { key: 'grid', label: '网格视图', icon: 'i-tabler-layout-grid' },
          ]"
        />
      </template>
    </ManageCollectionToolbar>
    <ManageActiveFilters
      :items="activeFilters"
      class="mb-4 px-1"
      @remove="removeFilter"
      @clear="clearFilters"
    />

    <SkeletonList v-if="!hydrated || pending" :rows="8" />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="图片加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <ManageEmpty
      v-else-if="!items.length"
      icon="i-tabler-photo-off"
      title="没有符合条件的图片"
      description="调整筛选或搜索条件后重试。"
    />

    <div
      v-else-if="view === 'list'"
      class="overflow-hidden rounded-xl border border-default"
    >
      <ManageRowShell
        v-for="image in items"
        :key="image.id"
        :selected="isSelected(image.id)"
        :selection-label="`选择图片：${image.title}`"
        @select="toggleOne(image.id)"
      >
        <template #media
          ><img
            :src="galleryRendition(image.assetId, 'thumbnail')"
            :alt="image.altText"
            class="aspect-[4/3] w-20 rounded-lg bg-elevated object-cover"
        /></template>
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
          <div class="mt-1 flex min-w-0 items-center gap-2 text-xs text-muted">
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
        <template #actions>
          <UButton
            color="primary"
            variant="soft"
            size="sm"
            icon="i-tabler-pencil"
            label="快速编辑"
            @click="openEdit(image)"
          />
          <UDropdownMenu
            v-if="moreItems(image).length"
            :items="moreItems(image)"
            ><UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-tabler-dots"
              square
              :aria-label="`更多操作：${image.title}`"
          /></UDropdownMenu>
        </template>
      </ManageRowShell>
    </div>

    <div
      v-else
      class="grid grid-cols-[repeat(auto-fill,minmax(13rem,1fr))] gap-4"
    >
      <article
        v-for="image in items"
        :key="image.id"
        class="group overflow-hidden rounded-xl border border-default bg-default"
        :class="isSelected(image.id) ? 'ring-2 ring-primary' : ''"
      >
        <div class="relative aspect-[4/3] overflow-hidden bg-elevated">
          <img
            :src="galleryRendition(image.assetId, 'grid-sm')"
            :alt="image.altText"
            class="size-full object-cover transition-transform duration-300 group-hover:scale-[1.02]"
          />
          <UCheckbox
            class="absolute left-2 top-2 rounded-md bg-default/85 p-1 backdrop-blur"
            :model-value="isSelected(image.id)"
            :aria-label="`选择图片：${image.title}`"
            @update:model-value="toggleOne(image.id)"
          />
          <UButton
            class="absolute right-2 top-2"
            color="primary"
            variant="solid"
            size="xs"
            icon="i-tabler-pencil"
            square
            aria-label="快速编辑"
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
      </article>
    </div>

    <ManageCollectionFooter
      v-if="items.length"
      v-model:page="page"
      v-model:size="size"
      :total="data.total"
      :total-pages="Math.max(1, data.totalPages)"
      :page-size-options="[12, 24, 48, 60]"
      label="图片选择、批量操作与分页"
    >
      <template #selection>
        <ManagePageSelection
          :model-value="isPageSelected"
          :indeterminate="isPageIndeterminate"
          label="选择当前页图片"
          @update:model-value="togglePage"
        />
        <div
          v-if="batchResult"
          class="flex items-center gap-2 rounded-lg bg-elevated px-2.5 py-1.5 text-xs"
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
          <span>{{
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
        <template v-if="selectionCount">
          <span class="text-sm">已选 {{ selectionCount }}</span
          ><span class="h-4 w-px bg-default" />
          <USelect
            v-model="batchAction"
            :items="batchOptions"
            placeholder="批量操作"
            size="sm"
            class="w-32"
          />
          <UButton
            size="sm"
            color="primary"
            variant="soft"
            :disabled="!batchAction"
            @click="prepareBatch"
            >应用</UButton
          >
          <UButton
            size="sm"
            color="neutral"
            variant="ghost"
            @click="clearSelection"
            >取消</UButton
          >
        </template>
        <span v-else class="text-xs">共 {{ data.total }} 张</span>
      </template>
    </ManageCollectionFooter>

    <UModal
      :open="Boolean(editing)"
      title="快速编辑图片"
      description="修改最常用的图片元数据。"
      @update:open="
        (open) => {
          if (!open) editing = undefined;
        }
      "
    >
      <template #body
        ><form class="space-y-4" @submit.prevent="saveEdit">
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
          <div class="flex justify-end gap-2">
            <UButton
              color="neutral"
              variant="outline"
              label="取消"
              @click="editing = undefined"
            /><UButton type="submit" label="保存" :loading="editPending" />
          </div></form
      ></template>
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
