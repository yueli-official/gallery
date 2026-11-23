<script setup lang="ts">
import { EditorCommandBar, EditorInspector } from "@yueli/ui/admin";
import type { FailureFeedback } from "@yueli/http-runtime";
import { classificationPath } from "~/utils/classification-tree";
import type { GalleryAdminImage, GallerySubmissionOptions, GalleryTagCandidate } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: ["auth", "admin"], key: route => String(route.params.imageId) });
useSeoMeta({ title: "编辑图片 · 图库管理" });
const route = useRoute();
const imageId = String(route.params.imageId);
const { call } = useGalleryApi();
const { can } = useGalleryMe();
const canUpdate = computed(() => can("gallery.image.update"));
const canHide = computed(() => can("gallery.image.hide"));
const canPublish = computed(() => can("gallery.submission.review"));
const immersive = useState("gallery-image-editor-immersive", () => false);
const settingsOpen = ref(false);
const baseline = shallowRef<GalleryAdminImage>();
const baselineForm = ref("");
const saving = ref(false);
const saved = ref(false);
const failure = ref<FailureFeedback | null>(null);
const fieldErrors = ref<Record<string, string>>({});
const deleting = ref(false);
const deleteOpen = ref(false);
const deleteFailure = ref<FailureFeedback | null>(null);
let leavingAfterDelete = false;
const form = reactive({
  publicationState: "draft" as GalleryAdminImage["publicationState"],
  title: "", altText: "", description: "", sourceUrl: "", primaryCategoryId: "",
  facetValueIds: {} as Record<string, string[]>, tagIds: [] as string[],
});
function snapshot() {
  return JSON.stringify({ ...form, facetValueIds: Object.fromEntries(Object.entries(form.facetValueIds).filter(([, values]) => values.length).sort(([a], [b]) => a.localeCompare(b)).map(([id, values]) => [id, [...values].sort()])), tagIds: [...form.tagIds].sort() });
}
const dirty = computed(() => Boolean(baseline.value) && snapshot() !== baselineForm.value);
watch(dirty, value => { if (value) saved.value = false; });

const { data, status, error, refresh } = await useAsyncData(
  `gallery-image-editor:${imageId}`,
  () => call<{ image: GalleryAdminImage }>(`/admin/images/${encodeURIComponent(imageId)}`),
  { server: false },
);
const image = computed(() => data.value?.image);
const hydrated = useClientHydrated();
const pending = useMinimumLoading(computed(() => !hydrated.value || status.value === "pending" || status.value === "idle"));
useGalleryMediaRefresh(data, refresh);
function hydrate(value: GalleryAdminImage) {
  baseline.value = value;
  Object.assign(form, {
    publicationState: value.publicationState, title: value.title, altText: value.altText,
    description: value.description, sourceUrl: value.sourceUrl, primaryCategoryId: value.primaryCategoryId,
    facetValueIds: value.facets.reduce<Record<string, string[]>>((groups, item) => {
      (groups[item.facetId] ||= []).push(item.valueId); return groups;
    }, {}),
    tagIds: value.tags.map(tag => tag.id),
  });
  baselineForm.value = snapshot();
  fieldErrors.value = {};
  failure.value = null;
}
// Refreshing expiring media must not replace an unsaved metadata draft or its version.
watch(image, value => { if (value && !baseline.value) hydrate(value); }, { immediate: true });
const { data: options, error: optionsError, refresh: refreshOptions } = await useFetch<GallerySubmissionOptions>("/api/gallery/submission-options", { server: false });
const categoryOptions = computed(() => {
  const items = (options.value?.categories || []).map(node => ({ label: classificationPath(node, options.value?.categories || []), value: node.id }));
  if (baseline.value?.primaryCategoryId && !items.some(item => item.value === baseline.value!.primaryCategoryId)) {
    items.push({ label: baseline.value.primaryCategory, value: baseline.value.primaryCategoryId });
  }
  return items;
});
const tagSearch = ref("");
const { data: suggestedTags, error: tagError, refresh: refreshTags } = await useAsyncData(
  `gallery-image-editor-tags:${imageId}`,
  () => call<{ items: GalleryTagCandidate[] }>("/tag-suggestions", { query: { q: tagSearch.value } }),
  { server: false, watch: [tagSearch] },
);
const tagOptions = computed(() => {
  const items = new Map((suggestedTags.value?.items || []).map(tag => [tag.id, { label: tag.name, value: tag.id }]));
  for (const tag of baseline.value?.tags || []) items.set(tag.id, { label: tag.name, value: tag.id });
  return [...items.values()];
});
const publicationOptions = computed(() => {
  const current = baseline.value;
  if (!current) return [];
  const names = { draft: "草稿", published: "已公开", hidden: "已隐藏", deleted: "已删除" };
  return [
    { label: names[current.publicationState], value: current.publicationState },
    ...(canPublish.value && !["published", "deleted"].includes(current.publicationState) ? [{ label: "公开", value: "published" as const }] : []),
    ...(canHide.value && current.publicationState === "published" ? [{ label: "下架（隐藏）", value: "hidden" as const }] : []),
  ];
});
function fieldError(key: string) { return fieldErrors.value[key] || failure.value?.fieldErrors[key]?.join(" "); }
function toggleFacet(group: GallerySubmissionOptions["facets"][number], id: string) {
  const selected = form.facetValueIds[group.id] || [];
  const next = selected.includes(id) ? selected.filter(value => value !== id) : group.maxValues === 1 ? [id] : [...selected, id];
  if (next.length > group.maxValues) { fieldErrors.value[`facet:${group.id}`] = `最多选择 ${group.maxValues} 项`; return; }
  delete fieldErrors.value[`facet:${group.id}`];
  form.facetValueIds[group.id] = next;
}
async function save() {
  if (!canUpdate.value || saving.value || deleting.value || !baseline.value?.updatedAt || !dirty.value) return;
  fieldErrors.value = {};
  if (!form.title.trim()) fieldErrors.value.title = "请填写标题";
  if (!form.altText.trim()) fieldErrors.value.altText = "请填写替代文本";
  if (!form.primaryCategoryId) fieldErrors.value.primaryCategoryId = "请选择主分类";
  if (Object.keys(fieldErrors.value).length) return;
  saving.value = true;
  failure.value = null;
  try {
    const response = await call<{ image: GalleryAdminImage }>(`/admin/images/${encodeURIComponent(baseline.value.id)}`, {
      method: "PATCH",
      body: {
        expectedUpdatedAt: baseline.value.updatedAt,
        publicationState: form.publicationState !== baseline.value.publicationState ? form.publicationState : undefined,
        title: form.title, description: form.description, altText: form.altText, sourceUrl: form.sourceUrl,
        classification: { primaryCategoryId: form.primaryCategoryId, facetValueIds: Object.values(form.facetValueIds).flat(), tagIds: form.tagIds },
      },
    });
    data.value = response;
    hydrate(response.image);
    saved.value = true;
  } catch (cause) {
    failure.value = galleryFailureFeedback(cause, "图片信息没有保存，请检查后重试。", {
      "/title": "title", "/altText": "altText", "/description": "description", "/sourceUrl": "sourceUrl",
      "/publicationState": "publicationState", "/classification/primaryCategoryId": "primaryCategoryId", "/classification/tagIds": "tagIds",
    });
  } finally { saving.value = false; }
}
async function reload() {
  if (dirty.value && !window.confirm("重新载入会替换尚未保存的修改，确定继续？")) return;
  await refresh();
  if (data.value?.image && !error.value) { hydrate(data.value.image); saved.value = false; }
}
async function deleteImage() {
  if (!canHide.value || !baseline.value?.updatedAt || saving.value || deleting.value) return;
  deleting.value = true;
  deleteFailure.value = null;
  try {
    await call(`/admin/images/${encodeURIComponent(baseline.value.id)}`, { method: "DELETE", body: { expectedUpdatedAt: baseline.value.updatedAt } });
    leavingAfterDelete = true;
    await navigateTo("/manage/images");
  } catch (cause) { deleteFailure.value = galleryFailureFeedback(cause, "图片没有删除，请检查后重试。"); }
  finally { deleting.value = false; }
}
onBeforeRouteLeave(() => {
  if (leavingAfterDelete) return true;
  if (saving.value || deleting.value) return false;
  return !dirty.value || window.confirm("修改尚未保存，确定离开？");
});
onBeforeUnmount(() => { immersive.value = false; });
</script>

<template>
  <div class="min-w-0" :class="{ 'xl:pr-[25rem]': settingsOpen }">
    <EditorCommandBar v-model:immersive="immersive" v-model:settings-open="settingsOpen" back-to="/manage/images" back-label="返回图片列表" title="编辑图片" settings-label="分类与图片设置" :controls-disabled="saving || deleting || !image">
      <template v-if="image" #title>
        <UFormField :error="fieldError('title')" class="min-w-0 flex-1">
          <UInput v-model="form.title" form="gallery-image-edit-form" aria-label="标题" placeholder="图片标题" maxlength="160" required class="w-full" :disabled="saving || deleting" />
        </UFormField>
      </template>
      <template #preview><UButton v-if="image?.publicationState === 'published'" :to="`/images/${image.publicId}`" target="_blank" rel="noopener" label="查看图片" icon="i-tabler-external-link" color="neutral" variant="ghost" size="sm" /></template>
      <template #actions><UButton form="gallery-image-edit-form" type="submit" :label="saved && !dirty ? '已保存' : '保存更改'" :loading="saving" :disabled="!canUpdate || !image || !dirty || deleting" size="sm" /></template>
    </EditorCommandBar>

    <main class="mx-auto grid w-full max-w-5xl gap-5 p-3 sm:p-5">
      <div v-if="pending && !image" class="space-y-4" aria-label="正在读取图片信息"><USkeleton class="h-64 rounded-lg" /><USkeleton class="h-24 rounded-lg" /></div>
      <UAlert v-else-if="error && !image" color="error" variant="subtle" title="图片信息未能读取" :description="galleryFailureMessage(error, '请检查权限或图片是否存在，然后重试。')"><template #actions><UButton label="重试" color="neutral" variant="outline" @click="refresh()" /></template></UAlert>
      <form v-else-if="image" id="gallery-image-edit-form" class="grid gap-5" @submit.prevent="save">
        <GalleryFailureNotice :feedback="failure" />
        <div v-if="failure" class="flex justify-end"><UButton label="重新载入" color="neutral" variant="outline" size="sm" :disabled="saving" @click="reload" /></div>
        <div data-image-editor-preview class="grid min-h-52 place-items-center overflow-hidden rounded-lg border border-default bg-elevated p-2"><img :src="galleryRendition(image, 'preview')" :alt="form.altText || image.title" class="max-h-[55svh] max-w-full object-contain" /></div>
        <UFormField label="说明" :error="fieldError('description')"><UTextarea v-model="form.description" :rows="4" class="w-full" :disabled="saving || deleting" /></UFormField>
        <UFormField label="替代文本" required :error="fieldError('altText')"><UInput v-model="form.altText" required class="w-full" :disabled="saving || deleting" /></UFormField>
        <UFormField label="来源地址" :error="fieldError('sourceUrl')"><UInput v-model="form.sourceUrl" type="url" placeholder="https://" class="w-full" :disabled="saving || deleting" /></UFormField>
      </form>
    </main>

    <EditorInspector v-model:open="settingsOpen" title="分类与图片设置">
      <div class="space-y-5">
        <UAlert v-if="optionsError" color="error" variant="subtle" title="分类选项未能加载" :description="galleryFailureMessage(optionsError, '请重试后再修改分类或图片特征。')"><template #actions><UButton label="重试" color="neutral" variant="outline" @click="refreshOptions()" /></template></UAlert>
        <UFormField label="发布状态" :error="fieldError('publicationState')"><USelect v-model="form.publicationState" :items="publicationOptions" class="w-full" :disabled="saving || deleting || publicationOptions.length < 2" /><p class="mt-1 text-xs text-muted">公开后显示在图库中；下架后停止公开展示。</p><p v-if="image && (image.processingState !== 'ready' || !['approved', 'not_required'].includes(image.reviewState) || image.safetyState !== 'safe')" class="mt-1 text-xs text-warning">图片须处理完成并通过审核后才能公开。</p></UFormField>
        <UFormField label="主分类" required :error="fieldError('primaryCategoryId')"><USelectMenu v-model="form.primaryCategoryId" :items="categoryOptions" value-key="value" aria-label="主分类" class="w-full" :disabled="saving || deleting" /></UFormField>
        <UFormField v-for="group in options?.facets || []" :key="group.id" :label="group.name" :error="fieldError(`facet:${group.id}`)">
          <fieldset :disabled="saving || deleting">
            <GalleryFeatureChoices v-if="group.slug === 'image-features'" :nodes="group.values" :selected-ids="form.facetValueIds[group.id] || []" @select="toggleFacet(group, $event.id)" />
            <GalleryWorkChoices v-else-if="group.slug === 'work'" :nodes="group.values" :selected-ids="form.facetValueIds[group.id] || []" @select="toggleFacet(group, $event.id)" />
            <USelectMenu v-else v-model="form.facetValueIds[group.id]" multiple :items="group.values.map(node => ({ label: classificationPath(node, group.values), value: node.id }))" value-key="value" class="w-full" placeholder="未设置" />
          </fieldset>
        </UFormField>
        <UFormField label="标签" :error="fieldError('tagIds') || (tagError ? galleryFailureMessage(tagError, '标签选项未能加载，请重试。') : undefined)"><USelectMenu v-model="form.tagIds" v-model:search-term="tagSearch" :items="tagOptions" value-key="value" multiple aria-label="标签" class="w-full" placeholder="选择标签" :disabled="saving || deleting" /><UButton v-if="tagError" label="重新加载标签" color="neutral" variant="ghost" size="xs" @click="refreshTags()" /></UFormField>
        <div v-if="canHide && image?.publicationState !== 'deleted'" class="border-t border-default pt-5"><UButton label="删除图片" icon="i-tabler-trash" color="error" variant="ghost" :disabled="saving || deleting" @click="void (deleteOpen = true)" /></div>
      </div>
    </EditorInspector>
    <UModal v-model:open="deleteOpen" title="删除这张图片？" description="删除后将不再公开展示，也会从专题中移除。" :dismissible="!deleting">
      <template #body><GalleryFailureNotice :feedback="deleteFailure" /><p class="break-words text-sm">{{ image?.title }}</p></template>
      <template #footer><UButton label="取消" color="neutral" variant="ghost" :disabled="deleting" @click="void (deleteOpen = false)" /><UButton label="确认删除" color="error" :loading="deleting" @click="deleteImage" /></template>
    </UModal>
  </div>
</template>
