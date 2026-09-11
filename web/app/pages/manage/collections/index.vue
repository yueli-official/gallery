<script setup lang="ts">
import { ManagePage } from "@yueli/ui/admin";
import { CollectionHeaderTools, CollectionPaginationBar } from "@yueli/ui/collection/pattern";
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
const createError = ref("");
const slugTouched = ref(false);
const form = reactive({
  name: "",
  slug: "",
  description: "",
  visibility: "private",
});
const { data, pending, error, refresh } = await useAsyncData(
  "gallery-manage-collections",
  () =>
    call<{ items: GalleryCollection[] }>(
      "/admin/collections",
    ),
  { server: false, default: () => ({ items: [] }) },
);
const search = ref("");
const visibility = ref<"all" | "public" | "private">("all");
const filteredCollections = computed(() => {
  const query = search.value.trim().toLowerCase();
  return data.value.items.filter(
    (item) =>
      (visibility.value === "all" || item.visibility === visibility.value) &&
      (!query ||
        `${item.name} ${item.description} ${item.slug || ""}`
          .toLowerCase()
          .includes(query)),
  );
});

const collectionPage = ref(1);
const collectionPageSize = ref(20);
const pagedCollections = computed(() => filteredCollections.value.slice((collectionPage.value - 1) * collectionPageSize.value, collectionPage.value * collectionPageSize.value));
watch([search, visibility, collectionPageSize], () => { collectionPage.value = 1; });
watch(() => filteredCollections.value.length, total => { collectionPage.value = Math.min(collectionPage.value, Math.max(1, Math.ceil(total / collectionPageSize.value))); });

function clientSlug(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "");
}

watch(
  () => form.name,
  (name) => {
    if (!slugTouched.value) form.slug = clientSlug(name);
  },
);

function openCreateCollection() {
  createError.value = "";
  slugTouched.value = false;
  Object.assign(form, {
    name: "",
    slug: "",
    description: "",
    visibility: "private",
  });
  createOpen.value = true;
}

function closeCreateCollection() {
  createOpen.value = false;
}

async function createCollection() {
  if (!canManageCollections.value || !form.name.trim() || !form.slug.trim())
    return;
  creating.value = true;
  createError.value = "";
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
  } catch (reason: any) {
    createError.value =
      galleryFailureMessage(reason, "专题没有创建，请重试。");
  } finally {
    creating.value = false;
  }
}
</script>

<template>
  <ManagePage id="collections" title="专题" icon="i-tabler-folders">
    <template #tools><CollectionHeaderTools v-model:search="search" label="专题搜索与筛选" search-placeholder="搜索名称、slug 或说明…" :filter-count="visibility === 'all' ? 0 : 1" :controls="[{kind:'select',id:'visibility',label:'可见性',value:visibility,options:[{label:'全部',value:'all'},{label:'公开',value:'public'},{label:'私密',value:'private'}]}]" @filters="values => { visibility = String(values.visibility) as typeof visibility }" /></template>
    <template #actions>
      <UButton
        v-if="canManageCollections"
        icon="i-tabler-plus"
        label="新建专题"
        @click="openCreateCollection"
      />
    </template>
    <SkeletonList v-if="!hydrated || pending" :rows="5" />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="专题加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <section
      v-else-if="data.items.length"
      class="overflow-hidden rounded-xl border border-default bg-default"
      aria-label="专题列表"
    >
      <div
        class="grid grid-cols-[minmax(0,1fr)_5rem_5rem] items-center gap-3 border-b border-default bg-elevated/50 px-4 py-2 text-xs font-medium text-muted md:grid-cols-[minmax(0,1fr)_6rem_8rem_5rem]"
      >
        <span>专题</span>
        <span>图片</span>
        <span class="hidden md:inline">状态</span>
        <span class="text-right">操作</span>
      </div>
      <article
        v-for="item in pagedCollections"
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
      <div class="border-t border-default p-3 sm:px-4">
        <CollectionPaginationBar :page="collectionPage" :page-size="collectionPageSize" :total="filteredCollections.length" :page-sizes="[20, 40, 60]" page-size-control="每页专题数量" label="专题分页" @page-change="collectionPage = $event" @page-size-change="collectionPageSize = $event" />
      </div>
    </section>
    <ManageEmpty
      v-else
      icon="i-tabler-folders"
      title="还没有专题"
      description="从私有专题开始整理，准备好后再公开。"
    />
    <USlideover
      v-if="canManageCollections"
      v-model:open="createOpen"
      title="新建专题"
    >
      <template #body>
        <form
          id="create-gallery-collection"
          class="space-y-4"
          @submit.prevent="createCollection"
        >
          <UAlert
            v-if="createError"
            color="error"
            variant="subtle"
            icon="i-tabler-alert-circle"
            title="创建失败"
            :description="createError"
          />
          <UFormField label="名称" required>
            <UInput
              v-model="form.name"
              placeholder="专题名称"
              class="w-full"
              autofocus
            />
          </UFormField>
          <UFormField
            label="Slug"
            help="英文名称会自动生成；中文名称请手动填写。"
            required
          >
            <UInput
              v-model="form.slug"
              placeholder="例如 night-colors"
              class="w-full"
              @input="slugTouched = true"
            />
          </UFormField>
          <UFormField label="说明">
            <UTextarea
              v-model="form.description"
              :rows="4"
              class="w-full"
            />
          </UFormField>
          <UFormField label="可见性">
            <USelect
              v-model="form.visibility"
              :items="[
                { label: '私有', value: 'private' },
                { label: '公开', value: 'public' },
              ]"
              value-key="value"
              class="w-full"
            />
          </UFormField>
        </form>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton
            color="neutral"
            variant="ghost"
            label="取消"
            @click="closeCreateCollection"
          />
          <UButton
            form="create-gallery-collection"
            type="submit"
            label="创建专题"
            :loading="creating"
            :disabled="!form.name.trim() || !form.slug.trim()"
          />
        </div>
      </template>
    </USlideover>
  </ManagePage>
</template>
