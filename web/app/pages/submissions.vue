<script setup lang="ts">
import { createGalleryNotifier } from "~/utils/feedback";
import type { GallerySubmission } from "~/types/gallery";

interface SubmissionPage {
  items: GallerySubmission[];
  total: number;
  page: number;
  size: number;

}

const { loggedIn, login } = useAuth();
const route = useRoute();
const router = useRouter();
const hydrated = useClientHydrated();
const { call } = useGalleryApi();
const toast = createGalleryNotifier(useToast());
const page = computed(() => Math.max(1, Number(route.query.page) || 1));
const outcome = computed(() => String(route.query.outcome || ""));
const processingState = computed(() =>
  String(route.query.processingState || ""),
);
const reviewState = computed(() => String(route.query.reviewState || ""));
const { data, error, pending, refresh } = await useAsyncData(
  "gallery-my-submissions",
  () =>
    call<SubmissionPage>("/me/submissions", {
      query: {
        page: page.value,
        size: 20,
        outcome: outcome.value || undefined,
        processingState: processingState.value || undefined,
        reviewState: reviewState.value || undefined,
      },
    }),
  {
    server: false,
    watch: [page, outcome, processingState, reviewState],
    default: () => ({
      items: [],
      total: 0,
      page: 1,
      size: 20,
    }),
  },
);
const withdrawing = ref("");
const outcomeItems = [
  { label: "全部结果", value: "all" },
  { label: "进行中", value: "pending" },
  { label: "审核通过（已发布）", value: "published" },
  { label: "审核通过（重复收录）", value: "duplicate" },
  { label: "审核未通过", value: "rejected" },
  { label: "处理失败", value: "failed" },
  { label: "已撤回", value: "withdrawn" },
];
const processingItems = [
  { label: "全部处理状态", value: "all" },
  { label: "排队中", value: "queued" },
  { label: "处理中", value: "processing" },
  { label: "处理完成", value: "ready" },
  { label: "处理失败", value: "failed" },
];
const reviewItems = [
  { label: "全部审核状态", value: "all" },
  { label: "无需人工审核", value: "not_required" },
  { label: "等待审核", value: "pending" },
  { label: "审核通过", value: "approved" },
  { label: "审核未通过", value: "rejected" },
];
function submissionSummary(submission: GallerySubmission) {
  if (submission.reviewNote) return submission.reviewNote;
  if (submission.failureCode) return ({
    unsupported_format: "文件格式不受支持，请更换图片后重试",
    animated_image: "暂不支持动画图片，请使用静态图片",
    derive_failed: "公开图片生成失败，请重新投稿",
    processing_failed: "图片处理失败，请重新投稿",
    safety_unavailable: "内容检查暂时不可用，请稍后重试",
  } as Record<string, string>)[submission.failureCode] || "图片处理失败，请重新投稿";
  if (submission.outcome === "duplicate") return "重复内容已收录到现有图片";
  if (submission.outcome === "published") return "已发布到图片目录";
  if (submissionStatus(submission) === "处理中") return "正在读取图片并生成预览，完成后进入等待审核";
  if (submissionStatus(submission) === "等待审核") return "媒体处理完成，等待管理员审核";
  return "";
}

function setQuery(
  key: "outcome" | "processingState" | "reviewState",
  value: string,
) {
  const query = Object.fromEntries(
    Object.entries(route.query).filter(
      ([entry]) => entry !== key && entry !== "page",
    ),
  );
  if (value !== "all") query[key] = value;
  void router.push({ query });
}
function setPage(value: number) {
  const query = { ...route.query };
  if (value <= 1) delete query.page;
  else query.page = String(value);
  void router.push({ query });
}
function clearFilters() {
  void router.push({ query: {} });
}

async function withdraw(id: string) {
  withdrawing.value = id;
  try {
    await call(
      `/me/submissions/${encodeURIComponent(id)}/withdraw`,
      { method: "POST" },
    );
    await refresh();
  } catch (reason: any) {
    toast.add({
      title: "无法撤回这条投稿",
      description: galleryFailureMessage(reason, "状态可能已经变化，请刷新后重试"),
      color: "error",
    });
    await refresh();
  } finally {
    withdrawing.value = "";
  }
}
useSeoMeta({ title: "我的投稿", robots: "noindex,nofollow" });
</script>

<template>
  <GalleryPublicPage class="max-w-6xl">
    <GalleryPageHeader
      title="我的投稿"
      description="跟踪每张图片的处理、审核与最终结果，失败原因会保留在对应记录中。"
    >
      <template #actions>
        <UButton
          to="/submit"
          icon="i-tabler-library-plus"
          label="批量投稿"
        />
      </template>
    </GalleryPageHeader>

    <UAlert
      v-if="!loggedIn"
      class="mb-6"
      color="neutral"
      variant="subtle"
      icon="i-tabler-clock-shield"
      title="正在查看此浏览器的临时投稿"
      description="临时身份保留 30 天。登录后会自动把这些投稿转入你的账号。"
    >
      <template #actions>
        <UButton label="登录并长期保留" @click="void login()" />
      </template>
    </UAlert>

    <section
      class="mb-6 rounded-xl border border-default bg-default/75 p-4"
      aria-label="筛选投稿记录"
    >
      <div class="grid gap-3 sm:grid-cols-3">
        <USelect
          :model-value="outcome || 'all'"
          :items="outcomeItems"
          value-key="value"
          aria-label="按最终结果筛选"
          @update:model-value="setQuery('outcome', String($event))"
        />
        <USelect
          :model-value="processingState || 'all'"
          :items="processingItems"
          value-key="value"
          aria-label="按处理状态筛选"
          @update:model-value="setQuery('processingState', String($event))"
        />
        <USelect
          :model-value="reviewState || 'all'"
          :items="reviewItems"
          value-key="value"
          aria-label="按审核状态筛选"
          @update:model-value="setQuery('reviewState', String($event))"
        />
      </div>
      <div
        class="mt-3 flex items-center justify-between gap-3 text-xs text-muted"
      >
        <span>共 {{ data.total }} 条记录</span>
        <UButton
          v-if="outcome || processingState || reviewState"
          color="neutral"
          variant="link"
          size="xs"
          label="清除筛选"
          @click="clearFilters"
        />
      </div>
    </section>

    <div v-if="!hydrated || pending" class="space-y-3">
      <USkeleton v-for="index in 6" :key="index" class="h-32 rounded-lg" />
    </div>
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="投稿记录加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <div
      v-else-if="data.items.length"
      class="overflow-hidden rounded-2xl border border-default bg-default/70 px-5 shadow-sm"
    >
      <article
        v-for="submission in data.items"
        :key="submission.id"
        class="grid gap-4 border-b border-default py-5 last:border-b-0 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-start"
      >
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h2 class="truncate font-medium text-highlighted">
              {{ submission.title }}
            </h2>
            <UBadge
              :color="submissionStatusColor(submission)"
              variant="soft"
              :label="submissionStatus(submission)"
            />
          </div>
          <p v-if="submissionSummary(submission)" class="mt-2 flex items-center gap-1.5 text-xs text-muted">
            <UIcon name="i-tabler-progress-check" class="size-4 shrink-0" />
            {{ submissionSummary(submission) }}
          </p>
          <p v-if="submission.createdAt" class="mt-3 text-xs text-dimmed">
            提交于 {{ new Date(submission.createdAt).toLocaleString("zh-CN") }}
          </p>
        </div>
        <div class="flex gap-2">
          <UButton
            v-if="submission.imageId"
            :to="`/images/${submission.imageId}`"
            color="neutral"
            variant="outline"
            size="sm"
            label="查看图片"
          />
          <UButton
            v-if="['pending', 'published'].includes(submission.outcome)"
            color="error"
            variant="ghost"
            size="sm"
            label="撤回"
            :loading="withdrawing === submission.id"
            @click="withdraw(submission.id)"
          />
        </div>
      </article>
    </div>
    <GalleryCompactEmpty
      v-else
      :icon="
        outcome || processingState || reviewState
          ? 'i-tabler-filter-off'
          : 'i-tabler-photo-up'
      "
      :title="
        outcome || processingState || reviewState
          ? '没有符合筛选条件的记录'
          : '还没有投稿记录'
      "
    >
      <template #actions>
        <UButton
          v-if="outcome || processingState || reviewState"
          class="mt-4"
          color="neutral"
          variant="outline"
          label="清除筛选"
          @click="clearFilters"
        /><UButton v-else to="/submit" class="mt-4" label="投稿图片" />
      </template>
    </GalleryCompactEmpty>

    <nav
      v-if="galleryPageCount(data) > 1"
      class="mt-8 flex items-center justify-center gap-3"
      aria-label="投稿记录分页"
    >
      <UButton
        color="neutral"
        variant="outline"
        icon="i-tabler-arrow-left"
        label="上一页"
        :disabled="page <= 1"
        @click="setPage(page - 1)"
      />
      <span class="text-sm tabular-nums text-muted"
        >{{ page }} / {{ galleryPageCount(data) }}</span
      >
      <UButton
        color="neutral"
        variant="outline"
        trailing-icon="i-tabler-arrow-right"
        label="下一页"
        :disabled="page >= galleryPageCount(data)"
        @click="setPage(page + 1)"
      />
    </nav>
  </GalleryPublicPage>
</template>
