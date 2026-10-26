<script setup lang="ts">
import { AssetVisualAssetField,AssetDraftNotice } from "@yueli/asset-nuxt/components";
import type { GalleryPresentationRule } from "~/types/presentation";

const model = defineModel<GalleryPresentationRule>({ required: true });
const props = defineProps<{ disabled?: boolean }>();
const emit = defineEmits<{ busy: [value: boolean] }>();
const { uploadPrepared,draft } = useGalleryAssetUpload(`gallery-visual:${useId()}`);
const file = shallowRef<File | null>(null);
const previewUrl = ref("");
const processing = ref(false);
const uploading = ref(false);
const progress = ref(0);
const error = ref("");
const field = ref<InstanceType<typeof AssetVisualAssetField>>();
const icon = computed({
  get: () => `i-tabler-${model.value.icon || 'photo'}`,
  set: (value: string) => { model.value.icon = value.replace(/^i-tabler-/, ""); },
});
const imageUrl = computed(() => model.value.assetId && model.value.delivery ? galleryRendition(model.value, "thumbnail") : "");
watch(() => processing.value || uploading.value, value => emit("busy", value));
function revokePreview() {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value);
  previewUrl.value = "";
}
function prepared(value: File) {
  revokePreview();
  file.value = value;
  previewUrl.value = URL.createObjectURL(value);
  model.value.mode = "image";
  error.value = "";
}
function clearImage() {
  revokePreview();
  file.value = null;
  model.value.imageId = "";
  model.value.assetId = "";
  model.value.delivery = undefined;
  model.value.mode = "icon";
  error.value = "";
}
async function prepare() {
  if (processing.value || uploading.value || props.disabled) return false;
  if (!field.value?.validate()) return false;
  if (model.value.mode === "image" && !file.value && !model.value.assetId) {
    error.value = "请选择一张封面，或改用图标展示。";
    return false;
  }
  if (!file.value) return true;
  uploading.value = true;
  error.value = "";
  try {
    const asset = await uploadPrepared(file.value, value => { progress.value = value; });
    model.value.assetId = asset.id;
    model.value.imageId = "";
    model.value.delivery = undefined;
    file.value = null;
    // Keep the local preview until the containing form closes or refreshes.
    return true;
  } catch (cause) {
    error.value = galleryFailureMessage(cause, "封面上传失败，请重试；图片草稿已保留。");
    return false;
  } finally { uploading.value = false; }
}
onBeforeUnmount(revokePreview);
defineExpose({ prepare });
</script>

<template>
  <div class="space-y-3">
    <AssetDraftNotice :expires-at="draft.expiresAt.value" :reset-required="draft.resetRequired.value" :disabled="processing || uploading" @restart="draft.restart" />
    <UFormField label="展示方式">
      <USelect v-model="model.mode" :items="[{label:'图标',value:'icon'},{label:'封面',value:'image'}]" :disabled="disabled || processing || uploading" class="w-full" />
    </UFormField>
    <AssetVisualAssetField ref="field" v-model:icon="icon" title="图标与封面" :image-url="imageUrl" :preview-url="previewUrl" :file="file"
      profile-key="gallery-submission" :image-options="{ aspectRatio: 1, maxOutputSide: 1024, outputType: 'image/webp', allowAnimation: false }"
      image-spec="WebP · 最长边 1024px" :disabled="disabled" :error="error" :uploading="uploading" :progress="progress"
      @processed="prepared" @clear-image="clearImage" @processing="processing = $event" />
  </div>
</template>
