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
    <ManageHeader title="专题集合"
      ><template #subtitle
        >Gallery-local Collection v0；每个集合只包含
        gallery.image，删除成员或集合都不会删除图片。</template
      ><template #actions
        ><UButton
          icon="i-tabler-plus"
          label="新建专题"
          @click="
            createOpen = true;
            void 0;
          " /></template
    ></ManageHeader>
    <UAlert
      class="mb-5"
      color="info"
      variant="subtle"
      icon="i-tabler-code"
      title="干净的 Collection kernel"
      description="slug、封面和 SEO 留在 Gallery extension；membership 以 expectedVersion 原子增删。重排使用独立操作，不把动态排序塞进核心表。"
    />
    <SkeletonList v-if="!hydrated || pending" :rows="5" />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="专题加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <div
      v-else-if="data.collections.length"
      class="divide-y divide-default border-y border-default"
    >
      <article
        v-for="item in data.collections"
        :key="item.id"
        class="grid gap-3 py-4 sm:grid-cols-[minmax(0,1fr)_auto]"
      >
        <div>
          <div class="flex items-center gap-2">
            <h2 class="font-medium text-highlighted">{{ item.name }}</h2>
            <UBadge
              :color="item.visibility === 'public' ? 'success' : 'neutral'"
              variant="soft"
              :label="item.visibility"
            />
          </div>
          <p class="mt-1 text-sm text-muted">
            {{ item.description || "暂无说明" }}
          </p>
          <p class="mt-2 text-xs text-dimmed">
            {{ item.itemCount }} 张 · version {{ item.version }}
          </p>
        </div>
        <UButton
          :to="`/collections/${item.slug}`"
          target="_blank"
          color="neutral"
          variant="ghost"
          icon="i-tabler-external-link"
          label="公开页"
        />
      </article>
    </div>
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
