<script setup lang="ts">
import {
  ManageCollectionToolbar,
  ManageEmpty,
  ManageHeader,
  ManageTabs,
  SkeletonList,
} from "@platform/manage/components";
import type {
  GalleryAdminSubmissionPage,
  GallerySubmission,
} from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "投稿与审核 · 图库管理" });
const route = useRoute();
const router = useRouter();
const { call } = useApi();
const hydrated = useClientHydrated();
const page = computed(() => Math.max(1, Number(route.query.page) || 1));
const q = computed(() => String(route.query.q || ""));
const qDraft = ref(q.value);
const sort = computed(() => String(route.query.sort || "oldest"));
const processingState = computed(() =>
  String(route.query.processingState || ""),
);
const reviewState = computed(() => String(route.query.reviewState || ""));
const safetyState = computed(() => String(route.query.safetyState || ""));
const outcome = computed(() => String(route.query.outcome || ""));
watch(q, (value) => {
  qDraft.value = value;
});

const note = ref<Record<string, string>>({});
const acting = ref("");
const actionErrors = ref<Record<string, string>>({});
const { data, pending, error, refresh } = await useAsyncData(
  "gallery-manage-submissions",
  () =>
    call<GalleryAdminSubmissionPage>("/api/v1/gallery/admin/submissions", {
      query: {
        q: q.value || undefined,
        sort: sort.value,
        page: page.value,
        size: 20,
        processingState: processingState.value || undefined,
        reviewState: reviewState.value || undefined,
        safetyState: safetyState.value || undefined,
        outcome: outcome.value || undefined,
      },
    }),
  {
    server: false,
    watch: [q, sort, page, processingState, reviewState, safetyState, outcome],
    default: () => ({
      items: [],
      page: 1,
      pageSize: 20,
      total: 0,
      totalPages: 0,
    }),
  },
);

const sortItems = [
  { label: "等待最久", value: "oldest" },
  { label: "最新投稿", value: "newest" },
  { label: "最近变化", value: "updated" },
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
const safetyItems = [
  { label: "全部安全状态", value: "all" },
  { label: "等待检查", value: "pending" },
  { label: "安全", value: "safe" },
  { label: "不确定", value: "uncertain" },
  { label: "已阻止", value: "blocked" },
  { label: "不可用", value: "unavailable" },
];
const outcomeItems = [
  { label: "全部结果", value: "all" },
  { label: "进行中", value: "pending" },
  { label: "已发布", value: "published" },
  { label: "重复内容", value: "duplicate" },
  { label: "已拒绝", value: "rejected" },
  { label: "已撤回", value: "withdrawn" },
  { label: "失败", value: "failed" },
];
const processingLabel = Object.fromEntries(
  processingItems.slice(1).map((item) => [item.value, item.label]),
);
const reviewLabel = Object.fromEntries(
  reviewItems.slice(1).map((item) => [item.value, item.label]),
);
const safetyLabel = Object.fromEntries(
  safetyItems.slice(1).map((item) => [item.value, item.label]),
);
const outcomeLabel = Object.fromEntries(
  outcomeItems.slice(1).map((item) => [item.value, item.label]),
);
const filterCount = computed(
  () =>
    [
      processingState.value,
      reviewState.value,
      safetyState.value,
      outcome.value,
    ].filter(Boolean).length,
);
const activePreset = computed(() => {
  if (
    reviewState.value === "pending" &&
    outcome.value === "pending" &&
    !processingState.value &&
    !safetyState.value
  )
    return "review";
  if (
    processingState.value === "failed" &&
    !reviewState.value &&
    !safetyState.value &&
    !outcome.value
  )
    return "failed";
  if (
    safetyState.value === "uncertain" &&
    outcome.value === "pending" &&
    !processingState.value &&
    !reviewState.value
  )
    return "uncertain";
  if (!filterCount.value) return "all";
  return "custom";
});
const presetModel = computed({
  get: () => activePreset.value,
  set: (value: string) => {
    if (["review", "failed", "uncertain", "all"].includes(value))
      preset(value as "review" | "failed" | "uncertain" | "all");
  },
});
const presetItems = [
  { key: "review", label: "待审核" },
  { key: "failed", label: "处理失败" },
  { key: "uncertain", label: "安全不确定" },
  { key: "all", label: "全部投稿" },
];

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
      (key === "sort" && value === "oldest")
    )
      continue;
    query[key] = String(value);
  }
  void router.push({ query });
}
function search() {
  setQuery({ q: qDraft.value.trim() || undefined });
}
function clearFilters() {
  qDraft.value = "";
  setQuery({
    q: undefined,
    processingState: undefined,
    reviewState: undefined,
    safetyState: undefined,
    outcome: undefined,
  });
}
function preset(kind: "review" | "failed" | "uncertain" | "all") {
  if (kind === "review")
    setQuery({
      processingState: undefined,
      reviewState: "pending",
      safetyState: undefined,
      outcome: "pending",
    });
  else if (kind === "failed")
    setQuery({
      processingState: "failed",
      reviewState: undefined,
      safetyState: undefined,
      outcome: undefined,
    });
  else if (kind === "uncertain")
    setQuery({
      processingState: undefined,
      reviewState: undefined,
      safetyState: "uncertain",
      outcome: "pending",
    });
  else
    setQuery({
      processingState: undefined,
      reviewState: undefined,
      safetyState: undefined,
      outcome: undefined,
    });
}
async function review(item: GallerySubmission, decision: "approve" | "reject") {
  acting.value = item.id;
  actionErrors.value = Object.fromEntries(
    Object.entries(actionErrors.value).filter(([id]) => id !== item.id),
  );
  try {
    await call(
      `/api/v1/gallery/admin/submissions/${encodeURIComponent(item.id)}/review`,
      { method: "POST", body: { decision, note: note.value[item.id] || "" } },
    );
    await refresh();
  } catch (reason: any) {
    actionErrors.value = {
      ...actionErrors.value,
      [item.id]:
        reason?.data?.message || "操作没有完成；状态可能已变化，请刷新后重试。",
    };
  } finally {
    acting.value = "";
  }
}
</script>

<template>
  <div>
    <ManageHeader title="投稿审核">
      <template #subtitle>
        先处理能进入目录的投稿；媒体失败和安全不确定保留为独立队列。
      </template>
    </ManageHeader>

    <ManageTabs v-model="presetModel" :items="presetItems" class="mb-4" />
    <ManageCollectionToolbar
      v-model:search="qDraft"
      search-placeholder="搜索标题、说明或来源…"
      :filter-count="filterCount"
      filter-label="状态筛选"
      class="mb-3"
    >
      <template #filters>
        <USelect
          :model-value="processingState || 'all'"
          :items="processingItems"
          value-key="value"
          @update:model-value="setQuery({ processingState: String($event) })"
        /><USelect
          :model-value="reviewState || 'all'"
          :items="reviewItems"
          value-key="value"
          @update:model-value="setQuery({ reviewState: String($event) })"
        /><USelect
          :model-value="safetyState || 'all'"
          :items="safetyItems"
          value-key="value"
          @update:model-value="setQuery({ safetyState: String($event) })"
        /><USelect
          :model-value="outcome || 'all'"
          :items="outcomeItems"
          value-key="value"
          @update:model-value="setQuery({ outcome: String($event) })"
        />
        <USelect
          :model-value="sort"
          :items="sortItems"
          value-key="value"
          aria-label="投稿排序"
          @update:model-value="setQuery({ sort: String($event) })"
        />
      </template>
      <template #actions>
        <UButton
          label="搜索"
          color="neutral"
          variant="outline"
          size="sm"
          @click="search"
        />
      </template>
      <template #mobile-actions>
        <UButton
          label="搜索"
          color="neutral"
          variant="outline"
          size="sm"
          @click="search"
        />
      </template>
    </ManageCollectionToolbar>
    <div
      class="mb-4 flex min-h-7 items-center justify-between gap-3 px-1 text-xs text-muted"
    >
      <span>共 {{ data.total }} 条投稿</span>
      <UButton
        v-if="filterCount || q || sort !== 'oldest'"
        color="neutral"
        variant="link"
        size="xs"
        label="清除筛选"
        @click="clearFilters"
      />
    </div>

    <SkeletonList v-if="!hydrated || pending" :rows="6" />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="投稿队列加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <section
      v-else-if="data.items.length"
      class="overflow-hidden rounded-xl border border-default bg-default"
      aria-label="投稿审核队列"
    >
      <article
        v-for="item in data.items"
        :key="item.id"
        class="grid gap-3 border-b border-default p-3 last:border-b-0 sm:grid-cols-[6.5rem_minmax(0,1fr)] sm:p-4 xl:grid-cols-[6.5rem_minmax(0,1fr)_19rem] xl:items-start"
      >
        <div class="relative overflow-hidden rounded-lg bg-elevated">
          <img
            :src="galleryRendition(item.assetId, 'thumbnail')"
            :alt="item.altText"
            class="aspect-[4/3] size-full object-cover"
          />
          <span
            class="absolute bottom-1.5 left-1.5 rounded bg-default/90 px-1.5 py-0.5 text-[11px] font-medium text-default backdrop-blur"
          >
            {{ outcomeLabel[item.outcome] }}
          </span>
        </div>
        <div class="min-w-0">
          <div class="flex min-w-0 items-center gap-2">
            <h2 class="truncate font-semibold text-highlighted">
              {{ item.title }}
            </h2>
            <UBadge
              v-if="item.safetyState !== 'safe'"
              :color="item.safetyState === 'blocked' ? 'error' : 'warning'"
              variant="soft"
              :label="safetyLabel[item.safetyState]"
            />
          </div>
          <p
            v-if="item.description"
            class="mt-1 line-clamp-2 text-sm leading-5 text-muted"
          >
            {{ item.description }}
          </p>
          <dl
            class="mt-3 grid grid-cols-2 gap-x-5 gap-y-2 text-xs sm:grid-cols-4 xl:grid-cols-2"
          >
            <div>
              <dt class="text-dimmed">媒体</dt>
              <dd class="mt-0.5 font-medium text-default">
                {{ processingLabel[item.processingState] }}
              </dd>
            </div>
            <div>
              <dt class="text-dimmed">人工审核</dt>
              <dd class="mt-0.5 font-medium text-default">
                {{ reviewLabel[item.reviewState] }}
              </dd>
            </div>
            <div>
              <dt class="text-dimmed">安全判断</dt>
              <dd class="mt-0.5 font-medium text-default">
                {{ safetyLabel[item.safetyState] }}
              </dd>
            </div>
            <div>
              <dt class="text-dimmed">最终结果</dt>
              <dd class="mt-0.5 font-medium text-default">
                {{ outcomeLabel[item.outcome] }}
              </dd>
            </div>
          </dl>
          <UAlert
            v-if="item.failureCode"
            class="mt-3"
            color="error"
            variant="subtle"
            icon="i-tabler-alert-triangle"
            title="媒体处理失败"
            :description="item.failureCode"
          />
          <p
            v-if="item.reviewNote"
            class="mt-3 border-l-2 border-default pl-3 text-sm text-muted"
          >
            审核记录：{{ item.reviewNote }}
          </p>
          <UAlert
            v-if="actionErrors[item.id]"
            class="mt-3"
            color="error"
            variant="subtle"
            title="本项操作失败"
            :description="actionErrors[item.id]"
          />
        </div>
        <div class="sm:col-start-2 xl:col-start-3">
          <div class="flex flex-wrap items-center gap-2 xl:justify-end">
            <UButton
              v-if="
                item.reviewState === 'pending' && item.outcome === 'pending'
              "
              label="批准进入目录"
              :loading="acting === item.id"
              :disabled="
                item.processingState !== 'ready' ||
                !['safe', 'uncertain'].includes(item.safetyState)
              "
              @click="review(item, 'approve')"
            />
            <UButton
              v-if="item.imageId"
              :to="`/images/${item.imageId}`"
              target="_blank"
              color="neutral"
              variant="ghost"
              icon="i-tabler-external-link"
              aria-label="查看公开图片"
            />
          </div>
          <details
            v-if="item.reviewState === 'pending' && item.outcome === 'pending'"
            class="group mt-3 rounded-lg border border-default bg-elevated/35 px-3 py-2"
          >
            <summary
              class="flex min-h-8 cursor-pointer list-none items-center justify-between gap-2 text-sm font-medium text-default"
            >
              备注或拒绝
              <UIcon
                name="i-tabler-chevron-down"
                class="size-4 text-muted transition group-open:rotate-180"
              />
            </summary>
            <div class="space-y-2 border-t border-default pt-3">
              <UTextarea
                v-model="note[item.id]"
                :rows="3"
                placeholder="记录判断；拒绝时必须填写原因"
              />
              <UButton
                color="error"
                variant="outline"
                label="拒绝"
                :loading="acting === item.id"
                :disabled="!note[item.id]?.trim()"
                @click="review(item, 'reject')"
              />
            </div>
          </details>
        </div>
      </article>
    </section>
    <ManageEmpty
      v-else
      icon="i-tabler-circle-check"
      title="当前筛选没有投稿"
      description="切换处理、安全或结果状态可以查看历史记录。"
    />

    <nav
      v-if="data.totalPages > 1"
      class="mt-6 flex items-center justify-center gap-3"
      aria-label="投稿分页"
    >
      <UButton
        color="neutral"
        variant="outline"
        label="上一页"
        :disabled="page <= 1"
        @click="setQuery({ page: page - 1 })"
      />
      <span class="text-sm tabular-nums text-muted"
        >{{ page }} / {{ data.totalPages }}</span
      >
      <UButton
        color="neutral"
        variant="outline"
        label="下一页"
        :disabled="page >= data.totalPages"
        @click="setQuery({ page: page + 1 })"
      />
    </nav>
  </div>
</template>
