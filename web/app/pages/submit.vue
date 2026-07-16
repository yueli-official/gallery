<script setup lang="ts">
import { createPlatformNotifier } from "@platform/ui/feedback";
import type {
  GallerySubmission,
  GallerySubmissionOptions,
  GalleryUploadedAsset,
} from "~/types/gallery";

const { loggedIn, login } = useAuth();
const { call } = useApi();
const { upload } = useGalleryAssetUpload();
const toast = createPlatformNotifier(useToast());
const { data: submissionOptions } = await useFetch<GallerySubmissionOptions>(
  "/api/gallery/submission-options",
);
const file = ref<File>();
const previewUrl = ref("");
const title = ref("");
const description = ref("");
const sourceUrl = ref("");
const primaryCategoryId = ref("");
const sceneValueIds = ref<string[]>([]);
const tags = ref("");
const progress = ref(0);
const pending = ref(false);
const completed = ref<GallerySubmission>();

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
const accepted =
  ".jpg,.jpeg,.png,.webp,.avif,.heic,.heif,image/jpeg,image/png,image/webp,image/avif,image/heic,image/heif";

useSeoMeta({
  title: "投稿一张图片",
  description: "提交一张静态图片；来源地址可以稍后补充。",
  robots: "noindex,follow",
});

function chooseFile(event: Event) {
  const next = (event.target as HTMLInputElement).files?.[0];
  if (!next) return;
  if (next.size > 20 * 1024 * 1024) {
    toast.add({ title: "文件超过 20 MiB", color: "error" });
    return;
  }
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value);
  file.value = next;
  previewUrl.value = URL.createObjectURL(next);
  if (!title.value) title.value = next.name.replace(/\.[^.]+$/, "");
}

async function submit() {
  if (!loggedIn.value) {
    await login();
    return;
  }
  if (
    !file.value ||
    !title.value.trim() ||
    !primaryCategoryId.value ||
    !sceneValueIds.value.length ||
    pending.value
  )
    return;
  pending.value = true;
  progress.value = 1;
  try {
    const asset = (await upload(file.value, (value) => {
      progress.value = value;
    })) as GalleryUploadedAsset;
    const response = await call<{ submission: GallerySubmission }>(
      "/api/v1/gallery/submissions",
      {
        method: "POST",
        body: {
          assetId: asset.id,
          title: title.value.trim(),
          description: description.value.trim(),
          sourceUrl: sourceUrl.value.trim(),
          altText: title.value.trim(),
          categoryIds: [primaryCategoryId.value],
          primaryCategoryId: primaryCategoryId.value,
          tags: tags.value
            .split(/[,，]/)
            .map((item) => item.trim())
            .filter(Boolean),
          facets: sceneFacet.value
            ? [{ facetId: sceneFacet.value.id, valueIds: sceneValueIds.value }]
            : [],
        },
      },
    );
    completed.value = response.submission;
    progress.value = 100;
  } catch (reason: any) {
    toast.add({
      title: "投稿没有完成",
      description: reason?.data?.message || reason?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    pending.value = false;
  }
}

onBeforeUnmount(() => {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value);
});
</script>

<template>
  <div class="gallery-page max-w-7xl">
    <header class="gallery-page-header max-w-4xl">
      <div>
        <p class="gallery-eyebrow">单张投稿</p>
        <h1 class="gallery-page-title mt-3">把一张好图放进图库</h1>
        <p class="gallery-page-copy">
          不需要创作者身份。图片通过处理与安全检查后即可展示。
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
      description="匿名投稿功能尚在接入中，目前不会创建临时身份。"
      ><template #actions
        ><UButton label="登录" @click="void login()" /></template
    ></UAlert>

    <div v-if="completed" class="gallery-submit-complete">
      <UIcon name="i-tabler-circle-check" class="size-8 text-success" />
      <h2 class="mt-4 text-xl font-semibold text-highlighted">
        投稿已进入处理队列
      </h2>
      <p class="mt-2 text-sm leading-6 text-muted">
        图片会先做格式、动画、安全和衍生版本检查。明确安全后按你的登录状态决定是否需要人工审核。
      </p>
      <div class="mt-5 flex flex-wrap gap-2">
        <UButton to="/submissions" label="查看我的投稿" /><UButton
          to="/"
          color="neutral"
          variant="outline"
          label="继续浏览"
        />
      </div>
    </div>

    <form
      v-else
      class="grid items-start gap-6 lg:grid-cols-[minmax(0,1.3fr)_minmax(20rem,.7fr)] xl:gap-9"
      @submit.prevent="submit"
    >
      <div>
        <label
          class="gallery-upload-field group focus-within:outline-2 focus-within:outline-offset-4 focus-within:outline-primary"
        >
          <input
            type="file"
            class="sr-only"
            :accept="accepted"
            :disabled="pending"
            @change="chooseFile"
          />
          <img
            v-if="previewUrl"
            :src="previewUrl"
            alt="待投稿图片预览"
            class="max-h-[65vh] max-w-full object-contain"
          />
          <span v-else class="max-w-sm px-6 text-center"
            ><UIcon
              name="i-tabler-photo-up"
              class="gallery-upload-icon mx-auto size-10"
            /><span class="mt-4 block font-semibold text-highlighted"
              >点击选择一张静态图片</span
            ><span class="mt-2 block text-sm leading-6 text-muted"
              >支持 JPEG、PNG、WebP、AVIF、HEIC/HEIF，最大 20 MiB。</span
            ></span
          >
        </label>
        <UProgress v-if="pending" class="mt-3" :model-value="progress" />
      </div>

      <div class="gallery-submit-form space-y-5">
        <div>
          <h2 class="text-lg font-semibold text-highlighted">图片信息</h2>
          <p class="mt-1 text-xs leading-5 text-muted">
            标题和分类会帮助其他人找到这张图片。
          </p>
        </div>
        <UFormField label="标题" required
          ><UInput
            v-model="title"
            maxlength="160"
            placeholder="简洁描述这张图片"
        /></UFormField>
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
            :rows="5"
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
          label="提交投稿"
          :loading="pending"
          :disabled="
            !loggedIn ||
            !file ||
            !title.trim() ||
            !primaryCategoryId ||
            !sceneValueIds.length
          "
        />
        <p class="text-xs leading-5 text-muted">
          投稿者不会显示在公开页面。已发布图片不能替换文件，如需更换请重新投稿。
        </p>
      </div>
    </form>
  </div>
</template>
