<script setup lang="ts">
import type { GallerySubmissionEditor } from "~/types/submission-editor";
import type { GalleryClassificationNode, GalleryFacet, GallerySubmissionOptions } from "~/types/gallery";
import { classificationPath } from "~/utils/classification-tree";

const model = defineModel<GallerySubmissionEditor>({ required: true });
const props = defineProps<{ options?: GallerySubmissionOptions; disabled?: boolean; batch?: boolean }>();
const inlineLimit = 3;
const categories = computed(() => props.options?.categories ?? []);
const groups = computed(() => (props.options?.facets ?? []).filter(group => group.values.length && group.maxValues > 0));

function menuItems(nodes: GalleryClassificationNode[]) {
  return nodes.map(node => ({ label: classificationPath(node, nodes), value: node.id, slug: node.slug }));
}

const categoryItems = computed(() => menuItems(categories.value));
const groupItems = computed(() => Object.fromEntries(groups.value
  .filter(group => group.values.length > inlineLimit && group.slug !== "image-features")
  .map(group => [group.id, menuItems(group.values)])));
const searchFields = ["label", "slug"];

function selected(group: GalleryFacet) {
  return group.slug === "scene" ? model.value.sceneValueIds : model.value.facetValues[group.id] ?? [];
}

function setSelection(group: GalleryFacet, ids: string[]) {
  if (ids.length > group.maxValues) return;
  if (group.slug === "scene") model.value.sceneValueIds = ids;
  else model.value.facetValues[group.id] = ids;
}

function multipleItems(group: GalleryFacet) {
  const ids = selected(group);
  return groupItems.value[group.id]?.map(item => ({
    ...item,
    disabled: ids.length >= group.maxValues && !ids.includes(item.value),
  })) ?? [];
}

function toggle(group: GalleryFacet, id: string) {
  const current = selected(group);
  if (current.includes(id)) setSelection(group, current.filter(value => value !== id));
  else setSelection(group, group.maxValues === 1 ? [id] : [...current, id]);
}
</script>
<template>
 <fieldset :disabled="disabled" class="submission-fields">
  <UFormField :label="batch ? '默认标题（选填）' : '标题'" :required="!batch"><UInput v-model="model.title" maxlength="160" class="w-full" :placeholder="batch ? '留空时保留各自标题' : '图片标题'" /></UFormField>
  <UFormField label="分类" required>
   <USelectMenu
    v-if="categories.length > inlineLimit" v-model="model.primaryCategoryId" :items="categoryItems"
    value-key="value" :filter-fields="searchFields" :search-input="{ placeholder: '搜索分类' }"
    placeholder="选择分类" class="w-full" :disabled="disabled"
   />
   <div v-else class="submission-visual-options">
    <GalleryVisualChoice v-for="node in categories" :key="node.id" kind="category" :node="node" :selected="model.primaryCategoryId === node.id" @select="model.primaryCategoryId = node.id" />
   </div>
  </UFormField>
  <UFormField v-for="group in groups" :key="group.id" :label="group.name" :required="group.minValues > 0" :hint="group.maxValues > 1 ? `最多 ${group.maxValues} 项` : undefined">
   <GalleryFeatureChoices v-if="group.slug === 'image-features'" :nodes="group.values" :selected-ids="selected(group)" @select="toggle(group, $event.id)" />
   <template v-else-if="group.values.length > inlineLimit">
    <USelectMenu
     v-if="group.maxValues === 1" :model-value="selected(group)[0]" :items="groupItems[group.id]"
     value-key="value" :filter-fields="searchFields" :search-input="{ placeholder: `搜索${group.name}` }"
     :placeholder="`选择${group.name}`" class="w-full" :disabled="disabled"
     @update:model-value="setSelection(group, $event ? [$event] : [])"
    />
    <USelectMenu
     v-else :model-value="selected(group)" :items="multipleItems(group)" multiple
     value-key="value" :filter-fields="searchFields" :search-input="{ placeholder: group.slug === 'work' ? '搜索作品或角色' : `搜索${group.name}` }"
     :placeholder="`选择${group.name}`" class="w-full" :disabled="disabled"
     @update:model-value="setSelection(group, $event)"
    />
   </template>
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
