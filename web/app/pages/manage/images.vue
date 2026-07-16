<script setup lang="ts">
import { createPlatformNotifier } from "@platform/ui/feedback";
import {
  ManageEmpty,
  ManageHeader,
  SkeletonList,
} from "@platform/manage/components";
import type { GalleryAdminImage, GalleryAdminImagePage } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "图片 · 图库管理" });
const route = useRoute();
const router = useRouter();
const { call } = useApi();
const hydrated = useClientHydrated();
const toast = createPlatformNotifier(useToast());
const allowed = <T extends string>(
  value: unknown,
  values: readonly T[],
  fallback: T,
) => (values.includes(String(value) as T) ? (String(value) as T) : fallback);
const page = computed(() => Math.max(1, Number(route.query.page) || 1));
const q = computed(() => String(route.query.q || ""));
const qDraft = ref(q.value);
const view = computed(() =>
  allowed(route.query.view, ["list", "grid"] as const, "list"),
);
const sort = computed(() =>
  allowed(
    route.query.sort,
    ["newest", "updated", "oldest", "title_asc", "title_desc"] as const,
    "newest",
  ),
);
const processingState = computed(() =>
  String(route.query.processingState || ""),
);
const reviewState = computed(() => String(route.query.reviewState || ""));
const publicationState = computed(() =>
  String(route.query.publicationState || ""),
);
const safetyState = computed(() => String(route.query.safetyState || ""));
watch(q, (value) => {
  qDraft.value = value;
});

const { data, pending, error, refresh } = await useAsyncData(
  "gallery-manage-images",
  () =>
    call<GalleryAdminImagePage>("/api/v1/gallery/admin/images", {
      query: {
        q: q.value || undefined,
        sort: sort.value,
        page: page.value,
        size: 24,
        processingState: processingState.value || undefined,
        reviewState: reviewState.value || undefined,
        publicationState: publicationState.value || undefined,
        safetyState: safetyState.value || undefined,
      },
    }),
  {
    server: false,
    watch: [
      q,
      sort,
      page,
      processingState,
      reviewState,
      publicationState,
      safetyState,
    ],
    default: () => ({
      items: [],
      page: 1,
      pageSize: 24,
      total: 0,
      totalPages: 0,
    }),
  },
);

const selectedIDs = ref<Set<string>>(new Set());
const bulkOpen = ref(false);
const bulkReason = ref("");
const bulkPending = ref(false);
const editing = shallowRef<GalleryAdminImage>();
const editPending = ref(false);
const editForm = reactive({
  title: "",
  description: "",
  altText: "",
  sourceUrl: "",
});
const filterCount = computed(
  () =>
    [
      processingState.value,
      reviewState.value,
      publicationState.value,
      safetyState.value,
    ].filter(Boolean).length,
);
const allPageSelected = computed(
  () =>
    data.value.items.length > 0 &&
    data.value.items.every((item) => selectedIDs.value.has(item.id)),
);

const sortItems = [
  { label: "最新创建", value: "newest" },
  { label: "最近更新", value: "updated" },
  { label: "最早创建", value: "oldest" },
  { label: "标题 A-Z", value: "title_asc" },
  { label: "标题 Z-A", value: "title_desc" },
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
const publicationItems = [
  { label: "全部发布状态", value: "all" },
  { label: "草稿", value: "draft" },
  { label: "已公开", value: "published" },
  { label: "已隐藏", value: "hidden" },
  { label: "已删除", value: "deleted" },
];
const safetyItems = [
  { label: "全部安全状态", value: "all" },
  { label: "等待检查", value: "pending" },
  { label: "安全", value: "safe" },
  { label: "不确定", value: "uncertain" },
  { label: "已阻止", value: "blocked" },
  { label: "不可用", value: "unavailable" },
];
const lifecycleLabel: Record<string, string> = {
  draft: "草稿",
  published: "已公开",
  hidden: "已隐藏",
  deleted: "已删除",
};
const safetyLabel: Record<string, string> = {
  pending: "等待检查",
  safe: "安全",
  uncertain: "不确定",
  blocked: "已阻止",
  unavailable: "不可用",
};
const processingLabel: Record<string, string> = {
  queued: "排队中",
  processing: "处理中",
  ready: "已就绪",
  failed: "处理失败",
};
const reviewLabel: Record<string, string> = {
  not_required: "无需审核",
  pending: "等待审核",
  approved: "已批准",
  rejected: "已拒绝",
};
const lifecycleColor = (value: string) =>
  (({ published: "success", hidden: "warning", deleted: "error" })[value] ||
    "neutral") as any;

function setQuery(values: Record<string, string | number | undefined>) {
  const resetsPage = !("page" in values);
  const query = Object.fromEntries(
    Object.entries(route.query).filter(
      ([key]) => !(key in values) && !(resetsPage && key === "page"),
    ),
  );
  for (const [key, value] of Object.entries(values)) {
    if (
      value === undefined ||
      value === "" ||
      value === "all" ||
      (key === "page" && value === 1) ||
      (key === "view" && value === "list") ||
      (key === "sort" && value === "newest")
    )
      continue;
    else query[key] = String(value);
  }
  void router.push({ query });
}
function search() {
  setQuery({ q: qDraft.value.trim() || undefined });
}
function clearFilters() {
  void router.push({ query: view.value === "grid" ? { view: "grid" } : {} });
}
function toggle(id: string) {
  const next = new Set(selectedIDs.value);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  selectedIDs.value = next;
}
function togglePage() {
  const next = new Set(selectedIDs.value);
  if (allPageSelected.value)
    data.value.items.forEach((item) => next.delete(item.id));
  else data.value.items.forEach((item) => next.add(item.id));
  selectedIDs.value = next;
}
function clearSelection() {
  selectedIDs.value = new Set();
}
function openBulkHide() {
  bulkOpen.value = true;
}
function closeBulkHide() {
  bulkOpen.value = false;
}
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
        body: {
          expectedUpdatedAt: editing.value.updatedAt,
          title: editForm.title.trim(),
          description: editForm.description.trim(),
          altText: editForm.altText.trim(),
          sourceUrl: editForm.sourceUrl.trim(),
        },
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
async function bulkHide() {
  if (!selectedIDs.value.size || !bulkReason.value.trim()) return;
  bulkPending.value = true;
  try {
    const response = await call<{
      results: Array<{ imageId: string; success: boolean; error?: string }>;
    }>("/api/v1/gallery/admin/images/bulk-hide", {
      method: "POST",
      body: {
        imageIds: [...selectedIDs.value],
        reason: bulkReason.value.trim(),
      },
    });
    const succeeded = response.results.filter((item) => item.success);
    const failed = response.results.length - succeeded.length;
    const next = new Set(selectedIDs.value);
    succeeded.forEach((item) => next.delete(item.imageId));
    selectedIDs.value = next;
    bulkOpen.value = false;
    bulkReason.value = "";
    toast.add({
      title: `已下架 ${succeeded.length} 张图片`,
      description: failed
        ? `${failed} 张状态不允许下架，已保留选择`
        : undefined,
      color: failed ? "warning" : "success",
    });
    await refresh();
  } catch (reason: any) {
    toast.add({
      title: "批量下架没有完成",
      description: reason?.data?.message || "请检查连接后重试，当前选择已保留",
      color: "error",
    });
  } finally {
    bulkPending.value = false;
  }
}
</script>

<template>
  <div>
    <ManageHeader title="图片资产台"
      ><template #subtitle
        >查看全部生命周期图片，按处理、审核、发布与安全状态筛选，并执行可恢复的元数据编辑和批量下架。</template
      ></ManageHeader
    >

    <section
      class="mb-5 space-y-3 rounded-xl border border-default bg-default p-4"
      aria-label="图片查询与筛选"
    >
      <div class="flex flex-wrap gap-2">
        <UInput
          v-model="qDraft"
          class="min-w-64 flex-1"
          icon="i-tabler-search"
          placeholder="搜索标题、说明或替代文本"
          @keyup.enter="search"
        />
        <UButton
          label="搜索"
          color="neutral"
          variant="outline"
          @click="search"
        />
        <USelect
          :model-value="sort"
          :items="sortItems"
          value-key="value"
          class="w-36"
          aria-label="图片排序"
          @update:model-value="setQuery({ sort: String($event) })"
        />
        <div
          class="flex rounded-md border border-default p-0.5"
          aria-label="图片视图"
        >
          <UButton
            color="neutral"
            :variant="view === 'list' ? 'soft' : 'ghost'"
            size="sm"
            icon="i-tabler-list"
            aria-label="列表视图"
            @click="setQuery({ view: 'list' })"
          />
          <UButton
            color="neutral"
            :variant="view === 'grid' ? 'soft' : 'ghost'"
            size="sm"
            icon="i-tabler-layout-grid"
            aria-label="网格视图"
            @click="setQuery({ view: 'grid' })"
          />
        </div>
      </div>
      <div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
        <USelect
          :model-value="processingState || 'all'"
          :items="processingItems"
          value-key="value"
          @update:model-value="setQuery({ processingState: String($event) })"
        />
        <USelect
          :model-value="reviewState || 'all'"
          :items="reviewItems"
          value-key="value"
          @update:model-value="setQuery({ reviewState: String($event) })"
        />
        <USelect
          :model-value="publicationState || 'all'"
          :items="publicationItems"
          value-key="value"
          @update:model-value="setQuery({ publicationState: String($event) })"
        />
        <USelect
          :model-value="safetyState || 'all'"
          :items="safetyItems"
          value-key="value"
          @update:model-value="setQuery({ safetyState: String($event) })"
        />
      </div>
      <div class="flex items-center justify-between gap-3 text-xs text-muted">
        <span
          >共 {{ data.total }} 张<span v-if="selectedIDs.size"
            >，已选择 {{ selectedIDs.size }} 张</span
          ></span
        >
        <UButton
          v-if="filterCount || q"
          color="neutral"
          variant="link"
          size="xs"
          label="清除查询"
          @click="clearFilters"
        />
      </div>
    </section>

    <div
      v-if="selectedIDs.size"
      class="sticky top-3 z-20 mb-4 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-primary/25 bg-default/95 p-3 shadow-lg backdrop-blur"
    >
      <div class="flex items-center gap-2">
        <UCheckbox
          :model-value="allPageSelected"
          label="选择本页"
          @update:model-value="togglePage"
        /><span class="text-sm text-muted"
          >已选择 {{ selectedIDs.size }} 张</span
        >
      </div>
      <div class="flex gap-2">
        <UButton
          color="neutral"
          variant="ghost"
          label="取消选择"
          @click="clearSelection"
        /><UButton
          color="error"
          variant="soft"
          icon="i-tabler-eye-off"
          label="批量下架"
          @click="openBulkHide"
        />
      </div>
    </div>

    <SkeletonList v-if="!hydrated || pending" :rows="8" />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="图片加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <div
      v-else-if="data.items.length && view === 'list'"
      class="divide-y divide-default border-y border-default"
    >
      <article
        v-for="image in data.items"
        :key="image.id"
        class="grid grid-cols-[auto_4rem_minmax(0,1fr)] items-center gap-3 py-3 sm:grid-cols-[auto_5rem_minmax(0,1fr)_auto]"
      >
        <UCheckbox
          :model-value="selectedIDs.has(image.id)"
          :aria-label="`选择 ${image.title}`"
          @update:model-value="toggle(image.id)"
        />
        <img
          :src="galleryRendition(image.assetId, 'thumbnail')"
          :alt="image.altText"
          class="aspect-[4/3] w-16 rounded-md bg-elevated object-cover sm:w-20"
        />
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <p class="truncate font-medium text-highlighted">
              {{ image.title }}
            </p>
            <UBadge
              :color="lifecycleColor(image.publicationState)"
              variant="soft"
              :label="lifecycleLabel[image.publicationState]"
            /><UBadge
              :color="
                image.safetyState === 'safe'
                  ? 'success'
                  : image.safetyState === 'blocked'
                    ? 'error'
                    : 'warning'
              "
              variant="subtle"
              :label="safetyLabel[image.safetyState]"
            />
          </div>
          <p class="mt-1 text-xs text-muted">
            {{ image.primaryCategory || "未分类" }} ·
            {{ processingLabel[image.processingState] }} /
            {{ reviewLabel[image.reviewState] }} ·
            {{ compactMetric(image.metrics.views) }} 次浏览
          </p>
        </div>
        <div class="col-start-3 flex gap-1 sm:col-auto">
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-tabler-edit"
            aria-label="快速编辑"
            @click="openEdit(image)"
          /><UButton
            v-if="image.publicationState === 'published'"
            :to="`/images/${image.id}`"
            target="_blank"
            color="neutral"
            variant="ghost"
            icon="i-tabler-external-link"
            aria-label="打开公开图片"
          />
        </div>
      </article>
    </div>
    <div
      v-else-if="data.items.length"
      class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4"
    >
      <article
        v-for="image in data.items"
        :key="image.id"
        class="overflow-hidden rounded-xl border border-default bg-default"
      >
        <div class="relative aspect-[4/3] bg-elevated">
          <img
            :src="galleryRendition(image.assetId, 'grid-sm')"
            :alt="image.altText"
            class="size-full object-cover"
          /><UCheckbox
            class="absolute left-3 top-3 rounded-md bg-default/90 p-1.5"
            :model-value="selectedIDs.has(image.id)"
            :aria-label="`选择 ${image.title}`"
            @update:model-value="toggle(image.id)"
          /><UBadge
            class="absolute right-3 top-3"
            :color="lifecycleColor(image.publicationState)"
            variant="solid"
            :label="lifecycleLabel[image.publicationState]"
          />
        </div>
        <div class="p-3">
          <div class="flex items-start justify-between gap-2">
            <div class="min-w-0">
              <h2 class="truncate font-medium text-highlighted">
                {{ image.title }}
              </h2>
              <p class="mt-1 text-xs text-muted">
                {{ safetyLabel[image.safetyState] }} ·
                {{ processingLabel[image.processingState] }}
              </p>
            </div>
            <UButton
              color="neutral"
              variant="ghost"
              size="xs"
              icon="i-tabler-edit"
              aria-label="快速编辑"
              @click="openEdit(image)"
            />
          </div>
        </div>
      </article>
    </div>
    <ManageEmpty
      v-else
      icon="i-tabler-photo-off"
      title="没有符合条件的图片"
      description="调整生命周期或质量筛选，或者清除搜索后重试。"
    />

    <nav
      v-if="data.totalPages > 1"
      class="mt-6 flex items-center justify-center gap-3"
      aria-label="管理图片分页"
    >
      <UButton
        color="neutral"
        variant="outline"
        label="上一页"
        :disabled="page <= 1"
        @click="setQuery({ page: page - 1 })"
      /><span class="text-sm tabular-nums text-muted"
        >{{ page }} / {{ data.totalPages }}</span
      ><UButton
        color="neutral"
        variant="outline"
        label="下一页"
        :disabled="page >= data.totalPages"
        @click="setQuery({ page: page + 1 })"
      />
    </nav>

    <UModal
      :open="Boolean(editing)"
      title="快速编辑图片"
      description="使用更新时间做并发保护；如果其他运营者先保存，当前提交会被拒绝。"
      @update:open="
        (value) => {
          if (!value) editing = undefined;
        }
      "
    >
      <template #body
        ><form class="space-y-4" @submit.prevent="saveEdit">
          <UFormField label="标题" required
            ><UInput v-model="editForm.title" maxlength="160" /></UFormField
          ><UFormField label="替代文本" required
            ><UInput v-model="editForm.altText" /></UFormField
          ><UFormField label="说明"
            ><UTextarea v-model="editForm.description" :rows="4" /></UFormField
          ><UFormField label="来源地址"
            ><UInput
              v-model="editForm.sourceUrl"
              type="url"
              placeholder="https://"
          /></UFormField>
          <div class="flex justify-end gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              label="取消"
              @click="editing = undefined"
            /><UButton type="submit" label="保存" :loading="editPending" />
          </div></form
      ></template>
    </UModal>

    <UModal
      v-model:open="bulkOpen"
      title="批量下架图片"
      :description="`将尝试下架 ${selectedIDs.size} 张图片；不处于公开状态的项目会保留为失败结果。`"
    >
      <template #body
        ><form class="space-y-4" @submit.prevent="bulkHide">
          <UFormField label="下架原因" required
            ><UTextarea
              v-model="bulkReason"
              :rows="4"
              placeholder="说明政策、版权或内容原因"
          /></UFormField>
          <div class="flex justify-end gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              label="取消"
              @click="closeBulkHide"
            /><UButton
              type="submit"
              color="error"
              label="确认批量下架"
              :loading="bulkPending"
              :disabled="!bulkReason.trim()"
            />
          </div></form
      ></template>
    </UModal>
  </div>
</template>
