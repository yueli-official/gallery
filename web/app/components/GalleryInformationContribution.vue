<script setup lang="ts">
import type { FailureFeedback } from "@yueli/http-runtime";
import type { GalleryImage, GalleryImageMetadataCorrection, GallerySubmissionOptions } from "~/types/gallery";
import { classificationPath } from "~/utils/classification-tree";
const props = defineProps<{ image: GalleryImage }>();
const open = defineModel<boolean>("open", { default: false });
const { call } = useGalleryApi();
const active = ref("information");
const pending = ref(false);
const tagPending = ref(false);
const busy = computed(() => pending.value || tagPending.value);
const result = ref("");
const failure = ref<FailureFeedback | null>(null);
const validation = ref("");
const options = shallowRef<GallerySubmissionOptions>();
const optionsError = ref("");
const loading = ref(false);
let revision = 0;
const initial = shallowRef<{ image: GalleryImage; categoryId: string }>();
const form = reactive({ title: "", description: "", altText: "", sourceUrl: "", primaryCategoryId: "", facets: {} as Record<string, string[]>, note: "" });
const tabs = computed(() => [{ label: "图片信息", value: "information", disabled: busy.value }, { label: "标签", value: "tags", disabled: busy.value }]);
const selectedFacets = computed(() => Object.values(form.facets).flat());
const classificationChanged = computed(() => Boolean(initial.value) && (
  form.primaryCategoryId !== initial.value!.categoryId ||
  JSON.stringify([...selectedFacets.value].sort()) !== JSON.stringify(initial.value!.image.facets.map(item => item.valueId).sort())
));
const changed = computed(() => Boolean(initial.value) && (
  (["title", "description", "altText", "sourceUrl"] as const).some(key => form[key].trim() !== initial.value!.image[key]) || classificationChanged.value
));
function reset(image: GalleryImage) {
  const value = { ...image, facets: image.facets.map(item => ({ ...item })), tags: [...image.tags] };
  const categoryId = options.value?.categories.find(node => node.slug === value.primaryCategorySlug)?.id || "";
  initial.value = { image: value, categoryId };
  Object.assign(form, { title: value.title, description: value.description, altText: value.altText, sourceUrl: value.sourceUrl, primaryCategoryId: categoryId, note: "", facets: value.facets.reduce<Record<string, string[]>>((groups, item) => { (groups[item.facetId] ||= []).push(item.valueId); return groups; }, {}) });
  failure.value = null; validation.value = "";
}
async function loadOptions() {
  const current = revision;
  loading.value = true; optionsError.value = "";
  try {
    const value = await call<GallerySubmissionOptions>("/submission-options");
    if (current !== revision) return;
    options.value = value;
    if (initial.value && !initial.value.categoryId) {
      const id = value.categories.find(node => node.slug === initial.value!.image.primaryCategorySlug)?.id || "";
      initial.value.categoryId = id; form.primaryCategoryId = id;
    }
  } catch (cause) { if (current === revision) optionsError.value = galleryFailureMessage(cause, "分类选项未能读取，请重试。"); }
  finally { if (current === revision) loading.value = false; }
}
watch(open, value => {
  revision++;
  if (!value) { tagPending.value = false; return; }
  result.value = ""; active.value = "information"; reset(props.image); void loadOptions();
});
watch(() => props.image.id, () => { open.value = false; result.value = ""; revision++; });
onBeforeUnmount(() => { revision++; });
function toggleFacet(group: GallerySubmissionOptions["facets"][number], id: string) {
  const current = form.facets[group.id] || [];
  const next = current.includes(id) ? current.filter(value => value !== id) : group.maxValues === 1 ? [id] : [...current, id];
  if (next.length > group.maxValues) { validation.value = `${group.name}最多选择 ${group.maxValues} 项`; return; }
  form.facets[group.id] = next; validation.value = "";
}
function submitted(message: string) { result.value = message; open.value = false; }
async function submit() {
  if (busy.value || !initial.value || !changed.value) return;
  validation.value = ""; failure.value = null;
  if (!form.title.trim() || !form.altText.trim()) { validation.value = "标题和替代文本不能为空"; return; }
  const source = initial.value.image;
  if (!source.updatedAt) { validation.value = "图片版本未能读取，请重新打开图片后提交。"; return; }
  const proposal: GalleryImageMetadataCorrection = { expectedImageUpdatedAt: source.updatedAt };
  for (const key of ["title", "description", "altText", "sourceUrl"] as const) if (form[key].trim() !== source[key]) proposal[key] = form[key].trim();
  if (classificationChanged.value) {
    if (!form.primaryCategoryId || !options.value || optionsError.value) { validation.value = "请先载入分类选项并选择主分类。"; return; }
    proposal.classification = { primaryCategoryId: form.primaryCategoryId, facetValueIds: [...selectedFacets.value] };
  }
  const current = revision;
  pending.value = true;
  try {
    await call(`/images/${encodeURIComponent(source.id)}/cases`, { method: "POST", body: { kind: "metadata_correction", description: form.note.trim(), proposedMetadata: proposal } });
    if (current === revision && props.image.id === source.id) submitted("信息建议已提交，审核后生效");
  } catch (cause) {
    if (current === revision) failure.value = galleryFailureFeedback(cause, "信息建议没有提交，请检查后重试。", { "/proposedMetadata/title": "title", "/proposedMetadata/altText": "altText", "/proposedMetadata/sourceUrl": "sourceUrl", "/proposedMetadata/description": "description" });
  } finally { pending.value = false; }
}
</script>

<template>
  <div>
    <UButton icon="i-tabler-plus" label="补充信息" color="neutral" variant="ghost" size="xs" @click="void (open = true)" />
    <p v-if="result" class="mt-1 text-xs text-muted" role="status">{{ result }}</p>
    <UModal v-model:open="open" title="补充图片信息" description="审核通过后生效。" :dismissible="!busy" :ui="{ content: 'sm:max-w-3xl' }">
      <template #body>
        <UTabs v-model="active" :items="tabs" :content="false" class="mb-4" />
        <GalleryTagContribution v-if="active === 'tags'" :image-id="image.id" :tags="image.tags" @submitted="submitted" @cancel="open = false" @pending="tagPending = $event" />
        <form v-else id="gallery-information-proposal" class="space-y-4" @submit.prevent="submit">
          <GalleryFailureNotice :feedback="failure" />
          <UAlert v-if="validation" color="error" variant="subtle" :title="validation" />
          <div class="grid gap-4 sm:grid-cols-2">
            <UFormField label="标题" :error="failure?.fieldErrors.title?.join(' ')"><UInput v-model="form.title" aria-label="建议标题" maxlength="160" class="w-full" :disabled="busy" /></UFormField>
            <UFormField label="替代文本" :error="failure?.fieldErrors.altText?.join(' ')"><UInput v-model="form.altText" aria-label="建议替代文本" maxlength="2000" class="w-full" :disabled="busy" /></UFormField>
          </div>
          <UFormField label="说明" :error="failure?.fieldErrors.description?.join(' ')"><UTextarea v-model="form.description" aria-label="建议说明" :rows="3" maxlength="10000" class="w-full" :disabled="busy" /></UFormField>
          <UFormField label="来源地址" :error="failure?.fieldErrors.sourceUrl?.join(' ')"><UInput v-model="form.sourceUrl" type="url" aria-label="建议来源地址" placeholder="https://" maxlength="2000" class="w-full" :disabled="busy" /></UFormField>
          <UAlert v-if="optionsError" color="error" variant="subtle" :title="optionsError"><template #actions><UButton label="重试" color="neutral" variant="outline" @click="loadOptions" /></template></UAlert>
          <USkeleton v-if="loading && !options" class="h-20 rounded-lg" />
          <fieldset v-else-if="options" class="space-y-4 border-t border-default pt-4" :disabled="busy">
            <UFormField label="主分类"><USelectMenu v-model="form.primaryCategoryId" :items="options.categories.map(node => ({ label: classificationPath(node, options!.categories), value: node.id }))" value-key="value" aria-label="建议主分类" class="w-full" /></UFormField>
            <UFormField v-for="group in options.facets" :key="group.id" :label="group.name">
              <GalleryFeatureChoices v-if="group.slug === 'image-features'" :nodes="group.values" :selected-ids="form.facets[group.id] || []" @select="toggleFacet(group, $event.id)" />
              <GalleryWorkChoices v-else-if="group.slug === 'work'" :nodes="group.values" :selected-ids="form.facets[group.id] || []" @select="toggleFacet(group, $event.id)" />
              <USelectMenu v-else v-model="form.facets[group.id]" :items="group.values.map(node => ({ label: classificationPath(node, group.values), value: node.id }))" multiple value-key="value" class="w-full" />
            </UFormField>
          </fieldset>
          <UFormField label="补充理由"><UTextarea v-model="form.note" maxlength="2000" :rows="2" class="w-full" :disabled="busy" /></UFormField>
        </form>
      </template>
      <template v-if="active !== 'tags'" #footer>
        <div class="flex w-full justify-end gap-2"><UButton label="取消" color="neutral" variant="ghost" :disabled="busy" @click="void (open = false)" /><UButton form="gallery-information-proposal" type="submit" label="提交审核" :loading="pending" :disabled="!changed || busy" /></div>
      </template>
    </UModal>
  </div>
</template>
