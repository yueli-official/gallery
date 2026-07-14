<script setup lang="ts">
import type { GalleryStudioArtwork } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "作品审核" });
const { isAdmin } = useAuth();
const { call } = useApi();
const toast = useToast();
const { data, pending, error, refresh } = await useAsyncData(
  "gallery-review-queue",
  () => call<{ artworks: GalleryStudioArtwork[] }>("/api/v1/gallery/admin/reviews?status=pending_review"),
  { server: false, default: () => ({ artworks: [] }) },
);
const notes = reactive<Record<string, string>>({});
const reviewing = ref("");

async function review(artwork: GalleryStudioArtwork, decision: "approve" | "reject") {
  if (decision === "reject" && !notes[artwork.id]?.trim()) {
    toast.add({ title: "请填写驳回原因", color: "warning" });
    return;
  }
  reviewing.value = artwork.id + decision;
  try {
    await call(`/api/v1/gallery/admin/reviews/${encodeURIComponent(artwork.id)}`, {
      method: "POST",
      body: { decision, note: notes[artwork.id] || "" },
    });
    toast.add({ title: decision === "approve" ? "作品已发布" : "已退回创作者", color: "success" });
    await refresh();
  } catch (error: any) {
    toast.add({ title: "审核失败", description: error?.data?.message || error?.message, color: "error" });
  } finally {
    reviewing.value = "";
  }
}
</script>

<template>
  <div class="mx-auto w-full max-w-6xl space-y-6 p-4 sm:p-6 lg:p-8">
    <header>
      <p class="text-sm font-medium text-primary">Moderation Queue</p>
      <h1 class="mt-1 text-2xl font-semibold text-highlighted">作品审核</h1>
      <p class="mt-2 text-sm text-muted">只有通过审核的公开作品才会进入发现页。</p>
    </header>

    <UAlert v-if="!isAdmin" color="error" variant="subtle" icon="i-tabler-lock" title="当前账号不是图库管理员" />
    <UAlert v-else-if="error" color="error" variant="subtle" icon="i-tabler-alert-circle" title="无法读取审核队列" description="队列状态未知，请刷新后重试。" />
    <div v-else-if="pending" class="space-y-4"><USkeleton v-for="i in 3" :key="i" class="h-64 rounded-xl" /></div>
    <div v-else-if="data?.artworks.length" class="space-y-5">
      <article v-for="artwork in data.artworks" :key="artwork.id" class="overflow-hidden rounded-2xl border border-default bg-default">
        <div class="grid lg:grid-cols-[minmax(0,1fr)_22rem]">
          <div class="grid grid-cols-2 gap-1 bg-elevated p-1 sm:grid-cols-3">
            <img v-for="asset in artwork.assets" :key="asset.id" :src="asset.cardUrl" :alt="artwork.title" class="aspect-[4/3] size-full rounded-lg object-cover" />
          </div>
          <div class="flex flex-col p-5">
            <div class="min-w-0">
              <p class="text-xs text-muted">{{ artwork.creator?.displayName || artwork.creatorId }}</p>
              <h2 class="mt-1 text-lg font-semibold text-highlighted">{{ artwork.title }}</h2>
              <p class="mt-2 line-clamp-4 text-sm leading-6 text-muted">{{ artwork.description || "未填写作品说明" }}</p>
              <div class="mt-4 flex flex-wrap gap-2">
                <UBadge color="neutral" variant="soft">{{ artwork.contentRating }}</UBadge>
                <UBadge color="neutral" variant="soft">{{ artwork.aiUsage }}</UBadge>
                <UBadge color="neutral" variant="soft">{{ artwork.rightsBasis }}</UBadge>
              </div>
            </div>
            <div class="mt-auto pt-6">
              <UTextarea v-model="notes[artwork.id]" :rows="3" placeholder="驳回时必须说明需要修改的内容" class="w-full" />
              <div class="mt-3 grid grid-cols-2 gap-2">
                <UButton color="error" variant="soft" label="退回修改" :loading="reviewing === artwork.id + 'reject'" @click="review(artwork, 'reject')" />
                <UButton icon="i-tabler-check" label="通过并发布" :loading="reviewing === artwork.id + 'approve'" @click="review(artwork, 'approve')" />
              </div>
            </div>
          </div>
        </div>
      </article>
    </div>
    <div v-else class="rounded-2xl border border-dashed border-default py-20 text-center">
      <UIcon name="i-tabler-circle-check" class="mx-auto size-8 text-success" />
      <p class="mt-3 font-medium text-highlighted">审核队列已清空</p>
      <p class="mt-1 text-sm text-muted">新的提交会自动出现在这里。</p>
    </div>
  </div>
</template>
