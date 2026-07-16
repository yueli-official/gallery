<script setup lang="ts">
import { ManageHeader, SkeletonList } from "@platform/manage/components";
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
const hydrated = useClientHydrated();
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
        Category 是主要浏览树，Facet
        是结构化筛选轴；所有变更先预览影响，再按同一 revision 原子执行。
      </template>
      <template #actions>
        <UBadge
          color="neutral"
          variant="soft"
          :label="`revision ${catalog.revision}`"
        />
      </template>
    </ManageHeader>

    <UAlert
      class="mb-5"
      color="info"
      variant="subtle"
      icon="i-tabler-shield-check"
      title="不会静默级联"
      description="停用保留历史关系；合并会迁移归属并建立直接 replacement；删除有关联时必须再次确认 delete-all-related。"
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

    <div v-else class="space-y-8">
      <ClassificationCategoryPanel
        :items="catalog.categories"
        @action="handleIdentityAction"
      />
      <ClassificationFacetPanel
        :facets="catalog.facets"
        @action="handleIdentityAction"
      />
      <ClassificationProposalPanel
        :proposals="proposalData.proposals"
        :total="proposalData.total"
        :tags="tagData.page.items"
        :pending="proposalsPending"
        :hydrated="hydrated"
        :reviewing-id="reviewingProposal"
        @review="reviewTagProposal"
      />
      <ClassificationTagPanel
        :tags="tagData.page.items"
        :pending="tagsPending"
        :hydrated="hydrated"
        :next-cursor="tagData.page.nextCursor"
        :loading-more="loadingMoreTags"
        @action="handleIdentityAction"
        @load-more="loadMoreTags"
      />
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
