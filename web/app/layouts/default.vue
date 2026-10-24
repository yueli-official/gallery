<script setup lang="ts">
import type { GallerySite } from "~/types/gallery";
const gallerySite = useGallerySiteSettings();
const { data: response } = await useFetch<{ site: GallerySite }>("/api/gallery/site", { key: "gallery-public-site-settings" });
watch(() => response.value?.site, value => { if (value) gallerySite.value = value; }, { immediate: true });
const mobileOpen = ref(false);
const { canManage, refreshMe } = useGalleryMe();
const { loggedIn } = useAuth();
const actions = computed(() => canManage.value ? [{ label: "图库管理", icon: "i-tabler-settings", to: "/manage" }] : []);
onMounted(() => { if (loggedIn.value) void refreshMe(); });
watch(loggedIn, value => { if (value) void refreshMe(); });
useHead({ bodyAttrs: { class: "gallery-public-active" } });
</script>
<template>
  <div class="gallery-workspace">
    <a href="#public-main" class="sr-only focus:not-sr-only">跳到主要内容</a>
    <aside class="gallery-workspace-sidebar"><GalleryPublicSidebar /></aside>
    <div class="gallery-workspace-panel">
      <header class="gallery-workspace-topbar">
        <UButton class="gallery-mobile-menu" icon="i-tabler-menu-2" color="neutral" variant="ghost" aria-label="打开导航" @click="void (mobileOpen = true)" />
        <GalleryGlobalSearch compact :placeholder="gallerySite?.searchPlaceholder" />
        <ConsumerAccountControl :context-actions="actions" />
      </header>
      <main id="public-main" tabindex="-1" class="gallery-workspace-main"><slot /></main>
    </div>
    <USlideover v-model:open="mobileOpen" side="left" title="图库导航" :ui="{ content: 'max-w-64', body: 'p-0' }">
      <template #body><GalleryPublicSidebar @navigate="mobileOpen = false" /></template>
    </USlideover>
  </div>
</template>
<style scoped>
.gallery-workspace{--gallery-content-max:120rem;min-height:100dvh;color:var(--gallery-ink)}.gallery-workspace-sidebar{position:fixed;inset:0 auto 0 0;width:12.5rem;z-index:30;overflow-y:auto}.gallery-workspace-panel{margin-left:12.5rem;min-width:0}.gallery-workspace-topbar{position:sticky;top:0;z-index:25;display:flex;align-items:center;justify-content:space-between;gap:1rem;padding:.75rem max(1.25rem,calc((100% - var(--gallery-content-max))/2 + 1.25rem));background:var(--gallery-canvas);border-bottom:1px solid var(--gallery-line);min-height:4rem}.gallery-workspace-main{width:100%;max-width:var(--gallery-content-max);margin-inline:auto;padding:1.15rem 1.25rem 2rem;outline:none;min-width:0}.gallery-mobile-menu{display:none}
@media(max-width:900px){.gallery-workspace-sidebar{display:none}.gallery-workspace-panel{margin-left:0}.gallery-mobile-menu{display:inline-flex}.gallery-workspace-topbar{gap:.5rem;padding:.65rem .75rem}.gallery-workspace-main{padding:.85rem .75rem 1.5rem}}
</style>
