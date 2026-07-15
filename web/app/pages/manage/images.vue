<script setup lang="ts">
import type { GalleryImageCard, GalleryImagePage } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "图片 · 图库管理" });
const { call } = useApi();
const q = ref("");
const page = ref(1);
const selected = ref<GalleryImageCard>();
const reason = ref("");
const hiding = ref(false);
const { data, pending, error, refresh } = await useAsyncData("gallery-manage-images", () => call<GalleryImagePage>(`/api/v1/gallery/images?q=${encodeURIComponent(q.value)}&page=${page.value}&size=24&sort=newest`), { server: false, watch: [page], default: () => ({ items: [], page: 1, pageSize: 24, total: 0, totalPages: 0 }) });
async function search() { page.value = 1; await refresh(); }
async function hide() {
  if (!selected.value || !reason.value.trim()) return;
  hiding.value = true;
  try { await call(`/api/v1/gallery/admin/images/${encodeURIComponent(selected.value.id)}/hide`, { method: "POST", body: { reason: reason.value.trim() } }); selected.value = undefined; reason.value = ""; await refresh(); }
  finally { hiding.value = false; }
}
</script>

<template>
  <div>
    <ManageHeader title="图片"><template #subtitle>检查当前公开目录；下架会立即从 Gallery 公开资格中移除，并建立已解决的 takedown Case。</template></ManageHeader>
    <div class="mb-5 flex gap-2"><UInput v-model="q" class="max-w-lg flex-1" icon="i-tabler-search" placeholder="搜索标题或说明" @keyup.enter="search" /><UButton label="搜索" color="neutral" variant="outline" @click="search" /></div>
    <SkeletonList v-if="pending" :rows="8" />
    <UAlert v-else-if="error" color="error" variant="subtle" title="图片加载失败"><template #actions><UButton label="重试" @click="() => refresh()" /></template></UAlert>
    <div v-else-if="data.items.length" class="divide-y divide-default border-y border-default">
      <article v-for="image in data.items" :key="image.id" class="grid grid-cols-[5rem_minmax(0,1fr)_auto] items-center gap-4 py-3">
        <img :src="galleryRendition(image.assetId, 'thumbnail')" :alt="image.altText" class="aspect-[4/3] w-20 rounded-md bg-elevated object-cover" />
		<div class="min-w-0"><p class="truncate font-medium text-highlighted">{{ image.title }}</p><p class="mt-1 text-xs text-muted">{{ image.primaryCategory || '未分类' }} · {{ compactMetric(image.metrics.views) }} 次浏览 · {{ compactMetric(image.metrics.favorites) }} 次收藏</p></div>
        <div class="flex gap-1"><UButton :to="`/images/${image.id}`" target="_blank" color="neutral" variant="ghost" icon="i-tabler-external-link" aria-label="打开公开图片" /><UButton color="error" variant="ghost" icon="i-tabler-eye-off" aria-label="下架图片" @click="() => { selected = image; }" /></div>
      </article>
    </div>
    <ManageEmpty v-else icon="i-tabler-photo-off" title="当前没有公开图片" description="只有满足处理、审核、安全和公开 rendition 条件的图片会显示。" />
    <div v-if="data.totalPages > 1" class="mt-5 flex justify-center gap-2"><UButton color="neutral" variant="outline" label="上一页" :disabled="page <= 1" @click="() => { page -= 1; }" /><span class="px-3 py-2 text-sm tabular-nums text-muted">{{ page }} / {{ data.totalPages }}</span><UButton color="neutral" variant="outline" label="下一页" :disabled="page >= data.totalPages" @click="() => { page += 1; }" /></div>
    <UModal :open="Boolean(selected)" title="下架图片" description="下架不会删除 private master，但公开列表会立即隐藏。" @update:open="value => { if (!value) selected = undefined }"><template #body><form class="space-y-4" @submit.prevent="hide"><UFormField label="原因" required><UTextarea v-model="reason" :rows="4" /></UFormField><div class="flex justify-end gap-2"><UButton color="neutral" variant="ghost" label="取消" @click="selected = undefined" /><UButton type="submit" color="error" label="确认下架" :loading="hiding" /></div></form></template></UModal>
  </div>
</template>
