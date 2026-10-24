<script setup lang="ts">
import { SettingSection } from "@yueli/ui/settings/pattern";
import type { GalleryAppearance, GalleryPresentationRule } from "~/types/presentation";
import { galleryPresentationIcons } from "~/types/presentation";
import type { GallerySubmissionOptions } from "~/types/gallery";
const model = defineModel<GalleryAppearance>({ required: true });
defineProps<{ disabled?: boolean }>();
const { data: options, error, refresh } = await useFetch<GallerySubmissionOptions>("/api/gallery/submission-options");
const adding = ref(false);
const candidate = ref<string>();
const modes = [{label:"图标",value:"icon"},{label:"图片",value:"image"}];
const sidebarModes = [{label:"简洁底纹",value:"plain"},{label:"图库图片",value:"image"}];
const iconLabels: Record<string,string> = { photo:"图片", "device-desktop":"屏幕",camera:"摄影",palette:"插画",user:"人物",mountain:"风景",building:"建筑","device-gamepad-2":"游戏",book:"作品",sparkles:"星光",category:"分类","device-mobile":"手机",leaf:"自然",heart:"心形",tag:"标签",square:"方形" };
const icons: Array<{ value: string; label: string; icon: string }> = galleryPresentationIcons.map(value => ({value,label:iconLabels[value]!,icon:`i-tabler-${value}`}));
const candidates = computed(() => [
  ...(options.value?.categories || []).map(node => ({value:`category:${node.id}`,label:`分类 · ${node.name}`,kind:"category" as const,nodeId:node.id})),
  ...(options.value?.facets || []).flatMap(group => group.values.map(node => ({value:`facet_value:${node.id}`,label:`${group.name} · ${node.name}`,kind:"facet_value" as const,nodeId:node.id}))),
]);
const available = computed(() => candidates.value.filter(item => !model.value.classification.some(rule => `${rule.kind}:${rule.nodeId}`===item.value)));
function label(rule: GalleryPresentationRule) { return candidates.value.find(item => item.value === `${rule.kind}:${rule.nodeId}`)?.label || "已停用的分类条目"; }
function add() {
  const item=available.value.find(item=>item.value===candidate.value);if(!item)return;
  model.value.classification.push({kind:item.kind,nodeId:item.nodeId,mode:"icon",icon:"photo",imageId:"",assetId:""});candidate.value=undefined;adding.value=false;
}
function modeChanged(rule: GalleryPresentationRule) { if(rule.mode==='icon'){rule.imageId="";rule.assetId="";} }
function sidebarChanged() { if(model.value.sidebar.mode==='plain'){model.value.sidebar.imageId="";model.value.sidebar.assetId="";} }
</script>
<template>
 <div class="space-y-5">
  <SettingSection title="左侧导航背景"><div class="grid gap-4 md:grid-cols-2"><UFormField label="背景样式"><USelect v-model="model.sidebar.mode" :items="sidebarModes" class="w-full" :disabled="disabled" @update:model-value="sidebarChanged" /></UFormField><UFormField label="背景遮罩" hint="55%–95%"><UInput v-model.number="model.sidebar.overlay" type="number" min="0.55" max="0.95" step="0.01" class="w-full" :disabled="disabled || model.sidebar.mode !== 'image'" /></UFormField><UFormField v-if="model.sidebar.mode === 'image'" label="背景图片" required class="md:col-span-2"><GalleryImagePicker :image-id="model.sidebar.imageId" :asset-id="model.sidebar.assetId" :disabled="disabled" @select="Object.assign(model.sidebar, $event)" /></UFormField></div></SettingSection>
  <SettingSection title="分类与维度展示" description="默认使用图标。为指定条目选择图片后，图库筛选和投稿选择器会使用同一张封面。"><template #actions><UButton icon="i-tabler-plus" label="添加展示条目" variant="outline" size="sm" :disabled="disabled || !available.length" @click="void (adding = true)" /></template>
   <UAlert v-if="error" title="分类选项加载失败" color="error"><template #actions><UButton label="重试" @click="refresh()" /></template></UAlert>
   <p v-if="!model.classification.length" class="text-sm text-muted">全部分类使用默认图标。添加条目可自定义图标或图片。</p>
   <div v-for="(rule,index) in model.classification" :key="`${rule.kind}:${rule.nodeId}`" class="grid items-center gap-3 border-b border-default py-4 first:pt-0 last:border-0 sm:grid-cols-[minmax(0,1fr)_7rem_minmax(0,1fr)_auto]">
    <span class="text-sm font-medium">{{ label(rule) }}</span><USelect v-model="rule.mode" :items="modes" :aria-label="`${label(rule)}展示方式`" :disabled="disabled" @update:model-value="modeChanged(rule)" /><GalleryImagePicker v-if="rule.mode === 'image'" :image-id="rule.imageId" :asset-id="rule.assetId" :disabled="disabled" @select="Object.assign(rule, $event)" /><USelect v-else v-model="rule.icon" :items="icons" :icon="`i-tabler-${rule.icon}`" :aria-label="`${label(rule)}图标`" :disabled="disabled" /><UButton icon="i-tabler-trash" color="neutral" variant="ghost" :aria-label="`移除${label(rule)}展示配置`" :disabled="disabled" @click="void (model.classification.splice(index,1))" />
   </div>
  </SettingSection>
  <UModal v-model:open="adding" title="选择展示条目"><template #body><UFormField label="分类或维度值"><USelectMenu v-model="candidate" :items="available" value-key="value" placeholder="搜索分类、作品或角色" class="w-full" /></UFormField></template><template #footer><UButton label="添加" :disabled="!candidate" @click="add" /></template></UModal>
 </div>
</template>
