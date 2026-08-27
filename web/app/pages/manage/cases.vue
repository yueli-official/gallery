<script setup lang="ts">
import { PageHeader } from "@yueli/ui/dashboard/pattern";
import {
  ManageEmpty,
  ManageTabs,
  SkeletonList,
} from "~/utils/manageComponents";
import {
  CollectionPagination,
  CollectionSortHeader,
  CollectionTableToolbar,
} from "@yueli/ui/collection/pattern";
import type { GalleryAdminCasePage, GalleryCase } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "处理单 · 图库管理" });
const route = useRoute();
const router = useRouter();
const { call } = useGalleryApi();
const { can } = useGalleryMe();
const hydrated = useClientHydrated();
const canResolveCases = computed(() => can("gallery.case.resolve"));
const page = computed(() => Math.max(1, Number(route.query.page) || 1));
const q = computed(() => String(route.query.q || ""));
const qDraft = ref(q.value);
const status = computed(() => String(route.query.status || "open"));
const kind = computed(() => String(route.query.kind || ""));
type CaseSortBy = "createdAt" | "updatedAt" | "kind" | "status";
type CaseSortOrder = "asc" | "desc";
const sortBy = computed<CaseSortBy>(() => {
  const value = String(route.query.sortBy || "createdAt");
  return ["createdAt", "updatedAt", "kind", "status"].includes(value)
    ? (value as CaseSortBy)
    : "createdAt";
});
const sortOrder = computed<CaseSortOrder>(() =>
  String(route.query.sortOrder || "asc") === "desc" ? "desc" : "asc",
);
watch(q, (value) => {
  qDraft.value = value;
});

const notes = ref<Record<string, string>>({});
const acting = ref("");
const actionErrors = ref<Record<string, string>>({});
const resolvingId = ref("");
const { data, pending, error, refresh } = await useAsyncData(
  "gallery-manage-cases",
  () =>
    call<GalleryAdminCasePage>("/admin/cases", {
      query: {
        q: q.value || undefined,
        sortBy: sortBy.value,
        sortOrder: sortOrder.value,
        status: status.value,
        kind: kind.value || undefined,
        page: page.value,
        size: 20,
      },
    }),
  {
    server: false,
    watch: [q, sortBy, sortOrder, status, kind, page],
    default: () => ({
      items: [],
      page: 1,
      pageSize: 20,
      total: 0,
      totalPages: 0,
    }),
  },
);

const statusTabs = [
  { label: "待处理", value: "open" },
  { label: "处理中", value: "reviewing" },
  { label: "已解决", value: "resolved" },
  { label: "已忽略", value: "dismissed" },
  { label: "全部", value: "all" },
];
const kindItems = [
  { label: "全部类型", value: "all" },
  { label: "内容举报", value: "report" },
  { label: "来源修正", value: "source_correction" },
  { label: "安全不确定", value: "safety_uncertain" },
  { label: "近重复", value: "near_duplicate" },
  { label: "下架调查", value: "takedown" },
];
const kindLabel = Object.fromEntries(
  kindItems.slice(1).map((item) => [item.value, item.label]),
);
const statusLabel = Object.fromEntries(
  statusTabs.slice(0, 4).map((item) => [item.value, item.label]),
);
const statusModel = computed({
  get: () => status.value,
  set: (value: string) => setQuery({ status: value }),
});
const tabItems = statusTabs.map((item) => ({
  key: item.value,
  label: item.label,
}));
const filterCount = computed(() => [kind.value].filter(Boolean).length);

function formatDate(value?: string) {
  if (!value) return "时间未知";
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

function closeResolution() {
  resolvingId.value = "";
}

function toggleResolution(itemId: string) {
  resolvingId.value = resolvingId.value === itemId ? "" : itemId;
}

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
      (key === "status" && value === "open") ||
      (key === "sortBy" && value === "createdAt") ||
      (key === "sortOrder" && value === "asc")
    )
      continue;
    query[key] = String(value);
  }
  if (values.status === "all") query.status = "all";
  void router.push({ query });
}
function search() {
  setQuery({ q: qDraft.value.trim() || undefined });
}
function changeColumnSort(nextSortBy: CaseSortBy) {
  if (sortBy.value === nextSortBy) {
    setQuery({ sortOrder: sortOrder.value === "asc" ? "desc" : "asc" });
    return;
  }
  setQuery({ sortBy: nextSortBy, sortOrder: "desc" });
}
async function resolve(
  item: GalleryCase,
  nextStatus: "reviewing" | "resolved" | "dismissed",
) {
  if (!canResolveCases.value || !item.updatedAt) return;
  acting.value = item.id;
  actionErrors.value = Object.fromEntries(
    Object.entries(actionErrors.value).filter(([id]) => id !== item.id),
  );
  try {
    await call(
      `/admin/cases/${encodeURIComponent(item.id)}/resolve`,
      {
        method: "POST",
        body: {
          expectedUpdatedAt: item.updatedAt,
          status: nextStatus,
          note: notes.value[item.id] || "",
        },
      },
    );
    await refresh();
  } catch (reason: any) {
    actionErrors.value = {
      ...actionErrors.value,
      [item.id]:
        reason?.data?.message ||
        "处理单可能已被其他运营者更新；当前备注已保留，请刷新后重试。",
    };
  } finally {
    acting.value = "";
  }
}
</script>

<template>
  <div>
    <PageHeader title="处理单" />

    <ManageTabs v-model="statusModel" :items="tabItems" class="mb-4" />
    <CollectionTableToolbar
      v-model:search="qDraft"
      label="处理单工具栏"
      search-placeholder="搜索原因、说明或处理结论…"
      search-action="搜索"
      :filter-count="filterCount"
      filter-label="筛选"
      class="mb-3"
      @search="search"
    >
      <template #filters>
        <div class="w-72 max-w-[calc(100vw-2rem)]">
          <UFormField label="处理单类型">
            <USelect
              :model-value="kind || 'all'"
              :items="kindItems"
              value-key="value"
              class="w-full"
              aria-label="处理单类型"
              @update:model-value="setQuery({ kind: String($event) })"
            />
          </UFormField>
        </div>
      </template>
    </CollectionTableToolbar>
    <SkeletonList v-if="!hydrated || pending" :rows="6" />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="处理单加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <section
      v-else-if="data.items.length"
      class="overflow-hidden rounded-xl border border-default bg-default"
      aria-label="信任处理单队列"
    >
      <div
        class="grid grid-cols-[minmax(0,1fr)_7rem] items-center gap-3 border-b border-default bg-elevated/50 px-4 py-2 text-xs font-medium text-muted md:grid-cols-[minmax(0,1fr)_8rem_8rem_7rem]"
      >
        <CollectionSortHeader
          label="处理单"
          :active="sortBy === 'createdAt'"
          :sort-order="sortOrder"
          @sort="changeColumnSort('createdAt')"
        />
        <CollectionSortHeader
          class="hidden md:inline-flex"
          label="类型"
          :active="sortBy === 'kind'"
          :sort-order="sortOrder"
          @sort="changeColumnSort('kind')"
        />
        <CollectionSortHeader
          class="hidden md:inline-flex"
          label="状态"
          :active="sortBy === 'status'"
          :sort-order="sortOrder"
          @sort="changeColumnSort('status')"
        />
        <span class="text-right">操作</span>
      </div>
      <article
        v-for="item in data.items"
        :key="item.id"
        class="grid gap-3 border-b border-default p-4 last:border-b-0 md:grid-cols-[minmax(0,1fr)_8rem_8rem_7rem] md:items-start"
      >
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
            <h2 class="font-semibold text-highlighted">
              {{ item.reason || kindLabel[item.kind] || "待运营复核" }}
            </h2>
            <span class="text-xs text-muted">{{ kindLabel[item.kind] }}</span>
          </div>
          <p
            v-if="item.description"
            class="mt-1 line-clamp-2 text-sm leading-5 text-muted"
          >
            {{ item.description }}
          </p>
          <div
            class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-dimmed"
          >
            <span>创建于 {{ formatDate(item.createdAt) }}</span>
            <a
              v-if="item.proposedSourceUrl"
              :href="item.proposedSourceUrl"
              target="_blank"
              rel="noopener noreferrer nofollow"
              class="max-w-72 truncate text-primary hover:underline"
              >查看建议来源</a
            >
          </div>
          <UAlert
            v-if="actionErrors[item.id]"
            class="mt-3"
            color="error"
            variant="subtle"
            title="本项操作失败"
            :description="actionErrors[item.id]"
          />
          <div
            v-if="canResolveCases && resolvingId === item.id"
            class="mt-4 rounded-lg border border-default bg-elevated/35 p-3"
          >
            <UFormField label="处理结论" required>
              <UTextarea
                v-model="notes[item.id]"
                :rows="3"
                placeholder="说明核查依据和最终判断"
              />
            </UFormField>
            <div class="mt-3 flex flex-wrap justify-end gap-2">
              <UButton
                color="neutral"
                variant="ghost"
                label="取消"
                @click="closeResolution"
              />
              <UButton
                color="error"
                variant="outline"
                label="忽略处理单"
                :loading="acting === item.id"
                :disabled="!item.updatedAt || !notes[item.id]?.trim()"
                @click="resolve(item, 'dismissed')"
              /><UButton
                label="标记已解决"
                :loading="acting === item.id"
                :disabled="!item.updatedAt || !notes[item.id]?.trim()"
                @click="resolve(item, 'resolved')"
              />
            </div>
          </div>
          <p v-else-if="item.resolutionNote" class="mt-3 text-sm text-muted">
            结论：{{ item.resolutionNote }}
          </p>
        </div>
        <span class="hidden text-xs text-muted md:block">{{ kindLabel[item.kind] }}</span>
        <span class="hidden text-xs text-muted md:block">{{ statusLabel[item.status] }}</span>
        <div class="flex flex-wrap items-center gap-1 md:justify-end">
          <UButton
            v-if="canResolveCases && item.status === 'open'"
            icon="i-tabler-hand-click"
            aria-label="接手处理"
            size="xs"
            square
            :loading="acting === item.id"
            :disabled="!item.updatedAt"
            @click="resolve(item, 'reviewing')"
          />
          <UButton
            v-else-if="canResolveCases && item.status === 'reviewing'"
            icon="i-tabler-check"
            aria-label="完成处理"
            size="xs"
            square
            @click="toggleResolution(item.id)"
          />
          <UButton
            v-if="item.imageId"
            :to="`/images/${item.imageId}`"
            target="_blank"
            color="neutral"
            variant="ghost"
            icon="i-tabler-photo"
            size="xs"
            square
            aria-label="查看关联图片"
          />
        </div>
      </article>
    </section>
    <ManageEmpty
      v-else
      icon="i-tabler-flag-off"
      title="当前队列为空"
      description="选择其他状态或类型可以查看已处理记录。"
    />

    <nav
      v-if="data.totalPages > 1"
      class="mt-6 flex items-center justify-center gap-3"
      aria-label="处理单分页"
    >
      <CollectionPagination
        :model-value="page"
        :total-pages="data.totalPages"
        @update:model-value="setQuery({ page: $event })"
      />
    </nav>
  </div>
</template>
