<script setup lang="ts">
import type { GallerySubmissionEditor } from "~/types/submission-editor";
import type { GalleryFacet, GallerySubmissionOptions } from "~/types/gallery";
const model = defineModel<GallerySubmissionEditor>({ required: true });
defineProps<{ options?: GallerySubmissionOptions; disabled?: boolean; batch?: boolean }>();
function selected(group: GalleryFacet) { return group.slug === 'scene' ? model.value.sceneValueIds : model.value.facetValues[group.id] || []; }
function toggle(group: GalleryFacet, id: string) {
  const current=selected(group);
  const next=current.includes(id) ? current.filter(value=>value!==id) : group.maxValues === 1 ? [id] : [...current,id];
  if(next.length>group.maxValues)return;
  if(group.slug==='scene') model.value.sceneValueIds=next;
  else model.value.facetValues[group.id]=next;
}
</script>
<template>
 <fieldset :disabled="disabled" class="submission-fields">
  <UFormField :label="batch ? '默认标题（选填）' : '标题'" :required="!batch"><UInput v-model="model.title" maxlength="160" class="w-full" :placeholder="batch ? '留空时保留各自标题' : '图片标题'" /></UFormField>
  <UFormField label="分类" required><div class="submission-visual-options"><GalleryVisualChoice v-for="node in options?.categories || []" :key="node.id" kind="category" :node="node" :selected="model.primaryCategoryId === node.id" @select="model.primaryCategoryId = node.id" /></div></UFormField>
  <UFormField v-for="group in (options?.facets || []).filter(item => item.values.length && item.maxValues > 0)" :key="group.id" :label="group.name" :required="group.minValues > 0" :hint="group.maxValues > 1 ? `最多 ${group.maxValues} 项` : undefined">
   <GalleryFeatureChoices v-if="group.slug === 'image-features'" :nodes="group.values" :selected-ids="selected(group)" @select="toggle(group, $event.id)" />
   <GalleryWorkChoices v-else-if="group.slug === 'work'" :nodes="group.values" :selected-ids="selected(group)" @select="toggle(group, $event.id)" />
   <div v-else class="submission-visual-options"><GalleryVisualChoice v-for="node in group.values" :key="node.id" :node="node" :selected="selected(group).includes(node.id)" @select="toggle(group, node.id)" /></div>
   <p v-if="group.minValues > selected(group).length" class="mt-1 text-xs text-muted">至少选择 {{ group.minValues }} 项</p>
  </UFormField>
  <UFormField label="来源（选填）"><UInput v-model="model.sourceUrl" type="url" placeholder="粘贴原始页面链接" class="w-full" /></UFormField>
  <UFormField label="标签（选填）" hint="逗号分隔"><UInput v-model="model.tags" placeholder="添加标签" class="w-full" /></UFormField>
  <UFormField label="说明（选填）"><UTextarea v-model="model.description" :rows="3" class="w-full" /></UFormField>
 </fieldset>
</template>
<style scoped>
.submission-fields{display:grid;gap:1rem;min-width:0}.submission-visual-options{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:.4rem;max-height:18rem;overflow:auto}
</style>
