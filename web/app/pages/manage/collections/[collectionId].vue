<script setup lang="ts">
import { ManageHeader, SkeletonList } from "@platform/manage/components";
import { createPlatformNotifier } from "@platform/ui/feedback";
import type {
  GalleryCollection,
  GalleryCollectionDetail,
  GalleryImageCard,
  GalleryImagePage,
} from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: "auth" });
const route = useRoute("/manage/collections/[collectionId]");
const { call } = useApi();
const toast = createPlatformNotifier(useToast());
const hydrated = useClientHydrated();
const collectionId = computed(() => String(route.params.collectionId));
const saving = ref(false);
const ordering = ref(false);
const metadataSaved = ref(false);
const orderSaved = ref(false);
const pickerOpen = ref(false);
const pickerPending = ref(false);
const imageSearch = ref("");
const pickerItems = ref<GalleryImageCard[]>([]);
const orderedIds = ref<string[]>([]);
const form = reactive({
  name: "",
  slug: "",
  description: "",
  visibility: "private" as "private" | "public",
  coverImageId: "",
  seoTitle: "",
  seoDescription: "",
});

const { data, pending, error, refresh } = await useAsyncData(
  "gallery-manage-collection-detail",
  () =>
    call<{ collection: GalleryCollectionDetail }>(
      `/api/v1/gallery/admin/collections/${encodeURIComponent(collectionId.value)}?page=1&size=60`,
    ),
  { server: false },
);
const collection = computed(() => data.value?.collection);
const members = computed(() => {
  const byID = new Map(
    (collection.value?.images || []).map((item) => [item.id, item]),
  );
  return orderedIds.value.flatMap((id) => {
    const item = byID.get(id);
    return item ? [item] : [];
  });
});
const canReorder = computed(() => (collection.value?.totalPages || 0) <= 1);

watch(
  collection,
  (value) => {
    if (!value) return;
    Object.assign(form, {
      name: value.name,
      slug: value.slug || "",
      description: value.description,
      visibility: value.visibility,
      coverImageId: value.coverImageId || "",
      seoTitle: value.seoTitle || "",
      seoDescription: value.seoDescription || "",
    });
    orderedIds.value = value.images.map((item) => item.id);
  },
  { immediate: true },
);

watch(
  form,
  () => {
    if (!saving.value) metadataSaved.value = false;
  },
  { deep: true },
);

useSeoMeta({ title: () => `${collection.value?.name || "专题"} · 图库管理` });

function message(reason: any): string {
  return reason?.data?.message || reason?.message || "请稍后重试";
}

async function saveMetadata(): Promise<void> {
  if (!collection.value || saving.value) return;
  saving.value = true;
  metadataSaved.value = false;
  try {
    await call<{ collection: GalleryCollection }>(
      `/api/v1/gallery/admin/collections/${encodeURIComponent(collection.value.id)}`,
      {
        method: "PATCH",
        body: { ...form, version: collection.value.version },
      },
    );
    await refresh();
    metadataSaved.value = true;
  } catch (reason) {
    toast.add({
      title: "专题设置没有保存",
      description: message(reason),
      color: "error",
    });
  } finally {
    saving.value = false;
  }
}

async function searchImages(): Promise<void> {
  pickerPending.value = true;
  try {
    const result = await call<GalleryImagePage>(
      `/api/v1/gallery/images?q=${encodeURIComponent(imageSearch.value.trim())}&page=1&size=12&sort=newest`,
    );
    const existing = new Set(
      collection.value?.images.map((item) => item.id) || [],
    );
    pickerItems.value = result.items.filter((item) => !existing.has(item.id));
  } catch (reason) {
    toast.add({
      title: "图片搜索失败",
      description: message(reason),
      color: "error",
    });
  } finally {
    pickerPending.value = false;
  }
}

async function mutateMembers(
  add: string[] = [],
  remove: string[] = [],
): Promise<void> {
  if (!collection.value) return;
  try {
    await call(
      `/api/v1/gallery/admin/collections/${encodeURIComponent(collection.value.id)}/members`,
      {
        method: "POST",
        body: { version: collection.value.version, add, remove },
      },
    );
    await refresh();
    if (pickerOpen.value) await searchImages();
  } catch (reason) {
    toast.add({
      title: "专题成员没有更新",
      description: message(reason),
      color: "error",
    });
  }
}

function moveMember(index: number, direction: -1 | 1): void {
  const target = index + direction;
  if (target < 0 || target >= orderedIds.value.length) return;
  const next = [...orderedIds.value];
  [next[index], next[target]] = [next[target]!, next[index]!];
  orderedIds.value = next;
  orderSaved.value = false;
}

function selectCover(imageId: string): void {
  form.coverImageId = imageId;
}

function clearCover(): void {
  form.coverImageId = "";
}

function memberMoreItems(item: GalleryImageCard) {
  return [
    [
      {
        label: form.coverImageId === item.id ? "当前封面" : "设为专题封面",
        icon: "i-tabler-photo-star",
        disabled: form.coverImageId === item.id,
        onSelect: () => selectCover(item.id),
      },
    ],
    [
      {
        label: "移出专题",
        icon: "i-tabler-trash",
        color: "error" as const,
        onSelect: () => mutateMembers([], [item.id]),
      },
    ],
  ];
}

async function saveOrder(): Promise<void> {
  if (!collection.value || !canReorder.value || ordering.value) return;
  ordering.value = true;
  orderSaved.value = false;
  try {
    await call(
      `/api/v1/gallery/admin/collections/${encodeURIComponent(collection.value.id)}/order`,
      {
        method: "PUT",
        body: { version: collection.value.version, imageIds: orderedIds.value },
      },
    );
    await refresh();
    orderSaved.value = true;
  } catch (reason) {
    toast.add({
      title: "图片顺序没有保存",
      description: message(reason),
      color: "error",
    });
  } finally {
    ordering.value = false;
  }
}

async function openPicker(): Promise<void> {
  pickerOpen.value = true;
  await searchImages();
}
</script>

<template>
  <div>
    <ManageHeader :title="collection?.name || '专题编辑'">
      <template #subtitle>先组织图片和顺序，再完成封面与公开叙事。</template>
      <template #actions>
        <UButton
          color="neutral"
          variant="ghost"
          to="/manage/collections"
          icon="i-tabler-arrow-left"
          label="返回列表"
        />
        <UButton
          v-if="collection?.slug"
          :to="`/collections/${collection.slug}`"
          target="_blank"
          color="neutral"
          variant="ghost"
          icon="i-tabler-external-link"
          label="公开预览"
        />
      </template>
    </ManageHeader>

    <SkeletonList v-if="!hydrated || (pending && !collection)" :rows="7" />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="专题加载失败"
    >
      <template #actions><UButton label="重试" @click="refresh()" /></template>
    </UAlert>

    <div
      v-else-if="collection"
      class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_23rem] xl:items-start"
    >
      <section
        class="gallery-manage-panel order-2 space-y-5 xl:sticky xl:top-20"
      >
        <div>
          <p
            class="text-xs font-semibold uppercase tracking-[.14em] text-primary"
          >
            Publish
          </p>
          <h2 class="mt-2 text-lg font-semibold text-highlighted">专题设置</h2>
          <p class="mt-1 text-sm text-muted">
            私有状态适合编排；公开前确认名称、封面和说明。
          </p>
        </div>
        <UFormField label="名称" required
          ><UInput v-model="form.name"
        /></UFormField>
        <UFormField label="Slug" required
          ><UInput v-model="form.slug"
        /></UFormField>
        <UFormField label="说明"
          ><UTextarea v-model="form.description" :rows="4"
        /></UFormField>
        <UFormField label="可见性">
          <USelect
            v-model="form.visibility"
            :items="[
              { label: '私有', value: 'private' },
              { label: '公开', value: 'public' },
            ]"
            value-key="value"
          />
        </UFormField>
        <div>
          <p class="mb-2 text-sm font-medium text-highlighted">当前封面</p>
          <div
            class="gallery-manage-cover"
            :style="{ backgroundColor: collection.coverColor || undefined }"
          >
            <img
              v-if="collection.coverAssetId"
              v-bind="
                galleryImageSources(collection.coverAssetId, 'grid', false)
              "
              :alt="collection.coverAltText || collection.name"
            />
            <UIcon v-else name="i-tabler-photo" class="size-9 text-muted" />
          </div>
          <UButton
            v-if="form.coverImageId"
            class="mt-2"
            color="neutral"
            variant="ghost"
            label="清除封面"
            @click="clearCover"
          />
        </div>
        <details class="group border-t border-default pt-4">
          <summary
            class="flex min-h-10 cursor-pointer list-none items-center justify-between text-sm font-medium text-default"
          >
            搜索与分享信息
            <UIcon
              name="i-tabler-chevron-down"
              class="size-4 text-muted transition group-open:rotate-180"
            />
          </summary>
          <div class="space-y-4 pt-3">
            <UFormField label="SEO 标题"
              ><UInput v-model="form.seoTitle"
            /></UFormField>
            <UFormField label="SEO 描述"
              ><UTextarea v-model="form.seoDescription" :rows="3"
            /></UFormField>
          </div>
        </details>
        <UButton
          block
          label="保存专题设置"
          icon="i-tabler-device-floppy"
          :loading="saving"
          @click="saveMetadata"
        />
        <p v-if="metadataSaved" class="text-sm text-success" role="status">
          专题设置已保存
        </p>
      </section>

      <section class="gallery-manage-panel order-1">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p
              class="text-xs font-semibold uppercase tracking-[.14em] text-primary"
            >
              Editorial sequence
            </p>
            <h2 class="mt-2 text-lg font-semibold text-highlighted">
              图片与顺序
            </h2>
            <p class="mt-1 text-sm text-muted">
              {{ collection.itemCount }} 张。顺序决定公开专题的阅读节奏。
            </p>
          </div>
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="outline"
              icon="i-tabler-arrows-sort"
              label="保存顺序"
              :loading="ordering"
              :disabled="!canReorder"
              @click="saveOrder"
            />
            <UButton
              icon="i-tabler-plus"
              label="添加图片"
              @click="openPicker"
            />
          </div>
        </div>
        <p v-if="orderSaved" class="mt-3 text-sm text-success" role="status">
          图片顺序已保存
        </p>
        <UAlert
          v-if="!canReorder"
          class="mt-4"
          color="warning"
          variant="subtle"
          title="大型专题暂不支持整组重排"
          description="当前编辑器只载入前 60 张；成员增删仍可用，整组重排将在规模化阶段升级为分段操作。"
        />
        <div
          v-if="members.length"
          class="mt-5 grid gap-3 sm:grid-cols-2 2xl:grid-cols-3"
        >
          <article
            v-for="(item, index) in members"
            :key="item.id"
            class="gallery-manage-member"
          >
            <img
              v-bind="galleryImageSources(item.assetId, 'grid', false)"
              :alt="item.altText || item.title"
            />
            <div class="min-w-0 flex-1">
              <h3 class="truncate text-sm font-semibold text-highlighted">
                {{ item.title }}
              </h3>
              <p class="mt-1 truncate text-xs text-muted">
                {{ item.primaryCategory || "未分类" }}
              </p>
              <div class="mt-2 flex flex-wrap gap-1">
                <UButton
                  color="neutral"
                  variant="ghost"
                  icon="i-tabler-arrow-up"
                  aria-label="上移"
                  :disabled="index === 0 || !canReorder"
                  @click="moveMember(index, -1)"
                />
                <UButton
                  color="neutral"
                  variant="ghost"
                  icon="i-tabler-arrow-down"
                  aria-label="下移"
                  :disabled="index === members.length - 1 || !canReorder"
                  @click="moveMember(index, 1)"
                />
                <UButton
                  v-if="form.coverImageId === item.id"
                  color="primary"
                  variant="soft"
                  icon="i-tabler-photo-star"
                  label="封面"
                  disabled
                />
                <UDropdownMenu :items="memberMoreItems(item)">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    icon="i-tabler-dots"
                    aria-label="更多图片操作"
                  />
                </UDropdownMenu>
              </div>
            </div>
          </article>
        </div>
        <div v-else class="gallery-compact-empty mt-5 min-h-64">
          <UIcon name="i-tabler-photo-plus" class="mx-auto size-8 text-muted" />
          <h3 class="mt-3 font-semibold text-highlighted">专题还没有图片</h3>
          <p class="mt-1 text-sm text-muted">添加已发布图片后再选择封面。</p>
        </div>
      </section>
    </div>

    <UModal
      v-model:open="pickerOpen"
      title="添加图片"
      description="搜索已发布且可公开收藏的图片。"
      :ui="{ content: 'sm:max-w-5xl' }"
    >
      <template #body>
        <form class="mb-5 flex gap-2" @submit.prevent="searchImages">
          <UInput
            v-model="imageSearch"
            class="flex-1"
            icon="i-tabler-search"
            placeholder="搜索标题或说明"
          />
          <UButton type="submit" label="搜索" :loading="pickerPending" />
        </form>
        <div v-if="pickerPending" class="grid gap-3 sm:grid-cols-3">
          <USkeleton
            v-for="index in 6"
            :key="index"
            class="aspect-[4/3] rounded-lg"
          />
        </div>
        <div v-else-if="pickerItems.length" class="grid gap-3 sm:grid-cols-3">
          <article
            v-for="item in pickerItems"
            :key="item.id"
            class="gallery-manage-picker-item"
          >
            <img
              v-bind="galleryImageSources(item.assetId, 'grid', false)"
              :alt="item.altText || item.title"
            />
            <div class="flex items-center justify-between gap-2 p-3">
              <p class="truncate text-sm font-medium text-highlighted">
                {{ item.title }}
              </p>
              <UButton
                size="sm"
                label="添加"
                @click="mutateMembers([item.id], [])"
              />
            </div>
          </article>
        </div>
        <p v-else class="py-10 text-center text-sm text-muted">
          没有可添加的图片
        </p>
      </template>
    </UModal>
  </div>
</template>
