<script setup lang="ts">
import type { GalleryCreatorProfile } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "创作者准入" });
const { isAdmin } = useAuth();
const { call } = useApi();
const toast = useToast();
const { data, pending, error, refresh } = await useAsyncData(
  "gallery-creator-applications",
  () => call<{ creators: GalleryCreatorProfile[] }>("/api/v1/gallery/admin/creators"),
  { server: false, default: () => ({ creators: [] }) },
);
const notes = reactive<Record<string, string>>({});
const reviewing = ref("");

async function review(creator: GalleryCreatorProfile, decision: "approve" | "reject") {
  reviewing.value = creator.id + decision;
  try {
    await call(`/api/v1/gallery/admin/creators/${encodeURIComponent(creator.id)}/review`, {
      method: "POST",
      body: { decision, note: notes[creator.id] || "" },
    });
    toast.add({ title: decision === "approve" ? "创作者已启用" : "申请已驳回", color: "success" });
    await refresh();
  } catch (error: any) {
    toast.add({ title: "操作失败", description: error?.data?.message || error?.message, color: "error" });
  } finally {
    reviewing.value = "";
  }
}
</script>

<template>
  <div class="mx-auto w-full max-w-5xl space-y-6 p-4 sm:p-6 lg:p-8">
    <header>
      <p class="text-sm font-medium text-primary">Creator Admission</p>
      <h1 class="mt-1 text-2xl font-semibold text-highlighted">创作者准入</h1>
      <p class="mt-2 text-sm text-muted">准入只授予本站的作品发布资格，不改变账号系统中的全局角色。</p>
    </header>
    <UAlert v-if="!isAdmin" color="error" variant="subtle" icon="i-tabler-lock" title="当前账号不是图库管理员" />
    <UAlert v-else-if="error" color="error" variant="subtle" icon="i-tabler-alert-circle" title="无法读取创作者申请" description="申请队列状态未知，请刷新后重试。" />
    <div v-else-if="pending" class="space-y-3"><USkeleton v-for="i in 4" :key="i" class="h-32 rounded-xl" /></div>
    <div v-else-if="data?.creators.length" class="space-y-3">
      <article v-for="creator in data.creators" :key="creator.id" class="rounded-2xl border border-default bg-default p-5">
        <div class="flex flex-col gap-5 sm:flex-row sm:items-start">
          <div class="grid size-12 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary"><UIcon name="i-tabler-user" class="size-6" /></div>
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <h2 class="font-semibold text-highlighted">{{ creator.displayName }}</h2>
              <span class="text-sm text-muted">@{{ creator.handle }}</span>
              <UBadge :color="creator.status === 'pending' ? 'warning' : 'error'" variant="soft">{{ creator.status === 'pending' ? '待审核' : '曾被驳回' }}</UBadge>
            </div>
            <p class="mt-3 text-sm leading-6 text-muted">{{ creator.applicationNote || "未填写申请说明" }}</p>
            <UTextarea v-model="notes[creator.id]" :rows="2" placeholder="审核备注（驳回时建议填写）" class="mt-4 w-full" />
          </div>
          <div class="grid shrink-0 grid-cols-2 gap-2 sm:w-48 sm:grid-cols-1">
            <UButton icon="i-tabler-check" label="通过申请" :loading="reviewing === creator.id + 'approve'" @click="review(creator, 'approve')" />
            <UButton color="error" variant="soft" label="驳回" :loading="reviewing === creator.id + 'reject'" @click="review(creator, 'reject')" />
          </div>
        </div>
      </article>
    </div>
    <div v-else class="rounded-2xl border border-dashed border-default py-20 text-center text-sm text-muted">当前没有待处理的创作者申请。</div>
  </div>
</template>
