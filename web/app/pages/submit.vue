<script setup lang="ts">
import { createGalleryNotifier } from "~/utils/feedback";
import type {
  GallerySubmission,
  GallerySubmissionOptions,
  GalleryUploadedAsset,
} from "~/types/gallery";
import {
  GALLERY_UPLOAD_ACCEPT,
  GALLERY_UPLOAD_BATCH_LIMIT,
  GALLERY_UPLOAD_FORMAT_LABEL,
  galleryUploadAnimationError,
  galleryUploadFileError,
} from "~/utils/galleryUpload";
import {
  applyGallerySubmissionDefaults,
  gallerySubmissionMetadataValid,
} from "~/utils/gallerySubmissionBatch";
import { gallerySubmissionFailure } from "~/utils/galleryAssetReadiness";

type QueueStatus =
  | "ready"
  | "uploading"
  | "checking"
  | "submitting"
  | "completed"
  | "duplicate"
  | "failed";
interface QueueItem {
  id: string;
  file: File;
  previewUrl: string;
  title: string;
  description: string;
  sourceUrl: string;
  primaryCategoryId: string;
  sceneValueIds: string[];
  tags: string;
  customized: boolean;
  status: QueueStatus;
  progress: number;
  error: string;
  assetId: string;
  submission?: GallerySubmission;
}

const { loggedIn, login } = useAuth();
const { call } = useApi();
const { upload, waitUntilReady } = useGalleryAssetUpload();
const toast = createGalleryNotifier(useToast());
const { data: submissionOptions } = await useFetch<GallerySubmissionOptions>(
  "/api/gallery/submission-options",
);
const queue = ref<QueueItem[]>([]);
const description = ref("");
const sourceUrl = ref("");
const primaryCategoryId = ref("");
const sceneValueIds = ref<string[]>([]);
const tags = ref("");
const sharedTitle = ref("");
const running = ref(false);
const runningAction = ref<"submit" | "retry" | "">("");
const editingId = ref("");

const categoryItems = computed(() =>
  (submissionOptions.value?.categories || []).map((item) => ({
    label: item.name,
    value: item.id,
  })),
);
const sceneFacet = computed(() =>
  submissionOptions.value?.facets.find((item) => item.slug === "scene"),
);
const sceneItems = computed(() =>
  (sceneFacet.value?.values || []).map((item) => ({
    label: item.name,
    value: item.id,
  })),
);
const completedCount = computed(
  () => queue.value.filter((item) => item.status === "completed").length,
);
const duplicateCount = computed(
  () => queue.value.filter((item) => item.status === "duplicate").length,
);
const failedCount = computed(
  () => queue.value.filter((item) => item.status === "failed").length,
);
const readyCount = computed(
  () => queue.value.filter((item) => item.status === "ready").length,
);
const validMetadata = computed(() =>
  queue.value
    .filter((item) => !isTerminalStatus(item.status))
    .every(gallerySubmissionMetadataValid),
);
useSeoMeta({
  title: "批量投稿图片",
  description: "一次提交多张静态图片，每张图片独立处理，失败项目可以单独重试。",
  robots: "noindex,follow",
});

async function chooseFiles(event: Event) {
  const input = event.target as HTMLInputElement;
  const files = [...(input.files || [])];
  input.value = "";
  const capacity = GALLERY_UPLOAD_BATCH_LIMIT - queue.value.length;
  if (files.length > capacity) {
    toast.add({
      title: `每批最多 ${GALLERY_UPLOAD_BATCH_LIMIT} 张`,
      description: `本次只加入前 ${Math.max(0, capacity)} 张`,
      color: "warning",
    });
  }
  for (const file of files.slice(0, Math.max(0, capacity))) {
    const duplicate = queue.value.some(
      (item) =>
        item.file.name === file.name &&
        item.file.size === file.size &&
        item.file.lastModified === file.lastModified,
    );
    if (duplicate) continue;
    const error =
      galleryUploadFileError(file) || (await galleryUploadAnimationError(file));
    if (error) {
      toast.add({ title: file.name, description: error, color: "error" });
      continue;
    }
    queue.value.push({
      id: crypto.randomUUID(),
      file,
      previewUrl: URL.createObjectURL(file),
      title: sharedTitle.value.trim() || file.name.replace(/\.[^.]+$/, ""),
      description: description.value,
      sourceUrl: sourceUrl.value,
      primaryCategoryId: primaryCategoryId.value,
      sceneValueIds: [...sceneValueIds.value],
      tags: tags.value,
      customized: false,
      status: "ready",
      progress: 0,
      error: "",
      assetId: "",
    });
  }
}

function applyDefaultsToItem(item: QueueItem, preserveTitle = false) {
  if (isTerminalStatus(item.status)) return;
  applyGallerySubmissionDefaults(
    item,
    {
      title: sharedTitle.value,
      description: description.value,
      sourceUrl: sourceUrl.value,
      primaryCategoryId: primaryCategoryId.value,
      sceneValueIds: sceneValueIds.value,
      tags: tags.value,
    },
    preserveTitle,
  );
}

function applyDefaults() {
  queue.value.forEach((item) => applyDefaultsToItem(item));
}

function isTerminalStatus(status: QueueStatus) {
  return status === "completed" || status === "duplicate";
}

function restoreDefaults(item: QueueItem) {
  applyDefaultsToItem(item);
}

function markCustomized(item: QueueItem) {
  if (!isTerminalStatus(item.status)) item.customized = true;
}

function toggleItemEditor(item: QueueItem) {
  editingId.value = editingId.value === item.id ? "" : item.id;
}

function removeItem(id: string) {
  const index = queue.value.findIndex((item) => item.id === id);
  if (index < 0) return;
  URL.revokeObjectURL(queue.value[index]!.previewUrl);
  queue.value.splice(index, 1);
}

function submissionBody(item: QueueItem) {
  const resolvedTitle = item.title.trim();
  return {
    assetId: item.assetId,
    title: resolvedTitle,
    description: item.description.trim(),
    sourceUrl: item.sourceUrl.trim(),
    altText: resolvedTitle,
    categoryIds: [item.primaryCategoryId],
    primaryCategoryId: item.primaryCategoryId,
    tags: item.tags
      .split(/[,，]/)
      .map((value) => value.trim())
      .filter(Boolean),
    facets: sceneFacet.value
      ? [{ facetId: sceneFacet.value.id, valueIds: item.sceneValueIds }]
      : [],
  };
}

async function submitItem(item: QueueItem) {
  item.error = "";
  try {
    let uploaded: GalleryUploadedAsset | undefined;
    if (!item.assetId) {
      item.status = "uploading";
      item.progress = 1;
      uploaded = (await upload(item.file, (value) => {
        item.progress = value;
      })) as GalleryUploadedAsset;
      item.assetId = uploaded.id;
    }
    item.status = "checking";
    await waitUntilReady(item.assetId, uploaded);
    item.status = "submitting";
    const response = await call<{ submission: GallerySubmission }>(
      "/api/v1/gallery/submissions",
      {
        method: "POST",
        body: submissionBody(item),
      },
    );
    item.submission = response.submission;
    item.status = "completed";
    item.progress = 100;
  } catch (reason: unknown) {
    const failure = gallerySubmissionFailure(reason);
    item.status =
      failure.kind === "already-submitted" ? "duplicate" : "failed";
    item.error = failure.message;
    if (item.status === "duplicate") item.progress = 100;
  }
}

async function run(items: QueueItem[], action: "submit" | "retry") {
  if (running.value || !validMetadata.value || !items.length) return;
  running.value = true;
  runningAction.value = action;
  for (const item of items) await submitItem(item);
  running.value = false;
  runningAction.value = "";
  if (failedCount.value) {
    toast.add({
      title: "部分投稿没有完成",
      description: `${failedCount.value} 张可保留现场后重试`,
      color: "warning",
    });
  }
}

function submitReady() {
  return run(
    queue.value.filter((item) => item.status === "ready"),
    "submit",
  );
}
function retryFailed() {
  return run(
    queue.value.filter((item) => item.status === "failed"),
    "retry",
  );
}

onBeforeUnmount(() =>
  queue.value.forEach((item) => URL.revokeObjectURL(item.previewUrl)),
);
</script>

<template>
  <div class="gallery-page max-w-7xl">
    <header class="mb-8 max-w-3xl">
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-highlighted">
          批量投稿图片
        </h1>
        <p class="mt-2 max-w-2xl text-sm leading-6 text-muted">
          一次最多 20
          张。每张图片独立上传和处理，部分失败不会影响已经完成的项目。
        </p>
      </div>
    </header>

    <UAlert
      v-if="!loggedIn"
      class="gallery-submit-notice mb-6"
      color="primary"
      variant="subtle"
      icon="i-tabler-shield-check"
      title="可以直接匿名投稿"
      description="浏览器会获得一个 30 天临时投稿身份；登录后可继续管理这些投稿。"
    >
      <template #actions
        ><UButton label="登录后投稿" @click="void login()"
      /></template>
    </UAlert>

    <form
      class="grid items-start gap-6 lg:grid-cols-[minmax(0,1.25fr)_minmax(20rem,.75fr)] xl:gap-9"
      @submit.prevent="submitReady"
    >
      <section class="space-y-4" aria-labelledby="upload-queue-title">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h2
              id="upload-queue-title"
              class="text-lg font-semibold text-highlighted"
            >
              文件队列
            </h2>
            <p class="mt-1 text-sm text-muted">
              {{ queue.length }} / {{ GALLERY_UPLOAD_BATCH_LIMIT }} 张<span
                v-if="completedCount"
                >，已完成 {{ completedCount }} 张</span
              >
            </p>
          </div>
          <label class="inline-flex cursor-pointer">
            <input
              type="file"
              class="sr-only"
              multiple
              :accept="GALLERY_UPLOAD_ACCEPT"
              :disabled="running || queue.length >= GALLERY_UPLOAD_BATCH_LIMIT"
              @change="chooseFiles"
            />
            <span
              class="inline-flex h-9 items-center gap-2 rounded-md border border-default bg-default px-3 text-sm font-medium text-highlighted transition-colors hover:bg-elevated"
              ><UIcon name="i-tabler-plus" class="size-4" />选择图片</span
            >
          </label>
        </div>

        <label
          v-if="!queue.length"
          class="gallery-upload-field group min-h-[18rem] focus-within:outline-2 focus-within:outline-offset-4 focus-within:outline-primary sm:min-h-[22rem]"
        >
          <input
            type="file"
            class="sr-only"
            multiple
            :accept="GALLERY_UPLOAD_ACCEPT"
            @change="chooseFiles"
          />
          <span class="max-w-md px-6 text-center">
            <UIcon
              name="i-tabler-library-plus"
              class="gallery-upload-icon mx-auto size-11"
            />
            <span class="mt-4 block font-semibold text-highlighted"
              >选择一组静态图片</span
            >
            <span class="mt-2 block text-sm leading-6 text-muted"
              >支持 {{ GALLERY_UPLOAD_FORMAT_LABEL }}，单张最大 20
              MiB。动画与重复文件会在上传前拦截。</span
            >
          </span>
        </label>

        <div v-else class="grid gap-3 xl:grid-cols-2">
          <article
            v-for="item in queue"
            :key="item.id"
            class="overflow-hidden rounded-xl border border-default bg-default shadow-sm"
          >
            <div class="relative aspect-[16/10] bg-elevated">
              <img
                :src="item.previewUrl"
                :alt="item.title"
                class="size-full object-cover"
              />
              <UBadge
                class="absolute left-3 top-3"
                :color="
                  item.status === 'completed'
                    ? 'success'
                    : item.status === 'duplicate'
                      ? 'success'
                    : item.status === 'failed'
                      ? 'error'
                      : 'neutral'
                "
                variant="solid"
                :label="
                  (
                    {
                      ready: '待上传',
                      uploading: '上传中',
                      checking: '安全检查',
                      submitting: '创建记录',
                      completed: '已完成',
                      duplicate: '已投稿',
                      failed: '失败',
                    } as const
                  )[item.status]
                "
              />
              <UButton
                v-if="
                  !['uploading', 'checking', 'submitting'].includes(item.status)
                "
                class="absolute right-2 top-2"
                color="neutral"
                variant="solid"
                size="xs"
                icon="i-tabler-x"
                aria-label="移除文件"
                @click="removeItem(item.id)"
              />
            </div>
            <div class="space-y-2 p-3">
              <UInput
                v-model="item.title"
                maxlength="160"
                aria-label="图片标题"
                :disabled="isTerminalStatus(item.status)"
                @update:model-value="markCustomized(item)"
              />
              <div class="flex items-center justify-between gap-2">
                <UBadge
                  v-if="item.customized"
                  color="primary"
                  variant="subtle"
                  label="已单独修改"
                />
                <span v-else class="text-xs text-dimmed">使用批量默认值</span>
                <UButton
                  v-if="!isTerminalStatus(item.status)"
                  type="button"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  :label="editingId === item.id ? '收起' : '单独编辑'"
                  :trailing-icon="
                    editingId === item.id
                      ? 'i-tabler-chevron-up'
                      : 'i-tabler-chevron-down'
                  "
                  @click="toggleItemEditor(item)"
                />
              </div>
              <div
                v-if="editingId === item.id"
                class="space-y-3 border-t border-default pt-3"
              >
                <UFormField label="主分类" required
                  ><USelect
                    v-model="item.primaryCategoryId"
                    :items="categoryItems"
                    value-key="value"
                    placeholder="选择分类"
                    class="w-full"
                    @update:model-value="markCustomized(item)"
                /></UFormField>
                <UFormField label="场景" required
                  ><USelect
                    v-model="item.sceneValueIds"
                    :items="sceneItems"
                    value-key="value"
                    multiple
                    placeholder="选择场景"
                    class="w-full"
                    @update:model-value="markCustomized(item)"
                /></UFormField>
                <details class="rounded-lg border border-default p-3">
                  <summary
                    class="cursor-pointer text-sm font-medium text-default"
                  >
                    更多信息
                  </summary>
                  <div class="mt-3 space-y-3">
                    <UFormField label="说明"
                      ><UTextarea
                        v-model="item.description"
                        :rows="3"
                        class="w-full"
                        @update:model-value="markCustomized(item)"
                    /></UFormField>
                    <UFormField label="来源地址"
                      ><UInput
                        v-model="item.sourceUrl"
                        type="url"
                        placeholder="https://"
                        class="w-full"
                        @update:model-value="markCustomized(item)"
                    /></UFormField>
                    <UFormField label="标签"
                      ><UInput
                        v-model="item.tags"
                        placeholder="夜景, 蓝色, 雨"
                        class="w-full"
                        @update:model-value="markCustomized(item)"
                    /></UFormField>
                  </div>
                </details>
                <UButton
                  v-if="item.customized"
                  type="button"
                  color="neutral"
                  variant="outline"
                  size="sm"
                  icon="i-tabler-restore"
                  label="恢复批量默认值"
                  @click="restoreDefaults(item)"
                />
              </div>
              <UProgress
                v-if="item.status === 'uploading'"
                size="xs"
                :model-value="item.progress"
              />
              <p
                v-if="item.error"
                class="text-xs leading-5"
                :class="
                  item.status === 'duplicate' ? 'text-success' : 'text-error'
                "
              >
                {{ item.error }}
              </p>
              <p v-else class="truncate text-xs text-muted">
                {{ item.file.name }} ·
                {{ (item.file.size / 1024 / 1024).toFixed(1) }} MiB
              </p>
            </div>
          </article>
        </div>
      </section>

      <aside class="gallery-submit-form space-y-5 lg:sticky lg:top-24">
        <div>
          <h2 class="text-lg font-semibold text-highlighted">批量默认值</h2>
          <p class="mt-1 text-xs leading-5 text-muted">
            设置后点击“应用到队列”；之后仍可逐张覆盖，已完成项目不会被改动。
          </p>
        </div>
        <UFormField label="统一标题" hint="可选；留空时保留各文件标题">
          <UInput
            v-model="sharedTitle"
            maxlength="160"
            placeholder="例如：城市夜景系列"
          />
        </UFormField>
        <UFormField label="主分类" required hint="用于主要浏览入口"
          ><USelect
            v-model="primaryCategoryId"
            :items="categoryItems"
            value-key="value"
            placeholder="选择壁纸、插画或摄影"
        /></UFormField>
        <UFormField label="场景" required hint="可选择 1-2 项"
          ><USelect
            v-model="sceneValueIds"
            :items="sceneItems"
            value-key="value"
            multiple
            placeholder="选择场景"
        /></UFormField>
        <details class="rounded-lg border border-default p-3">
          <summary class="cursor-pointer text-sm font-medium text-default">
            更多共同信息
          </summary>
          <div class="mt-4 space-y-4">
            <UFormField label="说明" hint="可选"
              ><UTextarea
                v-model="description"
                :rows="4"
                placeholder="补充画面、背景或整理说明"
            /></UFormField>
            <UFormField label="来源地址" hint="可选，不知道可以留空"
              ><UInput v-model="sourceUrl" type="url" placeholder="https://"
            /></UFormField>
            <UFormField label="标签" hint="可选，用逗号分隔"
              ><UInput v-model="tags" placeholder="夜景, 蓝色, 雨"
            /></UFormField>
          </div>
        </details>
        <UButton
          type="button"
          block
          color="primary"
          variant="soft"
          icon="i-tabler-copy-check"
          :label="`应用到队列中的 ${queue.filter((item) => !isTerminalStatus(item.status)).length} 张`"
          :disabled="!queue.some((item) => !isTerminalStatus(item.status))"
          @click="applyDefaults"
        />
        <div class="grid gap-3 border-t border-default pt-5">
          <UButton
            v-if="readyCount || runningAction === 'submit'"
            type="submit"
            block
            size="lg"
            icon="i-tabler-send"
            :label="readyCount > 1 ? `投稿 ${readyCount} 张图片` : '开始投稿'"
            :loading="runningAction === 'submit'"
            :disabled="!validMetadata"
          />
          <UButton
            v-if="failedCount || runningAction === 'retry'"
            type="button"
            block
            color="error"
            variant="soft"
            icon="i-tabler-refresh"
            :label="
              runningAction === 'retry'
                ? '正在重试'
                : `重试失败的 ${failedCount} 张`
            "
            :loading="runningAction === 'retry'"
            :disabled="running || !validMetadata"
            @click="retryFailed"
          />
          <UButton
            v-if="
              (completedCount || duplicateCount) && !readyCount && !failedCount
            "
            to="/submissions"
            block
            color="neutral"
            variant="outline"
            label="查看投稿记录"
          />
        </div>
        <p class="text-xs leading-5 text-muted">
          投稿者不会显示在公开页面。已发布图片不能替换文件，如需更换请重新投稿。
        </p>
      </aside>
    </form>
  </div>
</template>
