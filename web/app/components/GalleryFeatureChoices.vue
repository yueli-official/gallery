<script setup lang="ts">
import type { GalleryClassificationNode } from "~/types/gallery";

const props = defineProps<{
  nodes: GalleryClassificationNode[];
  selectedIds: string[];
  search?: string;
  counts?: boolean;
}>();
const emit = defineEmits<{ select: [node: GalleryClassificationNode] }>();
const trigger = useTemplateRef<HTMLElement>("trigger");
function remove(node: GalleryClassificationNode) {
  emit("select", node);
  nextTick(() => trigger.value?.focus());
}
const searchText = ref("");
const term = computed(() => (props.search ?? searchText.value).trim().toLocaleLowerCase());
const byId = computed(() => new Map(props.nodes.map(node => [node.id, node])));
const parents = computed(() => new Set(props.nodes.map(node => node.parentId).filter(Boolean)));
// Only concrete, non-root leaves can be assigned. Directory headings never
// become selected values, even when a filtered projection has no children.
const leaves = computed(() => props.nodes.filter(node => node.parentId && !parents.value.has(node.id)));
const selected = computed(() => leaves.value.filter(node => props.selectedIds.includes(node.id)));
function path(node: GalleryClassificationNode) {
  const names: string[] = [];
  const seen = new Set<string>();
  let current: GalleryClassificationNode | undefined = node;
  while (current && !seen.has(current.id)) {
    seen.add(current.id);
    names.unshift(current.name, current.slug);
    current = byId.value.get(current.parentId || "");
  }
  return names.join(" ").toLocaleLowerCase();
}
const groups = computed(() => props.nodes.filter(node => !node.parentId).map(root => {
  const belongs = (node: GalleryClassificationNode) => {
    const seen = new Set<string>();
    let current: GalleryClassificationNode | undefined = node;
    while (current && !seen.has(current.id)) {
      if (current.parentId === root.id) return true;
      seen.add(current.id);
      current = byId.value.get(current.parentId || "");
    }
    return false;
  };
  const values = leaves.value.filter(node => belongs(node) && (!term.value || path(node).includes(term.value)));
  const sections = props.nodes.filter(node => node.parentId && belongs(node) && parents.value.has(node.id)).map(node => ({
    node, values: values.filter(value => value.parentId === node.id),
  })).filter(section => section.values.length);
  const direct = values.filter(value => value.parentId === root.id);
  return { root, sections, direct, values };
}).filter(group => group.values.length));
</script>

<template>
  <div class="feature-choices">
    <div v-if="selected.length" class="feature-selected" aria-label="已选图片特征">
      <button v-for="node in selected" :key="node.id" type="button" class="feature-chip selected" :aria-label="`移除特征：${node.name}`" @click="remove(node)">{{ node.name }} <span aria-hidden="true">×</span></button>
    </div>
    <details class="feature-picker" :open="Boolean(search?.trim())">
      <summary ref="trigger">{{ selected.length ? `调整特征（${selected.length}）` : '选择特征' }}</summary>
      <div class="feature-picker-body">
        <UInput v-if="search === undefined" v-model="searchText" size="sm" icon="i-tabler-search" placeholder="搜索图片特征" aria-label="搜索图片特征" class="w-full" @keydown.enter.prevent />
        <details v-for="group in groups" :key="group.root.id" class="feature-group" :open="Boolean(term)">
          <summary>{{ group.root.name }}<span v-if="group.values.some(node => selectedIds.includes(node.id))" class="feature-count">{{ group.values.filter(node => selectedIds.includes(node.id)).length }} 已选</span></summary>
          <div class="feature-sections">
            <section v-for="section in group.sections" :key="section.node.id">
              <h4>{{ section.node.name }}</h4>
              <div class="feature-values">
                <button v-for="node in section.values" :key="node.id" type="button" class="feature-chip" :class="{ selected: selectedIds.includes(node.id) }" :aria-pressed="selectedIds.includes(node.id)" @click="$emit('select', node)">{{ node.name }}<span v-if="counts" class="feature-count">{{ node.count }}</span></button>
              </div>
            </section>
            <div v-if="group.direct.length" class="feature-values">
              <button v-for="node in group.direct" :key="node.id" type="button" class="feature-chip" :class="{ selected: selectedIds.includes(node.id) }" :aria-pressed="selectedIds.includes(node.id)" @click="$emit('select', node)">{{ node.name }}<span v-if="counts" class="feature-count">{{ node.count }}</span></button>
            </div>
          </div>
        </details>
        <p v-if="!groups.length" class="text-xs text-muted">没有匹配的特征</p>
      </div>
    </details>
  </div>
</template>

<style scoped>
.feature-choices{display:grid;gap:.5rem;min-width:0;color:var(--gallery-ink,var(--ui-text))}.feature-selected,.feature-values{display:flex;flex-wrap:wrap;gap:.35rem}.feature-picker,.feature-group{min-width:0;border:1px solid var(--gallery-line,var(--ui-border));border-radius:.4rem}.feature-picker summary,.feature-group summary{cursor:pointer;padding:.5rem .6rem;font-size:.75rem;overflow-wrap:anywhere}.feature-picker-body{display:grid;gap:.5rem;padding:0 .5rem .5rem}.feature-group{border-radius:.3rem}.feature-group summary{font-weight:600}.feature-sections{display:grid;gap:.7rem;padding:0 .6rem .7rem}.feature-sections h4{margin:0 0 .35rem;font-size:.7rem;color:var(--gallery-muted,var(--ui-text-muted))}.feature-chip{display:inline-flex;align-items:center;gap:.3rem;max-width:100%;padding:.3rem .5rem;border:1px solid var(--gallery-line,var(--ui-border));border-radius:.3rem;font-size:.7rem;text-align:left;overflow-wrap:anywhere;background:var(--gallery-panel-strong,var(--ui-bg))}.feature-chip:hover{border-color:var(--gallery-accent,var(--ui-primary))}.feature-chip.selected{color:var(--gallery-accent,var(--ui-primary));border-color:currentColor;background:var(--gallery-accent-soft,var(--ui-bg-accented))}.feature-chip:focus-visible,summary:focus-visible{outline:2px solid var(--gallery-accent,var(--ui-primary));outline-offset:2px}.feature-count{margin-left:.35rem;font-size:.65rem;color:var(--gallery-muted,var(--ui-text-muted))}
</style>
