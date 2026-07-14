<script setup lang="ts">
import type { GalleryDiscovery } from "~/types/gallery";

const { data, error, status, refresh } = await useFetch<GalleryDiscovery>(
  "/api/gallery/discovery",
  { key: "gallery-discovery-shell" },
);

useSeoMeta({
  title: () => data.value?.site.name ?? "图片社区",
  description: () => data.value?.site.description ?? "发现和收藏创作者作品。",
});
</script>

<template>
  <div class="space-y-16">
    <section v-if="status === 'pending'" aria-label="正在加载作品">
      <div class="mb-8 space-y-3">
        <USkeleton class="h-9 w-48" />
        <USkeleton class="h-5 w-full max-w-xl" />
      </div>
      <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
        <USkeleton v-for="index in 10" :key="index" class="aspect-[4/5] rounded-xl" />
      </div>
    </section>

    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      icon="i-tabler-alert-circle"
      title="图片社区尚未完成初始化"
      description="发现内容暂时不可用，请检查 Gallery API 与站点配置。"
    >
      <template #actions>
        <UButton
          color="error"
          variant="soft"
          label="重新加载"
          @click="() => refresh()"
        />
      </template>
    </UAlert>

    <template v-else-if="data">
      <header class="max-w-3xl space-y-4">
        <p class="text-sm font-medium text-primary">多创作者图片社区</p>
        <h1 class="font-display text-4xl font-semibold tracking-tight text-highlighted sm:text-5xl">
          {{ data.site.title }}
        </h1>
        <p class="max-w-2xl text-base leading-7 text-toned sm:text-lg">
          {{ data.site.description }}
        </p>
      </header>

      <section v-if="data.featured.length" aria-labelledby="featured-heading">
        <div class="mb-6 flex items-end justify-between gap-4">
          <div>
            <p class="text-sm font-medium text-primary">有来源的人工策展</p>
            <h2 id="featured-heading" class="mt-1 text-2xl font-semibold text-highlighted">
              编辑精选
            </h2>
          </div>
        </div>
        <GalleryArtworkGrid :items="data.featured" priority />
      </section>

      <section aria-labelledby="latest-heading">
        <div class="mb-6">
          <p class="text-sm font-medium text-primary">按发布时间排序</p>
          <h2 id="latest-heading" class="mt-1 text-2xl font-semibold text-highlighted">
            最新发布
          </h2>
        </div>
        <GalleryArtworkGrid v-if="data.latest.length" :items="data.latest" />
        <div v-else class="rounded-xl border border-dashed border-default px-6 py-16 text-center">
          <UIcon name="i-tabler-photo-plus" class="mx-auto size-8 text-dimmed" />
          <h3 class="mt-4 font-semibold text-highlighted">等待第一件作品</h3>
          <p class="mt-2 text-sm text-muted">完成创作者准入与审核后，已发布作品会出现在这里。</p>
        </div>
      </section>
    </template>
  </div>
</template>
