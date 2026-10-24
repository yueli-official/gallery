<script setup lang="ts">
import type { GallerySavedFilter } from "~/types/presentation";
defineEmits<{ navigate: [] }>();
const site = useGallerySiteSettings();
const route = useRoute();
const colorMode = useColorMode();
const { loggedIn } = useAuth();
const { items: savedFilters, error, refresh, remove } = useGallerySavedFilters();
const deleteError = ref("");
const navigation = [
  { to: "/images", label: "全部图片", icon: "i-tabler-layout-grid" },
  { to: "/collections", label: "专题集合", icon: "i-tabler-folders" },
  { to: "/rankings", label: "图片排行", icon: "i-tabler-chart-bar" },
  { to: "/favorites", label: "我的收藏", icon: "i-tabler-heart" },
  { to: "/submissions", label: "我的投稿", icon: "i-tabler-history" },
  { to: "/submit", label: "投稿图片", icon: "i-tabler-square-plus" },
];
const background = computed(() => {
  const config = site.value?.appearance.sidebar;
  if (config?.mode !== "image") return {};
  return { backgroundImage: `linear-gradient(rgba(12,17,25,${config.overlay}), rgba(12,17,25,${config.overlay})), url("${galleryRendition(config.assetId, "display")}")` };
});
const imageBackground = computed(() => site.value?.appearance.sidebar.mode === "image");
function link(item: GallerySavedFilter) {
  const q = item.query;
  return { path: "/images", query: { q: q.q || undefined, categories: q.categories || undefined, facets: q.facets || undefined,
    tag: q.tag || undefined, matchAll: q.matchAll ? "true" : undefined, orientation: q.orientation || undefined, minLongEdge: q.minLongEdge || undefined } };
}
async function deleteFilter(id: string) {
  deleteError.value = "";
  try { await remove(id); } catch (reason) { deleteError.value = galleryFailureMessage(reason, "筛选没有删除"); }
}
onMounted(() => void refresh());
watch(loggedIn, () => void refresh());
</script>

<template>
  <div class="public-sidebar" :class="{ 'has-background': imageBackground }" :style="background">
    <NuxtLink to="/" class="public-brand" :aria-label="`${site?.name || '月离图库'}首页`" @click="$emit('navigate')">
      <img src="/favicon.svg?v=20261024-brand" class="brand-symbol" width="36" height="36" alt="" />
      <span class="brand-copy"><span class="brand-wordmark">Gallery</span><small>{{ site?.name || '月离图库' }}</small></span>
    </NuxtLink>
    <nav class="public-navigation" aria-label="主要导航">
      <NuxtLink v-for="item in navigation" :key="item.to" :to="item.to" :class="{ active: route.path === item.to || (item.to === '/images' && (route.path === '/' || route.path.startsWith('/images/'))) }" @click="$emit('navigate')">
        <UIcon :name="item.icon" /><span>{{ item.label }}</span>
      </NuxtLink>
    </nav>
    <div class="sidebar-saved">
      <div class="sidebar-section-heading"><span>我的筛选</span><NuxtLink to="/images?save=true" aria-label="保存当前筛选" @click="$emit('navigate')"><UIcon name="i-tabler-plus" /></NuxtLink></div>
      <p v-if="error || deleteError" class="text-xs text-error" role="alert">{{ error || deleteError }}</p>
      <div v-for="item in savedFilters" :key="item.id" class="saved-filter-row">
        <NuxtLink :to="link(item)" @click="$emit('navigate')"><UIcon name="i-tabler-bookmark" /><span>{{ item.name }}</span></NuxtLink>
        <UButton icon="i-tabler-x" color="neutral" variant="ghost" size="xs" :aria-label="`删除筛选：${item.name}`" @click="deleteFilter(item.id)" />
      </div>
      <p v-if="!savedFilters.length && !error" class="sidebar-empty">{{ loggedIn ? '保存常用组合，下次直接找到。' : '登录后保存常用筛选。' }}</p>
    </div>
    <div class="sidebar-footer"><NuxtLink to="/manage" aria-label="图库管理"><UIcon name="i-tabler-settings" /></NuxtLink><button type="button" aria-label="切换明暗主题" @click="colorMode.preference = colorMode.value === 'dark' ? 'light' : 'dark'"><UIcon name="i-tabler-moon-stars" /></button></div>
  </div>
</template>

<style scoped>
.public-sidebar{display:flex;flex-direction:column;min-height:100%;padding:1.2rem .75rem;background:var(--gallery-panel);background-position:center;background-size:cover;border-right:1px solid var(--gallery-line);color:var(--gallery-ink)}
:global(.dark) .public-sidebar{background:radial-gradient(ellipse at 0 100%,var(--gallery-accent-soft),transparent 62%),var(--gallery-panel)}
.public-sidebar.has-background{--gallery-ink:#f4f1e9;--gallery-muted:#b7bec8;--gallery-accent:#f1f5f9;--gallery-accent-soft:#ffffff14;--gallery-line:#ffffff30;color:#f4f1e9}:global(.dark) .public-sidebar.has-background{--gallery-accent:#e7c68e;--gallery-accent-soft:#342f27}.public-brand{display:flex;align-items:center;gap:.65rem;min-width:0;padding:.15rem .45rem 1.5rem}.brand-symbol{display:block;width:2.25rem;height:2.25rem;flex-shrink:0}.brand-copy{display:flex;flex-direction:column;gap:.25rem;min-width:0}.brand-wordmark{font-family:var(--font-sans);font-size:1.4rem;font-weight:650;line-height:1;letter-spacing:-.055em;color:var(--gallery-ink)}.brand-copy small{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:.57rem;line-height:1.4;letter-spacing:.1em;color:var(--gallery-muted)}
.public-navigation{display:grid;gap:.25rem}.public-navigation>a{display:flex;gap:.75rem;align-items:center;padding:.7rem .8rem;border:1px solid transparent;border-radius:.4rem;font-size:.78rem}.public-navigation>a.active{border-color:color-mix(in srgb,var(--gallery-accent) 38%,transparent);color:var(--gallery-accent);background:var(--gallery-accent-soft)}.public-navigation>a:hover{background:var(--gallery-accent-soft)}.public-navigation .iconify{font-size:1.05rem;flex-shrink:0}
.sidebar-saved{border-top:1px solid var(--gallery-line);margin-top:1.5rem;padding:.9rem .5rem;min-height:0;overflow:auto}.sidebar-section-heading{display:flex;justify-content:space-between;align-items:center;font-size:.72rem;color:var(--gallery-muted);margin-bottom:.7rem}.sidebar-section-heading a{display:grid;place-items:center;width:1.75rem;height:1.75rem;color:var(--gallery-accent)}.sidebar-empty{font-size:.68rem;line-height:1.8;color:var(--gallery-muted)}.saved-filter-row{display:flex;align-items:center;gap:.2rem}.saved-filter-row>a{display:flex;align-items:center;gap:.5rem;flex:1;min-width:0;font-size:.72rem;padding:.6rem 0}.saved-filter-row>a span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.sidebar-footer{margin-top:auto;padding:.8rem .6rem 0;display:flex;gap:1rem;color:var(--gallery-muted)}.sidebar-footer a,.sidebar-footer button{display:grid;place-items:center;min-width:1.75rem;min-height:1.75rem;cursor:pointer}
@media(pointer:coarse){.sidebar-footer a,.sidebar-footer button,.sidebar-section-heading a{min-width:2.75rem;min-height:2.75rem}}
</style>
