<script setup lang="ts">
import type { GalleryCase } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "处理单 · 图库管理" });
const { call } = useApi();
const statusFilter = ref("open");
const notes = ref<Record<string, string>>({});
const acting = ref("");
const { data, pending, error, refresh } = await useAsyncData("gallery-manage-cases", () => call<{ cases: GalleryCase[]; total: number }>(`/api/v1/gallery/admin/cases?status=${statusFilter.value}&page=1&size=60`), { server: false, watch: [statusFilter], default: () => ({ cases: [], total: 0 }) });
const tabs = [{ label: "待处理", value: "open" }, { label: "处理中", value: "reviewing" }, { label: "已解决", value: "resolved" }, { label: "已忽略", value: "dismissed" }];
async function resolve(item: GalleryCase, status: "reviewing" | "resolved" | "dismissed") {
  acting.value = item.id;
  try { await call(`/api/v1/gallery/admin/cases/${encodeURIComponent(item.id)}/resolve`, { method: "POST", body: { status, note: notes.value[item.id] || "" } }); await refresh(); }
  finally { acting.value = ""; }
}
</script>

<template>
  <div>
    <ManageHeader title="处理单"><template #subtitle>举报、来源修正、安全不确定、近重复与下架调查使用同一工作队列；举报次数不会自动改变发布状态。</template></ManageHeader>
    <div class="mb-5 flex gap-1 overflow-x-auto"><UButton v-for="tab in tabs" :key="tab.value" color="neutral" :variant="statusFilter === tab.value ? 'soft' : 'ghost'" :label="tab.label" @click="() => { statusFilter = tab.value; }" /></div>
    <SkeletonList v-if="pending" :rows="6" />
    <UAlert v-else-if="error" color="error" variant="subtle" title="处理单加载失败"><template #actions><UButton label="重试" @click="() => refresh()" /></template></UAlert>
    <div v-else-if="data.cases.length" class="divide-y divide-default border-y border-default">
      <article v-for="item in data.cases" :key="item.id" class="grid gap-3 py-4 sm:grid-cols-[9rem_minmax(0,1fr)]"><div><UBadge color="neutral" variant="soft" :label="item.kind" /><p class="mt-2 text-xs text-muted">{{ item.status }}</p></div><div><div class="flex items-start justify-between gap-3"><div><p class="font-medium text-highlighted">{{ item.reason || (item.kind === 'source_correction' ? '来源修正' : '待运营复核') }}</p><p v-if="item.description" class="mt-1 text-sm text-toned">{{ item.description }}</p><a v-if="item.proposedSourceUrl" :href="item.proposedSourceUrl" target="_blank" rel="noopener noreferrer nofollow" class="mt-2 block break-all text-sm text-primary hover:underline">{{ item.proposedSourceUrl }}</a></div><UButton v-if="item.imageId" :to="`/images/${item.imageId}`" target="_blank" color="neutral" variant="ghost" icon="i-tabler-external-link" label="查看" /></div><div v-if="['open', 'reviewing'].includes(item.status)" class="mt-4 grid gap-2 lg:grid-cols-[minmax(0,1fr)_auto]"><UInput v-model="notes[item.id]" placeholder="处理结论（解决或忽略时必填）" /><div class="flex flex-wrap gap-2"><UButton v-if="item.status === 'open'" color="neutral" variant="outline" label="开始处理" :loading="acting === item.id" @click="resolve(item, 'reviewing')" /><UButton color="neutral" variant="outline" label="忽略" :loading="acting === item.id" @click="resolve(item, 'dismissed')" /><UButton label="解决" :loading="acting === item.id" @click="resolve(item, 'resolved')" /></div></div><p v-else-if="item.resolutionNote" class="mt-3 text-sm text-muted">结论：{{ item.resolutionNote }}</p></div></article>
    </div>
    <ManageEmpty v-else icon="i-tabler-flag-off" title="当前队列为空" description="选择其他状态可以查看已处理记录。" />
  </div>
</template>
