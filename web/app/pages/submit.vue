<script setup lang="ts">
import { createPlatformNotifier } from "@platform/ui/feedback";
import type {
  GallerySubmission,
  GallerySubmissionOptions,
  GalleryUploadedAsset,
} from "~/types/gallery";
import {
  GALLERY_UPLOAD_BATCH_LIMIT,
  galleryUploadAnimationError,
  galleryUploadFileError,
} from "~/utils/galleryUpload";

type QueueStatus =
  "ready" | "uploading" | "submitting" | "completed" | "failed";
interface QueueItem {
  id: string;
  file: File;
  previewUrl: string;
  title: string;
  status: QueueStatus;
  progress: number;
  error: string;
  assetId: string;
  submission?: GallerySubmission;
}

const { loggedIn, login } = useAuth();
const { call } = useApi();
const { upload } = useGalleryAssetUpload();
const toast = createPlatformNotifier(useToast());
const { data: submissionOptions } = await useFetch<GallerySubmissionOptions>(
  "/api/gallery/submission-options",
);
const queue = ref<QueueItem[]>([]);
const description = ref("");
const sourceUrl = ref("");
const primaryCategoryId = ref("");
const sceneValueIds = ref<string[]>([]);
const tags = ref("");
const sharedTitleEnabled = ref(false);
const sharedTitle = ref("");
const running = ref(false);

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
const failedCount = computed(
  () => queue.value.filter((item) => item.status === "failed").length,
);
const readyCount = computed(
  () => queue.value.filter((item) => item.status === "ready").length,
);
const validMetadata = computed(() =>
  Boolean(
    primaryCategoryId.value &&
    sceneValueIds.value.length &&
    (sharedTitleEnabled.value
      ? sharedTitle.value.trim()
      : queue.value.every((item) => item.title.trim())),
  ),
);
const accepted =
  ".jpg,.jpeg,.png,.webp,.avif,.heic,.heif,image/jpeg,image/png,image/webp,image/avif,image/heic,image/heif";

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
      title: file.name.replace(/\.[^.]+$/, ""),
      status: "ready",
      progress: 0,
      error: "",
      assetId: "",
    });
  }
}

function removeItem(id: string) {
  const index = queue.value.findIndex((item) => item.id === id);
  if (index < 0) return;
  URL.revokeObjectURL(queue.value[index]!.previewUrl);
  queue.value.splice(index, 1);
}

function submissionBody(item: QueueItem) {
  const resolvedTitle = sharedTitleEnabled.value
    ? sharedTitle.value.trim()
    : item.title.trim();
  return {
    assetId: item.assetId,
    title: resolvedTitle,
    description: description.value.trim(),
    sourceUrl: sourceUrl.value.trim(),
    altText: resolvedTitle,
    categoryIds: [primaryCategoryId.value],
    primaryCategoryId: primaryCategoryId.value,
    tags: tags.value
      .split(/[,，]/)
      .map((value) => value.trim())
      .filter(Boolean),
    facets: sceneFacet.value
      ? [{ facetId: sceneFacet.value.id, valueIds: sceneValueIds.value }]
      : [],
  };
}

async function submitItem(item: QueueItem) {
  item.error = "";
  try {
    if (!item.assetId) {
      item.status = "uploading";
      item.progress = 1;
      const asset = (await upload(item.file, (value) => {
        item.progress = value;
      })) as GalleryUploadedAsset;
      item.assetId = asset.id;
    }
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
  } catch (reason: any) {
    item.status = "failed";
    item.error = reason?.data?.message || reason?.message || "请稍后重试";
  }
}

async function run(items: QueueItem[]) {
  if (!loggedIn.value) {
    await login();
    return;
  }
  if (running.value || !validMetadata.value || !items.length) return;
  running.value = true;
  for (const item of items) await submitItem(item);
  running.value = false;
  if (failedCount.value) {
    toast.add({
      title: "部分投稿没有完成",
      description: `${failedCount.value} 张可保留现场后重试`,
      color: "warning",
    });
  }
}

function submitReady() {
  return run(queue.value.filter((item) => item.status === "ready"));
}
function retryFailed() {
  return run(queue.value.filter((item) => item.status === "failed"));
}

onBeforeUnmount(() =>
  queue.value.forEach((item) => URL.revokeObjectURL(item.previewUrl)),
);
</script>

<template>
  <div class="gallery-page max-w-7xl">
    <header class="gallery-page-header max-w-4xl">
      <div>
        <p class="gallery-eyebrow">批量投稿</p>
        <h1 class="gallery-page-title mt-3">把一组好图放进图库</h1>
        <p class="gallery-page-copy">
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
      icon="i-tabler-login-2"
      title="登录后即可投稿"
      description="登录状态用于保存投稿进度和处理结果。"
    >
      <template #actions
        ><UButton label="登录" @click="void login()"
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
              :accept="accepted"
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
          class="gallery-upload-field group min-h-[26rem] focus-within:outline-2 focus-within:outline-offset-4 focus-within:outline-primary"
        >
          <input
            type="file"
            class="sr-only"
            multiple
            :accept="accepted"
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
              >支持 JPEG、PNG、WebP、AVIF、HEIC/HEIF，单张最大 20
              MiB。动画与重复文件会在上传前拦截。</span
            >
          </span>
        </label>

        <div v-else class="grid gap-3 sm:grid-cols-2">
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
                      submitting: '创建记录',
                      completed: '已完成',
                      failed: '失败',
                    } as const
                  )[item.status]
                "
              />
              <UButton
                v-if="!['uploading', 'submitting'].includes(item.status)"
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
                v-if="sharedTitleEnabled"
                :model-value="sharedTitle"
                aria-label="统一图片标题"
                disabled
              />
              <UInput
                v-else
                v-model="item.title"
                maxlength="160"
                aria-label="图片标题"
                :disabled="item.status === 'completed'"
              />
              <UProgress
                v-if="item.status === 'uploading'"
                size="xs"
                :model-value="item.progress"
              />
              <p v-if="item.error" class="text-xs leading-5 text-error">
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
          <h2 class="text-lg font-semibold text-highlighted">
            这一批的共同信息
          </h2>
          <p class="mt-1 text-xs leading-5 text-muted">
            分类、场景和说明会用于队列中的每一张图片，标题可以逐张修改。
          </p>
        </div>
        <UCheckbox
          v-model="sharedTitleEnabled"
          label="这一批使用统一标题"
          description="关闭时可在左侧逐张修改标题"
        />
        <UFormField v-if="sharedTitleEnabled" label="统一标题" required>
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
        <UButton
          type="submit"
          block
          size="lg"
          icon="i-tabler-send"
          :label="readyCount > 1 ? `投稿 ${readyCount} 张图片` : '开始投稿'"
          :loading="running"
          :disabled="!loggedIn || !readyCount || !validMetadata"
        />
        <UButton
          v-if="failedCount"
          type="button"
          block
          color="error"
          variant="soft"
          icon="i-tabler-refresh"
          :label="`重试失败的 ${failedCount} 张`"
          :disabled="running || !validMetadata"
          @click="retryFailed"
        />
        <UButton
          v-if="completedCount && !readyCount && !failedCount"
          to="/submissions"
          block
          color="neutral"
          variant="outline"
          label="查看投稿记录"
        />
        <p class="text-xs leading-5 text-muted">
          投稿者不会显示在公开页面。已发布图片不能替换文件，如需更换请重新投稿。
        </p>
      </aside>
    </form>
  </div>
</template>
