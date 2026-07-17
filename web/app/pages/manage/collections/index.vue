<script setup lang="ts">
import {
  ManageEmpty,
  ManageHeader,
  SkeletonList,
} from "@platform/manage/components";
import type { GalleryCollection } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
useSeoMeta({ title: "专题集合 · 图库管理" });
const { call } = useApi();
const hydrated = useClientHydrated();
const createOpen = ref(false);
const creating = ref(false);
const form = reactive({
  name: "",
  slug: "",
  description: "",
  visibility: "private",
});
const { data, pending, error, refresh } = await useAsyncData(
  "gallery-manage-collections",
  () =>
    call<{ collections: GalleryCollection[] }>(
      "/api/v1/gallery/admin/collections",
    ),
  { server: false, default: () => ({ collections: [] }) },
);
const publicCount = computed(
  () =>
    data.value.collections.filter((item) => item.visibility === "public")
      .length,
);
const imageCount = computed(() =>
  data.value.collections.reduce((total, item) => total + item.itemCount, 0),
);
async function createCollection() {
  if (!form.name.trim() || !form.slug.trim()) return;
  creating.value = true;
  try {
    await call("/api/v1/gallery/admin/collections", {
      method: "POST",
      body: form,
    });
    createOpen.value = false;
    Object.assign(form, {
      name: "",
      slug: "",
      description: "",
      visibility: "private",
    });
    await refresh();
  } finally {
    creating.value = false;
  }
}
</script>

<template>
  <div>
    <ManageHeader title="专题策展"
      ><template #subtitle
        >把已经通过审核的图片组织成有封面、有顺序、有公开叙事的专题。</template
      ><template #actions
        ><UButton
          icon="i-tabler-plus"
          label="新建专题"
          @click="
            createOpen = true;
            void 0;
          " /></template
    ></ManageHeader>
    <div
      v-if="hydrated && !pending && !error"
      class="mb-5 grid overflow-hidden rounded-xl border border-default bg-default sm:grid-cols-3"
      aria-label="专题概况"
    >
      <div class="px-4 py-3 sm:border-r sm:border-default">
        <p class="text-xs text-muted">专题总数</p>
        <p class="mt-1 text-xl font-semibold tabular-nums text-highlighted">
          {{ data.collections.length }}
        </p>
      </div>
      <div class="border-t border-default px-4 py-3 sm:border-r sm:border-t-0">
        <p class="text-xs text-muted">已公开</p>
        <p class="mt-1 text-xl font-semibold tabular-nums text-highlighted">
          {{ publicCount }}
        </p>
      </div>
      <div class="border-t border-default px-4 py-3 sm:border-t-0">
        <p class="text-xs text-muted">收录关系</p>
        <p class="mt-1 text-xl font-semibold tabular-nums text-highlighted">
          {{ imageCount }}
        </p>
      </div>
    </div>
    <SkeletonList v-if="!hydrated || pending" :rows="5" />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="专题加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <section
      v-else-if="data.collections.length"
      class="overflow-hidden rounded-xl border border-default bg-default"
      aria-label="专题列表"
    >
      <article
        v-for="item in data.collections"
        :key="item.id"
        class="grid gap-4 border-b border-default p-3 last:border-b-0 sm:grid-cols-[10rem_minmax(0,1fr)_auto] sm:items-center sm:p-4"
      >
        <NuxtLink
          :to="`/manage/collections/${encodeURIComponent(item.id)}`"
          class="group relative block h-28 overflow-hidden rounded-lg bg-elevated"
          :aria-label="`编辑专题：${item.name}`"
        >
          <img
            v-if="item.coverAssetId"
            v-bind="galleryImageSources(item.coverAssetId, 'grid', false)"
            :alt="item.coverAltText || item.name"
            class="size-full object-cover transition duration-300 group-hover:scale-[1.02]"
          />
          <span v-else class="grid size-full place-items-center text-muted">
            <UIcon name="i-tabler-photo-plus" class="size-7" />
          </span>
          <span
            class="absolute bottom-2 left-2 rounded bg-default/90 px-2 py-1 text-[11px] font-medium text-default backdrop-blur"
          >
            {{ item.itemCount }} 张
          </span>
        </NuxtLink>
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h2 class="truncate font-semibold text-highlighted">
              {{ item.name }}
            </h2>
            <span
              class="inline-flex items-center gap-1 text-xs"
              :class="
                item.visibility === 'public' ? 'text-success' : 'text-muted'
              "
            >
              <span class="size-1.5 rounded-full bg-current" />
              {{ item.visibility === "public" ? "公开" : "私有草稿" }}
            </span>
          </div>
          <p class="mt-1 line-clamp-2 text-sm leading-5 text-muted">
            {{ item.description || "暂无说明" }}
          </p>
          <p class="mt-2 truncate font-mono text-xs text-dimmed">
            /collections/{{ item.slug || "未设置-slug" }}
          </p>
        </div>
        <div class="flex items-center gap-1 sm:justify-end">
          <UButton
            :to="`/manage/collections/${encodeURIComponent(item.id)}`"
            icon="i-tabler-pencil"
            label="继续策展"
          />
          <UButton
            v-if="item.visibility === 'public' && item.slug"
            :to="`/collections/${item.slug}`"
            target="_blank"
            color="neutral"
            variant="ghost"
            icon="i-tabler-external-link"
            aria-label="查看公开专题"
          />
        </div>
      </article>
    </section>
    <ManageEmpty
      v-else
      icon="i-tabler-folders"
      title="还没有专题"
      description="从私有专题开始整理，准备好后再公开。"
    />
    <UModal
      v-model:open="createOpen"
      title="新建专题"
      description="专题默认可以保持私有；公开后会出现在前台。"
      ><template #body
        ><form class="space-y-4" @submit.prevent="createCollection">
          <UFormField label="名称" required
            ><UInput v-model="form.name" /></UFormField
          ><UFormField label="Slug" required
            ><UInput
              v-model="form.slug"
              placeholder="night-colors" /></UFormField
          ><UFormField label="说明"
            ><UTextarea v-model="form.description" :rows="4" /></UFormField
          ><UFormField label="可见性"
            ><USelect
              v-model="form.visibility"
              :items="[
                { label: '私有', value: 'private' },
                { label: '公开', value: 'public' },
              ]"
              value-key="value"
          /></UFormField>
          <div class="flex justify-end gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              label="取消"
              @click="
                createOpen = false;
                void 0;
              "
            /><UButton type="submit" label="创建专题" :loading="creating" />
          </div></form></template
    ></UModal>
  </div>
</template>
