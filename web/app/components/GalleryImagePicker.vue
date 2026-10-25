<script setup lang="ts">
import { CollectionPagination } from "@yueli/ui/collection/pattern";
import type { GalleryImageCard, GalleryImagePage } from "~/types/gallery";
const props = defineProps<{ imageId: string; assetId: string; disabled?: boolean }>();
const emit = defineEmits<{ select: [value: { imageId: string; assetId: string }] }>();
const open = ref(false);
const search = ref("");
const page = ref(1);
const query = ref("");
const { data: selected, error: selectedError, refresh: refreshSelected } = await useAsyncData(
  computed(() => `gallery-picker-image-${props.imageId}`),
  () => props.imageId ? $fetch<{ image?: GalleryImageCard }>(`/api/gallery/images/${encodeURIComponent(props.imageId)}`) : Promise.resolve({ image: undefined }),
);
useGalleryMediaRefresh(selected, refreshSelected);
const { data, error, status, execute } = await useFetch<GalleryImagePage>("/api/gallery/images", { query: computed(() => ({ q: query.value, page: page.value, size: 12 })), immediate: false, watch: false });
useGalleryMediaRefresh(data, execute);
watch(open, value => { if (value) void execute(); });
async function searchImages() { query.value=search.value.trim(); page.value=1; await execute(); }
async function changePage(value: number) { page.value=value;await execute(); }
function choose(image: GalleryImageCard) { emit("select", { imageId: image.id, assetId: image.assetId }); open.value=false; }
</script>
<template>
  <div class="flex items-center gap-3">
    <img v-if="selected?.image" :src="galleryRendition(selected.image, 'thumbnail')" alt="当前展示图片" class="size-14 rounded-md object-cover" />
    <UButton v-else-if="selectedError" label="重试预览" color="neutral" variant="ghost" size="sm" @click="refreshSelected()" />
    <UButton icon="i-tabler-photo" :label="imageId ? '更换图片' : '选择图片'" color="neutral" variant="outline" size="sm" :disabled="disabled" @click="void (open = true)" />
    <UModal v-model:open="open" title="从图库选择图片" description="只显示已公开且可正常访问的图片。" :ui="{ content: 'max-w-3xl' }">
      <template #body><div class="space-y-4"><form class="flex gap-2" @submit.prevent="searchImages"><UInput v-model="search" icon="i-tabler-search" placeholder="搜索图片标题" aria-label="搜索图库图片" class="flex-1" /><UButton type="submit" label="搜索" color="neutral" variant="outline" /></form><div v-if="status === 'pending'" class="grid grid-cols-3 gap-3"><USkeleton v-for="i in 6" :key="i" class="aspect-square" /></div><UAlert v-else-if="error" title="图片加载失败" color="error"><template #actions><UButton label="重试" @click="execute()" /></template></UAlert><p v-else-if="!data?.items.length" class="text-sm text-muted">没有可选图片，请先投稿并发布图片。</p><div v-else class="grid grid-cols-3 gap-3"><button v-for="image in data.items" :key="image.id" type="button" class="overflow-hidden rounded-md border border-default text-left hover:border-primary" @click="choose(image)"><img :src="galleryRendition(image, 'thumbnail')" :alt="image.title" class="aspect-square w-full object-cover" /><span class="block truncate p-2 text-xs">{{ image.title }}</span></button></div><CollectionPagination v-if="data && data.total > data.size" :model-value="data.page" :total-pages="galleryPageCount(data)" @update:model-value="changePage" /></div></template>
    </UModal>
  </div>
</template>
