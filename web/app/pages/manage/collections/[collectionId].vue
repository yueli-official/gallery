<script setup lang="ts">
import { AdminRowActions, EditorInspector, PageHeader } from "@yueli/ui/admin";
import {
  CollectionHeaderTools,
  CollectionPanel,
  CollectionViewToggle,
} from "@yueli/ui/collection/pattern";
import type { CollectionPanelMessages } from "@yueli/ui/collection/pattern";
import { useActionFeedback } from "@yueli/ui/feedback";
import { ActionFeedbackButton } from "@yueli/ui/feedback/pattern";
import { SkeletonList } from "~/utils/manageComponents";
import { createGalleryNotifier } from "~/utils/feedback";
import type {
  GalleryCollection,
  GalleryCollectionDetail,
  GalleryImageCard,
} from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
const route = useRoute("/manage/collections/[collectionId]");
const router = useRouter();
const { call } = useGalleryApi();
const { can } = useGalleryMe();
const toast = createGalleryNotifier(useToast());
const hydrated = useClientHydrated();
const canManageCollections = computed(() => can("gallery.collection.manage"));
const collectionId = computed(() => String(route.params.collectionId));
const settingsOpen = ref(false);
const settingsSection = ref<"content" | "presentation" | "search">("content");
const memberSearch = ref("");
const orderedIds = ref<string[]>([]);
const selectedMemberIds = ref<Set<string>>(new Set());
const bulkRemoveOpen = ref(false);
const bulkRemovePending = ref(false);
const draggingMemberId = ref("");
const dragTargetMemberId = ref("");
type MemberSortPreset =
  | "manual"
  | "published_desc"
  | "published_asc"
  | "updated_desc"
  | "updated_asc";
const memberSortPreset = ref<MemberSortPreset>("manual");
const metadataTouched = ref(false);
const orderTouched = ref(false);
let syncingForm = false;

const form = reactive({
  name: "",
  slug: "",
  description: "",
  visibility: "private" as "private" | "public",
  coverImageId: "",
  seoTitle: "",
  seoDescription: "",
});

const {
  status: metadataSaveStatus,
  pending: markMetadataSaving,
  success: markMetadataSaved,
  reset: resetMetadataSave,
} = useActionFeedback();
const {
  status: orderSaveStatus,
  pending: markOrderSaving,
  success: markOrderSaved,
  reset: resetOrderSave,
} = useActionFeedback();
const saving = computed(() => metadataSaveStatus.value === "pending");
const ordering = computed(() => orderSaveStatus.value === "pending");

const { data, pending, error, refresh } = await useAsyncData(
  "gallery-manage-collection-detail",
  async () => {
    const result = await call<{ collection: GalleryCollectionDetail }>(`/admin/collections/${encodeURIComponent(collectionId.value)}?page=1&size=60`);
    const items = [...result.collection.items];
    const pages = galleryPageCount(result.collection);
    for (let nextPage = 2; nextPage <= pages; nextPage++) {
      const next = await call<{ collection: GalleryCollectionDetail }>(`/admin/collections/${encodeURIComponent(collectionId.value)}?page=${nextPage}&size=60`);
      items.push(...next.collection.items);
    }
    return { collection: { ...result.collection, items } };
  },
  { server: false },
);
const collection = computed(() => data.value?.collection);
const members = computed(() => {
  const byID = new Map(
    (collection.value?.items || []).map((item) => [item.id, item]),
  );
  return orderedIds.value.flatMap((id) => {
    const item = byID.get(id);
    return item ? [item] : [];
  });
});
const visibleMembers = computed(() => {
  const query = memberSearch.value.trim().toLocaleLowerCase("zh-CN");
  if (!query) return members.value;
  return members.value.filter((item) =>
    [item.title, item.primaryCategory, item.primaryCategorySlug]
      .join(" ")
      .toLocaleLowerCase("zh-CN")
      .includes(query),
  );
});
const selectedCover = computed(() =>
  members.value.find((item) => item.id === form.coverImageId),
);
const memberPage = ref(1);
const memberPageSize = ref(20);
const pagedMembers = computed(() => visibleMembers.value.slice((memberPage.value - 1) * memberPageSize.value, memberPage.value * memberPageSize.value));
watch([memberSearch, memberPageSize], () => { memberPage.value = 1; });
watch(() => visibleMembers.value.length, total => { memberPage.value = Math.min(memberPage.value, Math.max(1, Math.ceil(total / memberPageSize.value))); });
const canReorder = computed(() => (galleryPageCount(collection.value) || 0) <= 1);
const canArrangeMembers = computed(
  () => canReorder.value && !memberSearch.value.trim(),
);
const selectedMemberCount = computed(() => selectedMemberIds.value.size);
const isMemberPageSelected = computed(
  () =>
    pagedMembers.value.length > 0 &&
    pagedMembers.value.every((item) => selectedMemberIds.value.has(item.id)),
);
const isMemberPageIndeterminate = computed(() => {
  const count = pagedMembers.value.filter((item) =>
    selectedMemberIds.value.has(item.id),
  ).length;
  return count > 0 && count < pagedMembers.value.length;
});
const memberSortOptions = [
  { label: "手动排序", value: "manual" },
  { label: "发布时间 · 新到旧", value: "published_desc" },
  { label: "发布时间 · 旧到新", value: "published_asc" },
  { label: "修改时间 · 新到旧", value: "updated_desc" },
  { label: "修改时间 · 旧到新", value: "updated_asc" },
];

function sameOrder(left: readonly string[], right: readonly string[]) {
  return (
    left.length === right.length &&
    left.every((value, index) => value === right[index])
  );
}

const metadataDirty = computed(() => {
  const value = collection.value;
  if (!value || !metadataTouched.value) return false;
  return (
    form.name !== value.name ||
    form.slug !== (value.slug || "") ||
    form.description !== value.description ||
    form.visibility !== value.visibility ||
    form.coverImageId !== (value.coverImageId || "") ||
    form.seoTitle !== (value.seoTitle || "") ||
    form.seoDescription !== (value.seoDescription || "")
  );
});
const orderDirty = computed(() => {
  const serverOrder = collection.value?.items.map((item) => item.id) || [];
  return orderTouched.value && !sameOrder(orderedIds.value, serverOrder);
});
const hasUnsavedChanges = computed(
  () => metadataDirty.value || orderDirty.value,
);

const viewMode = computed<"list" | "grid">({
  get: () => (route.query.view === "list" ? "list" : "grid"),
  set: (value) => {
    void router.replace({
      query: {
        ...route.query,
        view: value === "grid" ? undefined : value,
      },
    });
  },
});

const settingsTabs = [
  { label: "内容", value: "content", icon: "i-tabler-file-text" },
  { label: "展示", value: "presentation", icon: "i-tabler-photo-cog" },
  { label: "SEO", value: "search", icon: "i-tabler-search" },
];
const collectionMessages: CollectionPanelMessages = {
  searchPlaceholder: "搜索专题内图片…",
  searchAction: "搜索",
  filtersAction: "筛选",
  activeFilters: (count) => `筛选（${count}）`,
  clearFilters: "清除筛选",
  selectPage: "选择当前页图片",
  selectItem: (label) => `选择图片：${label}`,
  bulkRegion: "专题图片批量操作",
  selected: (count) => `已选择 ${count} 张图片`,
  selectAllResults: "选择全部结果",
  clearSelection: "取消选择",
  emptyTitle: "没有匹配的图片",
  emptyDescription: "调整搜索内容，或向专题添加图片。",
  errorTitle: "专题图片加载失败",
  retry: "重新加载",
  showing: (first, last, total) => `显示 ${first}–${last}，共 ${total} 张`,
  pageSize: "每页",
  pageSizeControl: "每页图片数量",
  pageSizeOption: (value) => `${value} 张`,
};

watch(
  collection,
  (value) => {
    if (!value) return;
    if (!metadataTouched.value) {
      syncingForm = true;
      Object.assign(form, {
        name: value.name,
        slug: value.slug || "",
        description: value.description,
        visibility: value.visibility,
        coverImageId: value.coverImageId || "",
        seoTitle: value.seoTitle || "",
        seoDescription: value.seoDescription || "",
      });
      syncingForm = false;
    }
    if (!orderTouched.value)
      orderedIds.value = value.items.map((item) => item.id);
  },
  { immediate: true },
);

watch(
  form,
  () => {
    if (syncingForm) return;
    metadataTouched.value = true;
    if (!saving.value) resetMetadataSave();
  },
  { deep: true, flush: "sync" },
);

useSeoMeta({ title: () => `${collection.value?.name || "专题"} · 图库管理` });

function message(reason: any): string {
  return galleryFailureMessage(reason, "请稍后重试");
}

function submitMemberSearch(value: string) {
  memberSearch.value = value;
}

function memberKey(item: GalleryImageCard) {
  return item.id;
}

function memberLabel(item: GalleryImageCard) {
  return item.title || "未命名图片";
}

function memberIndex(imageId: string) {
  return orderedIds.value.indexOf(imageId);
}

function isMemberSelected(imageId: string) {
  return selectedMemberIds.value.has(imageId);
}

function toggleMember(imageId: string) {
  const next = new Set(selectedMemberIds.value);
  if (next.has(imageId)) next.delete(imageId);
  else next.add(imageId);
  selectedMemberIds.value = next;
}

function toggleMemberPage(selected: boolean) {
  const next = new Set(selectedMemberIds.value);
  for (const item of pagedMembers.value) {
    if (selected) next.add(item.id);
    else next.delete(item.id);
  }
  selectedMemberIds.value = next;
}

function clearMemberSelection() {
  selectedMemberIds.value = new Set();
}

function openBulkRemove(): void {
  bulkRemoveOpen.value = true;
}

function closeBulkRemove(): void {
  bulkRemoveOpen.value = false;
}

function toggleSettings() {
  settingsOpen.value = !settingsOpen.value;
}

function handleShortcut(event: KeyboardEvent) {
  if ((event.metaKey || event.ctrlKey) && event.key === ",") {
    event.preventDefault();
    toggleSettings();
  }
}

onMounted(() => window.addEventListener("keydown", handleShortcut));
onBeforeUnmount(() => window.removeEventListener("keydown", handleShortcut));

async function saveMetadata(): Promise<void> {
  if (
    !canManageCollections.value ||
    !collection.value ||
    saving.value ||
    !metadataDirty.value
  )
    return;
  markMetadataSaving();
  try {
    await call<{ collection: GalleryCollection }>(
      `/admin/collections/${encodeURIComponent(collection.value.id)}`,
      {
        method: "PATCH",
        body: { ...form, version: collection.value.version },
      },
    );
    metadataTouched.value = false;
    await refresh();
    markMetadataSaved();
  } catch (reason) {
    resetMetadataSave();
    toast.add({
      title: "专题设置没有保存",
      description: message(reason),
      color: "error",
    });
  }
}

async function mutateMembers(
  add: string[] = [],
  remove: string[] = [],
): Promise<boolean> {
  if (!canManageCollections.value || !collection.value) return false;
  const previousOrder = [...orderedIds.value];
  try {
    await call(
      `/admin/collections/${encodeURIComponent(collection.value.id)}/members`,
      {
        method: "POST",
        body: { version: collection.value.version, add, remove },
      },
    );
    await refresh();
    const serverOrder = collection.value?.items.map((item) => item.id) || [];
    const serverIDs = new Set(serverOrder);
    const preserved = previousOrder.filter((id) => serverIDs.has(id));
    orderedIds.value = [
      ...preserved,
      ...serverOrder.filter((id) => !preserved.includes(id)),
    ];
    orderTouched.value = !sameOrder(orderedIds.value, serverOrder);
    selectedMemberIds.value = new Set(
      [...selectedMemberIds.value].filter((id) => serverIDs.has(id)),
    );
    resetOrderSave();
    return true;
  } catch (reason) {
    toast.add({
      title: "专题成员没有更新",
      description: message(reason),
      color: "error",
    });
    return false;
  }
}

function moveMember(imageId: string, direction: -1 | 1): void {
  if (!canManageCollections.value || !canArrangeMembers.value) return;
  const index = memberIndex(imageId);
  const target = index + direction;
  if (index < 0 || target < 0 || target >= orderedIds.value.length) return;
  const next = [...orderedIds.value];
  [next[index], next[target]] = [next[target]!, next[index]!];
  orderedIds.value = next;
  orderTouched.value = true;
  memberSortPreset.value = "manual";
  resetOrderSave();
}

function moveMemberTo(imageId: string, target: "top" | "bottom"): void {
  if (!canManageCollections.value || !canArrangeMembers.value) return;
  const index = memberIndex(imageId);
  if (index < 0) return;
  const next = [...orderedIds.value];
  const [moved] = next.splice(index, 1);
  if (!moved) return;
  if (target === "top") next.unshift(moved);
  else next.push(moved);
  orderedIds.value = next;
  orderTouched.value = true;
  memberSortPreset.value = "manual";
  resetOrderSave();
}

function startMemberDrag(event: DragEvent, imageId: string): void {
  if (!canManageCollections.value || !canArrangeMembers.value) {
    event.preventDefault();
    return;
  }
  draggingMemberId.value = imageId;
  dragTargetMemberId.value = "";
  event.dataTransfer?.setData("text/plain", imageId);
  if (event.dataTransfer) event.dataTransfer.effectAllowed = "move";
}

function markMemberDragTarget(imageId: string): void {
  if (draggingMemberId.value && draggingMemberId.value !== imageId)
    dragTargetMemberId.value = imageId;
}

function dropMember(imageId: string): void {
  const sourceID = draggingMemberId.value;
  if (!sourceID || sourceID === imageId || !canArrangeMembers.value) {
    endMemberDrag();
    return;
  }
  const next = [...orderedIds.value];
  const sourceIndex = next.indexOf(sourceID);
  if (sourceIndex < 0) {
    endMemberDrag();
    return;
  }
  next.splice(sourceIndex, 1);
  const targetIndex = next.indexOf(imageId);
  if (targetIndex < 0) {
    endMemberDrag();
    return;
  }
  next.splice(targetIndex, 0, sourceID);
  orderedIds.value = next;
  orderTouched.value = true;
  memberSortPreset.value = "manual";
  resetOrderSave();
  endMemberDrag();
}

function endMemberDrag(): void {
  draggingMemberId.value = "";
  dragTargetMemberId.value = "";
}

function applyMemberSort(): void {
  if (
    !canManageCollections.value ||
    !canArrangeMembers.value ||
    memberSortPreset.value === "manual"
  )
    return;
  const [field, direction] = memberSortPreset.value.split("_") as [
    "published" | "updated",
    "asc" | "desc",
  ];
  const originalIndex = new Map(
    orderedIds.value.map((id, index) => [id, index]),
  );
  const timeOf = (item: GalleryImageCard) => {
    const value = field === "published" ? item.publishedAt : item.updatedAt;
    const time = value ? Date.parse(value) : Number.NaN;
    return Number.isFinite(time) ? time : 0;
  };
  orderedIds.value = [...members.value]
    .sort((left, right) => {
      const delta = timeOf(left) - timeOf(right);
      if (delta) return direction === "asc" ? delta : -delta;
      return (
        (originalIndex.get(left.id) || 0) - (originalIndex.get(right.id) || 0)
      );
    })
    .map((item) => item.id);
  orderTouched.value = true;
  resetOrderSave();
}

function selectCover(imageId: string): void {
  if (!canManageCollections.value) return;
  form.coverImageId = imageId;
}

function clearCover(): void {
  if (!canManageCollections.value) return;
  form.coverImageId = "";
}

function memberMoreItems(item: GalleryImageCard) {
  const index = memberIndex(item.id);
  return [
    [
      {
        id: "move-top",
        label: "移到顶部",
        icon: "i-tabler-arrow-bar-to-up",
        disabled: !canArrangeMembers.value || index === 0,
        onSelect: () => moveMemberTo(item.id, "top"),
      },
      {
        id: "move-up",
        label: "上移",
        icon: "i-tabler-arrow-up",
        disabled: !canArrangeMembers.value || index === 0,
        onSelect: () => moveMember(item.id, -1),
      },
      {
        id: "move-down",
        label: "下移",
        icon: "i-tabler-arrow-down",
        disabled:
          !canArrangeMembers.value || index === orderedIds.value.length - 1,
        onSelect: () => moveMember(item.id, 1),
      },
      {
        id: "move-bottom",
        label: "移到底部",
        icon: "i-tabler-arrow-bar-to-down",
        disabled:
          !canArrangeMembers.value || index === orderedIds.value.length - 1,
        onSelect: () => moveMemberTo(item.id, "bottom"),
      },
    ],
    [
      {
        id: "cover",
        label: form.coverImageId === item.id ? "当前封面" : "设为专题封面",
        icon: "i-tabler-photo-star",
        disabled: form.coverImageId === item.id,
        onSelect: () => selectCover(item.id),
      },
    ],
    [
      {
        id: "remove",
        label: "移出专题",
        icon: "i-tabler-trash",
        tone: "danger" as const,
        onSelect: () => mutateMembers([], [item.id]),
      },
    ],
  ];
}

async function runBulkRemove(): Promise<void> {
  if (!selectedMemberCount.value || bulkRemovePending.value) return;
  bulkRemovePending.value = true;
  const removed = selectedMemberCount.value;
  const success = await mutateMembers([], [...selectedMemberIds.value]);
  if (success) {
    clearMemberSelection();
    bulkRemoveOpen.value = false;
    toast.add({
      title: `已从专题移出 ${removed} 张图片`,
      color: "success",
    });
  }
  bulkRemovePending.value = false;
}

async function saveOrder(): Promise<void> {
  if (
    !canManageCollections.value ||
    !collection.value ||
    !canReorder.value ||
    ordering.value ||
    !orderDirty.value
  )
    return;
  markOrderSaving();
  try {
    await call(
      `/admin/collections/${encodeURIComponent(collection.value.id)}/order`,
      {
        method: "PUT",
        body: { version: collection.value.version, imageIds: orderedIds.value },
      },
    );
    orderTouched.value = false;
    await refresh();
    markOrderSaved();
  } catch (reason) {
    resetOrderSave();
    toast.add({
      title: "图片顺序没有保存",
      description: message(reason),
      color: "error",
    });
  }
}

</script>

<template>
  <div
    class="yueli-admin-canvas min-h-full min-w-0"
    data-gallery-collection-editor
  >
    <div
      class="sticky top-0 z-30 flex min-h-16 items-center justify-between gap-2 border-b border-default bg-default px-3 py-1.5 sm:gap-4 sm:px-4 sm:py-2 lg:px-8"
      data-gallery-collection-commandbar
    >
      <div class="flex min-w-0 items-center gap-2">
        <UDashboardSidebarToggle class="size-11 sm:size-8 lg:hidden" />
        <UTooltip text="返回专题列表">
          <UButton
            to="/manage/collections"
            icon="i-tabler-arrow-left"
            color="neutral"
            variant="ghost"
            square
            class="size-11 sm:size-8"
            aria-label="返回专题列表"
          />
        </UTooltip>
        <span
          class="hidden max-w-[min(32vw,28rem)] truncate text-sm font-semibold text-toned sm:block"
        >
          专题编辑
        </span>
        <template v-if="collection">
          <span class="hidden h-5 w-px bg-accented md:block" />
          <span
            class="hidden max-w-[min(28vw,24rem)] truncate text-xs text-muted md:block"
          >
            {{ collection.name }}
          </span>
          <span
            v-if="hasUnsavedChanges"
            class="hidden text-xs text-warning lg:inline"
          >
            未保存
          </span>
        </template>
      </div>

      <div v-if="collection" class="flex shrink-0 items-center gap-1.5">
        <UTooltip v-if="collection.slug" text="公开预览">
          <UButton
            :to="`/collections/${collection.slug}`"
            target="_blank"
            icon="i-tabler-eye"
            color="neutral"
            variant="ghost"
            square
            class="hidden size-11 sm:inline-flex sm:size-8"
            aria-label="公开预览"
          />
        </UTooltip>
        <UTooltip text="专题设置 (⌘/Ctrl ,)">
          <UButton
            icon="i-tabler-adjustments-horizontal"
            :color="settingsOpen ? 'primary' : 'neutral'"
            :variant="settingsOpen ? 'soft' : 'ghost'"
            square
            class="size-11 sm:size-8"
            aria-label="专题设置"
            :aria-pressed="settingsOpen"
            @click="toggleSettings"
          />
        </UTooltip>
        <ActionFeedbackButton
          v-if="canManageCollections"
          :status="metadataSaveStatus"
          idle-label="保存"
          pending-label="保存中"
          success-label="已保存"
          class="min-h-11 sm:min-h-8"
          :disabled="!metadataDirty"
          @click="saveMetadata"
        />
      </div>
    </div>

    <div
      v-if="!hydrated || (pending && !collection)"
      class="px-4 py-8 sm:px-6 lg:px-8"
    >
      <SkeletonList :rows="7" />
    </div>
    <div v-else-if="error" class="px-4 py-8 sm:px-6 lg:px-8">
      <UAlert
        color="error"
        variant="subtle"
        title="专题加载失败"
        :actions="[{ label: '重试', onClick: () => refresh() }]"
      />
    </div>

    <main
      v-else-if="collection"
      class="min-w-0 px-4 pb-12 pt-6 transition-[padding] duration-200 ease-out sm:px-6 sm:pb-16 sm:pt-8 lg:px-8"
      :class="settingsOpen ? 'xl:pr-[27rem]' : ''"
      data-gallery-collection-workspace
    >
      <div class="min-w-0 space-y-5">
        <PageHeader :title="collection.name" :description="`图片与顺序 · ${collection.itemCount} 张`">
          <template #actions>
          <div v-if="canManageCollections" class="flex flex-wrap gap-2">
            <ActionFeedbackButton
              :status="orderSaveStatus"
              idle-label="保存顺序"
              pending-label="保存中"
              success-label="已保存"
              idle-icon="i-tabler-arrows-sort"
              color="neutral"
              variant="outline"
              :disabled="!orderDirty || !canReorder"
              @click="saveOrder"
            />
            <UButton
              icon="i-tabler-plus"
              label="前往图片列表添加"
              to="/manage/images?status=published"
            />
          </div>

          </template>
        </PageHeader>

        <UAlert
          v-if="!canReorder"
          color="warning"
          variant="subtle"
          title="大型专题暂不支持整组重排"
          description="当前编辑器只载入前 60 张；成员增删仍可用，整组重排将在规模化阶段升级为分段操作。"
        />

        <CollectionPanel

          compact-pagination
          class="gallery-compact-collection"
          v-model:search="memberSearch"
          :items="pagedMembers"
          :item-key="memberKey"
          :item-label="memberLabel"
          :messages="collectionMessages"
          state="ready"
          :total="visibleMembers.length"
          :page="memberPage"
          :page-size="memberPageSize"
          :page-sizes="[20, 40, 60]"
          @page-change="memberPage = $event"
          @page-size-change="memberPageSize = $event"
          :layout="viewMode === 'grid' ? 'grid' : 'rows'"
          :selection-count="selectedMemberCount"
          :page-selected="isMemberPageSelected"
          :page-indeterminate="isMemberPageIndeterminate"
          :is-selected="isMemberSelected"
          :selectable="canManageCollections"
          label="专题图片"
          @search="submitMemberSearch"
          @retry="refresh"
          @toggle-page="toggleMemberPage"
          @toggle-item="toggleMember"
          @clear-selection="clearMemberSelection"
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
        <template #columns>
            <span class="text-xs text-muted">选择本页</span>
          </template>


          <template #bulk-actions>
            <UButton
              v-if="canManageCollections"
              color="error"
              variant="soft"
              size="xs"
              icon="i-tabler-trash"
              label="批量移出专题"
              @click="openBulkRemove"
            />
          </template>

          <template #item="{ item }">
            <div
              v-if="viewMode === 'list'"
              class="grid min-w-0 grid-cols-[3.5rem_3.5rem_minmax(0,1fr)_auto] items-center gap-3 rounded-lg transition-shadow"
              :class="
                dragTargetMemberId === item.id
                  ? 'ring-2 ring-primary ring-offset-2 ring-offset-default'
                  : ''
              "
              data-gallery-collection-list-item
              :data-dragging="draggingMemberId === item.id ? 'true' : undefined"
              @dragover.prevent="markMemberDragTarget(item.id)"
              @drop.prevent="dropMember(item.id)"
            >
              <div class="flex items-center gap-1">
                <span class="w-4 text-center text-xs tabular-nums text-dimmed">
                  {{ memberIndex(item.id) + 1 }}
                </span>
                <UTooltip
                  :text="
                    canArrangeMembers
                      ? '拖动调整顺序'
                      : memberSearch.trim()
                        ? '清除搜索后可拖动排序'
                        : '当前专题暂不支持整组重排'
                  "
                >
                  <UButton
                    icon="i-tabler-grip-vertical"
                    color="neutral"
                    variant="ghost"
                    size="xs"
                    square
                    :draggable="canArrangeMembers"
                    :disabled="!canArrangeMembers"
                    :aria-label="`拖动排序：${item.title}`"
                    @dragstart="startMemberDrag($event, item.id)"
                    @dragend="endMemberDrag"
                  />
                </UTooltip>
              </div>
              <div
                class="size-14 overflow-hidden rounded-lg bg-elevated"
                :style="{ backgroundColor: item.dominantColor || undefined }"
              >
                <img
                  v-bind="galleryImageSources(item.assetId, 'grid', false)"
                  :alt="item.altText || item.title"
                  class="size-full object-cover"
                />
              </div>
              <div class="min-w-0">
                <div class="flex min-w-0 items-center gap-2">
                  <p class="truncate text-sm font-medium text-highlighted">
                    {{ item.title }}
                  </p>
                  <UBadge
                    v-if="form.coverImageId === item.id"
                    label="封面"
                    color="primary"
                    variant="soft"
                    size="xs"
                  />
                </div>
                <p class="mt-0.5 truncate text-xs text-muted">
                  {{ item.primaryCategory || "未分类" }}
                </p>
              </div>
              <div
                v-if="canManageCollections"
                class="flex items-center justify-end"
              >
                <AdminRowActions
                  presentation="overflow"
                  :label="`图片操作：${item.title}`"
                  :items="memberMoreItems(item)"
                />
              </div>
            </div>

            <div
              v-else
              class="group -m-4 overflow-hidden rounded-lg transition-shadow"
              :class="
                dragTargetMemberId === item.id
                  ? 'ring-2 ring-primary ring-offset-2 ring-offset-default'
                  : ''
              "
              data-gallery-collection-grid-item
              :data-dragging="draggingMemberId === item.id ? 'true' : undefined"
              @dragover.prevent="markMemberDragTarget(item.id)"
              @drop.prevent="dropMember(item.id)"
            >
              <div
                class="relative aspect-[4/3] overflow-hidden bg-elevated"
                :style="{ backgroundColor: item.dominantColor || undefined }"
              >
                <img
                  v-bind="galleryImageSources(item.assetId, 'grid', false)"
                  :alt="item.altText || item.title"
                  class="size-full object-cover transition-transform duration-200 group-hover:scale-[1.02]"
                />
                <span
                  class="absolute right-2.5 top-2.5 grid size-5 place-items-center rounded-lg bg-default/90 text-xs font-semibold tabular-nums text-toned shadow-sm backdrop-blur"
                >
                  {{ memberIndex(item.id) + 1 }}
                </span>
                <UBadge
                  v-if="form.coverImageId === item.id"
                  class="absolute left-9 top-2.5"
                  label="封面"
                  color="primary"
                  variant="solid"
                  size="xs"
                />
              </div>
              <div class="p-2.5">
                <div class="flex min-w-0 items-start justify-between gap-1">
                  <div class="min-w-0">
                    <p class="truncate text-sm font-semibold text-highlighted">
                      {{ item.title }}
                    </p>
                    <p class="mt-1 truncate text-xs text-muted">
                      {{ item.primaryCategory || "未分类" }}
                    </p>
                  </div>
                  <div
                    v-if="canManageCollections"
                    class="flex shrink-0 items-center gap-1"
                  >
                    <UTooltip
                      :text="
                        canArrangeMembers
                          ? '拖动调整顺序'
                          : memberSearch.trim()
                            ? '清除搜索后可拖动排序'
                            : '当前专题暂不支持整组重排'
                      "
                    >
                      <UButton
                        icon="i-tabler-grip-vertical"
                        color="neutral"
                        variant="ghost"
                        size="xs"
                        square
                        :draggable="canArrangeMembers"
                        :disabled="!canArrangeMembers"
                        :aria-label="`拖动排序：${item.title}`"
                        @dragstart="startMemberDrag($event, item.id)"
                        @dragend="endMemberDrag"
                      />
                    </UTooltip>
                    <AdminRowActions
                      presentation="overflow"
                      :label="`图片操作：${item.title}`"
                      :items="memberMoreItems(item)"
                    />
                  </div>
                </div>
              </div>
            </div>
          </template>
        </CollectionPanel>

        <p v-if="!canManageCollections" class="text-sm text-muted" role="note">
          当前角色可以查看专题，但不能修改专题。
        </p>
      </div>
    </main>

    <EditorInspector v-model:open="settingsOpen" title="专题设置">
      <template #default="{ docked }">
        <div
          class="min-w-0"
          data-gallery-collection-inspector
          :data-inspector-mode="docked ? 'docked' : 'overlay'"
        >
          <UTabs
            v-model="settingsSection"
            :items="settingsTabs"
            :content="false"
            value-key="value"
            variant="pill"
            color="neutral"
            class="w-full"
            :ui="{
              list: 'w-full rounded-xl bg-elevated/70 p-1',
              indicator: 'rounded-lg bg-default ring-1 ring-default shadow-xs',
              trigger:
                'min-h-9 flex-1 justify-center gap-2 rounded-lg data-[state=active]:text-highlighted',
              leadingIcon: 'size-4.5 shrink-0',
            }"
          />

          <fieldset
            :disabled="!canManageCollections"
            class="mt-5 min-w-0 space-y-5"
          >
            <section v-if="settingsSection === 'content'" class="space-y-5">
              <UFormField label="名称" required>
                <UInput v-model="form.name" class="w-full" />
              </UFormField>
              <UFormField label="Slug" required>
                <UInput v-model="form.slug" class="w-full" />
              </UFormField>
              <UFormField label="说明">
                <UTextarea
                  v-model="form.description"
                  :rows="5"
                  class="w-full"
                />
              </UFormField>
            </section>

            <section
              v-else-if="settingsSection === 'presentation'"
              class="space-y-5"
            >
              <UFormField label="可见性">
                <USelect
                  v-model="form.visibility"
                  :items="[
                    { label: '私有', value: 'private' },
                    { label: '公开', value: 'public' },
                  ]"
                  value-key="value"
                  class="w-full"
                />
              </UFormField>
              <UFormField
                label="图片排序"
                help="按时间生成一次当前顺序；应用后仍需在页面保存顺序。"
              >
                <div class="space-y-2">
                  <USelect
                    v-model="memberSortPreset"
                    :items="memberSortOptions"
                    value-key="value"
                    class="w-full"
                  />
                  <UButton
                    block
                    color="neutral"
                    variant="outline"
                    icon="i-tabler-arrows-sort"
                    label="应用排序"
                    :disabled="
                      memberSortPreset === 'manual' || !canArrangeMembers
                    "
                    @click="applyMemberSort"
                  />
                </div>
              </UFormField>
              <div>
                <div class="mb-3 flex items-center justify-between gap-3">
                  <h2
                    class="flex items-center gap-2 text-sm font-semibold text-highlighted"
                  >
                    <UIcon name="i-tabler-photo" class="size-4 text-muted" />
                    专题封面
                  </h2>
                  <UButton
                    v-if="form.coverImageId"
                    color="neutral"
                    variant="ghost"
                    size="xs"
                    label="清除"
                    @click="clearCover"
                  />
                </div>
                <div
                  class="gallery-manage-cover grid aspect-video max-w-[26rem] place-items-center overflow-hidden rounded-xl border border-default bg-default [&_img]:size-full [&_img]:object-cover"
                  :style="{
                    backgroundColor:
                      selectedCover?.dominantColor ||
                      collection?.coverColor ||
                      undefined,
                  }"
                >
                  <img
                    v-if="selectedCover"
                    v-bind="
                      galleryImageSources(selectedCover.assetId, 'grid', false)
                    "
                    :alt="selectedCover.altText || selectedCover.title"
                  />
                  <img
                    v-else-if="collection?.coverAssetId"
                    v-bind="
                      galleryImageSources(
                        collection?.coverAssetId || '',
                        'grid',
                        false,
                      )
                    "
                    :alt="
                      collection?.coverAltText || collection?.name || '专题封面'
                    "
                  />
                  <UIcon
                    v-else
                    name="i-tabler-photo"
                    class="size-9 text-muted"
                  />
                </div>
                <p class="mt-2 text-xs leading-5 text-muted">
                  从专题图片的更多操作中设为封面。
                </p>
              </div>
            </section>

            <section v-else class="space-y-5">
              <UFormField label="SEO 标题">
                <UInput v-model="form.seoTitle" class="w-full" />
              </UFormField>
              <UFormField label="SEO 描述">
                <UTextarea
                  v-model="form.seoDescription"
                  :rows="4"
                  class="w-full"
                />
              </UFormField>
            </section>

            <p v-if="!canManageCollections" class="text-sm text-muted">
              当前角色可以查看专题设置，但不能修改。
            </p>
          </fieldset>
        </div>
      </template>

      <template #footer>
        <ActionFeedbackButton
          v-if="canManageCollections"
          :status="metadataSaveStatus"
          idle-label="保存"
          pending-label="保存中"
          success-label="已保存"
          block
          :disabled="!metadataDirty"
          @click="saveMetadata"
        />
      </template>
    </EditorInspector>

    <UModal
      v-if="canManageCollections"
      v-model:open="bulkRemoveOpen"
      title="批量移出专题"
      :description="`将从「${collection?.name || '当前专题'}」移出选中的 ${selectedMemberCount} 张图片；图片本身不会被删除。`"
    >
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton
            color="neutral"
            variant="outline"
            label="取消"
            @click="closeBulkRemove"
          />
          <UButton
            color="error"
            icon="i-tabler-trash"
            label="确认移出"
            :loading="bulkRemovePending"
            :disabled="!selectedMemberCount"
            @click="runBulkRemove"
          />
        </div>
      </template>
    </UModal>


  </div>
</template>
