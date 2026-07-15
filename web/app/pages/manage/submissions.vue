<script setup lang="ts">
import type { GallerySubmission } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "投稿与审核 · 图库管理" });
const { call } = useApi();
const page = ref(1);
const note = ref<Record<string, string>>({});
const acting = ref("");
const { data, pending, error, refresh } = await useAsyncData("gallery-manage-submissions", () => call<{ submissions: GallerySubmission[]; total: number }>(`/api/v1/gallery/admin/submissions?page=${page.value}&size=20`), { server: false, watch: [page], default: () => ({ submissions: [], total: 0 }) });
async function review(item: GallerySubmission, decision: "approve" | "reject") {
  acting.value = item.id;
  try { await call(`/api/v1/gallery/admin/submissions/${encodeURIComponent(item.id)}/review`, { method: "POST", body: { decision, note: note.value[item.id] || "" } }); await refresh(); }
  finally { acting.value = ""; }
}
</script>

<template>
  <div>
    <ManageHeader title="投稿与审核"><template #subtitle>Guest 投稿、安全不确定和近重复在这里人工处理；登录投稿只有不确定项进入队列。</template></ManageHeader>
    <UAlert class="mb-5" color="info" variant="subtle" icon="i-tabler-info-circle" title="发布仍受媒体门禁约束" description="只有 processing=ready、public rendition ready 且安全结论可接受的投稿才能批准；接口会拒绝提前发布。" />
    <SkeletonList v-if="pending" :rows="6" />
    <UAlert v-else-if="error" color="error" variant="subtle" title="审核队列加载失败"><template #actions><UButton label="重试" @click="() => refresh()" /></template></UAlert>
    <div v-else-if="data.submissions.length" class="space-y-3">
      <article v-for="item in data.submissions" :key="item.id" class="rounded-lg border border-default bg-default p-4">
        <div class="flex flex-wrap items-start justify-between gap-3"><div><h2 class="font-medium text-highlighted">{{ item.title }}</h2><p class="mt-1 text-xs text-muted">{{ item.processingState }} · {{ item.safetyState }} · {{ item.reviewState }}</p></div><UBadge :color="item.processingState === 'ready' ? 'success' : 'warning'" variant="soft" :label="item.processingState === 'ready' ? '媒体已就绪' : '等待媒体处理'" /></div>
        <p v-if="item.description" class="mt-3 text-sm leading-6 text-toned">{{ item.description }}</p>
        <div class="mt-4 grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto]"><UInput v-model="note[item.id]" placeholder="拒绝原因或审核备注" /><div class="flex gap-2"><UButton color="error" variant="outline" label="拒绝" :loading="acting === item.id" @click="review(item, 'reject')" /><UButton label="批准" :loading="acting === item.id" :disabled="item.processingState !== 'ready'" @click="review(item, 'approve')" /></div></div>
      </article>
    </div>
    <ManageEmpty v-else icon="i-tabler-circle-check" title="没有待审投稿" description="明确安全的已验证用户投稿可直接发布，不会进入这个队列。" />
  </div>
</template>
