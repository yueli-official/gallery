<script setup lang="ts">
import {
  ManageEmpty,
  ManageHeader,
  SkeletonList,
} from "@platform/manage/components";
import type { GalleryAdminCasePage, GalleryCase } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "处理单 · 图库管理" });
const route = useRoute();
const router = useRouter();
const { call } = useApi();
const hydrated = useClientHydrated();
const page = computed(() => Math.max(1, Number(route.query.page) || 1));
const q = computed(() => String(route.query.q || ""));
const qDraft = ref(q.value);
const status = computed(() => String(route.query.status || "open"));
const kind = computed(() => String(route.query.kind || ""));
const sort = computed(() => String(route.query.sort || "oldest"));
watch(q, (value) => {
  qDraft.value = value;
});

const notes = ref<Record<string, string>>({});
const acting = ref("");
const actionErrors = ref<Record<string, string>>({});
const { data, pending, error, refresh } = await useAsyncData(
  "gallery-manage-cases",
  () =>
    call<GalleryAdminCasePage>("/api/v1/gallery/admin/cases", {
      query: {
        q: q.value || undefined,
        sort: sort.value,
        status: status.value,
        kind: kind.value || undefined,
        page: page.value,
        size: 20,
      },
    }),
  {
    server: false,
    watch: [q, sort, status, kind, page],
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
const sortItems = [
  { label: "等待最久", value: "oldest" },
  { label: "最新创建", value: "newest" },
  { label: "最近变化", value: "updated" },
];
const kindLabel = Object.fromEntries(
  kindItems.slice(1).map((item) => [item.value, item.label]),
);
const statusLabel = Object.fromEntries(
  statusTabs.slice(0, 4).map((item) => [item.value, item.label]),
);

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
      (key === "sort" && value === "oldest")
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
function clearQuery() {
  qDraft.value = "";
  void router.push({
    query: status.value === "open" ? {} : { status: status.value },
  });
}
async function resolve(
  item: GalleryCase,
  nextStatus: "reviewing" | "resolved" | "dismissed",
) {
  if (!item.updatedAt) return;
  acting.value = item.id;
  actionErrors.value = Object.fromEntries(
    Object.entries(actionErrors.value).filter(([id]) => id !== item.id),
  );
  try {
    await call(
      `/api/v1/gallery/admin/cases/${encodeURIComponent(item.id)}/resolve`,
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
    <ManageHeader title="处理单工作台"
      ><template #subtitle
        >举报、来源修正、安全不确定、近重复与下架调查共享同一队列；状态变化使用更新时间避免覆盖其他运营者。</template
      ></ManageHeader
    >

    <section
      class="mb-5 space-y-3 rounded-xl border border-default bg-default p-4"
      aria-label="处理单筛选"
    >
      <div class="flex gap-1 overflow-x-auto pb-1" aria-label="处理单状态">
        <UButton
          v-for="tab in statusTabs"
          :key="tab.value"
          color="neutral"
          :variant="status === tab.value ? 'soft' : 'ghost'"
          :label="tab.label"
          @click="setQuery({ status: tab.value })"
        />
      </div>
      <div class="flex flex-wrap gap-2">
        <UInput
          v-model="qDraft"
          class="min-w-64 flex-1"
          icon="i-tabler-search"
          placeholder="搜索原因、说明或处理结论"
          @keyup.enter="search"
        /><UButton
          label="搜索"
          color="neutral"
          variant="outline"
          @click="search"
        /><USelect
          :model-value="kind || 'all'"
          :items="kindItems"
          value-key="value"
          class="w-40"
          aria-label="处理单类型"
          @update:model-value="setQuery({ kind: String($event) })"
        /><USelect
          :model-value="sort"
          :items="sortItems"
          value-key="value"
          class="w-36"
          aria-label="处理单排序"
          @update:model-value="setQuery({ sort: String($event) })"
        />
      </div>
      <div class="flex items-center justify-between text-xs text-muted">
        <span>共 {{ data.total }} 个处理单</span>
        <UButton
          v-if="q || kind || sort !== 'oldest'"
          color="neutral"
          variant="link"
          size="xs"
          label="清除查询"
          @click="clearQuery"
        />
      </div>
    </section>

    <SkeletonList v-if="!hydrated || pending" :rows="6" />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="处理单加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <div
      v-else-if="data.items.length"
      class="divide-y divide-default border-y border-default"
    >
      <article
        v-for="item in data.items"
        :key="item.id"
        class="grid gap-3 py-4 sm:grid-cols-[9rem_minmax(0,1fr)]"
      >
        <div>
          <UBadge
            color="neutral"
            variant="soft"
            :label="kindLabel[item.kind]"
          />
          <p class="mt-2 text-xs text-muted">{{ statusLabel[item.status] }}</p>
        </div>
        <div class="min-w-0">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="font-medium text-highlighted">
                {{ item.reason || kindLabel[item.kind] || "待运营复核" }}
              </p>
              <p
                v-if="item.description"
                class="mt-1 text-sm leading-6 text-toned"
              >
                {{ item.description }}
              </p>
              <a
                v-if="item.proposedSourceUrl"
                :href="item.proposedSourceUrl"
                target="_blank"
                rel="noopener noreferrer nofollow"
                class="mt-2 block break-all text-sm text-primary hover:underline"
                >{{ item.proposedSourceUrl }}</a
              >
            </div>
            <div class="flex gap-1">
              <UButton
                v-if="item.imageId"
                :to="`/images/${item.imageId}`"
                target="_blank"
                color="neutral"
                variant="ghost"
                icon="i-tabler-photo"
                aria-label="查看关联图片"
              />
            </div>
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
            v-if="['open', 'reviewing'].includes(item.status)"
            class="mt-4 grid gap-2 lg:grid-cols-[minmax(0,1fr)_auto]"
          >
            <UInput
              v-model="notes[item.id]"
              placeholder="处理结论（解决或忽略时必填）"
            />
            <div class="flex flex-wrap gap-2">
              <UButton
                v-if="item.status === 'open'"
                color="neutral"
                variant="outline"
                label="开始处理"
                :loading="acting === item.id"
                :disabled="!item.updatedAt"
                @click="resolve(item, 'reviewing')"
              /><UButton
                color="neutral"
                variant="outline"
                label="忽略"
                :loading="acting === item.id"
                :disabled="!item.updatedAt || !notes[item.id]?.trim()"
                @click="resolve(item, 'dismissed')"
              /><UButton
                label="解决"
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
      </article>
    </div>
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
