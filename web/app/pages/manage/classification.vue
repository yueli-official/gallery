<script setup lang="ts">
import { createPlatformNotifier } from "@platform/ui/feedback";
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
type ManagedIdentity = GalleryClassificationCatalogNode | GalleryClassificationCatalogFacet | GalleryClassificationTag;

const { call } = useApi();
const toast = createPlatformNotifier(useToast());
const { data, pending, error, refresh } = await useAsyncData(
  "gallery-manage-classification",
  () => call<{ catalog: GalleryClassificationCatalog }>("/api/v1/gallery/admin/classification"),
  {
    server: false,
    default: () => ({ catalog: { revision: 0, categories: [], facets: [] } }),
  },
);
const { data: tagData, pending: tagsPending, refresh: refreshTags } = await useAsyncData(
  "gallery-manage-classification-tags",
  () => call<{ page: GalleryClassificationTagPage }>("/api/v1/gallery/admin/classification/tags?size=40"),
  { server: false, default: () => ({ page: { items: [], nextCursor: "" } }) },
);
const { data: proposalData, pending: proposalsPending, refresh: refreshProposals } = await useAsyncData(
  "gallery-manage-classification-tag-proposals",
  () => call<{ proposals: GalleryClassificationTagProposal[]; total: number }>("/api/v1/gallery/admin/classification/tag-proposals?status=pending&page=1&size=30"),
  { server: false, default: () => ({ proposals: [], total: 0 }) },
);

const previewOpen = ref(false);
const previewing = ref(false);
const executing = ref(false);
const actionError = ref("");
const preview = ref<GalleryClassificationGovernancePreview>();
const pendingCommand = ref<GalleryClassificationGovernanceCommand>();
const editorOpen = ref(false);
const editorMode = ref<"reparent" | "merge">("reparent");
const editorKind = ref<IdentityKind>("category");
const editorIdentity = ref<ManagedIdentity>();
const editorTargetID = ref("");
const loadingMoreTags = ref(false);
const proposalTargets = ref<Record<string, string>>({});
const reviewingProposal = ref("");

const catalog = computed(() => data.value.catalog);
const destructivePreview = computed(() => pendingCommand.value?.operation === "delete" || pendingCommand.value?.operation === "merge");

function valuesForFacet(facet: GalleryClassificationCatalogFacet): GalleryClassificationCatalogNode[] {
  return facet.values;
}

function childrenOf(kind: IdentityKind, id: string): GalleryClassificationCatalogNode[] {
  if (kind === "category") return catalog.value.categories.filter((item) => item.parentId === id);
  if (kind === "facet_value") {
    return catalog.value.facets.flatMap((facet) => facet.values).filter((item) => item.parentId === id);
  }
  return [];
}

function owningFacet(id: string): GalleryClassificationCatalogFacet | undefined {
  return catalog.value.facets.find((facet) => facet.values.some((value) => value.id === id));
}

const editorTargets = computed<ManagedIdentity[]>(() => {
  const source = editorIdentity.value;
  if (!source) return [];
  if (editorKind.value === "category") {
    return catalog.value.categories.filter((item) => item.id !== source.id && item.status === "active");
  }
  if (editorKind.value === "facet_value") {
    return (owningFacet(source.id)?.values || []).filter((item) => item.id !== source.id && item.status === "active");
  }
  if (editorKind.value === "tag") {
    return tagData.value.page.items.filter((item) => item.id !== source.id && item.status === "active");
  }
  return [];
});

async function requestPreview(command: GalleryClassificationGovernanceCommand) {
  previewing.value = true;
  actionError.value = "";
  pendingCommand.value = command;
  previewOpen.value = true;
  try {
    const response = await call<{ preview: GalleryClassificationGovernancePreview }>(
      "/api/v1/gallery/admin/classification/governance/preview",
      { method: "POST", body: { command } },
    );
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
  return requestPreview({ operation: "delete", kind, id: item.id, childPlan: [], deleteAllRelated: false });
}

function openEditor(mode: "reparent" | "merge", kind: IdentityKind, item: ManagedIdentity) {
  editorMode.value = mode;
  editorKind.value = kind;
  editorIdentity.value = item;
  editorTargetID.value = "";
  editorOpen.value = true;
}

function previewEditor() {
  const source = editorIdentity.value;
  if (!source || !editorTargetID.value) return;
  const command: GalleryClassificationGovernanceCommand = {
    operation: editorMode.value,
    kind: editorKind.value,
    id: source.id,
    childPlan: [],
    deleteAllRelated: false,
  };
  if (editorMode.value === "reparent") {
    command.parentId = editorTargetID.value === "root" ? "" : editorTargetID.value;
  } else {
    command.targetId = editorTargetID.value;
    command.childPlan = childrenOf(editorKind.value, source.id).map((child) => ({
      childId: child.id,
      parentId: editorTargetID.value,
    }));
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
    // feedback-contract: execution closes the preview modal and refreshes multiple catalog sections.
    toast.add({ title: "分类治理已执行", description: operation === "delete" ? "相关标识和已确认依赖已删除。" : "目录 revision 已更新。", color: "success" });
  } catch (cause) {
    actionError.value = cause instanceof Error ? cause.message : "治理执行失败，请重新预览";
  } finally {
    executing.value = false;
  }
}

async function loadMoreTags() {
  const cursor = tagData.value.page.nextCursor;
  if (!cursor || loadingMoreTags.value) return;
  loadingMoreTags.value = true;
  try {
    const response = await call<{ page: GalleryClassificationTagPage }>(`/api/v1/gallery/admin/classification/tags?size=40&cursor=${encodeURIComponent(cursor)}`);
    tagData.value.page = {
      items: [...tagData.value.page.items, ...response.page.items],
      nextCursor: response.page.nextCursor,
    };
  } finally {
    loadingMoreTags.value = false;
  }
}

async function reviewTagProposal(item: GalleryClassificationTagProposal, decision: "approve" | "reject") {
  reviewingProposal.value = item.id;
  actionError.value = "";
  try {
    await call(`/api/v1/gallery/admin/classification/tag-proposals/${encodeURIComponent(item.id)}/review`, {
      method: "POST",
      body: { decision, targetTagId: decision === "approve" ? proposalTargets.value[item.id] || "" : "" },
    });
    await Promise.all([refresh(), refreshTags(), refreshProposals()]);
    // feedback-contract: the reviewed proposal leaves its current queue row after refresh.
    toast.add({ title: decision === "approve" ? "Tag 提案已批准" : "Tag 提案已拒绝", color: "success" });
  } catch (cause) {
    actionError.value = cause instanceof Error ? cause.message : "Tag 提案处理失败";
  } finally {
    reviewingProposal.value = "";
  }
}

function needsDeleteConfirmation(): boolean {
  return Boolean(preview.value?.diagnostics.some((item) => item.code === "govern.delete_confirmation_required"));
}

function stepLabel(kind: string): string {
  const labels: Record<string, string> = {
    change_status: "切换状态",
    move_child: "移动子节点",
    migrate_assignments: "迁移对象归属",
    migrate_primary: "迁移主分类",
    migrate_aliases: "迁移别名",
    set_replacement: "建立直接替代",
    clear_primary_assignments: "清除主分类",
    delete_assignments: "删除对象归属",
    delete_aliases: "删除别名",
    delete_references: "删除历史引用",
    delete_identity: "删除标识",
  };
  return labels[kind] || kind;
}
</script>

<template>
  <div>
    <ManageHeader title="分类与维度">
      <template #subtitle>Category 是主要浏览树，Facet 是结构化筛选轴；所有变更先预览影响，再按同一 revision 原子执行。</template>
      <template #actions><UBadge color="neutral" variant="soft" :label="`revision ${catalog.revision}`" /></template>
    </ManageHeader>

    <UAlert class="mb-5" color="info" variant="subtle" icon="i-tabler-shield-check" title="不会静默级联" description="停用保留历史关系；合并会迁移归属并建立直接 replacement；删除有关联时必须再次确认 delete-all-related。" />
    <UAlert v-if="actionError && !previewOpen" class="mb-5" color="error" variant="subtle" title="操作失败" :description="actionError" />
    <SkeletonList v-if="pending" :rows="6" />
    <UAlert v-else-if="error" color="error" variant="subtle" title="分类目录加载失败"><template #actions><UButton class="min-h-11" label="重试" @click="() => refresh()" /></template></UAlert>

    <div v-else class="space-y-8">
      <section>
        <div class="mb-3 flex items-end justify-between gap-3">
          <div><h2 class="font-semibold text-highlighted">Category</h2><p class="mt-1 text-sm text-muted">多归属、单父层级；主分类保存在独立 companion 关系中。</p></div>
          <span class="text-xs text-dimmed">{{ catalog.categories.length }} 项</span>
        </div>
        <div class="divide-y divide-default border-y border-default">
          <article v-for="item in catalog.categories" :key="item.id" class="grid gap-3 py-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2"><h3 class="font-medium text-highlighted">{{ item.name }}</h3><UBadge :color="item.status === 'active' ? 'success' : item.status === 'replaced' ? 'warning' : 'neutral'" variant="soft" :label="item.status" /></div>
              <p class="mt-1 truncate text-sm text-muted">{{ item.slug }} · {{ item.id }}</p>
              <p v-if="item.parentId" class="mt-1 text-xs text-dimmed">父节点 {{ item.parentId }}</p>
            </div>
            <div class="flex flex-wrap gap-2">
              <UButton v-if="item.status !== 'replaced'" class="min-h-11" color="neutral" variant="outline" size="sm" :label="item.status === 'active' ? '停用' : '启用'" @click="setStatus('category', item)" />
              <UButton v-if="item.status !== 'replaced'" class="min-h-11" color="neutral" variant="ghost" size="sm" label="移动" @click="openEditor('reparent', 'category', item)" />
              <UButton v-if="item.status !== 'replaced'" class="min-h-11" color="neutral" variant="ghost" size="sm" label="合并" @click="openEditor('merge', 'category', item)" />
              <UButton class="min-h-11" color="error" variant="ghost" size="sm" label="删除" @click="deleteIdentity('category', item)" />
            </div>
          </article>
        </div>
      </section>

      <section>
        <div class="mb-3"><h2 class="font-semibold text-highlighted">Facet</h2><p class="mt-1 text-sm text-muted">Facet 本身只启停或删除；需要合并轴时，先逐个治理 Value，再删除空轴。</p></div>
        <div class="grid gap-4 xl:grid-cols-2">
          <article v-for="facet in catalog.facets" :key="facet.id" class="rounded-lg border border-default bg-default p-4 sm:p-5">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div><div class="flex items-center gap-2"><h3 class="font-semibold text-highlighted">{{ facet.name }}</h3><UBadge :color="facet.status === 'active' ? 'success' : facet.status === 'replaced' ? 'warning' : 'neutral'" variant="soft" :label="facet.status" /></div><p class="mt-1 text-xs text-muted">{{ facet.slug }} · {{ facet.id }}</p></div>
              <div class="flex gap-2"><UButton v-if="facet.status !== 'replaced'" class="min-h-11" color="neutral" variant="ghost" size="sm" :label="facet.status === 'active' ? '停用' : '启用'" @click="setStatus('facet', facet)" /><UButton class="min-h-11" color="error" variant="ghost" size="sm" label="删除" @click="deleteIdentity('facet', facet)" /></div>
            </div>
            <div class="mt-4 divide-y divide-default">
              <div v-for="value in valuesForFacet(facet)" :key="value.id" class="py-3">
                <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                  <div class="min-w-0"><div class="flex items-center gap-2"><span class="text-sm font-medium text-highlighted">{{ value.name }}</span><UBadge size="xs" :color="value.status === 'active' ? 'success' : value.status === 'replaced' ? 'warning' : 'neutral'" variant="soft" :label="value.status" /></div><p class="mt-1 truncate text-xs text-dimmed">{{ value.slug }} · {{ value.id }}</p></div>
                  <div class="flex flex-wrap gap-2"><UButton v-if="value.status !== 'replaced'" class="min-h-11" color="neutral" variant="ghost" size="xs" :label="value.status === 'active' ? '停用' : '启用'" @click="setStatus('facet_value', value)" /><UButton v-if="value.status !== 'replaced'" class="min-h-11" color="neutral" variant="ghost" size="xs" label="移动" @click="openEditor('reparent', 'facet_value', value)" /><UButton v-if="value.status !== 'replaced'" class="min-h-11" color="neutral" variant="ghost" size="xs" label="合并" @click="openEditor('merge', 'facet_value', value)" /><UButton class="min-h-11" color="error" variant="ghost" size="xs" label="删除" @click="deleteIdentity('facet_value', value)" /></div>
                </div>
              </div>
            </div>
          </article>
        </div>
      </section>

      <section>
        <div class="mb-3 flex items-end justify-between gap-3"><div><h2 class="font-semibold text-highlighted">Tag Proposal</h2><p class="mt-1 text-sm text-muted">批准时先重新解析最新 Lookup Registry；可创建 canonical Tag，或显式归并为已有 Tag 的 Alias。</p></div><span class="text-xs text-dimmed">{{ proposalData.total }} 条待审</span></div>
        <SkeletonList v-if="proposalsPending" :rows="3" />
        <div v-else-if="proposalData.proposals.length" class="divide-y divide-default border-y border-default">
          <article v-for="item in proposalData.proposals" :key="item.id" class="grid gap-3 py-4 lg:grid-cols-[minmax(0,1fr)_minmax(14rem,22rem)_auto] lg:items-center">
            <div class="min-w-0"><p class="font-medium text-highlighted">{{ item.inputValue }}</p><p class="mt-1 truncate text-xs text-muted">lookup: {{ item.lookupKey }} · submission {{ item.submissionId }}</p></div>
            <UFormField label="批准方式">
              <select v-model="proposalTargets[item.id]" class="h-11 w-full rounded-md border border-default bg-default px-3 text-sm text-highlighted" :aria-label="`选择 ${item.inputValue} 的 Tag 处理方式`">
                <option value="">批准并创建新 canonical Tag</option>
                <option v-for="tag in tagData.page.items.filter((entry) => entry.status === 'active')" :key="tag.id" :value="tag.id">作为 {{ tag.name }} 的 Alias</option>
              </select>
            </UFormField>
            <div class="flex gap-2"><UButton class="min-h-11" color="neutral" variant="outline" label="拒绝" :loading="reviewingProposal === item.id" @click="reviewTagProposal(item, 'reject')" /><UButton class="min-h-11" label="批准" :loading="reviewingProposal === item.id" @click="reviewTagProposal(item, 'approve')" /></div>
          </article>
        </div>
        <ManageEmpty v-else icon="i-tabler-tag-off" title="没有待审 Tag" description="未知投稿词会进入独立 Proposal，不会提前污染公开目录。" />
      </section>

      <section>
        <div class="mb-3 flex items-end justify-between gap-3"><div><h2 class="font-semibold text-highlighted">Tag</h2><p class="mt-1 text-sm text-muted">扁平长尾词；Alias 和 replacement 直接解析到 canonical Tag，不允许链。</p></div><span class="text-xs text-dimmed">keyset cursor</span></div>
        <SkeletonList v-if="tagsPending" :rows="4" />
        <div v-else-if="tagData.page.items.length" class="divide-y divide-default border-y border-default">
          <article v-for="tag in tagData.page.items" :key="tag.id" class="grid gap-3 py-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
            <div class="min-w-0"><div class="flex flex-wrap items-center gap-2"><h3 class="font-medium text-highlighted">{{ tag.name }}</h3><UBadge :color="tag.status === 'active' ? 'success' : tag.status === 'replaced' ? 'warning' : 'neutral'" variant="soft" :label="tag.status" /></div><p class="mt-1 truncate text-xs text-muted">{{ tag.slug }} · {{ tag.id }}</p><p class="mt-1 text-xs text-dimmed">{{ tag.assignmentCount }} 个关系 · {{ tag.aliasCount }} 个 Alias</p></div>
            <div class="flex flex-wrap gap-2"><UButton v-if="tag.status !== 'replaced'" class="min-h-11" color="neutral" variant="ghost" size="sm" :label="tag.status === 'active' ? '停用' : '启用'" @click="setStatus('tag', tag)" /><UButton v-if="tag.status !== 'replaced'" class="min-h-11" color="neutral" variant="ghost" size="sm" label="合并" @click="openEditor('merge', 'tag', tag)" /><UButton class="min-h-11" color="error" variant="ghost" size="sm" label="删除" @click="deleteIdentity('tag', tag)" /></div>
          </article>
          <div v-if="tagData.page.nextCursor" class="flex justify-center py-4"><UButton class="min-h-11" color="neutral" variant="outline" label="加载更多 Tag" :loading="loadingMoreTags" @click="loadMoreTags" /></div>
        </div>
        <ManageEmpty v-else icon="i-tabler-tags" title="还没有 canonical Tag" description="批准第一个 Proposal 后会在这里出现。" />
      </section>
    </div>

    <UModal v-model:open="editorOpen" :title="editorMode === 'merge' ? '合并标识' : '移动节点'" description="系统会根据最新目录生成计划；这里只选择目标，不直接写数据库。">
      <template #body>
        <form class="space-y-4" @submit.prevent="previewEditor">
          <div class="rounded-md bg-elevated p-3 text-sm"><p class="font-medium text-highlighted">{{ editorIdentity?.name }}</p><p class="mt-1 break-all text-xs text-muted">{{ editorIdentity?.id }}</p></div>
          <UFormField :label="editorMode === 'merge' ? '合并目标' : '新父节点'" required>
            <select v-model="editorTargetID" class="h-11 w-full rounded-md border border-default bg-default px-3 text-sm text-highlighted">
              <option value="" disabled>请选择</option>
              <option v-if="editorMode === 'reparent'" value="root">设为根节点</option>
              <option v-for="target in editorTargets" :key="target.id" :value="target.id">{{ target.name }} · {{ target.slug }}</option>
            </select>
          </UFormField>
          <p v-if="editorMode === 'merge' && editorIdentity && childrenOf(editorKind, editorIdentity.id).length" class="text-xs text-muted">其直接子节点将显式移动到目标；预览会再次验证完整性和环。</p>
          <div class="flex justify-end gap-2"><UButton class="min-h-11" color="neutral" variant="ghost" label="取消" @click="() => { editorOpen = false; }" /><UButton class="min-h-11" type="submit" label="生成预览" :disabled="!editorTargetID" /></div>
        </form>
      </template>
    </UModal>

    <UModal v-model:open="previewOpen" title="治理影响预览" description="执行时会重算 revision 和 impact token；任何变化都会拒绝旧计划。">
      <template #body>
        <div class="space-y-4">
          <div v-if="previewing" class="py-8"><SkeletonList :rows="4" /></div>
          <UAlert v-else-if="actionError" color="error" variant="subtle" title="操作失败" :description="actionError" />
          <template v-else-if="preview">
            <div class="flex items-center justify-between gap-3"><UBadge :color="preview.outcome === 'planned' ? 'success' : 'warning'" variant="soft" :label="preview.outcome" /><span class="text-xs text-muted">revision {{ preview.catalogRevision }}</span></div>
            <div v-if="preview.diagnostics.length" class="space-y-2"><UAlert v-for="item in preview.diagnostics" :key="`${item.code}:${item.reference}`" color="warning" variant="subtle" :title="item.code" :description="item.reference || item.path.join('.')" /></div>
            <ol v-if="preview.plan.steps.length" class="divide-y divide-default rounded-lg border border-default px-4">
              <li v-for="(step, index) in preview.plan.steps" :key="`${index}:${step.kind}:${step.sourceId}`" class="flex items-start justify-between gap-3 py-3 text-sm"><div><p class="font-medium text-highlighted">{{ index + 1 }}. {{ stepLabel(step.kind) }}</p><p class="mt-1 break-all text-xs text-muted">{{ step.sourceId }}<template v-if="step.targetId"> → {{ step.targetId }}</template></p></div><UBadge v-if="step.affectedCount" color="neutral" variant="soft" :label="`${step.affectedCount} 项`" /></li>
            </ol>
          </template>
          <div class="flex flex-wrap justify-end gap-2"><UButton class="min-h-11" color="neutral" variant="ghost" label="关闭" @click="() => { previewOpen = false; }" /><UButton v-if="needsDeleteConfirmation()" class="min-h-11" color="error" variant="outline" label="生成全部删除计划" :loading="previewing" @click="confirmDeleteAllRelated" /><UButton v-if="preview?.outcome === 'planned'" class="min-h-11" :color="destructivePreview ? 'error' : 'primary'" :label="destructivePreview ? '确认并执行' : '执行计划'" :loading="executing" @click="executePreview" /></div>
        </div>
      </template>
    </UModal>
  </div>
</template>
