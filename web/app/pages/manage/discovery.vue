<script setup lang="ts">
import { PageHeader } from "@yueli/ui/dashboard/pattern";
import { SkeletonList } from "~/utils/manageComponents";
import { createGalleryNotifier } from "~/utils/feedback";
import type {
  GalleryDiscovery,
  GalleryHomeSection,
  GalleryHomeSectionKey,
  GallerySite,
} from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "站点与首页 · 图库管理" });

const { call } = useApi();
const { can } = useGalleryMe();
const hydrated = useClientHydrated();
const toast = createGalleryNotifier(useToast());
const canManageSettings = computed(() => can("gallery.discovery.manage"));
const saving = ref(false);
const saved = ref(false);

const sectionNames: Record<GalleryHomeSectionKey, string> = {
  random: "随机图片",
  collections: "专题入口",
  latest: "最新图片",
  trending: "趋势排行",
};

const form = reactive<GallerySite>({
  name: "",
  title: "",
  description: "",
  searchPlaceholder: "",
  footerTagline: "",
  randomBatchSize: 24,
  randomCandidateSize: 240,
  homeSections: [],
});

const [
  { data: settingsData, pending, error, refresh },
  { data: preview, refresh: refreshPreview },
] = await Promise.all([
  useAsyncData(
    "gallery-manage-site-settings",
    () =>
      call<{ site: GallerySite }>("/api/v1/gallery/admin/site-settings"),
    { server: false },
  ),
  useAsyncData(
    "gallery-manage-discovery-preview",
    () =>
      call<GalleryDiscovery>(
        "/api/v1/gallery/discovery?seed=operator-preview",
      ),
    { server: false },
  ),
]);

watch(
  () => settingsData.value?.site,
  (site) => {
    if (!site) return;
    Object.assign(form, {
      name: site.name,
      title: site.title,
      description: site.description,
      searchPlaceholder: site.searchPlaceholder,
      footerTagline: site.footerTagline,
      randomBatchSize: site.randomBatchSize,
      randomCandidateSize: site.randomCandidateSize,
    });
    form.homeSections.splice(
      0,
      form.homeSections.length,
      ...[...site.homeSections]
        .sort((left, right) => left.position - right.position)
        .map((section) => ({ ...section })),
    );
  },
  { immediate: true },
);

watch(
  form,
  () => {
    if (!saving.value) saved.value = false;
  },
  { deep: true },
);

function sectionLimit(section: GalleryHomeSection): { min: number; max: number } {
  if (section.key === "random") return { min: 12, max: 60 };
  if (section.key === "collections") return { min: 1, max: 12 };
  return { min: 1, max: 24 };
}

function moveSection(index: number, offset: -1 | 1): void {
  const target = index + offset;
  if (target < 0 || target >= form.homeSections.length) return;
  const current = form.homeSections[index];
  const replacement = form.homeSections[target];
  if (!current || !replacement) return;
  form.homeSections.splice(index, 1, replacement);
  form.homeSections.splice(target, 1, current);
}

function message(reason: unknown): string {
  const value = reason as {
    data?: { message?: string };
    message?: string;
  };
  return value?.data?.message || value?.message || "请稍后重试";
}

async function saveSettings(): Promise<void> {
  if (!canManageSettings.value || saving.value) return;
  saving.value = true;
  saved.value = false;
  try {
    const result = await call<{ site: GallerySite }>(
      "/api/v1/gallery/admin/site-settings",
      {
        method: "PATCH",
        body: {
          name: form.name,
          title: form.title,
          description: form.description,
          searchPlaceholder: form.searchPlaceholder,
          footerTagline: form.footerTagline,
          randomCandidateSize: Number(form.randomCandidateSize),
          homeSections: form.homeSections.map((section, position) => ({
            ...section,
            position,
            itemLimit: Number(section.itemLimit),
          })),
        },
      },
    );
    settingsData.value = result;
    await refreshPreview();
    saved.value = true;
  } catch (reason) {
    toast.add({
      title: "站点与首页设置没有保存",
      description: message(reason),
      color: "error",
    });
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <div>
    <PageHeader title="站点与首页">
      <template #subtitle>
        前台标题、说明、板块顺序和展示数量都从这里发布，不再修改页面代码。
      </template>
      <template #actions>
        <UButton
          v-if="canManageSettings"
          icon="i-tabler-device-floppy"
          label="保存设置"
          :loading="saving"
          @click="saveSettings"
        />
      </template>
    </PageHeader>

    <SkeletonList v-if="!hydrated || pending" :rows="6" />
    <UAlert
      v-else-if="error"
      color="error"
      variant="subtle"
      title="站点设置加载失败"
    >
      <template #actions>
        <UButton label="重试" @click="refresh()" />
      </template>
    </UAlert>

    <form
      v-else
      class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_22rem] xl:items-start"
      @submit.prevent="saveSettings"
    >
      <div class="space-y-6">
        <fieldset
          :disabled="!canManageSettings"
          class="gallery-manage-panel space-y-5"
        >
          <div>
            <h2 class="text-lg font-semibold text-highlighted">站点内容</h2>
            <p class="mt-1 text-sm leading-6 text-muted">
              名称用于品牌与页脚，页面标题和描述用于首页及搜索分享信息。
            </p>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <UFormField label="站点名称" required>
              <UInput v-model="form.name" maxlength="80" />
            </UFormField>
            <UFormField label="首页标题" required>
              <UInput v-model="form.title" maxlength="120" />
            </UFormField>
          </div>
          <UFormField label="首页描述">
            <UTextarea v-model="form.description" :rows="3" maxlength="320" />
          </UFormField>
          <div class="grid gap-4 md:grid-cols-2">
            <UFormField label="搜索框提示" required>
              <UInput v-model="form.searchPlaceholder" maxlength="120" />
            </UFormField>
            <UFormField label="页脚说明">
              <UInput v-model="form.footerTagline" maxlength="240" />
            </UFormField>
          </div>
        </fieldset>

        <fieldset
          :disabled="!canManageSettings"
          class="gallery-manage-panel space-y-5"
        >
          <div>
            <h2 class="text-lg font-semibold text-highlighted">首页板块</h2>
            <p class="mt-1 text-sm leading-6 text-muted">
              从上到下就是前台顺序。关闭后内容仍保留，只是不在首页显示。
            </p>
          </div>

          <div class="divide-y divide-default border-y border-default">
            <section
              v-for="(section, index) in form.homeSections"
              :key="section.key"
              class="py-5 first:pt-3 last:pb-3"
            >
              <div class="flex flex-wrap items-center gap-3">
                <div class="min-w-0 flex-1">
                  <p class="text-sm font-semibold text-highlighted">
                    {{ sectionNames[section.key] }}
                  </p>
                  <p class="mt-0.5 text-xs text-muted">
                    第 {{ index + 1 }} 个板块
                  </p>
                </div>
                <div class="flex items-center gap-1">
                  <UButton
                    type="button"
                    color="neutral"
                    variant="ghost"
                    icon="i-tabler-arrow-up"
                    aria-label="向上移动"
                    :disabled="index === 0"
                    @click="moveSection(index, -1)"
                  />
                  <UButton
                    type="button"
                    color="neutral"
                    variant="ghost"
                    icon="i-tabler-arrow-down"
                    aria-label="向下移动"
                    :disabled="index === form.homeSections.length - 1"
                    @click="moveSection(index, 1)"
                  />
                  <USwitch v-model="section.enabled" label="显示" />
                </div>
              </div>

              <div class="mt-4 grid gap-4 md:grid-cols-2">
                <UFormField label="标题" required>
                  <UInput v-model="section.title" maxlength="80" />
                </UFormField>
                <UFormField label="右侧按钮文字">
                  <UInput v-model="section.actionLabel" maxlength="40" />
                </UFormField>
                <UFormField label="说明">
                  <UTextarea
                    v-model="section.description"
                    :rows="2"
                    maxlength="240"
                  />
                </UFormField>
                <UFormField
                  :label="
                    section.key === 'collections'
                      ? '专题数量'
                      : '图片数量'
                  "
                  :hint="`${sectionLimit(section).min}–${sectionLimit(section).max}`"
                >
                  <UInput
                    v-model.number="section.itemLimit"
                    type="number"
                    :min="sectionLimit(section).min"
                    :max="sectionLimit(section).max"
                  />
                </UFormField>
              </div>
            </section>
          </div>
        </fieldset>
      </div>

      <aside class="space-y-6 xl:sticky xl:top-6">
        <fieldset
          :disabled="!canManageSettings"
          class="gallery-manage-panel space-y-4"
        >
          <div>
            <h2 class="text-lg font-semibold text-highlighted">随机发现</h2>
            <p class="mt-1 text-sm leading-6 text-muted">
              图片按 seed 稳定随机，并优先打散相同分类；不是按热度排序。
            </p>
          </div>
          <UFormField label="候选池数量" hint="40–2000">
            <UInput
              v-model.number="form.randomCandidateSize"
              type="number"
              min="40"
              max="2000"
            />
          </UFormField>
          <p class="text-xs leading-5 text-muted">
            首页实际展示数量由“随机图片”板块控制，默认 24 张。
          </p>
        </fieldset>

        <section class="gallery-manage-panel overflow-hidden p-0">
          <div
            class="grid h-48 grid-cols-4 grid-rows-2 gap-1 bg-elevated p-1"
            aria-label="随机图片预览"
          >
            <div
              v-for="(item, index) in preview?.images.slice(0, 4)"
              :key="item.id"
              class="overflow-hidden bg-muted"
              :class="
                index === 0
                  ? 'col-span-2 row-span-2'
                  : index === 1
                    ? 'col-span-2'
                    : ''
              "
            >
              <img
                v-bind="galleryImageSources(item.assetId, 'grid', false)"
                :alt="item.altText || item.title"
                class="size-full object-cover"
              />
            </div>
          </div>
          <div class="p-4">
            <p class="text-sm font-semibold text-highlighted">当前随机结果</p>
            <p class="mt-1 text-xs leading-5 text-muted">
              保存后刷新预览。单张图片内容仍在“图片”中管理。
            </p>
            <UButton
              class="mt-4"
              to="/?seed=operator-preview"
              target="_blank"
              color="neutral"
              variant="outline"
              icon="i-tabler-external-link"
              label="打开首页"
            />
          </div>
        </section>

        <UButton
          v-if="canManageSettings"
          type="submit"
          block
          icon="i-tabler-device-floppy"
          label="保存站点与首页"
          :loading="saving"
        />
        <p v-if="saved" class="text-sm text-success" role="status">
          设置已保存，前台刷新后生效。
        </p>
        <p v-if="!canManageSettings" class="text-sm leading-6 text-muted">
          当前角色可以查看设置，但不能修改。
        </p>
      </aside>
    </form>
  </div>
</template>
