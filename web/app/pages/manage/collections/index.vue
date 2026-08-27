<script setup lang="ts">
import { PageHeader } from "@yueli/ui/dashboard/pattern";
import { CollectionTableToolbar } from "@yueli/ui/collection/pattern";
import { ManageEmpty, SkeletonList } from "~/utils/manageComponents";
import type { GalleryCollection } from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "专题集合 · 图库管理" });
const { call } = useGalleryApi();
const { can } = useGalleryMe();
const hydrated = useClientHydrated();
const canManageCollections = computed(() => can("gallery.collection.manage"));
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
      "/admin/collections",
    ),
  { server: false, default: () => ({ collections: [] }) },
);
const search = ref("");
const visibility = ref<"all" | "public" | "private">("all");
const filteredCollections = computed(() => {
  const query = search.value.trim().toLowerCase();
  return data.value.collections.filter(
    (item) =>
      (visibility.value === "all" || item.visibility === visibility.value) &&
      (!query ||
        `${item.name} ${item.description} ${item.slug || ""}`
          .toLowerCase()
          .includes(query)),
  );
});
async function createCollection() {
  if (!canManageCollections.value || !form.name.trim() || !form.slug.trim())
    return;
  creating.value = true;
  try {
    await call("/admin/collections", {
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
    <PageHeader title="专题策展"
      ><template #actions
        ><UButton
          v-if="canManageCollections"
          icon="i-tabler-plus"
          label="新建专题"
          @click="
            createOpen = true;
            void 0;
          " /></template
    ></PageHeader>
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
      <CollectionTableToolbar
        v-model:search="search"
        label="专题列表工具栏"
        search-placeholder="搜索名称、slug 或说明…"
        search-action="搜索"
        filter-label="筛选"
        :filter-count="visibility === 'all' ? 0 : 1"
      >
        <template #filters>
          <div class="w-64 max-w-[calc(100vw-2rem)]">
            <UFormField label="可见性">
              <USelect
                v-model="visibility"
                :items="[
                  { label: '全部', value: 'all' },
                  { label: '公开', value: 'public' },
                  { label: '私有', value: 'private' },
                ]"
                value-key="value"
                class="w-full"
              />
            </UFormField>
          </div>
        </template>
      </CollectionTableToolbar>
      <div
        class="grid grid-cols-[minmax(0,1fr)_5rem_5rem] items-center gap-3 border-b border-default bg-elevated/50 px-4 py-2 text-xs font-medium text-muted md:grid-cols-[minmax(0,1fr)_6rem_8rem_5rem]"
      >
        <span>专题</span>
        <span>图片</span>
        <span class="hidden md:inline">状态</span>
        <span class="text-right">操作</span>
      </div>
      <article
        v-for="item in filteredCollections"
        :key="item.id"
        class="grid grid-cols-[minmax(0,1fr)_5rem_5rem] items-center gap-3 border-b border-default p-4 last:border-b-0 md:grid-cols-[minmax(0,1fr)_6rem_8rem_5rem]"
      >
        <div class="flex min-w-0 items-center gap-3">
          <div class="hidden aspect-[4/3] w-16 shrink-0 overflow-hidden rounded-lg bg-elevated sm:block">
            <img
              v-if="item.coverAssetId"
              v-bind="galleryImageSources(item.coverAssetId, 'grid', false)"
              :alt="item.coverAltText || item.name"
              class="size-full object-cover"
            />
            <span v-else class="grid size-full place-items-center text-muted">
              <UIcon name="i-tabler-photo-plus" class="size-5" />
            </span>
          </div>
          <div class="min-w-0">
            <NuxtLink
              :to="`/manage/collections/${encodeURIComponent(item.id)}`"
              class="block truncate text-sm font-medium text-highlighted hover:text-primary"
              :aria-label="`编辑专题：${item.name}`"
            >
              {{ item.name }}
            </NuxtLink>
            <p class="mt-1 truncate text-xs text-muted">
              /collections/{{ item.slug || "未设置-slug" }}
            </p>
          </div>
        </div>
        <span class="text-xs tabular-nums text-muted">{{ item.itemCount }}</span>
        <span class="hidden text-xs text-muted md:inline">
          {{ item.visibility === "public" ? "公开" : "私有" }}
        </span>
        <div class="flex items-center justify-end gap-1">
          <UTooltip
            v-if="item.visibility === 'public' && item.slug"
            text="查看公开专题"
          >
            <UButton
              :to="`/collections/${item.slug}`"
              target="_blank"
              rel="noopener"
              color="neutral"
              variant="ghost"
              size="xs"
              square
              icon="i-tabler-external-link"
              :aria-label="`查看公开专题：${item.name}`"
            />
          </UTooltip>
          <UTooltip :text="canManageCollections ? '编辑专题' : '查看专题'">
            <UButton
              :to="`/manage/collections/${encodeURIComponent(item.id)}`"
              :icon="canManageCollections ? 'i-tabler-pencil' : 'i-tabler-eye'"
              color="neutral"
              variant="ghost"
              size="xs"
              square
              :aria-label="`${canManageCollections ? '编辑' : '查看'}专题：${item.name}`"
            />
          </UTooltip>
        </div>
      </article>
      <div
        v-if="!filteredCollections.length"
        class="grid min-h-48 place-items-center px-6 py-10 text-center text-sm text-muted"
      >
        没有匹配的专题。
      </div>
    </section>
    <ManageEmpty
      v-else
      icon="i-tabler-folders"
      title="还没有专题"
      description="从私有专题开始整理，准备好后再公开。"
    />
    <UModal
      v-if="canManageCollections"
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
