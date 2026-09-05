<script setup lang="ts">
import { ManagePage, TabbedSurface } from "@yueli/ui/admin";
import { bindSettingsBeforeUnload } from "@yueli/ui/settings/browser";
import { useVueSettingsWorkflow } from "@yueli/ui/settings/vue";
import { SettingSection } from "@yueli/ui/settings/pattern";
import { useSettingsLeaveGuard } from "@yueli/ui/settings/vue-router";
import { onMounted, onScopeDispose } from "vue";
import { SkeletonList } from "~/utils/manageComponents";
import { createGalleryNotifier } from "~/utils/feedback";
import type {
  GalleryHomeSection,
  GalleryHomeSectionKey,
  GallerySite,
} from "~/types/gallery";

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "站点设置 · 图库管理" });

const { call } = useGalleryApi();
const { can } = useGalleryMe();
const route = useRoute();
const router = useRouter();
const hydrated = useClientHydrated();
const toast = createGalleryNotifier(useToast());
const canManageSettings = computed(() => can("gallery.discovery.manage"));
const saving = ref(false);
const saved = ref(false);
const saveError = ref("");
type SettingsSection = "site" | "home" | "discovery";
const settingsSections = [
  {
    value: "site",
    label: "站点",
    icon: "i-tabler-adjustments-horizontal",
  },
  { value: "home", label: "首页", icon: "i-tabler-layout-dashboard" },
  { value: "discovery", label: "发现", icon: "i-tabler-refresh" },
] as const;
const settingsSectionKeys = settingsSections.map((item) => item.value);
const activeSection = computed<SettingsSection>({
  get: () => {
    const value = String(route.query.section || "site");
    return settingsSectionKeys.includes(value as SettingsSection)
      ? (value as SettingsSection)
      : "site";
  },
  set: (value) => {
    void router.replace({
      query: {
        ...route.query,
        section: value === "site" ? undefined : value,
      },
    });
  },
});

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

function snapshotForm(): GallerySite {
  return {
    name: form.name,
    title: form.title,
    description: form.description,
    searchPlaceholder: form.searchPlaceholder,
    footerTagline: form.footerTagline,
    randomBatchSize: Number(form.randomBatchSize),
    randomCandidateSize: Number(form.randomCandidateSize),
    homeSections: form.homeSections.map((section) => ({ ...section })),
  };
}

function restoreForm(site: GallerySite): void {
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
}

const settingsWorkflow = useVueSettingsWorkflow<GallerySite>({
  snapshot: snapshotForm,
  restore: restoreForm,
});
const dirty = settingsWorkflow.dirty;

let unbindBeforeUnload: (() => void) | undefined;
onMounted(() => {
  unbindBeforeUnload = bindSettingsBeforeUnload({
    isDirty: () => dirty.value,
  });
});
onScopeDispose(() => unbindBeforeUnload?.());
useSettingsLeaveGuard({
  isDirty: () => dirty.value,
  confirm: () =>
    window.confirm("有未保存的站点设置，确定离开当前页面吗？"),
});

const { data: settingsData, pending, error, refresh } = await useAsyncData(
  "gallery-manage-site-settings",
  () => call<{ site: GallerySite }>("/admin/site-settings"),
  { server: false },
);

watch(
  () => settingsData.value?.site,
  (site) => {
    if (!site) return;
    restoreForm(site);
    settingsWorkflow.capture();
  },
  { immediate: true, flush: "sync" },
);

watch(
  form,
  () => {
    if (!saving.value) {
      saved.value = false;
      saveError.value = "";
    }
  },
  { deep: true, flush: "sync" },
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
  return galleryFailureMessage(value, "请稍后重试");
}

async function saveSettings(): Promise<void> {
  if (!canManageSettings.value || saving.value) return;
  saving.value = true;
  saved.value = false;
  saveError.value = "";
  try {
    const result = await call<{ site: GallerySite }>(
      "/admin/site-settings",
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
    restoreForm(result.site);
    settingsWorkflow.capture();
    saved.value = true;
  } catch (reason) {
    saveError.value = message(reason);
    toast.add({
      title: "站点设置没有保存",
      description: saveError.value,
      color: "error",
    });
  } finally {
    saving.value = false;
  }
}

</script>

<template>
  <ManagePage id="discovery" title="站点设置" icon="i-tabler-settings">
    <template #actions>
      <UButton
        v-if="canManageSettings"
        icon="i-tabler-device-floppy"
        label="保存设置"
        :loading="saving"
        @click="saveSettings"
      />
    </template>

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

    <form v-else @submit.prevent="saveSettings">
      <UAlert
        v-if="!canManageSettings"
        class="mb-5"
        color="neutral"
        variant="subtle"
        icon="i-tabler-lock"
        title="只读设置"
        description="当前角色可以查看站点设置，但不能修改。"
      />

      <TabbedSurface
        v-model="activeSection"
        :items="settingsSections"
        navigation-label="站点设置"
        data-manage-surface="site-settings"
      >
        <fieldset
          v-if="activeSection === 'site'"
          :disabled="!canManageSettings"
          class="min-w-0 p-4 sm:p-5"
        >
          <SettingSection title="站点信息">
            <div class="grid gap-4 md:grid-cols-2">
              <UFormField label="站点名称" required>
                <UInput v-model="form.name" maxlength="80" class="w-full" />
              </UFormField>
              <UFormField label="页脚说明">
                <UInput v-model="form.footerTagline" maxlength="240" class="w-full" />
              </UFormField>
            </div>
          </SettingSection>
        </fieldset>

        <fieldset
          v-else-if="activeSection === 'home'"
          :disabled="!canManageSettings"
          class="min-w-0 space-y-5 p-4 sm:p-5"
        >
          <SettingSection title="首页信息">
            <div class="grid gap-4 md:grid-cols-2">
              <UFormField label="首页标题" required>
                <UInput
                  v-model="form.title"
                  maxlength="120"
                  class="w-full"
                />
              </UFormField>
              <UFormField label="搜索框提示" required>
                <UInput
                  v-model="form.searchPlaceholder"
                  maxlength="120"
                  class="w-full"
                />
              </UFormField>
              <UFormField label="首页描述" class="md:col-span-2">
                <UTextarea
                  v-model="form.description"
                  :rows="3"
                  maxlength="320"
                  class="w-full"
                />
              </UFormField>
            </div>
          </SettingSection>

          <SettingSection title="首页板块">
            <div class="divide-y divide-default border-y border-default">
              <section
                v-for="(section, index) in form.homeSections"
                :key="section.key"
                class="py-5 first:pt-4 last:pb-4"
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
                    <UInput
                      v-model="section.title"
                      maxlength="80"
                      class="w-full"
                    />
                  </UFormField>
                  <UFormField label="右侧按钮文字">
                    <UInput
                      v-model="section.actionLabel"
                      maxlength="40"
                      class="w-full"
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
                      class="w-full"
                    />
                  </UFormField>
                </div>
              </section>
            </div>
          </SettingSection>
        </fieldset>

        <div v-else class="min-w-0 p-4 sm:p-5">
          <fieldset
            :disabled="!canManageSettings"
            class="min-w-0"
          >
            <SettingSection title="随机发现">
              <UFormField label="候选池数量" hint="40–2000">
                <UInput v-model.number="form.randomCandidateSize" type="number" min="40" max="2000" class="w-full max-w-xs" />
              </UFormField>
            </SettingSection>
          </fieldset>

        </div>

        <div
          v-if="saved"
          class="flex items-center gap-2 border-t border-default px-4 py-3 text-sm text-success sm:px-5"
          role="status"
        >
          <UIcon name="i-tabler-circle-check" class="size-4" />
          <span>设置已保存，前台刷新后生效。</span>
        </div>
      </TabbedSurface>
    </form>
  </ManagePage>
</template>
