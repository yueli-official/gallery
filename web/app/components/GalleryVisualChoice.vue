<script setup lang="ts">
import type { GalleryClassificationNode } from "~/types/gallery";
const props = withDefaults(defineProps<{ node: GalleryClassificationNode; kind?: "category" | "facet_value"; selected?: boolean; count?: number; compact?: boolean }>(), { kind: "facet_value" });
defineEmits<{ select: [] }>();
const site = useGallerySiteSettings();
const rule = computed(() => site.value?.appearance.classification.find(item => item.kind === props.kind && item.nodeId === props.node.id));
const defaultIcons: Record<string, string> = { wallpaper: "device-desktop", illustration: "palette", photography: "camera", people: "user", nature: "leaf", landscape: "mountain", architecture: "building", portrait: "user", square: "square" };
const icon = computed(() => `i-tabler-${rule.value?.icon || defaultIcons[props.node.slug] || 'photo'}`);
</script>
<template>
  <button type="button" class="visual-choice" :class="{ selected, 'with-image': rule?.mode === 'image', 'is-compact': compact }" :aria-pressed="compact ? undefined : Boolean(selected)" :title="node.name" @click="$emit('select')">
    <span class="visual-choice-media"><img v-if="rule?.mode === 'image'" :src="galleryRendition(rule, 'thumbnail')" :alt="node.name" loading="lazy" /><UIcon v-else :name="icon" /><span v-if="selected" class="choice-check"><UIcon name="i-tabler-check" /></span></span>
    <span class="visual-choice-name">{{ node.name }}</span><span v-if="count !== undefined" class="visual-choice-count">{{ count }}</span>
  </button>
</template>
<style scoped>
.visual-choice{display:flex;flex-direction:column;min-width:0;gap:.32rem;text-align:center;color:var(--gallery-ink);border:1px solid var(--gallery-line);border-radius:.4rem;background:var(--gallery-panel-strong);padding:.45rem .2rem;transition:border-color .15s}.visual-choice:hover{border-color:var(--gallery-accent)}.visual-choice.selected{border-color:var(--gallery-accent);background:var(--gallery-accent-soft);color:var(--gallery-accent)}.visual-choice-media{display:grid;place-items:center;position:relative;width:100%;height:2.4rem}.visual-choice-media>.iconify{width:1.45rem;height:1.45rem}.visual-choice-media img{width:100%;height:100%;object-fit:cover;border-radius:.3rem}.with-image{padding:0 0 .35rem;border-color:transparent;background:transparent}.with-image .visual-choice-media{height:auto;aspect-ratio:1.15}.with-image.selected{border-color:var(--gallery-accent)}.visual-choice-name{font-size:.65rem;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:100%;padding:0 .1rem}.visual-choice-count{font-size:.57rem;color:var(--gallery-muted)}.choice-check{position:absolute;top:.2rem;right:.2rem;display:grid;place-items:center;background:var(--gallery-accent);color:var(--gallery-canvas);width:1rem;height:1rem;border-radius:50%;font-size:.7rem}
.is-compact{flex-direction:row;align-items:center;gap:.35rem;max-width:100%;padding:.35rem .5rem;text-align:left}.is-compact .visual-choice-media{flex:none;width:1.15rem;height:1.15rem;aspect-ratio:auto}.is-compact .visual-choice-media>.iconify{width:1rem;height:1rem}.is-compact .visual-choice-media img{border-radius:.2rem}.is-compact .visual-choice-name{padding:0;font-size:.7rem}.is-compact.with-image{border-color:var(--gallery-line);background:var(--gallery-panel-strong)}.is-compact.with-image:hover{border-color:var(--gallery-accent)}
</style>
