<script setup lang="ts">
import type { GalleryClassificationNode } from "~/types/gallery";
import { classificationFamilies } from "~/utils/classification-tree";

const props = defineProps<{
  nodes: GalleryClassificationNode[];
  selectedIds: string[];
  search?: string;
  counts?: boolean;
}>();
defineEmits<{ select: [node: GalleryClassificationNode] }>();
const families = computed(() => classificationFamilies(props.nodes, props.search));
</script>

<template>
  <div class="work-choices">
    <section v-for="family in families" :key="family.root.id" class="work-family" :aria-label="family.root.name">
      <GalleryVisualChoice :node="family.root" compact :aria-pressed="selectedIds.includes(family.root.id)" :selected="selectedIds.includes(family.root.id)" :count="counts ? family.root.count : undefined" @select="$emit('select', family.root)" />
      <div v-if="family.children.length" class="work-characters">
        <GalleryVisualChoice v-for="node in family.children" :key="node.id" :node="node" compact :aria-pressed="selectedIds.includes(node.id)" :selected="selectedIds.includes(node.id)" :count="counts ? node.count : undefined" @select="$emit('select', node)" />
      </div>
    </section>
    <p v-if="!families.length" class="text-xs text-muted">没有匹配的作品或角色</p>
  </div>
</template>

<style scoped>
.work-choices{display:grid;gap:.7rem;min-width:0}.work-family{display:grid;justify-items:start;gap:.4rem;min-width:0}.work-characters{display:flex;flex-wrap:wrap;gap:.4rem;margin-left:.6rem;padding-left:.6rem;border-left:1px solid var(--gallery-line,var(--ui-border))}
</style>
