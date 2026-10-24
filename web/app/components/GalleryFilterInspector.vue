<script setup lang="ts">
import type { GalleryClassificationNode, GalleryFacet, GalleryTagCandidate } from "~/types/gallery";
defineProps<{ categories: GalleryClassificationNode[]; facets: GalleryFacet[]; tags: GalleryTagCandidate[]; selectedCategories: string[]; selectedFacets: string[]; selectedTag: string; matchAll?: boolean; orientation?: string; minLongEdge?: number }>();
defineEmits<{ category: [slug: string]; facet: [key: string]; tag: [slug: string]; technical: [value: { matchAll?: boolean; orientation?: string; minLongEdge?: number }]; close: [] }>();
const search = ref("");
const expanded = ref<string[]>([]);
function values(group: GalleryFacet) { const term=search.value.trim().toLocaleLowerCase(); const found=group.values.filter(node => !term || node.name.toLocaleLowerCase().includes(term)); return term || expanded.value.includes(group.id) ? found : found.slice(0,6); }
</script>
<template>
  <div class="filter-inspector">
    <div class="inspector-heading"><h2>图片筛选</h2><UButton icon="i-tabler-x" color="neutral" variant="ghost" size="xs" aria-label="收起筛选" @click="$emit('close')" /></div>
    <UInput v-model="search" icon="i-tabler-search" placeholder="查找分类或维度" aria-label="查找分类或维度" size="sm" class="w-full" />
    <section class="filter-group"><h3>分类</h3><div class="visual-options"><GalleryVisualChoice v-for="node in categories.filter(item => !search || item.name.includes(search))" :key="node.id" :node="node" kind="category" :selected="selectedCategories.includes(node.slug)" :count="node.count" @select="$emit('category', node.slug)" /></div></section>
    <section v-for="group in facets.filter(item => item.values.length && item.slug !== 'orientation')" :key="group.id" class="filter-group">
      <div class="filter-group-heading"><h3>{{ group.name }}</h3><button v-if="group.values.length > 6" type="button" @click="expanded = expanded.includes(group.id) ? expanded.filter(id => id !== group.id) : [...expanded,group.id]">{{ expanded.includes(group.id) ? '收起' : '查看全部' }}</button></div>
      <div class="visual-options"><GalleryVisualChoice v-for="node in values(group)" :key="node.id" :node="node" :selected="selectedFacets.includes(`${group.slug}:${node.slug}`)" :count="node.count" @select="$emit('facet', `${group.slug}:${node.slug}`)" /></div>
    </section>
    <section v-if="selectedFacets.length > 1" class="filter-group"><h3>同一维度多选时</h3><div class="technical-options"><button type="button" :class="{ selected: !matchAll }" @click="$emit('technical',{ matchAll: false })">任一匹配</button><button type="button" :class="{ selected: matchAll }" @click="$emit('technical',{ matchAll: true })">同时出现</button></div></section>
    <section class="filter-group"><h3>图片方向</h3><div class="technical-options"><button v-for="item in [{ label:'全部',value:'' },{label:'横图',value:'landscape'},{label:'竖图',value:'portrait'},{label:'方图',value:'square'}]" :key="item.value" type="button" :class="{ selected: (orientation || '') === item.value }" @click="$emit('technical',{ orientation: item.value || undefined })">{{ item.label }}</button></div></section>
    <section class="filter-group"><h3>分辨率</h3><div class="technical-options"><button v-for="item in [{ label:'不限',value:0 },{label:'2K+',value:2560},{label:'4K+',value:3840}]" :key="item.value" type="button" :class="{ selected: (minLongEdge || 0) === item.value }" @click="$emit('technical',{ minLongEdge: item.value || undefined })">{{ item.label }}</button></div><p class="filter-help">按图片长边筛选，竖图同样适用。</p></section>
    <section v-if="tags.length" class="filter-group"><h3>标签</h3><div class="filter-tags"><button v-for="tag in tags" :key="tag.id" type="button" :class="{ selected: selectedTag === tag.slug }" @click="$emit('tag', tag.slug)">{{ tag.name }}</button></div></section>
  </div>
</template>
<style scoped>
.filter-inspector{min-width:0}.inspector-heading,.filter-group-heading{display:flex;align-items:center;justify-content:space-between;gap:.5rem}.inspector-heading{margin-bottom:1rem}.inspector-heading h2{font-size:1.08rem;font-weight:650}.filter-group{padding:1.15rem 0;border-bottom:1px solid var(--gallery-line)}.filter-group h3{font-size:.78rem;font-weight:600;margin-bottom:.7rem}.filter-group-heading h3{margin:0}.filter-group-heading{margin-bottom:.7rem}.filter-group-heading button{font-size:.63rem;color:var(--gallery-accent)}.visual-options{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:.45rem}.technical-options{display:flex;gap:.3rem}.technical-options button{flex:1;font-size:.65rem;border:1px solid var(--gallery-line);border-radius:.3rem;padding:.45rem .3rem;background:var(--gallery-panel-strong)}.technical-options button.selected,.filter-tags button.selected{color:var(--gallery-accent);border-color:var(--gallery-accent);background:var(--gallery-accent-soft)}.filter-help{font-size:.6rem;color:var(--gallery-muted);margin-top:.5rem}.filter-tags{display:flex;flex-wrap:wrap;gap:.35rem}.filter-tags button{border:1px solid var(--gallery-line);border-radius:.3rem;font-size:.65rem;padding:.3rem .5rem}
</style>
