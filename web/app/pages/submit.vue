<script setup lang="ts">
import { createPlatformNotifier } from "@platform/ui/feedback";
import type { GalleryDiscovery, GallerySubmission, GalleryUploadedAsset } from "~/types/gallery";

const { loggedIn, login } = useAuth();
const { call } = useApi();
const { upload } = useGalleryAssetUpload();
const toast = createPlatformNotifier(useToast());
const { data: discovery } = await useFetch<GalleryDiscovery>("/api/gallery/discovery", { query: { seed: "submission-form" } });
const file = ref<File>();
const previewUrl = ref("");
const title = ref("");
const description = ref("");
const sourceUrl = ref("");
const topicId = ref("");
const tags = ref("");
const progress = ref(0);
const pending = ref(false);
const completed = ref<GallerySubmission>();

const topicFacet = computed(() => discovery.value?.facets.find(item => item.slug === "topic"));
const topicItems = computed(() => (discovery.value?.facetValues || []).filter(item => item.facetId === topicFacet.value?.id).map(item => ({ label: item.name, value: item.id })));
const accepted = ".jpg,.jpeg,.png,.webp,.avif,.heic,.heif,image/jpeg,image/png,image/webp,image/avif,image/heic,image/heif";

useSeoMeta({ title: "投稿一张图片", description: "提交一张静态图片；来源地址可以稍后补充。", robots: "noindex,follow" });

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
  if (!file.value || !title.value.trim() || !topicId.value || pending.value) return;
  pending.value = true;
  progress.value = 1;
  try {
    const asset = await upload(file.value, value => { progress.value = value; }) as GalleryUploadedAsset;
    const response = await call<{ submission: GallerySubmission }>("/api/v1/gallery/submissions", {
      method: "POST",
      body: {
        assetId: asset.id,
        title: title.value.trim(),
        description: description.value.trim(),
        sourceUrl: sourceUrl.value.trim(),
        altText: title.value.trim(),
        topicId: topicId.value,
        tags: tags.value.split(/[,，]/).map(item => item.trim()).filter(Boolean),
        facets: [],
      },
    });
    completed.value = response.submission;
    progress.value = 100;
  } catch (reason: any) {
    toast.add({ title: "投稿没有完成", description: reason?.data?.message || reason?.message || "请稍后重试", color: "error" });
  } finally {
    pending.value = false;
  }
}

onBeforeUnmount(() => { if (previewUrl.value) URL.revokeObjectURL(previewUrl.value); });
</script>

<template>
  <div class="gallery-page max-w-5xl">
    <header class="mb-8 max-w-2xl">
      <p class="text-xs font-semibold uppercase tracking-[.18em] text-primary">One image at a time</p>
      <h1 class="mt-2 text-3xl font-semibold tracking-tight text-highlighted">投稿一张图片</h1>
      <p class="mt-2 text-sm leading-6 text-muted">不需要创作者身份。登录投稿在安全处理明确通过后可直接展示；匿名投稿会在 Guest 服务上线后进入人工审核。</p>
    </header>

    <UAlert v-if="!loggedIn" class="mb-6" color="warning" variant="subtle" icon="i-tabler-user" title="当前请先登录后投稿" description="Identity Guest Subject 尚未升级完成，因此不会在 Gallery 内伪造临时用户。"><template #actions><UButton label="登录" @click="() => { void login(); }" /></template></UAlert>

    <div v-if="completed" class="rounded-lg border border-success/30 bg-success/5 p-6">
      <UIcon name="i-tabler-circle-check" class="size-8 text-success" />
      <h2 class="mt-4 text-xl font-semibold text-highlighted">投稿已进入处理队列</h2>
      <p class="mt-2 text-sm leading-6 text-muted">图片会先做格式、动画、安全和衍生版本检查。明确安全后按你的登录状态决定是否需要人工审核。</p>
      <div class="mt-5 flex flex-wrap gap-2"><UButton to="/submissions" label="查看我的投稿" /><UButton to="/" color="neutral" variant="outline" label="继续浏览" /></div>
    </div>

    <form v-else class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]" @submit.prevent="submit">
      <div>
        <label class="group grid min-h-80 cursor-pointer place-items-center overflow-hidden rounded-lg border border-dashed border-default bg-elevated/35 focus-within:outline-2 focus-within:outline-offset-4 focus-within:outline-primary">
          <input type="file" class="sr-only" :accept="accepted" :disabled="pending" @change="chooseFile" />
          <img v-if="previewUrl" :src="previewUrl" alt="待投稿图片预览" class="max-h-[65vh] max-w-full object-contain" />
          <span v-else class="max-w-sm px-6 text-center"><UIcon name="i-tabler-photo-up" class="mx-auto size-10 text-dimmed" /><span class="mt-4 block font-semibold text-highlighted">选择一张静态图片</span><span class="mt-2 block text-sm leading-6 text-muted">JPEG、PNG、WebP、AVIF、HEIC/HEIF；最大 20 MiB。不支持 GIF 或其他动画。</span></span>
        </label>
        <UProgress v-if="pending" class="mt-3" :model-value="progress" />
      </div>

      <div class="space-y-5">
        <UFormField label="标题" required><UInput v-model="title" maxlength="160" placeholder="简洁描述这张图片" /></UFormField>
        <UFormField label="分类" required hint="每张图片选择一个主主题"><USelect v-model="topicId" :items="topicItems" value-key="value" placeholder="选择分类" /></UFormField>
        <UFormField label="说明" hint="可选"><UTextarea v-model="description" :rows="5" placeholder="补充画面、背景或整理说明" /></UFormField>
        <UFormField label="来源地址" hint="可选，不知道可以留空"><UInput v-model="sourceUrl" type="url" placeholder="https://" /></UFormField>
        <UFormField label="标签" hint="可选，用逗号分隔"><UInput v-model="tags" placeholder="夜景, 蓝色, 雨" /></UFormField>
        <UButton type="submit" block size="lg" icon="i-tabler-send" label="提交图片" :loading="pending" :disabled="!loggedIn || !file || !title.trim() || !topicId" />
        <p class="text-xs leading-5 text-muted">投稿者不会显示在公开页面。已发布图片不能替换像素；如需更换，请撤回后重新提交。</p>
      </div>
    </form>
  </div>
</template>
