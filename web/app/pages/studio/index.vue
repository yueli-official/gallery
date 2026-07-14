<script setup lang="ts">
import type { GalleryCreatorProfile, GalleryStudioArtwork } from "~/types/gallery";

definePageMeta({ middleware: "auth" });
useSeoMeta({ title: "创作中心" });

const { user } = useAuth();
const { call } = useApi();
const toast = useToast();

const { data: creatorData, pending, error: creatorError, refresh: refreshCreator } = await useAsyncData(
  "gallery-my-creator",
  () => call<{ creator: GalleryCreatorProfile | null }>("/api/v1/gallery/me/creator"),
  { server: false, default: () => ({ creator: null }) },
);
const creator = computed(() => creatorData.value?.creator ?? null);

const { data: artworkData, error: artworksError, refresh: refreshArtworks } = await useAsyncData(
  "gallery-my-artworks",
  async () => creator.value?.status === "active"
    ? call<{ artworks: GalleryStudioArtwork[] }>("/api/v1/gallery/me/artworks")
    : { artworks: [] },
  { server: false, watch: [() => creator.value?.status], default: () => ({ artworks: [] }) },
);

const application = reactive({
  handle: "",
  displayName: user.value?.name ?? "",
  applicationNote: "",
});
const submitting = ref(false);
const creating = ref(false);

async function apply() {
  submitting.value = true;
  try {
    await call("/api/v1/gallery/me/creator", { method: "POST", body: application });
    await refreshCreator();
    toast.add({ title: "申请已提交", description: "管理员审核通过后即可创建作品。", color: "success" });
  } catch (error: any) {
    toast.add({ title: "提交失败", description: error?.data?.message || error?.message || "请稍后重试", color: "error" });
  } finally {
    submitting.value = false;
  }
}

async function createArtwork() {
  creating.value = true;
  try {
    const value = await call<{ artwork: GalleryStudioArtwork }>("/api/v1/gallery/me/artworks", { method: "POST" });
    await navigateTo(`/studio/artworks/${value.artwork.id}`);
  } catch (error: any) {
    toast.add({ title: "创建失败", description: error?.data?.message || error?.message || "请稍后重试", color: "error" });
    creating.value = false;
  }
}

const statusMeta: Record<string, { label: string; color: "neutral" | "warning" | "success" | "error" }> = {
  draft: { label: "草稿", color: "neutral" },
  pending_review: { label: "审核中", color: "warning" },
  published: { label: "已发布", color: "success" },
  rejected: { label: "需修改", color: "error" },
};
</script>

<template>
  <div class="mx-auto w-full max-w-6xl space-y-8 py-4 sm:py-8">
    <header class="flex flex-col gap-4 border-b border-default pb-8 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p class="text-sm font-medium text-primary">Gallery Studio</p>
        <h1 class="mt-2 font-display text-3xl font-semibold text-highlighted">创作中心</h1>
        <p class="mt-2 text-sm text-muted">管理图片、作品信息与审核状态。</p>
      </div>
      <UButton v-if="creator?.status === 'active'" icon="i-tabler-plus" label="创建作品" :loading="creating" @click="createArtwork" />
    </header>

    <div v-if="pending" class="space-y-3">
      <USkeleton class="h-28 rounded-xl" />
      <USkeleton class="h-56 rounded-xl" />
    </div>

    <UAlert v-else-if="creatorError" color="error" variant="subtle" icon="i-tabler-alert-circle" title="无法读取创作者状态" description="请刷新页面；如果问题持续存在，请检查 Gallery API。" />

    <section v-else-if="!creator" class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]">
      <div class="rounded-2xl border border-default bg-default p-6 sm:p-8">
        <h2 class="text-xl font-semibold text-highlighted">申请成为创作者</h2>
        <p class="mt-2 text-sm leading-6 text-muted">创作者名称会显示在公开作品页；申请通过前不会开放发布权限。</p>
        <form class="mt-6 space-y-5" @submit.prevent="apply">
          <UFormField label="创作者标识" description="用于公开主页地址，仅建议使用小写字母、数字和短横线。" required>
            <UInput v-model="application.handle" pattern="[a-z0-9][a-z0-9-]{2,31}" minlength="3" maxlength="32" placeholder="例如 moonlit-studio" class="w-full" />
          </UFormField>
          <UFormField label="展示名称" required>
            <UInput v-model="application.displayName" placeholder="你的公开名称" class="w-full" />
          </UFormField>
          <UFormField label="申请说明" description="简单介绍计划发布的内容。">
            <UTextarea v-model="application.applicationNote" :rows="4" class="w-full" />
          </UFormField>
          <UButton type="submit" label="提交申请" :loading="submitting" />
        </form>
      </div>
      <aside class="rounded-2xl bg-elevated/60 p-6">
        <UIcon name="i-tabler-sparkles" class="size-6 text-primary" />
        <h2 class="mt-4 font-semibold text-highlighted">这里不限定创作主题</h2>
        <p class="mt-2 text-sm leading-6 text-muted">插画、摄影、设计、3D 与其他视觉作品都可以通过统一维度归类。</p>
      </aside>
    </section>

    <UAlert
      v-else-if="creator.status === 'pending'"
      color="warning"
      variant="subtle"
      icon="i-tabler-clock-hour-4"
      title="创作者申请审核中"
      :description="`@${creator.handle} 的申请已进入队列，审核通过后会在这里开放作品编辑器。`"
    />

    <div v-else-if="creator.status === 'rejected'" class="space-y-4">
      <UAlert color="error" variant="subtle" icon="i-tabler-alert-circle" title="创作者申请暂未通过" :description="creator.reviewNote || '管理员尚未填写原因。'" />
      <p class="text-sm text-muted">目前请联系站点管理员补充资料；重新申请流程会在后续版本开放。</p>
    </div>

    <UAlert v-else-if="creator.status === 'suspended'" color="error" variant="subtle" icon="i-tabler-user-off" title="创作者资格已暂停" description="当前不能新建或编辑作品，请联系站点管理员。" />

    <template v-else-if="creator.status === 'active'">
      <UAlert v-if="artworksError" color="error" variant="subtle" icon="i-tabler-alert-circle" title="无法读取作品列表" description="作品状态未知，请刷新后重试。" />
      <section class="rounded-2xl border border-default bg-default">
        <div class="flex items-center justify-between border-b border-default px-5 py-4">
          <div>
            <h2 class="font-semibold text-highlighted">我的作品</h2>
            <p class="mt-1 text-xs text-muted">@{{ creator.handle }} · {{ artworkData?.artworks.length ?? 0 }} 件</p>
          </div>
          <UButton color="neutral" variant="ghost" icon="i-tabler-refresh" aria-label="刷新作品" @click="() => refreshArtworks()" />
        </div>
        <div v-if="artworkData?.artworks.length" class="divide-y divide-default">
          <NuxtLink
            v-for="artwork in artworkData.artworks"
            :key="artwork.id"
            :to="`/studio/artworks/${artwork.id}`"
            class="group grid grid-cols-[4.5rem_minmax(0,1fr)_auto] items-center gap-4 px-5 py-4 transition hover:bg-elevated/50"
          >
            <div class="aspect-square overflow-hidden rounded-lg bg-elevated">
              <img v-if="artwork.assets[0]" :src="artwork.assets[0].thumbnailUrl" :alt="artwork.title" class="size-full object-cover" />
              <UIcon v-else name="i-tabler-photo" class="m-auto size-6 text-dimmed" />
            </div>
            <div class="min-w-0">
              <p class="truncate font-medium text-highlighted group-hover:text-primary">{{ artwork.title || "未命名作品" }}</p>
              <p v-if="artwork.reviewNote" class="mt-1 truncate text-xs text-error">{{ artwork.reviewNote }}</p>
              <p v-else class="mt-1 text-xs text-muted">{{ artwork.assets.length }} 张图片</p>
            </div>
            <UBadge :color="statusMeta[artwork.status]?.color || 'neutral'" variant="soft">{{ statusMeta[artwork.status]?.label || artwork.status }}</UBadge>
          </NuxtLink>
        </div>
        <div v-else class="px-6 py-16 text-center">
          <UIcon name="i-tabler-photo-plus" class="mx-auto size-8 text-dimmed" />
          <p class="mt-3 text-sm text-muted">还没有作品，创建第一份草稿吧。</p>
        </div>
      </section>
    </template>
  </div>
</template>
