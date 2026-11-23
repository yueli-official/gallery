<script setup lang="ts">
import type { GalleryImageCard } from "~/types/gallery";
const props = withDefaults(defineProps<{ items: GalleryImageCard[]; priority?: boolean; headingLevel?: "h2" | "h3"; ranked?: boolean; numbered?: boolean; view?: "grid" | "masonry" | "related" }>(), { view: "masonry" });
</script>
<template>
  <div class="gallery-image-grid" :class="{ 'is-grid': view === 'grid', 'is-related': view === 'related' }">
    <article v-for="(image,index) in items" :key="image.id" class="gallery-image-card group">
      <NuxtLink :to="`/images/${encodeURIComponent(image.publicId)}`" class="gallery-image-link">
        <img v-bind="galleryImageSources(image, 'masonry', Boolean(priority && index < 6))" :alt="image.altText || image.title" :width="image.width" :height="image.height" />
        <span v-if="ranked" class="image-rank" :class="{ 'is-podium': index < 3 }">{{ String(index + 1).padStart(2, "0") }}</span>
        <div class="image-caption"><component :is="props.headingLevel || 'h2'">{{ image.title }}</component><p><span v-if="image.primaryCategory" class="image-category">{{ image.primaryCategory }}</span><span>{{ image.width }} × {{ image.height }}</span></p></div>
      </NuxtLink>
      <GalleryFavoriteButton :image-id="image.id" :title="image.title" />
    </article>
  </div>
</template>
<style scoped>
.gallery-image-grid{columns:15rem;column-gap:.65rem}.gallery-image-card{position:relative;break-inside:avoid;margin-bottom:.65rem;border:1px solid var(--gallery-line);border-radius:.45rem;overflow:hidden;background:var(--gallery-panel)}.gallery-image-link{position:relative;display:block;overflow:hidden}.gallery-image-link img{display:block;width:100%;height:auto;transition:transform .4s}.gallery-image-card:hover img{transform:scale(1.015)}.image-caption{position:absolute;bottom:0;left:0;right:0;padding:2.8rem .65rem .6rem;background:linear-gradient(transparent,rgba(7,12,19,.88));color:#f8f5ed}.image-caption h2,.image-caption h3{font-size:.73rem;font-weight:550;line-height:1.5;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.image-caption p{display:flex;align-items:center;justify-content:space-between;gap:.35rem;font-size:.58rem;margin-top:.3rem;color:#d5d9df}.image-category{border:1px solid #ffffff38;padding:.08rem .3rem;border-radius:.2rem;background:#ffffff12}.image-rank{position:absolute;top:.55rem;left:.55rem;display:grid;place-items:center;min-width:1.75rem;height:1.75rem;border:1px solid #ffffff40;border-radius:.3rem;background:#0c1118b3;color:#fff;font:500 .72rem var(--font-mono);backdrop-filter:blur(8px)}.image-rank.is-podium{background:var(--gallery-accent);border-color:var(--gallery-accent);color:var(--gallery-on-accent)}.is-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(15rem,1fr));align-items:start;gap:.65rem;column-count:unset}.is-grid .gallery-image-card{margin-bottom:0}.is-grid img{aspect-ratio:3/4;object-fit:contain;background:var(--gallery-viewer-canvas)}
.is-related{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:.65rem;column-count:unset}.is-related .gallery-image-card{margin-bottom:0}.is-related img{aspect-ratio:4/3;object-fit:cover}
@media(max-width:700px){.gallery-image-grid{column-width:auto;column-count:2}.is-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.image-caption{padding:2.4rem .45rem .4rem}.image-caption h2,.image-caption h3{font-size:.66rem}.image-caption p{font-size:.53rem}}
@container gallery-related (max-width:51.99rem){.is-related{grid-template-columns:repeat(3,minmax(0,1fr))}}
@container gallery-related (max-width:35.99rem){.is-related{grid-template-columns:repeat(2,minmax(0,1fr))}}
@container gallery-related (min-width:66rem){.is-related{grid-template-columns:repeat(6,minmax(0,1fr))}}
</style>
