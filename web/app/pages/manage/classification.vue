<script setup lang="ts">
import {
  ManageCollectionToolbar,
  ManageHeader,
  SkeletonList,
} from "@platform/manage/components";
import type {
  GalleryClassificationCatalog,
  GalleryClassificationCatalogFacet,
  GalleryClassificationCatalogNode,
  GalleryClassificationGovernanceCommand,
  GalleryClassificationGovernancePreview,
  GalleryClassificationTag,
  GalleryClassificationTagPage,
  GalleryClassificationTagProposal,
} from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "分类与维度 · 图库管理" });

type IdentityKind = GalleryClassificationGovernanceCommand["kind"];
type ManagedIdentity =
  | GalleryClassificationCatalogNode
  | GalleryClassificationCatalogFacet
  | GalleryClassificationTag;
type IdentityOperation = "status" | "reparent" | "merge" | "delete";

const { call } = useApi();
const route = useRoute();
const router = useRouter();
const hydrated = useClientHydrated();
const sectionKeys = ["categories", "facets", "tags", "proposals"] as const;
const section = computed({
  get: () => {
    const value = String(route.query.section || "categories");
    return sectionKeys.includes(value as (typeof sectionKeys)[number])
      ? value
      : "categories";
  },
  set: (value: string) => {
    void router.replace({
      query: {
        ...route.query,
        section: value === "categories" ? undefined : value,
        q: undefined,
      },
    });
  },
});
const searchInput = ref(String(route.query.q || ""));
let searchTimer: ReturnType<typeof setTimeout> | undefined;
watch(
  () => route.query.q,
  (value) => {
    searchInput.value = String(value || "");
  },
);
watch(searchInput, (value) => {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(() => {
    void router.replace({
      query: { ...route.query, q: value.trim() || undefined },
    });
  }, 250);
});
const { data, pending, error, refresh } = await useAsyncData(
  "gallery-manage-classification",
  () =>
    call<{ catalog: GalleryClassificationCatalog }>(
      "/api/v1/gallery/admin/classification",
    ),
  {
    server: false,
    default: () => ({ catalog: { revision: 0, categories: [], facets: [] } }),
  },
);
const {
  data: tagData,
  pending: tagsPending,
  refresh: refreshTags,
} = await useAsyncData(
  "gallery-manage-classification-tags",
  () =>
    call<{ page: GalleryClassificationTagPage }>(
      "/api/v1/gallery/admin/classification/tags?size=40",
    ),
  { server: false, default: () => ({ page: { items: [], nextCursor: "" } }) },
);
const {
  data: proposalData,
  pending: proposalsPending,
  refresh: refreshProposals,
} = await useAsyncData(
  "gallery-manage-classification-tag-proposals",
  () =>
    call<{ proposals: GalleryClassificationTagProposal[]; total: number }>(
      "/api/v1/gallery/admin/classification/tag-proposals?status=pending&page=1&size=30",
    ),
  { server: false, default: () => ({ proposals: [], total: 0 }) },
);

const previewOpen = ref(false);
const previewing = ref(false);
const executing = ref(false);
const actionError = ref("");
const actionSuccess = ref("");
const preview = ref<GalleryClassificationGovernancePreview>();
const pendingCommand = ref<GalleryClassificationGovernanceCommand>();
const editorOpen = ref(false);
const editorMode = ref<"reparent" | "merge">("reparent");
const editorKind = ref<IdentityKind>("category");
const editorIdentity = ref<ManagedIdentity>();
const loadingMoreTags = ref(false);
const reviewingProposal = ref("");

const catalog = computed(() => data.value.catalog);
const query = computed(() =>
  String(route.query.q || "")
    .trim()
    .toLowerCase(),
);
const matches = (name: string, slug: string) =>
  !query.value || `${name} ${slug}`.toLowerCase().includes(query.value);
const visibleCategories = computed(() =>
  catalog.value.categories.filter((item) => matches(item.name, item.slug)),
);
const visibleFacets = computed(() =>
  catalog.value.facets
    .filter(
      (facet) =>
        matches(facet.name, facet.slug) ||
        facet.values.some((value) => matches(value.name, value.slug)),
    )
    .map((facet) => ({
      ...facet,
      values: query.value
        ? facet.values.filter((value) => matches(value.name, value.slug))
        : facet.values,
    })),
);
const visibleTags = computed(() =>
  tagData.value.page.items.filter((item) => matches(item.name, item.slug)),
);
const visibleProposals = computed(() =>
  proposalData.value.proposals.filter((item) =>
    matches(item.inputValue, item.lookupKey),
  ),
);
const tabs = computed(() => [
  {
    key: "categories",
    label: "分类树",
    description: "维护公开浏览的主路径",
    icon: "i-tabler-sitemap",
    count: catalog.value.categories.length,
  },
  {
    key: "facets",
    label: "筛选维度",
    description: "维护结构化筛选轴和值",
    icon: "i-tabler-adjustments-horizontal",
    count: catalog.value.facets.length,
  },
  {
    key: "tags",
    label: "规范标签",
    description: "治理长尾词与同义关系",
    icon: "i-tabler-tags",
    count: tagData.value.page.items.length,
  },
  {
    key: "proposals",
    label: "待审词",
    description: "决定新词创建或归并",
    icon: "i-tabler-tag-starred",
    count: proposalData.value.total,
  },
]);
const searchPlaceholder = computed(
  () =>
    ({
      categories: "搜索分类名称或标识…",
      facets: "搜索维度或维度值…",
      tags: "搜索标签…",
      proposals: "搜索待审词…",
    })[section.value],
);
const destructivePreview = computed(
  () =>
    pendingCommand.value?.operation === "delete" ||
    pendingCommand.value?.operation === "merge",
);

function childrenOf(
  kind: IdentityKind,
  id: string,
): GalleryClassificationCatalogNode[] {
  if (kind === "category")
    return catalog.value.categories.filter((item) => item.parentId === id);
  if (kind === "facet_value")
    return catalog.value.facets
      .flatMap((facet) => facet.values)
      .filter((item) => item.parentId === id);
  return [];
}

function owningFacet(id: string) {
  return catalog.value.facets.find((facet) =>
    facet.values.some((value) => value.id === id),
  );
}

const editorTargets = computed<ManagedIdentity[]>(() => {
  const source = editorIdentity.value;
  if (!source) return [];
  if (editorKind.value === "category")
    return catalog.value.categories.filter(
      (item) => item.id !== source.id && item.status === "active",
    );
  if (editorKind.value === "facet_value")
    return (owningFacet(source.id)?.values || []).filter(
      (item) => item.id !== source.id && item.status === "active",
    );
  if (editorKind.value === "tag")
    return tagData.value.page.items.filter(
      (item) => item.id !== source.id && item.status === "active",
    );
  return [];
});
const editorChildCount = computed(() =>
  editorIdentity.value
    ? childrenOf(editorKind.value, editorIdentity.value.id).length
    : 0,
);

async function requestPreview(command: GalleryClassificationGovernanceCommand) {
  previewing.value = true;
  actionError.value = "";
  actionSuccess.value = "";
  pendingCommand.value = command;
  previewOpen.value = true;
  try {
    const response = await call<{
      preview: GalleryClassificationGovernancePreview;
    }>("/api/v1/gallery/admin/classification/governance/preview", {
      method: "POST",
      body: { command },
    });
    preview.value = response.preview;
  } catch (cause) {
    preview.value = undefined;
    actionError.value = cause instanceof Error ? cause.message : "治理预览失败";
  } finally {
    previewing.value = false;
  }
}

function setStatus(kind: IdentityKind, item: ManagedIdentity) {
  return requestPreview({
    operation: "set_status",
    kind,
    id: item.id,
    status: item.status === "active" ? "inactive" : "active",
    childPlan: [],
    deleteAllRelated: false,
  });
}

function deleteIdentity(kind: IdentityKind, item: ManagedIdentity) {
  return requestPreview({
    operation: "delete",
    kind,
    id: item.id,
    childPlan: [],
    deleteAllRelated: false,
  });
}

function openEditor(
  mode: "reparent" | "merge",
  kind: IdentityKind,
  item: ManagedIdentity,
) {
  editorMode.value = mode;
  editorKind.value = kind;
  editorIdentity.value = item;
  editorOpen.value = true;
}

function handleIdentityAction(
  operation: IdentityOperation,
  kind: IdentityKind,
  item: ManagedIdentity,
) {
  if (operation === "status") return setStatus(kind, item);
  if (operation === "delete") return deleteIdentity(kind, item);
  return openEditor(operation, kind, item);
}

function previewEditor(targetID: string) {
  const source = editorIdentity.value;
  if (!source) return;
  const command: GalleryClassificationGovernanceCommand = {
    operation: editorMode.value,
    kind: editorKind.value,
    id: source.id,
    childPlan: [],
    deleteAllRelated: false,
  };
  if (editorMode.value === "reparent")
    command.parentId = targetID === "root" ? "" : targetID;
  else {
    command.targetId = targetID;
    command.childPlan = childrenOf(editorKind.value, source.id).map(
      (child) => ({
        childId: child.id,
        parentId: targetID,
      }),
    );
  }
  editorOpen.value = false;
  return requestPreview(command);
}

function confirmDeleteAllRelated() {
  if (!pendingCommand.value) return;
  return requestPreview({ ...pendingCommand.value, deleteAllRelated: true });
}

async function executePreview() {
  if (!pendingCommand.value || preview.value?.outcome !== "planned") return;
  executing.value = true;
  actionError.value = "";
  actionSuccess.value = "";
  const operation = pendingCommand.value.operation;
  try {
    await call("/api/v1/gallery/admin/classification/governance/execute", {
      method: "POST",
      body: {
        command: pendingCommand.value,
        expectedCatalogRevision: preview.value.plan.expectedCatalogRevision,
        expectedRequestToken: preview.value.plan.expectedRequestToken,
        expectedImpactToken: preview.value.plan.expectedImpactToken,
      },
    });
    previewOpen.value = false;
    preview.value = undefined;
    pendingCommand.value = undefined;
    await Promise.all([refresh(), refreshTags(), refreshProposals()]);
    actionSuccess.value =
      operation === "delete"
        ? "分类治理已执行；相关标识和已确认依赖已删除。"
        : "分类治理已执行；目录 revision 已更新。";
  } catch (cause) {
    actionError.value =
      cause instanceof Error ? cause.message : "治理执行失败，请重新预览";
  } finally {
    executing.value = false;
  }
}

async function loadMoreTags() {
  const cursor = tagData.value.page.nextCursor;
  if (!cursor || loadingMoreTags.value) return;
  loadingMoreTags.value = true;
  try {
    const response = await call<{ page: GalleryClassificationTagPage }>(
      `/api/v1/gallery/admin/classification/tags?size=40&cursor=${encodeURIComponent(cursor)}`,
    );
    tagData.value.page = {
      items: [...tagData.value.page.items, ...response.page.items],
      nextCursor: response.page.nextCursor,
    };
  } finally {
    loadingMoreTags.value = false;
  }
}

async function reviewTagProposal(
  item: GalleryClassificationTagProposal,
  decision: "approve" | "reject",
  targetTagId: string,
) {
  reviewingProposal.value = item.id;
  actionError.value = "";
  actionSuccess.value = "";
  try {
    await call(
      `/api/v1/gallery/admin/classification/tag-proposals/${encodeURIComponent(item.id)}/review`,
      {
        method: "POST",
        body: {
          decision,
          targetTagId: decision === "approve" ? targetTagId : "",
        },
      },
    );
    await Promise.all([refresh(), refreshTags(), refreshProposals()]);
    actionSuccess.value =
      decision === "approve" ? "Tag 提案已批准" : "Tag 提案已拒绝";
  } catch (cause) {
    actionError.value =
      cause instanceof Error ? cause.message : "Tag 提案处理失败";
  } finally {
    reviewingProposal.value = "";
  }
}
</script>

<template>
  <div>
    <ManageHeader title="分类与维度">
      <template #subtitle>
        维护公开目录的路径、筛选轴和检索词；高风险变更先预演影响。
      </template>
    </ManageHeader>

    <USelect
      v-model="section"
      :items="
        tabs.map((item) => ({
          label: `${item.label} · ${item.count}`,
          value: item.key,
        }))
      "
      value-key="value"
      icon="i-tabler-category"
      class="mb-4 w-full lg:hidden"
      aria-label="选择目录治理任务"
    />
    <div class="grid gap-6 lg:grid-cols-[13.5rem_minmax(0,1fr)] lg:items-start">
      <aside class="hidden lg:block lg:sticky lg:top-20">
        <nav class="space-y-1" aria-label="目录治理任务">
          <button
            v-for="item in tabs"
            :key="item.key"
            type="button"
            class="group grid w-full grid-cols-[2rem_minmax(0,1fr)_auto] gap-2 rounded-lg px-2.5 py-3 text-left transition"
            :class="
              section === item.key
                ? 'bg-elevated text-highlighted ring-1 ring-default'
                : 'text-muted hover:bg-elevated/55 hover:text-default'
            "
            @click="section = item.key"
          >
            <UIcon :name="item.icon" class="mt-0.5 size-4.5 text-primary" />
            <span class="min-w-0">
              <span class="block text-sm font-semibold">{{ item.label }}</span>
              <span class="mt-0.5 block text-xs leading-4 text-muted">{{
                item.description
              }}</span>
            </span>
            <span class="text-xs tabular-nums text-dimmed">{{
              item.count
            }}</span>
          </button>
        </nav>
        <details
          class="group mt-5 border-t border-default pt-3 text-xs text-muted"
        >
          <summary
            class="flex cursor-pointer list-none items-center justify-between py-2"
          >
            目录协议
            <UIcon
              name="i-tabler-chevron-down"
              class="size-4 transition group-open:rotate-180"
            />
          </summary>
          <p class="mt-2 leading-5">
            revision
            {{
              catalog.revision
            }}。停用保留历史关系，合并迁移归属，关联删除必须二次确认。
          </p>
        </details>
      </aside>

      <div class="min-w-0">
        <ManageCollectionToolbar
          v-model:search="searchInput"
          :search-placeholder="searchPlaceholder"
          compact-filters
          class="mb-4"
        />
        <UAlert
          v-if="actionError && !previewOpen"
          class="mb-5"
          color="error"
          variant="subtle"
          title="操作失败"
          :description="actionError"
        />
        <UAlert
          v-else-if="actionSuccess"
          class="mb-5"
          color="success"
          variant="subtle"
          title="操作完成"
          :description="actionSuccess"
        />
        <SkeletonList v-if="!hydrated || pending" :rows="6" />
        <UAlert
          v-else-if="error"
          color="error"
          variant="subtle"
          title="分类目录加载失败"
        >
          <template #actions>
            <UButton class="min-h-11" label="重试" @click="refresh()" />
          </template>
        </UAlert>

        <div v-else>
          <ClassificationCategoryPanel
            v-if="section === 'categories'"
            :items="visibleCategories"
            @action="handleIdentityAction"
          />
          <ClassificationFacetPanel
            v-else-if="section === 'facets'"
            :facets="visibleFacets"
            @action="handleIdentityAction"
          />
          <ClassificationProposalPanel
            v-else-if="section === 'proposals'"
            :proposals="visibleProposals"
            :total="proposalData.total"
            :tags="tagData.page.items"
            :pending="proposalsPending"
            :hydrated="hydrated"
            :reviewing-id="reviewingProposal"
            @review="reviewTagProposal"
          />
          <ClassificationTagPanel
            v-else
            :tags="visibleTags"
            :pending="tagsPending"
            :hydrated="hydrated"
            :next-cursor="tagData.page.nextCursor"
            :loading-more="loadingMoreTags"
            @action="handleIdentityAction"
            @load-more="loadMoreTags"
          />
        </div>
      </div>
    </div>

    <ClassificationEditorModal
      v-model:open="editorOpen"
      :mode="editorMode"
      :identity="editorIdentity"
      :targets="editorTargets"
      :child-count="editorChildCount"
      @preview="previewEditor"
    />
    <ClassificationPreviewModal
      v-model:open="previewOpen"
      :previewing="previewing"
      :executing="executing"
      :error="actionError"
      :preview="preview"
      :destructive="destructivePreview"
      @delete-all="confirmDeleteAllRelated"
      @execute="executePreview"
    />
  </div>
</template>
