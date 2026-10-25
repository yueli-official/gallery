<script setup lang="ts">
import { SettingSection } from "@yueli/ui/settings/pattern";
import type { GalleryAppearance } from "~/types/presentation";
const model = defineModel<GalleryAppearance>({ required: true });
defineProps<{ disabled?: boolean }>();
const sidebarModes = [{ label: "简洁底纹", value: "plain" }, { label: "图库图片", value: "image" }];
function sidebarChanged() {
  if (model.value.sidebar.mode === "plain") {
    model.value.sidebar.imageId = "";
    model.value.sidebar.assetId = "";
  }
}
</script>

<template>
  <div class="space-y-5">
    <SettingSection title="左侧导航背景">
      <div class="grid gap-4 md:grid-cols-2">
        <UFormField label="背景样式"><USelect v-model="model.sidebar.mode" :items="sidebarModes" class="w-full" :disabled="disabled" @update:model-value="sidebarChanged" /></UFormField>
        <UFormField label="背景遮罩" hint="55%–95%"><UInput v-model.number="model.sidebar.overlay" type="number" min="0.55" max="0.95" step="0.01" class="w-full" :disabled="disabled || model.sidebar.mode !== 'image'" /></UFormField>
        <UFormField v-if="model.sidebar.mode === 'image'" label="背景图片" required class="md:col-span-2"><GalleryImagePicker :image-id="model.sidebar.imageId" :asset-id="model.sidebar.assetId" :disabled="disabled" @select="Object.assign(model.sidebar, $event)" /></UFormField>
      </div>
    </SettingSection>
    <SettingSection title="分类与维度展示">
      <UButton to="/manage/classification" label="管理分类图标与封面" icon="i-tabler-category" color="neutral" variant="outline" />
    </SettingSection>
  </div>
</template>
