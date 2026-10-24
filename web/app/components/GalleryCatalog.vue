<script setup lang="ts">
import { CollectionPagination } from "@yueli/ui/collection/pattern";
import type { GalleryImagePage } from "~/types/gallery";
import type { GalleryCatalogSort, GalleryCatalogState } from "~/utils/catalog";
import type { GallerySavedFilter } from "~/types/presentation";
const route = useRoute();
const catalog = useGalleryCatalogState();
const { state, request, selectedCategories, selectedFacets, selectedTag } = catalog;
const { data, status, error, refresh } = await useFetch<GalleryImagePage>("/api/gallery/images", { query: request });
const pending = useMinimumLoading(computed(() => status.value === "pending" || status.value === "idle"));
const navigation = useGalleryViewerNavigation();
watch([data, request], ([page, query]) => { if (page && page.page === query.page) navigation.value = createCatalogViewerNavigationSession(page.items, query, galleryPageCount(page)); }, { immediate: true });
const filtersOpen = ref(true);
const mobileOpen = ref(false);
const saveOpen = ref(Boolean(route.query.save));
const saveName = ref("");
const saving = ref(false);
const saveError = ref("");
const { loggedIn, login } = useAuth();
const { call } = useGalleryApi();
const { refresh: refreshSaved } = useGallerySavedFilters();
const sorts = [{ label:"最新收录",value:"newest" }, { label:"最早收录",value:"oldest" },{label:"标题 A–Z",value:"title_asc"}];
const sortModel = computed({get: () => state.value.sort, set: (value: string) => void catalog.setSort(value as GalleryCatalogSort)});
const refinements = computed(() => [
  ...state.value.categories.map(value => ({key:`category:${value}`,label:data.value?.categories.find(item => item.slug === value)?.name || value,kind:"category",value})),
  ...state.value.facets.map(value => {const [group,key]=value.split(":");const facet=data.value?.facets.find(item => item.slug === group);return {key:`facet:${value}`,label:`${facet?.name || group}：${facet?.values.find(item => item.slug === key)?.name || key}`,kind:"facet",value};}),
  ...(state.value.q ? [{key:"q",label:state.value.q,kind:"q",value:state.value.q}] : []),
  ...(state.value.tag ? [{key:"tag",label:`#${state.value.tag}`,kind:"tag",value:state.value.tag}] : []),
  ...(state.value.orientation ? [{key:"orientation",label:({landscape:"横图",portrait:"竖图",square:"方图"} as Record<string,string>)[state.value.orientation]!,kind:"orientation",value:state.value.orientation}] : []),
  ...(state.value.minLongEdge ? [{key:"minLongEdge",label:`长边 ≥ ${state.value.minLongEdge}px`,kind:"minLongEdge",value:String(state.value.minLongEdge)}] : []),
]);
function category(value: string) {catalog.toggleCategory(value);void catalog.apply();}
function facet(value: string) {catalog.toggleFacet(value);void catalog.apply();}
function tag(value: string) {catalog.toggleTag(value);void catalog.apply();}
function technical(value: Partial<GalleryCatalogState>) {void catalog.apply(value);}
function remove(item: {kind:string;value:string}) {
 if(item.kind === "category") void catalog.removeCategory(item.value);
 else if(item.kind === "facet") void catalog.removeFacet(item.value);
 else if(item.kind === "tag") void catalog.removeTag();
 else if(item.kind === "q") void catalog.removeSearch();
 else void catalog.apply({ [item.kind]: undefined });
}
async function save() {
 if (!loggedIn.value) {await login();return;}
 if (!saveName.value.trim() || saving.value) return;
 saving.value=true;saveError.value="";
 const q=state.value;
 try {
  await call<{filter:GallerySavedFilter}>("/me/saved-filters",{method:"POST",body:{name:saveName.value.trim(),query:{q:q.q,categories:q.categories.join(","),facets:q.facets.join(","),tag:q.tag,matchAll:Boolean(q.matchAll),orientation:q.orientation || "",minLongEdge:q.minLongEdge || 0}}});
  await refreshSaved();saveOpen.value=false;saveName.value="";
 } catch(reason) {saveError.value=galleryFailureMessage(reason,"筛选没有保存，请检查名称后重试");}
 finally {saving.value=false;}
}
</script>
<template>
 <div class="gallery-catalog" :class="{ 'without-inspector': !filtersOpen }">
  <section class="catalog-collection" aria-label="图片集合">
   <div class="catalog-toolbar"><div class="catalog-title"><h1>{{ refinements.length ? '筛选结果' : '全部图片' }}</h1><span v-if="data">{{ data.total }} 张</span></div><div class="catalog-actions"><USelect v-model="sortModel" :items="sorts" size="sm" aria-label="图片排序" /><UTooltip text="网格"><UButton icon="i-tabler-layout-grid" size="sm" color="neutral" :variant="state.view === 'grid' ? 'soft' : 'ghost'" aria-label="网格视图" @click="catalog.setView('grid')" /></UTooltip><UTooltip text="瀑布流"><UButton icon="i-tabler-columns-3" size="sm" color="neutral" :variant="state.view === 'masonry' ? 'soft' : 'ghost'" aria-label="瀑布流视图" @click="catalog.setView('masonry')" /></UTooltip><UButton class="desktop-filter-toggle" icon="i-tabler-adjustments-horizontal" size="sm" :variant="filtersOpen ? 'outline' : 'ghost'" label="图片筛选" @click="void (filtersOpen = !filtersOpen)" /><UButton class="mobile-filter-toggle" icon="i-tabler-adjustments-horizontal" size="sm" variant="outline" label="筛选" @click="void (mobileOpen = true)" /></div></div>
   <div v-if="refinements.length" class="catalog-refinements"><button v-for="item in refinements" :key="item.key" type="button" :aria-label="`移除筛选：${item.label}`" @click="remove(item)">{{ item.label }}<UIcon name="i-tabler-x" /></button><UButton label="清空" size="xs" color="neutral" variant="ghost" @click="catalog.clear()" /><UButton icon="i-tabler-bookmark-plus" label="保存筛选" size="xs" variant="outline" @click="void (saveOpen = true)" /></div>
   <div v-if="pending" class="catalog-skeleton" aria-label="正在加载图片"><USkeleton v-for="i in 9" :key="i" class="aspect-[3/4] rounded-md" /></div>
   <UAlert v-else-if="error" color="error" variant="subtle" title="图片没有加载成功"><template #actions><UButton label="重试" @click="refresh()" /></template></UAlert>
   <div v-else-if="!data?.items.length" class="catalog-empty"><UIcon name="i-tabler-photo-search" /><h2>{{ refinements.length ? '没有符合条件的图片' : '还没有公开图片' }}</h2><p>{{ refinements.length ? '减少一些条件，再看看。' : '从投稿开始收集第一张图片。' }}</p><UButton v-if="refinements.length" label="清除筛选" variant="outline" @click="catalog.clear()" /><UButton v-else to="/submit" label="投稿图片" variant="outline" /></div>
   <GalleryImageGrid v-else :items="data.items" priority :view="state.view" />
   <CollectionPagination v-if="data && data.total > data.size" :model-value="data.page" :total-pages="galleryPageCount(data)" @update:model-value="catalog.setPage" />
  </section>
  <aside v-if="filtersOpen" class="catalog-desktop-inspector" aria-label="图片筛选"><GalleryFilterInspector :categories="data?.categories || []" :facets="data?.facets || []" :tags="data?.tags || []" :selected-categories="selectedCategories" :selected-facets="selectedFacets" :selected-tag="selectedTag" :match-all="state.matchAll" :orientation="state.orientation" :min-long-edge="state.minLongEdge" @category="category" @facet="facet" @tag="tag" @technical="technical" @close="filtersOpen = false" /></aside>
  <USlideover v-model:open="mobileOpen" title="图片筛选"><template #body><GalleryFilterInspector :categories="data?.categories || []" :facets="data?.facets || []" :tags="data?.tags || []" :selected-categories="selectedCategories" :selected-facets="selectedFacets" :selected-tag="selectedTag" :match-all="state.matchAll" :orientation="state.orientation" :min-long-edge="state.minLongEdge" @category="category" @facet="facet" @tag="tag" @technical="technical" @close="mobileOpen = false" /></template></USlideover>
  <UModal v-model:open="saveOpen" title="保存常用筛选" description="保存条件组合，后续收录的图片也会出现在结果中。"><template #body><form id="gallery-save-filter" class="space-y-4" @submit.prevent="save"><UFormField label="名称" required :error="saveError"><UInput v-model="saveName" maxlength="60" placeholder="例如：喜欢的角色壁纸" class="w-full" autofocus /></UFormField></form></template><template #footer><UButton label="取消" color="neutral" variant="ghost" @click="void (saveOpen = false)" /><UButton form="gallery-save-filter" type="submit" :label="loggedIn ? '保存筛选' : '登录后保存'" :loading="saving" /></template></UModal>
 </div>
</template>
<style scoped>
.gallery-catalog{display:grid;grid-template-columns:minmax(0,1fr) 17rem;gap:1.25rem;align-items:start}.gallery-catalog.without-inspector{grid-template-columns:minmax(0,1fr)}.catalog-collection{min-width:0}.catalog-toolbar{display:flex;align-items:center;justify-content:space-between;gap:.65rem;margin-bottom:1rem;flex-wrap:wrap}.catalog-title{display:flex;align-items:baseline;gap:.6rem}.catalog-title h1{font-size:1.3rem;font-weight:650;letter-spacing:-.025em}.catalog-title>span{font-size:.72rem;color:var(--gallery-muted)}.catalog-actions{display:flex;align-items:center;gap:.35rem}.catalog-refinements{display:flex;align-items:center;flex-wrap:wrap;gap:.35rem;margin-bottom:1rem}.catalog-refinements>button{display:flex;align-items:center;gap:.45rem;padding:.35rem .5rem;border:1px solid var(--gallery-line);border-radius:.3rem;background:var(--gallery-panel-strong);font-size:.65rem}.catalog-refinements .iconify{font-size:.7rem}.catalog-desktop-inspector{position:sticky;top:5rem;max-height:calc(100dvh - 6rem);overflow:auto;scrollbar-width:thin;padding-left:1.2rem;border-left:1px solid var(--gallery-line)}.catalog-skeleton{display:grid;grid-template-columns:repeat(auto-fill,minmax(15rem,1fr));gap:.65rem}.catalog-empty{display:flex;flex-direction:column;align-items:center;justify-content:center;gap:.8rem;padding:5rem 1rem;border:1px dashed var(--gallery-line);border-radius:.5rem;color:var(--gallery-muted)}.catalog-empty>.iconify{font-size:2rem}.catalog-empty h2{font-size:.9rem;color:var(--gallery-ink)}.catalog-empty p{font-size:.75rem}.mobile-filter-toggle{display:none}
@media(max-width:1150px){.gallery-catalog{grid-template-columns:minmax(0,1fr)}.catalog-desktop-inspector,.desktop-filter-toggle{display:none}.mobile-filter-toggle{display:inline-flex}}@media(max-width:500px){.catalog-title h1{font-size:1.15rem}.catalog-toolbar{gap:.8rem}.catalog-actions{margin-left:auto}}
@media(max-width:700px){.catalog-skeleton{grid-template-columns:repeat(2,minmax(0,1fr))}}
</style>
