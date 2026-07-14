<script setup lang="ts">
import type { GalleryDiscovery, GalleryStudioArtwork } from "~/types/gallery";

definePageMeta({ middleware: "auth" });
const route = useRoute();
const artworkId = String(route.params.artworkId);
const { call } = useApi();
const { upload } = useGalleryAssetUpload();
const toast = useToast();

const { data: discovery } = await useFetch<GalleryDiscovery>("/api/gallery/discovery");
const { data, pending, error: artworkError, refresh } = await useAsyncData(
  `gallery-studio-${artworkId}`,
  () => call<{ artwork: GalleryStudioArtwork }>(`/api/v1/gallery/me/artworks/${encodeURIComponent(artworkId)}`),
  { server: false },
);
const artwork = computed(() => data.value?.artwork);

const form = reactive({
  title: "",
  description: "",
  visibility: "public",
  contentRating: "general",
  aiUsage: "none",
  aiTrainingPermission: "unspecified",
  rightsBasis: "original",
  license: "all_rights_reserved",
});
const selected = reactive<Record<string, string[]>>({});
const initialized = ref(false);

watch([artwork, discovery], ([value]) => {
  if (!value || initialized.value) return;
  Object.assign(form, {
    title: value.title,
    description: value.description,
    visibility: value.visibility,
    contentRating: value.contentRating,
    aiUsage: value.aiUsage,
    aiTrainingPermission: value.aiTrainingPermission,
    rightsBasis: value.rightsBasis,
    license: value.license,
  });
  for (const assignment of value.facetAssignments) {
    (selected[assignment.facetId] ||= []).push(assignment.valueId);
  }
  initialized.value = true;
}, { immediate: true });

const editable = computed(() => ["draft", "rejected"].includes(artwork.value?.status ?? ""));
const saving = ref(false);
const submitting = ref(false);
const uploadProgress = ref<Record<string, number>>({});

function toggleFacet(facetId: string, valueId: string, multiple: boolean) {
  if (!editable.value) return;
  const current = selected[facetId] ?? [];
  if (!multiple) {
    selected[facetId] = current.includes(valueId) ? [] : [valueId];
    return;
  }
  selected[facetId] = current.includes(valueId)
    ? current.filter((id) => id !== valueId)
    : [...current, valueId];
}

function payload() {
  return {
    ...form,
    facetSelections: Object.entries(selected).map(([facetId, valueIds]) => ({ facetId, valueIds })),
  };
}

async function save(silent = false) {
  saving.value = true;
  try {
    await call(`/api/v1/gallery/me/artworks/${encodeURIComponent(artworkId)}`, { method: "PATCH", body: payload() });
    await refresh();
    if (!silent) toast.add({ title: "草稿已保存", color: "success" });
  } catch (error: any) {
    toast.add({ title: "保存失败", description: error?.data?.message || error?.message, color: "error" });
    throw error;
  } finally {
    saving.value = false;
  }
}

async function imageSize(file: File) {
  try {
    const bitmap = await createImageBitmap(file);
    const size = { width: bitmap.width, height: bitmap.height };
    bitmap.close();
    return size;
  } catch {
    return { width: 0, height: 0 };
  }
}

async function chooseFiles(event: Event) {
  const input = event.target as HTMLInputElement;
  const files = Array.from(input.files ?? []);
  for (const file of files) {
    uploadProgress.value[file.name] = 0;
    try {
      const [asset, dimensions] = await Promise.all([
        upload(file, (value) => { uploadProgress.value[file.name] = value; }),
        imageSize(file),
      ]);
      await call(`/api/v1/gallery/me/artworks/${encodeURIComponent(artworkId)}/assets`, {
        method: "POST",
        body: {
          assetId: asset.id,
          width: asset.width || dimensions.width,
          height: asset.height || dimensions.height,
          format: file.type,
          altText: form.title,
        },
      });
    } catch (error: any) {
      toast.add({ title: `${file.name} 上传失败`, description: error?.data?.message || error?.message, color: "error" });
    } finally {
      uploadProgress.value = Object.fromEntries(Object.entries(uploadProgress.value).filter(([name]) => name !== file.name));
    }
  }
  input.value = "";
  await refresh();
}

async function removeAsset(id: string) {
  await call(`/api/v1/gallery/me/artworks/${encodeURIComponent(artworkId)}/assets/${encodeURIComponent(id)}`, { method: "DELETE" });
  await refresh();
}

async function submit() {
  submitting.value = true;
  try {
    await save(true);
    await call(`/api/v1/gallery/me/artworks/${encodeURIComponent(artworkId)}/submit`, { method: "POST" });
    await refresh();
    toast.add({ title: "已提交审核", description: "审核期间作品内容将暂时锁定。", color: "success" });
  } catch (error: any) {
    toast.add({ title: "暂时不能提交", description: error?.data?.message || error?.message || "请检查必填信息", color: "error" });
  } finally {
    submitting.value = false;
  }
}

useSeoMeta({ title: () => artwork.value?.title || "编辑作品" });
</script>

<template>
  <div class="mx-auto w-full max-w-7xl py-4 sm:py-8">
    <div v-if="pending" class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_23rem]">
      <USkeleton class="h-[34rem] rounded-2xl" />
      <USkeleton class="h-[26rem] rounded-2xl" />
    </div>

    <UAlert v-else-if="artworkError" color="error" variant="subtle" icon="i-tabler-alert-circle" title="无法读取这份作品" description="作品不存在、不属于当前创作者，或 Gallery API 暂时不可用。" />

    <template v-else-if="artwork">
      <header class="mb-8 flex flex-col gap-4 border-b border-default pb-6 sm:flex-row sm:items-center sm:justify-between">
        <div class="flex min-w-0 items-center gap-3">
          <UButton to="/studio" icon="i-tabler-arrow-left" color="neutral" variant="ghost" aria-label="返回创作中心" />
          <div class="min-w-0">
            <p class="truncate font-semibold text-highlighted">{{ form.title || "未命名作品" }}</p>
            <p class="text-xs text-muted">{{ editable ? "草稿可编辑" : artwork.status === "pending_review" ? "审核中，内容已锁定" : "当前版本只读" }}</p>
          </div>
        </div>
        <div class="flex gap-2">
          <UButton v-if="editable" color="neutral" variant="soft" label="保存草稿" :loading="saving" @click="save()" />
          <UButton v-if="editable" icon="i-tabler-send" label="提交审核" :loading="submitting" @click="submit" />
          <UButton v-else-if="artwork.status === 'published'" :to="`/artworks/${artwork.id}`" icon="i-tabler-external-link" label="查看作品" />
        </div>
      </header>

      <UAlert v-if="artwork.status === 'rejected'" class="mb-6" color="error" variant="subtle" icon="i-tabler-alert-circle" title="作品需要修改" :description="artwork.reviewNote || '请根据审核要求完善后重新提交。'" />

      <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_23rem] lg:items-start">
        <main class="space-y-6">
          <section class="rounded-2xl border border-default bg-default p-5 sm:p-6">
            <div class="flex items-center justify-between gap-4">
              <div>
                <h2 class="font-semibold text-highlighted">作品图片</h2>
                <p class="mt-1 text-xs text-muted">支持多图，第一张作为封面；单张最大 50 MB。</p>
              </div>
              <label v-if="editable" class="cursor-pointer rounded-lg focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-primary">
                <input class="sr-only" type="file" accept="image/jpeg,image/png,image/webp,image/gif" multiple @change="chooseFiles" />
                <span class="inline-flex h-9 items-center gap-2 rounded-lg bg-primary px-3 text-sm font-medium text-inverted"><UIcon name="i-tabler-upload" class="size-4" />添加图片</span>
              </label>
            </div>
            <div v-if="artwork.assets.length" class="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-3">
              <figure v-for="asset in artwork.assets" :key="asset.id" class="group relative overflow-hidden rounded-xl border border-default bg-elevated">
                <img :src="asset.cardUrl" :alt="asset.altText || form.title" class="aspect-[4/3] size-full object-cover" />
                <figcaption class="absolute inset-x-0 bottom-0 flex items-center justify-between bg-black/55 px-2.5 py-2 text-xs text-white opacity-100 backdrop-blur transition sm:opacity-0 sm:group-hover:opacity-100 sm:group-focus-within:opacity-100">
                  <span>{{ asset.width }} × {{ asset.height }}</span>
                  <UButton v-if="editable" icon="i-tabler-trash" color="error" variant="soft" size="xs" aria-label="移除图片" @click="removeAsset(asset.id)" />
                </figcaption>
              </figure>
            </div>
            <div v-else class="mt-5 rounded-xl border border-dashed border-default py-16 text-center text-sm text-muted">添加至少一张图片后才能提交审核。</div>
            <div v-if="Object.keys(uploadProgress).length" class="mt-4 space-y-3">
              <div v-for="(progress, name) in uploadProgress" :key="name">
                <div class="mb-1 flex justify-between text-xs text-muted"><span class="truncate">{{ name }}</span><span>{{ progress }}%</span></div>
                <UProgress :model-value="progress" />
              </div>
            </div>
          </section>

          <section class="space-y-5 rounded-2xl border border-default bg-default p-5 sm:p-6">
            <h2 class="font-semibold text-highlighted">作品信息</h2>
            <UFormField label="标题" required>
              <UInput v-model="form.title" :disabled="!editable" placeholder="给作品一个清晰的标题" class="w-full" />
            </UFormField>
            <UFormField label="描述">
              <UTextarea v-model="form.description" :disabled="!editable" :rows="7" placeholder="创作背景、内容说明或使用工具……" class="w-full" />
            </UFormField>
          </section>

          <section class="rounded-2xl border border-default bg-default p-5 sm:p-6">
            <div>
              <h2 class="font-semibold text-highlighted">分类维度</h2>
              <p class="mt-1 text-xs text-muted">维度由平台统一维护，同一维度内可单选或多选。</p>
            </div>
            <div class="mt-6 space-y-6">
              <fieldset v-for="facet in discovery?.facets" :key="facet.id">
                <legend class="text-sm font-medium text-highlighted">{{ facet.name }} <span v-if="facet.requiredOnPublish" class="text-error">*</span></legend>
                <p v-if="facet.description" class="mt-1 text-xs text-muted">{{ facet.description }}</p>
                <div class="mt-3 flex flex-wrap gap-2">
                  <button
                    v-for="value in discovery?.facetValues.filter((item) => item.facetId === facet.id)"
                    :key="value.id"
                    type="button"
                    class="rounded-full border px-3 py-1.5 text-sm transition focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
                    :class="selected[facet.id]?.includes(value.id) ? 'border-primary bg-primary/10 text-primary' : 'border-default text-muted hover:border-primary/40 hover:text-default'"
                    :disabled="!editable"
                    :aria-pressed="selected[facet.id]?.includes(value.id) || false"
                    @click="toggleFacet(facet.id, value.id, facet.selectionMode === 'multiple')"
                  >{{ value.name }}</button>
                </div>
              </fieldset>
            </div>
          </section>
        </main>

        <aside class="space-y-6 lg:sticky lg:top-24">
          <section class="space-y-5 rounded-2xl border border-default bg-default p-5">
            <h2 class="font-semibold text-highlighted">发布设置</h2>
            <UFormField label="可见范围"><USelect v-model="form.visibility" :disabled="!editable" :items="[{ label: '公开', value: 'public' }, { label: '不列出', value: 'unlisted' }, { label: '私密', value: 'private' }]" class="w-full" /></UFormField>
            <UFormField label="内容分级"><USelect v-model="form.contentRating" :disabled="!editable" :items="[{ label: '全年龄', value: 'general' }, { label: '敏感内容', value: 'sensitive' }, { label: '成人内容', value: 'adult' }]" class="w-full" /></UFormField>
            <UFormField label="AI 使用"><USelect v-model="form.aiUsage" :disabled="!editable" :items="[{ label: '未使用', value: 'none' }, { label: '辅助创作', value: 'assistive' }, { label: '主要由 AI 生成', value: 'mostly_generated' }]" class="w-full" /></UFormField>
            <UFormField label="训练授权"><USelect v-model="form.aiTrainingPermission" :disabled="!editable" :items="[{ label: '未声明', value: 'unspecified' }, { label: '允许', value: 'allow' }, { label: '不允许', value: 'disallow' }]" class="w-full" /></UFormField>
            <UFormField label="权利基础"><USelect v-model="form.rightsBasis" :disabled="!editable" :items="[{ label: '原创', value: 'original' }, { label: '授权转载', value: 'authorized_repost' }, { label: '公有领域', value: 'public_domain' }, { label: '已获许可素材', value: 'licensed_material' }]" class="w-full" /></UFormField>
          </section>
          <UAlert color="neutral" variant="subtle" icon="i-tabler-info-circle" title="提交前检查" description="标题、至少一张图片，以及标注为 * 的分类维度必须完整。" />
        </aside>
      </div>
    </template>
  </div>
</template>
