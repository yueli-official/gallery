<script setup lang="ts">
import { createGalleryNotifier } from "~/utils/feedback";
import type { GalleryCollectionDetail } from "~/types/gallery";

definePageMeta({ middleware: "auth" });
const route = useRoute();
const router = useRouter();
const hydrated = useClientHydrated();
const { call } = useGalleryApi();
const toast = createGalleryNotifier(useToast());
const page = computed(() => Math.max(1, Number(route.query.page) || 1));
const sort = computed(() =>
  ["newest", "oldest", "title_asc", "title_desc"].includes(
    String(route.query.sort),
  )
    ? String(route.query.sort)
    : "newest",
);
const sortItems = [
  { label: "最近收藏", value: "newest" },
  { label: "最早收藏", value: "oldest" },
  { label: "标题 A-Z", value: "title_asc" },
  { label: "标题 Z-A", value: "title_desc" },
];
const { data, error, pending, refresh } = await useAsyncData(
  "gallery-my-favorites",
  () =>
    call<{ collection: GalleryCollectionDetail }>(
      "/me/favorites",
      { query: { page: page.value, size: 24, sort: sort.value } },
    ),
  { server: false, watch: [page, sort] },
);
const collection = computed(() => data.value?.collection);
const images = computed(() => collection.value?.items || []);
const removing = ref("");

function updateQuery(next: { page?: number; sort?: string }) {
  const query = { ...route.query };
  if (next.page !== undefined) {
    if (next.page <= 1) delete query.page;
    else query.page = String(next.page);
  }
  if (next.sort !== undefined) {
    if (next.sort === "newest") delete query.sort;
    else query.sort = next.sort;
    delete query.page;
  }
  void router.push({ query });
}

async function removeFavorite(imageId: string) {
  if (!collection.value || removing.value) return;
  removing.value = imageId;
  try {
    await call(`/me/favorites/${encodeURIComponent(imageId)}`, {
      method: "DELETE",
      query: { version: collection.value.version },
    });
    // feedback-contract: the removed card immediately leaves the list, so its inline surface no longer exists.
    toast.add({ title: "已取消收藏", color: "success" });
    if (images.value.length === 1 && page.value > 1)
      updateQuery({ page: page.value - 1 });
    else await refresh();
  } catch (reason: any) {
    await refresh();
    toast.add({
      title: "收藏状态已变化",
      description: galleryFailureMessage(reason, "列表已刷新，请再试一次"),
      color: "warning",
    });
  } finally {
    removing.value = "";
  }
}

useSeoMeta({ title: "我的收藏", robots: "noindex,nofollow" });
</script>

<template>
  <GalleryPublicPage>
    <GalleryPageHeader
      title="我的收藏"
      description="只对你可见。按收藏时间或标题整理，点击图片查看完整详情。"
    >
      <template #actions>
        <div v-if="collection" class="flex items-center gap-3">
          <span class="text-sm tabular-nums text-muted"
            >{{ collection.itemCount }} 张</span
          >
          <USelect
            :model-value="sort"
            :items="sortItems"
            value-key="value"
            class="w-36"
            aria-label="收藏排序"
            @update:model-value="updateQuery({ sort: String($event) })"
          />
        </div>
      </template>
    </GalleryPageHeader>
    <div v-if="!hydrated || pending" class="gallery-grid">
      <USkeleton
        v-for="index in 12"
        :key="index"
        class="aspect-[4/3] rounded-lg"
      />
    </div>
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="收藏加载失败"
      ><template #actions><UButton label="重试" @click="refresh()" /></template
    ></UAlert>
    <GalleryFavoriteGrid
      v-else-if="images.length"
      :items="images"
      :removing="removing"
      @remove="removeFavorite"
    />
    <GalleryCompactEmpty
      v-else
      icon="i-tabler-heart"
      title="还没有收藏图片"
      description="遇到想再看的图片时，点一下收藏。"
    >
      <template #actions>
        <UButton
          to="/images"
          class="mt-4"
          color="neutral"
          variant="outline"
          label="去浏览"
        />
      </template>
    </GalleryCompactEmpty>

    <nav
      v-if="collection && galleryPageCount(collection) > 1"
      class="mt-10 flex items-center justify-center gap-3"
      aria-label="收藏分页"
    >
      <UButton
        color="neutral"
        variant="outline"
        icon="i-tabler-arrow-left"
        label="上一页"
        :disabled="page <= 1"
        @click="updateQuery({ page: page - 1 })"
      />
      <span class="text-sm tabular-nums text-muted"
        >{{ page }} / {{ galleryPageCount(collection) }}</span
      >
      <UButton
        color="neutral"
        variant="outline"
        trailing-icon="i-tabler-arrow-right"
        label="下一页"
        :disabled="page >= galleryPageCount(collection)"
        @click="updateQuery({ page: page + 1 })"
      />
    </nav>
  </GalleryPublicPage>
</template>
