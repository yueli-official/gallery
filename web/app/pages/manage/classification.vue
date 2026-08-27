<script setup lang="ts">
import { ManagePage, TabbedSurface } from "@yueli/ui/admin";
import { SkeletonList } from "~/utils/manageComponents";
import { classificationMutationErrorMessage } from "~/utils/classificationMutation";
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

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "分类与维度 · 图库管理" });

type IdentityKind = GalleryClassificationGovernanceCommand["kind"];
type ManagedIdentity =
  | GalleryClassificationCatalogNode
  | GalleryClassificationCatalogFacet
  | GalleryClassificationTag;
type IdentityOperation = "status" | "reparent" | "merge" | "delete";

const { call } = useGalleryApi();
const { can } = useGalleryMe();
const route = useRoute();
const router = useRouter();
const hydrated = useClientHydrated();
const canGovern = computed(() => can("gallery.classification.govern"));
const canReviewProposals = computed(() =>
  can("gallery.classification.proposal_review"),
);
const sectionKeys = ["categories", "facets", "tags", "proposals"] as const;
const section = computed({
  get: () => {
    const value = String(route.query.section || "categories");
    return sectionKeys.includes(value as (typeof sectionKeys)[number]) &&
      (value !== "proposals" || canReviewProposals.value)
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
      "/admin/classification",
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
      "/admin/classification/tags?size=40",
    ),
  { server: false, default: () => ({ page: { items: [], nextCursor: "" } }) },
);
const {
  data: proposalData,
  pending: proposalsPending,
  refresh: refreshProposals,
} = await useAsyncData(
  "gallery-manage-classification-tag-proposals",
  async () => {
    if (!canReviewProposals.value) return { proposals: [], total: 0 };
    return await call<{
      proposals: GalleryClassificationTagProposal[];
      total: number;
    }>(
      "/admin/classification/tag-proposals?status=pending&page=1&size=30",
    );
  },
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
const createOpen = ref(false);
const creating = ref(false);
const createError = ref("");
const identityFormMode = ref<"create" | "edit">("create");
const editingIdentityID = ref("");
const createKind = ref<"category" | "facet" | "facet_value" | "tag">(
  "category",
);
const createFacetID = ref("");
const createForm = reactive({ name: "", slug: "", parentId: "__root__" });

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
const tabs = computed(() =>
  [
    {
      value: "categories",
      label: "分类",
      icon: "i-tabler-sitemap",
    },
    {
      value: "facets",
      label: "维度",
      icon: "i-tabler-adjustments-horizontal",
    },
    {
      value: "tags",
      label: "标签",
      icon: "i-tabler-tags",
    },
    canReviewProposals.value
      ? {
          value: "proposals",
          label: "标签提案",
          icon: "i-tabler-tag-starred",
        }
      : undefined,
  ].filter((item): item is NonNullable<typeof item> => Boolean(item)),
);

const createTitle = computed(() =>
  `${identityFormMode.value === "create" ? "新增" : "编辑"}${
    createKind.value === "category"
      ? "分类"
      : createKind.value === "facet"
        ? "维度"
        : createKind.value === "facet_value"
          ? "维度值"
          : "标签"
  }`,
);
const identityFormHint = computed(() =>
  identityFormMode.value === "edit"
    ? "名称和标识会同步用于后台检索与公开筛选。"
    : createKind.value === "tag"
      ? "新标签创建后立即可用于图片分类。"
      : "新建内容先保存为草稿，确认后再从列表启用。",
);
const createParentItems = computed(() => {
  const root = [{ label: "顶级", value: "__root__" }];
  if (createKind.value === "category") {
    return [
      ...root,
      ...catalog.value.categories
        .filter((item) => item.status !== "replaced")
        .map((item) => ({ label: item.name, value: item.id })),
    ];
  }
  if (createKind.value === "facet_value") {
    const facet = catalog.value.facets.find(
      (item) => item.id === createFacetID.value,
    );
    return [
      ...root,
      ...(facet?.values || [])
        .filter((item) => item.status !== "replaced")
        .map((item) => ({ label: item.name, value: item.id })),
    ];
  }
  return root;
});

function openCreateIdentity(
  kind: "category" | "facet" | "facet_value" | "tag",
  facetID = "",
) {
  identityFormMode.value = "create";
  editingIdentityID.value = "";
  createKind.value = kind;
  createFacetID.value = facetID;
  createError.value = "";
  Object.assign(createForm, { name: "", slug: "", parentId: "__root__" });
  createOpen.value = true;
}

function openEditIdentity(kind: IdentityKind, item: ManagedIdentity) {
  if (!canGovern.value || !["category", "facet", "facet_value", "tag"].includes(kind))
    return;
  identityFormMode.value = "edit";
  editingIdentityID.value = item.id;
  createKind.value = kind as "category" | "facet" | "facet_value" | "tag";
  createFacetID.value = "";
  createError.value = "";
  Object.assign(createForm, {
    name: item.name,
    slug: item.slug,
    parentId: "__root__",
  });
  createOpen.value = true;
}

function closeCreateIdentity() {
  createOpen.value = false;
}

async function saveIdentity() {
  if (
    !canGovern.value ||
    creating.value ||
    !createForm.name.trim() ||
    !createForm.slug.trim()
  )
    return;
  creating.value = true;
  createError.value = "";
  try {
    await call(
      identityFormMode.value === "create"
        ? "/admin/classification/identities"
        : `/admin/classification/identities/${encodeURIComponent(editingIdentityID.value)}`,
      {
      method: identityFormMode.value === "create" ? "POST" : "PATCH",
      body: {
        kind: createKind.value,
        name: createForm.name.trim(),
        slug: createForm.slug.trim(),
        ...(identityFormMode.value === "create"
          ? {
              parentId:
                createForm.parentId === "__root__" ? "" : createForm.parentId,
              facetId: createFacetID.value,
            }
          : {}),
      },
    },
    );
    createOpen.value = false;
    await (createKind.value === "tag" ? refreshTags() : refresh());
    actionSuccess.value = `${createTitle.value}完成。`;
  } catch (cause) {
    createError.value = classificationMutationErrorMessage(
      cause,
      `${createTitle.value}失败`,
    );
  } finally {
    creating.value = false;
  }
}
const searchPlaceholder = computed(
  () =>
    ({
      categories: "搜索分类名称或标识…",
      facets: "搜索维度或维度值…",
      tags: "搜索标签…",
      proposals: "搜索标签提案…",
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
  if (!canGovern.value) return;
  previewing.value = true;
  actionError.value = "";
  actionSuccess.value = "";
  pendingCommand.value = command;
  previewOpen.value = true;
  try {
    const response = await call<{
      preview: GalleryClassificationGovernancePreview;
    }>("/admin/classification/governance/preview", {
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
  if (!canGovern.value) return;
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
  if (
    !canGovern.value ||
    !pendingCommand.value ||
    preview.value?.outcome !== "planned"
  )
    return;
  executing.value = true;
  actionError.value = "";
  actionSuccess.value = "";
  const operation = pendingCommand.value.operation;
  try {
    await call("/admin/classification/governance/execute", {
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
    await Promise.all([
      refresh(),
      refreshTags(),
      ...(canReviewProposals.value ? [refreshProposals()] : []),
    ]);
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
      `/admin/classification/tags?size=40&cursor=${encodeURIComponent(cursor)}`,
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
  if (!canReviewProposals.value) return;
  reviewingProposal.value = item.id;
  actionError.value = "";
  actionSuccess.value = "";
  try {
    await call(
      `/admin/classification/tag-proposals/${encodeURIComponent(item.id)}/review`,
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
  <ManagePage id="classification" title="分类与维度" icon="i-tabler-category">
    <template #actions>
      <UButton
        v-if="canGovern && section === 'categories'"
        icon="i-tabler-plus"
        label="新增分类"
        @click="openCreateIdentity('category')"
      />
      <UButton
        v-else-if="canGovern && section === 'facets'"
        icon="i-tabler-plus"
        label="新增维度"
        @click="openCreateIdentity('facet')"
      />
      <UButton
        v-else-if="canGovern && section === 'tags'"
        icon="i-tabler-plus"
        label="新增标签"
        @click="openCreateIdentity('tag')"
      />
    </template>

    <TabbedSurface
      v-model="section"
      :items="tabs"
      navigation-label="分类治理"
      data-manage-surface="classification"
    >
      <div
        data-classification-search
        class="border-b border-default p-3 sm:p-4"
      >
        <UInput
          v-model="searchInput"
          icon="i-tabler-search"
          size="sm"
          :placeholder="searchPlaceholder"
          class="w-full"
        />
      </div>

      <div class="min-w-0 p-4 sm:p-5">
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
            :can-govern="canGovern"
            @action="handleIdentityAction"
            @edit="openEditIdentity"
          />
          <ClassificationFacetPanel
            v-else-if="section === 'facets'"
            :facets="visibleFacets"
            :can-govern="canGovern"
            @action="handleIdentityAction"
            @create-value="openCreateIdentity('facet_value', $event.id)"
            @edit="openEditIdentity"
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
            :can-govern="canGovern"
            @action="handleIdentityAction"
            @load-more="loadMoreTags"
            @edit="openEditIdentity"
          />
        </div>
      </div>
    </TabbedSurface>

    <UModal v-model:open="createOpen" :title="createTitle">
      <template #body>
        <form id="classification-create-form" class="space-y-4" @submit.prevent="saveIdentity">
          <UAlert
            v-if="createError"
            color="error"
            variant="subtle"
            title="创建失败"
            :description="createError"
          />
          <UFormField label="名称" required>
            <UInput v-model="createForm.name" class="w-full" autofocus />
          </UFormField>
          <UFormField label="标识" required>
            <UInput
              v-model="createForm.slug"
              class="w-full"
              placeholder="例如 landscape"
            />
          </UFormField>
          <UFormField
            v-if="identityFormMode === 'create' && !['facet', 'tag'].includes(createKind)"
            :label="createKind === 'category' ? '上级分类' : '上级维度值'"
          >
            <USelect
              v-model="createForm.parentId"
              :items="createParentItems"
              value-key="value"
              class="w-full"
            />
          </UFormField>
          <p class="text-xs leading-5 text-muted">
            {{ identityFormHint }}
          </p>
        </form>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton
            label="取消"
            color="neutral"
            variant="ghost"
            @click="closeCreateIdentity"
          />
          <UButton
            form="classification-create-form"
            type="submit"
            :label="createTitle"
            :loading="creating"
            :disabled="!createForm.name.trim() || !createForm.slug.trim()"
          />
        </div>
      </template>
    </UModal>

    <ClassificationEditorModal
      v-if="canGovern"
      v-model:open="editorOpen"
      :mode="editorMode"
      :identity="editorIdentity"
      :targets="editorTargets"
      :child-count="editorChildCount"
      @preview="previewEditor"
    />
    <ClassificationPreviewModal
      v-if="canGovern"
      v-model:open="previewOpen"
      :previewing="previewing"
      :executing="executing"
      :error="actionError"
      :preview="preview"
      :destructive="destructivePreview"
      @delete-all="confirmDeleteAllRelated"
      @execute="executePreview"
    />
  </ManagePage>
</template>
