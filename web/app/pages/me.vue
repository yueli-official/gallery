<script setup lang="ts">
useSeoMeta({ title: "我的" });
const { loggedIn } = useAuth();
const colorMode = useColorMode();
const appearanceOptions = [
  { label: "跟随系统", value: "system" },
  { label: "浅色", value: "light" },
  { label: "深色", value: "dark" },
];

const site = useGallerySiteSettings();
const { isAdministrator, canManage } = useGalleryMe();
const actions = computed(() => [
  { label: "我的收藏", icon: "i-tabler-heart", to: "/favorites" },
  { label: "我的投稿", icon: "i-tabler-history", to: "/submissions" },
  { label: "投稿图片", icon: "i-tabler-square-plus", to: "/submit" },
  ...(canManage.value ? [{ label: "图库管理", icon: "i-tabler-settings", to: "/manage" }] : []),
]);
</script>

<template>
  <div class="mx-auto my-5 w-full max-w-xl space-y-4 px-4">
    <ConsumerAccountPanel :site-name="site?.name || '月离图库'" :can-manage="isAdministrator" :context-actions="actions" />
    <UFormField v-if="!loggedIn" label="外观" class="rounded-xl border border-default bg-default p-4">
      <USelect v-model="colorMode.preference" :items="appearanceOptions" class="w-full" />
    </UFormField>
  </div>
</template>
