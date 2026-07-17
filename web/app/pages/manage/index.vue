<script setup lang="ts">
import { ManageHeader } from "@platform/manage/components";
import type { GalleryAdminOverview } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "今日运营 · 图库管理" });

const { user, isAdmin } = useAuth();
const { call } = useApi();
const hydrated = useClientHydrated();
const { data, pending, error, refresh } = await useAsyncData(
  "gallery-admin-overview",
  () =>
    call<{ overview: GalleryAdminOverview }>("/api/v1/gallery/admin/overview"),
  { server: false },
);
const overview = computed(() => data.value?.overview);
const attentionTotal = computed(
  () =>
    (overview.value?.pendingSubmissions || 0) +
    (overview.value?.openCases || 0) +
    (overview.value?.failedProcessing || 0),
);
const signals = computed(() => [
  {
    label: "待审投稿",
    value: overview.value?.pendingSubmissions || 0,
    tone: "primary",
  },
  {
    label: "开放处理单",
    value: overview.value?.openCases || 0,
    tone: "warning",
  },
  {
    label: "处理失败",
    value: overview.value?.failedProcessing || 0,
    tone: "error",
  },
  {
    label: "公开图片",
    value: overview.value?.publishedImages || 0,
    tone: "neutral",
  },
]);
const priorities = computed(() => [
  {
    key: "submissions",
    label: "审核新投稿",
    description: "确认内容质量、分类与安全判断，让合格图片进入公开目录。",
    count: overview.value?.pendingSubmissions || 0,
    icon: "i-tabler-photo-check",
    to: "/manage/submissions?reviewState=pending&outcome=pending",
    action: "进入审核",
    tone: "primary",
  },
  {
    key: "cases",
    label: "跟进信任问题",
    description: "处理举报、来源修正和安全不确定，优先解决等待时间最长的项目。",
    count: overview.value?.openCases || 0,
    icon: "i-tabler-shield-check",
    to: "/manage/cases",
    action: "查看处理单",
    tone: "warning",
  },
  {
    key: "failed",
    label: "恢复失败处理",
    description: "检查媒体处理失败原因，保留已有投稿信息并重新推进。",
    count: overview.value?.failedProcessing || 0,
    icon: "i-tabler-alert-triangle",
    to: "/manage/submissions?processingState=failed",
    action: "检查失败项",
    tone: "error",
  },
]);
const workspaces = [
  {
    label: "图片资产",
    description: "浏览完整生命周期与异常状态",
    icon: "i-tabler-photo",
    to: "/manage/images",
  },
  {
    label: "专题策展",
    description: "组织封面、成员与公开叙事",
    icon: "i-tabler-folders",
    to: "/manage/collections",
  },
  {
    label: "目录治理",
    description: "维护分类、维度和标签质量",
    icon: "i-tabler-category",
    to: "/manage/classification",
  },
  {
    label: "发现策略",
    description: "检查随机发现与排行信号",
    icon: "i-tabler-sparkles",
    to: "/manage/discovery",
  },
];
</script>

<template>
  <div>
    <ManageHeader title="今日运营">
      <template #subtitle
        >{{
          user?.name || user?.email || "运营者"
        }}，先处理会阻塞发布或影响信任的工作。</template
      >
    </ManageHeader>

    <div v-if="!hydrated || pending" class="space-y-6">
      <USkeleton class="h-24 rounded-xl" />
      <div class="grid gap-6 lg:grid-cols-[minmax(0,1.5fr)_minmax(18rem,.5fr)]">
        <USkeleton class="h-96 rounded-xl" /><USkeleton
          class="h-96 rounded-xl"
        />
      </div>
    </div>

    <UAlert
      v-else-if="!isAdmin"
      color="error"
      variant="subtle"
      icon="i-tabler-lock"
      title="当前账号没有图库管理权限"
    />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      icon="i-tabler-alert-circle"
      title="运营数据暂时不可用"
      description="页面没有丢失你的操作，可以稍后重试。"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>

    <template v-else>
      <section
        aria-label="今日运营信号"
        class="mb-7 grid grid-cols-2 overflow-hidden rounded-xl border border-default bg-default lg:grid-cols-4"
      >
        <NuxtLink
          v-for="(signal, index) in signals"
          :key="signal.label"
          :to="
            index === 3
              ? '/manage/images?status=published'
              : priorities[index]?.to
          "
          class="group min-w-0 border-default px-4 py-4 transition-colors hover:bg-elevated/60 lg:px-5"
          :class="[
            index % 2 ? '' : 'border-r',
            index > 1 ? 'border-t lg:border-t-0' : '',
            index > 0 ? 'lg:border-l' : '',
          ]"
        >
          <p class="text-xs font-medium text-muted">{{ signal.label }}</p>
          <div class="mt-2 flex items-end justify-between gap-2">
            <strong
              class="font-display text-2xl font-semibold tabular-nums text-highlighted"
              >{{ signal.value }}</strong
            ><UIcon
              name="i-tabler-arrow-up-right"
              class="mb-1 size-4 text-dimmed transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5"
            />
          </div>
        </NuxtLink>
      </section>

      <div
        class="grid items-start gap-7 lg:grid-cols-[minmax(0,1.55fr)_minmax(19rem,.65fr)]"
      >
        <section aria-labelledby="priority-heading">
          <div class="mb-3 flex items-end justify-between gap-4">
            <div>
              <h2
                id="priority-heading"
                class="text-lg font-semibold text-highlighted"
              >
                需要你处理
              </h2>
              <p class="mt-1 text-sm text-muted">
                {{
                  attentionTotal
                    ? `共 ${attentionTotal} 项需要关注`
                    : "当前没有阻塞事项"
                }}
              </p>
            </div>
          </div>
          <div
            v-if="attentionTotal"
            class="overflow-hidden rounded-xl border border-default bg-default"
          >
            <article
              v-for="priority in priorities.filter((item) => item.count)"
              :key="priority.key"
              class="grid gap-4 border-b border-default p-4 last:border-b-0 sm:grid-cols-[auto_minmax(0,1fr)_auto] sm:items-center sm:p-5"
            >
              <span
                class="grid size-11 place-items-center rounded-lg"
                :class="
                  priority.tone === 'error'
                    ? 'bg-error/10 text-error'
                    : priority.tone === 'warning'
                      ? 'bg-warning/10 text-warning'
                      : 'bg-primary/10 text-primary'
                "
                ><UIcon :name="priority.icon" class="size-5"
              /></span>
              <div class="min-w-0">
                <div class="flex items-baseline gap-2">
                  <h3 class="font-semibold text-highlighted">
                    {{ priority.label }}
                  </h3>
                  <span
                    class="text-sm font-semibold tabular-nums"
                    :class="
                      priority.tone === 'error'
                        ? 'text-error'
                        : priority.tone === 'warning'
                          ? 'text-highlighted'
                          : 'text-primary'
                    "
                    >{{ priority.count }}</span
                  >
                </div>
                <p class="mt-1 max-w-2xl text-sm leading-6 text-muted">
                  {{ priority.description }}
                </p>
              </div>
              <UButton
                :to="priority.to"
                :label="priority.action"
                color="neutral"
                variant="outline"
                trailing-icon="i-tabler-arrow-right"
                class="justify-self-start sm:justify-self-end"
              />
            </article>
          </div>
          <div
            v-else
            class="rounded-xl border border-default bg-default px-5 py-10 text-center"
          >
            <span
              class="mx-auto grid size-11 place-items-center rounded-full bg-success/10 text-success"
              ><UIcon name="i-tabler-check" class="size-5"
            /></span>
            <h3 class="mt-3 font-semibold text-highlighted">队列已经清空</h3>
            <p class="mt-1 text-sm text-muted">
              可以继续整理图片、专题或目录。
            </p>
          </div>

          <div class="mt-8">
            <h2 class="text-lg font-semibold text-highlighted">继续工作</h2>
            <div
              class="mt-3 grid overflow-hidden rounded-xl border border-default bg-default sm:grid-cols-2"
            >
              <NuxtLink
                v-for="(workspace, index) in workspaces"
                :key="workspace.to"
                :to="workspace.to"
                class="group flex min-h-24 items-center gap-3 border-default px-4 py-3 transition-colors hover:bg-elevated/60"
                :class="[
                  index % 2 === 0 ? 'sm:border-r' : '',
                  index > 1
                    ? 'border-t'
                    : index === 1
                      ? 'border-t sm:border-t-0'
                      : '',
                ]"
                ><UIcon
                  :name="workspace.icon"
                  class="size-5 shrink-0 text-primary" /><span class="min-w-0"
                  ><strong
                    class="block text-sm font-semibold text-highlighted"
                    >{{ workspace.label }}</strong
                  ><span class="mt-1 block text-xs leading-5 text-muted">{{
                    workspace.description
                  }}</span></span
                ><UIcon
                  name="i-tabler-chevron-right"
                  class="ml-auto size-4 text-dimmed transition-transform group-hover:translate-x-0.5"
              /></NuxtLink>
            </div>
          </div>
        </section>

        <aside class="space-y-5">
          <section
            class="rounded-xl border border-default bg-default px-5 py-5"
          >
            <div class="flex items-center justify-between gap-3">
              <p class="text-sm font-semibold text-highlighted">图库运行正常</p>
              <span class="size-2 rounded-full bg-success" />
            </div>
            <p class="mt-2 text-sm leading-6 text-muted">
              API 可用，运营数据已同步。只有出现异常时，这里才会升级为处理入口。
            </p>
            <UButton
              to="/"
              target="_blank"
              class="mt-4"
              color="neutral"
              variant="outline"
              size="sm"
              icon="i-tabler-arrow-up-right"
              label="查看公开站点"
            />
          </section>
          <section class="rounded-xl border border-default bg-default p-5">
            <h2 class="text-sm font-semibold text-highlighted">工作原则</h2>
            <dl class="mt-4 space-y-4 text-sm">
              <div>
                <dt class="font-medium text-default">先解除阻塞</dt>
                <dd class="mt-1 leading-5 text-muted">
                  失败处理和待审投稿优先于内容整理。
                </dd>
              </div>
              <div>
                <dt class="font-medium text-default">再处理信任</dt>
                <dd class="mt-1 leading-5 text-muted">
                  举报、来源和安全问题保留完整上下文。
                </dd>
              </div>
              <div>
                <dt class="font-medium text-default">最后优化发现</dt>
                <dd class="mt-1 leading-5 text-muted">
                  专题、分类和排行不会覆盖内容真实性。
                </dd>
              </div>
            </dl>
          </section>
        </aside>
      </div>
    </template>
  </div>
</template>
