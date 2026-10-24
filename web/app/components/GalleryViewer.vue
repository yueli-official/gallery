<script setup lang="ts">
import { assetUploadURL } from "@yueli/asset-nuxt/upload";
import type {
  GalleryImage,
  GalleryImageCard,
  GalleryRelatedImage,
} from "~/types/gallery";

const props = withDefaults(
  defineProps<{
    image?: GalleryImage;
    status?: "idle" | "pending" | "success" | "error";
    failed?: boolean;
    previous?: GalleryImageCard;
    next?: GalleryImageCard;
    related?: GalleryRelatedImage[];
    relatedPending?: boolean;
  }>(),
  {
    status: "idle",
    failed: false,
    related: () => [],
    relatedPending: false,
  },
);
const emit = defineEmits<{
  close: [];
  retry: [];
  navigate: [imageId: string];
}>();

const image = computed(() => props.image);
const reportOpen = ref(false);
const reportKind = ref<"report" | "source_correction">("report");
const zoom = ref(1);
const fullscreen = ref(false);
const moreDetails = ref<HTMLDetailsElement>();
const pointerStart = ref<{ x: number; y: number }>();
let qualifiedViewTimer: ReturnType<typeof setTimeout> | undefined;
const { favoritePending, toggleFavorite, shareImage, track } =
  useGalleryImageActions(image);

const canZoomOut = computed(() => zoom.value > 1);
const canZoomIn = computed(() => zoom.value < 2.5);
const viewerLabel = computed(() =>
  props.image ? `图片详情：${props.image.title}` : "图片详情",
);
const mediaStyle = computed(() => ({
  "--gallery-image-matte": props.image?.dominantColor || "var(--gallery-panel)",
}));
const filmstripItems = computed<GalleryImageCard[]>(() => {
  const candidates = [
    props.previous,
    props.image,
    props.next,
    ...props.related,
  ].filter((item): item is GalleryImageCard => Boolean(item));
  const unique = new Map<string, GalleryImageCard>();

  for (const item of candidates) {
    if (!unique.has(item.id)) unique.set(item.id, item);
  }

  return [...unique.values()].slice(0, 12);
});

function resetMedia(): void {
  zoom.value = 1;
}

function changeZoom(delta: number): void {
  zoom.value = Math.min(2.5, Math.max(1, zoom.value + delta));
}

async function toggleFullscreen(): Promise<void> {
  if (!import.meta.client) return;
  if (document.fullscreenElement) {
    await document.exitFullscreen();
  } else {
    await document.documentElement.requestFullscreen();
  }
}

function onFullscreenChange(): void {
  fullscreen.value = Boolean(document.fullscreenElement);
}

function navigate(target?: GalleryImageCard): void {
  if (!target) return;
  resetMedia();
  emit("navigate", target.id);
}

function share(): void {
  if (!props.image || !import.meta.client) return;
  const url = new URL(
    `/images/${encodeURIComponent(props.image.id)}`,
    window.location.origin,
  ).href;
  void shareImage(url);
}

function closeMoreDetails(): void {
  if (moreDetails.value) moreDetails.value.open = false;
}

function onDocumentPointerDown(event: PointerEvent): void {
  const details = moreDetails.value;
  const target = event.target;
  if (
    !details?.open ||
    !(target instanceof Node) ||
    details.contains(target)
  )
    return;
  closeMoreDetails();
}

function onKeydown(event: KeyboardEvent): void {
  if (event.target instanceof Element && event.target.closest("input,textarea,select,[contenteditable=true]")) return;
  if (reportOpen.value) return;
  if (event.key === "Escape" && moreDetails.value?.open) {
    event.preventDefault();
    closeMoreDetails();
  } else if (event.key === "ArrowLeft" && props.previous) {
    event.preventDefault();
    navigate(props.previous);
  } else if (event.key === "ArrowRight" && props.next) {
    event.preventDefault();
    navigate(props.next);
  } else if ((event.key === "+" || event.key === "=") && canZoomIn.value) {
    event.preventDefault();
    changeZoom(0.5);
  } else if (event.key === "-" && canZoomOut.value) {
    event.preventDefault();
    changeZoom(-0.5);
  } else if (event.key === "0") {
    event.preventDefault();
    resetMedia();
  }
}

function onPointerDown(event: PointerEvent): void {
  if (zoom.value !== 1 || event.pointerType === "mouse") return;
  pointerStart.value = { x: event.clientX, y: event.clientY };
}

function onPointerUp(event: PointerEvent): void {
  const start = pointerStart.value;
  pointerStart.value = undefined;
  if (!start || zoom.value !== 1) return;
  const deltaX = event.clientX - start.x;
  const deltaY = event.clientY - start.y;
  if (Math.abs(deltaX) < 70 || Math.abs(deltaX) < Math.abs(deltaY) * 1.25)
    return;
  navigate(deltaX > 0 ? props.previous : props.next);
}

function beginQualifiedView(): void {
  if (qualifiedViewTimer) clearTimeout(qualifiedViewTimer);
  if (!props.image) return;
  qualifiedViewTimer = setTimeout(() => void track("qualified_view"), 2200);
}

watch(
  () => props.image?.id,
  () => {
    resetMedia();
    closeMoreDetails();
    beginQualifiedView();
  },
);

onMounted(() => {
  window.addEventListener("keydown", onKeydown);
  document.addEventListener("pointerdown", onDocumentPointerDown);
  document.addEventListener("fullscreenchange", onFullscreenChange);
  onFullscreenChange();
  beginQualifiedView();
});
onBeforeUnmount(() => {
  window.removeEventListener("keydown", onKeydown);
  document.removeEventListener("pointerdown", onDocumentPointerDown);
  document.removeEventListener("fullscreenchange", onFullscreenChange);
  if (qualifiedViewTimer) clearTimeout(qualifiedViewTimer);
});

const router = useRouter();
const { data: classificationOptions } = await useFetch<import("~/types/gallery").GallerySubmissionOptions>("/api/gallery/submission-options");
const assignedGroups = computed(() => (classificationOptions.value?.facets || []).map(group => ({ group, values: group.values.filter(node => image.value?.facets.some(item => item.facetId === group.id && item.valueId === node.id)) })).filter(item => item.values.length));
const category = computed(() => classificationOptions.value?.categories.find(node => node.slug === image.value?.primaryCategorySlug));
const fileFacts = ref<{url:string;mime:string;sizeBytes:number}>();
const factsFailed = ref(false);
const downloadPending = ref(false);
const downloadError = ref("");
const { can } = useGalleryMe();
async function loadFacts(){if(!image.value)return;factsFailed.value=false;try{fileFacts.value=await $fetch(`/api/gallery/images/${encodeURIComponent(image.value.id)}/download`);}catch{factsFailed.value=true;}}
onMounted(()=>void loadFacts());
watch(()=>image.value?.id,()=>{fileFacts.value=undefined;downloadError.value="";if(import.meta.client)void loadFacts();});
function facetLink(slug:string,value:string){void router.push({path:"/images",query:{facets:`${slug}:${value}`}});}
function gcd(a:number,b:number):number {while(b){const n=a%b;a=b;b=n;}return a||1;}
const ratio = computed(()=>{if(!image.value)return "";const divisor=gcd(image.value.width,image.value.height);return `${image.value.width/divisor}:${image.value.height/divisor}`;});
async function download(){
 if(!image.value||downloadPending.value)return;downloadPending.value=true;downloadError.value="";
 try{
  const facts=await $fetch<{url:string;mime:string;sizeBytes:number}>(`/api/gallery/images/${encodeURIComponent(image.value.id)}/download`);fileFacts.value=facts;
  const response=await fetch(assetUploadURL(facts.url));if(!response.ok)throw new Error("original_download_failed");
  const blob=await response.blob();const url=URL.createObjectURL(blob);const anchor=document.createElement("a");
  const extension=({"image/jpeg":"jpg","image/png":"png","image/webp":"webp","image/gif":"gif"} as Record<string,string>)[facts.mime] || "image";
  anchor.href=url;anchor.download=`${image.value.title.replace(/[\\/:*?"<>|]/g,"_")}.${extension}`;document.body.appendChild(anchor);anchor.click();anchor.remove();URL.revokeObjectURL(url);
 }catch(reason){downloadError.value=galleryFailureMessage(reason,"原图下载失败，请稍后重试");}finally{downloadPending.value=false;}
}
</script>
<template>
 <article class="gallery-detail" :aria-label="viewerLabel">
  <div class="detail-toolbar"><button type="button" @click="emit('close')"><UIcon name="i-tabler-arrow-left" />返回图片库</button><div><UButton icon="i-tabler-chevron-left" label="上一张" color="neutral" variant="ghost" size="sm" :disabled="!previous" @click="navigate(previous)" /><UButton trailing-icon="i-tabler-chevron-right" label="下一张" color="neutral" variant="ghost" size="sm" :disabled="!next" @click="navigate(next)" /></div></div>
  <div v-if="status==='pending' || status==='idle'" class="detail-loading"><USkeleton class="h-[65vh] rounded-md" /><USkeleton class="h-64 rounded-md" /></div>
  <UAlert v-else-if="failed || !image" color="error" variant="subtle" title="图片暂时无法查看"><template #actions><UButton label="重试" @click="emit('retry')" /><UButton to="/images" label="返回图库" color="neutral" variant="outline" /></template></UAlert>
  <template v-else>
   <div class="detail-layout">
    <section class="detail-viewer-column">
     <div class="detail-image-stage" :class="{ zoomed:zoom>1 }" :style="mediaStyle" @pointerdown="onPointerDown" @pointerup="onPointerUp"><img v-bind="galleryImageSources(image.assetId,'display',true)" :alt="image.altText || image.title" :width="image.width" :height="image.height" :style="{transform:`scale(${zoom})`}" /></div>
     <div class="detail-viewer-tools"><UTooltip text="缩小"><UButton icon="i-tabler-minus" color="neutral" variant="ghost" aria-label="缩小图片" :disabled="!canZoomOut" @click="changeZoom(-.5)" /></UTooltip><span>{{ Math.round(zoom*100) }}%</span><UTooltip text="放大"><UButton icon="i-tabler-plus" color="neutral" variant="ghost" aria-label="放大图片" :disabled="!canZoomIn" @click="changeZoom(.5)" /></UTooltip><UButton label="适应屏幕" color="neutral" variant="ghost" size="sm" @click="resetMedia" /><UTooltip text="全屏"><UButton icon="i-tabler-arrows-maximize" color="neutral" variant="ghost" aria-label="全屏查看" @click="toggleFullscreen" /></UTooltip></div>
     <div v-if="filmstripItems.length>1" class="detail-filmstrip" aria-label="相邻与相关图片"><button v-for="item in filmstripItems" :key="item.id" type="button" :class="{active:item.id===image.id}" :aria-label="item.id===image.id ? `当前图片：${item.title}` : `查看图片：${item.title}`" @click="navigate(item)"><img :src="galleryRendition(item.assetId,'thumbnail')" :alt="item.title" /></button></div>
    </section>
    <aside class="detail-information">
     <header><h1>{{ image.title }}</h1><p v-if="image.description">{{ image.description }}</p></header>
     <div class="detail-main-actions"><UButton icon="i-tabler-download" label="下载原图" :loading="downloadPending" @click="download" /><UButton :icon="image.favorited ? 'i-tabler-heart-filled' : 'i-tabler-heart'" :label="image.favorited ? '已收藏' : '收藏'" color="neutral" variant="outline" :loading="favoritePending" @click="toggleFavorite" /><UTooltip text="分享图片"><UButton icon="i-tabler-share-3" color="neutral" variant="ghost" aria-label="分享图片" @click="share" /></UTooltip></div>
     <p v-if="downloadError" class="text-xs text-error" role="alert">{{ downloadError }}</p>
     <section v-if="category" class="detail-metadata-group"><h2>分类</h2><div class="detail-choices"><GalleryVisualChoice :node="category" kind="category" @select="void router.push({path:'/images',query:{categories:category.slug}})" /></div></section>
     <section v-for="item in assignedGroups" :key="item.group.id" class="detail-metadata-group"><h2>{{ item.group.name }}</h2><div class="detail-choices"><GalleryVisualChoice v-for="node in item.values" :key="node.id" :node="node" @select="facetLink(item.group.slug,node.slug)" /></div></section>
     <section class="detail-metadata-group"><h2>文件信息</h2><dl><div><dt>分辨率</dt><dd>{{ image.width }} × {{ image.height }}</dd></div><div><dt>比例</dt><dd>{{ ratio }}</dd></div><template v-if="fileFacts"><div><dt>格式</dt><dd>{{ fileFacts.mime.replace('image/','').toUpperCase() }}</dd></div><div><dt>文件大小</dt><dd>{{ (fileFacts.sizeBytes/1024/1024).toFixed(2) }} MiB</dd></div></template></dl><p v-if="factsFailed" class="detail-facts-error">文件信息未能读取 <button type="button" @click="loadFacts">重试</button></p></section>
     <section v-if="image.tags.length" class="detail-metadata-group"><h2>标签</h2><div class="detail-tags"><NuxtLink v-for="tag in image.tags" :key="tag" :to="{path:'/images',query:{tag}}">{{ tag }}</NuxtLink></div></section>
     <section class="detail-metadata-group"><h2>来源</h2><a v-if="image.sourceUrl" :href="image.sourceUrl" target="_blank" rel="noopener noreferrer nofollow" referrerpolicy="no-referrer" class="detail-source">查看原始页面 <UIcon name="i-tabler-arrow-up-right" /></a><p v-else class="detail-facts-error">未填写</p></section>
     <div class="detail-secondary-actions"><UButton v-if="can('gallery.image.update')" :to="`/manage/images?q=${encodeURIComponent(image.title)}`" icon="i-tabler-edit" label="编辑信息" color="neutral" variant="ghost" size="xs" /><UButton icon="i-tabler-flag" label="报告问题" color="neutral" variant="ghost" size="xs" @click="reportKind='report';reportOpen=true" /><UButton icon="i-tabler-link" label="补充来源" color="neutral" variant="ghost" size="xs" @click="reportKind='source_correction';reportOpen=true" /></div>
    </aside>
   </div>
   <section v-if="related.length" class="detail-related"><div class="gallery-section-heading"><h2>相关图片</h2><NuxtLink to="/images" class="text-xs text-muted">查看全部 →</NuxtLink></div><GalleryImageGrid :items="related.slice(0,8)" heading-level="h3" /></section>
   <GalleryCommentSection :image-id="image.id" />
   <GalleryReportDialog v-model:open="reportOpen" :image-id="image.id" :kind="reportKind" />
  </template>
 </article>
</template>
<style scoped>
.detail-toolbar{display:flex;align-items:center;justify-content:space-between;gap:1rem;margin-bottom:1rem}.detail-toolbar>button{display:flex;align-items:center;gap:.5rem;font-size:.75rem;color:var(--gallery-muted)}.detail-toolbar>div{display:flex;gap:.25rem}.detail-layout,.detail-loading{display:grid;grid-template-columns:minmax(0,1fr) 19rem;gap:1.3rem;align-items:start}.detail-viewer-column{min-width:0}.detail-image-stage{display:flex;align-items:center;justify-content:center;overflow:hidden;border:1px solid var(--gallery-line);border-radius:.45rem .45rem 0 0;background:var(--gallery-viewer-canvas);height:min(68dvh,55rem);touch-action:pan-y}.detail-image-stage img{display:block;width:100%;height:100%;object-fit:contain;transform-origin:center;transition:transform .18s}.detail-image-stage.zoomed{overflow:auto}.detail-image-stage.zoomed img{object-fit:contain}.detail-viewer-tools{display:flex;align-items:center;gap:.3rem;padding:.4rem .6rem;border:1px solid var(--gallery-line);border-top:0;border-radius:0 0 .4rem .4rem;background:var(--gallery-panel)}.detail-viewer-tools>span{font-size:.7rem;min-width:3rem;text-align:center}.detail-filmstrip{display:flex;gap:.5rem;overflow:auto;margin-top:1rem;padding-bottom:.5rem}.detail-filmstrip button{flex:0 0 5.5rem;aspect-ratio:1;border:1px solid transparent;border-radius:.35rem;overflow:hidden;opacity:.75}.detail-filmstrip button.active{border-color:var(--gallery-accent);opacity:1}.detail-filmstrip img{width:100%;height:100%;object-fit:cover}.detail-information{position:sticky;top:5rem;max-height:calc(100dvh - 6rem);overflow:auto;scrollbar-width:thin;padding-left:1rem;border-left:1px solid var(--gallery-line)}.detail-information h1{font-size:1.2rem;font-weight:650;line-height:1.5}.detail-information header p{font-size:.73rem;line-height:1.8;color:var(--gallery-muted);margin-top:.6rem;white-space:pre-line;overflow-wrap:anywhere}.detail-main-actions{display:flex;gap:.4rem;flex-wrap:wrap;margin:1rem 0}.detail-metadata-group{padding:1rem 0;border-bottom:1px solid var(--gallery-line)}.detail-metadata-group h2{font-size:.78rem;font-weight:600;margin-bottom:.65rem}.detail-choices{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:.4rem}.detail-metadata-group dl{display:grid;gap:.5rem;font-size:.7rem}.detail-metadata-group dl>div{display:flex;justify-content:space-between;gap:1rem}.detail-metadata-group dt{color:var(--gallery-muted)}.detail-tags{display:flex;flex-wrap:wrap;gap:.35rem}.detail-tags>a{border:1px solid var(--gallery-line);background:var(--gallery-panel);border-radius:.3rem;padding:.25rem .5rem;font-size:.65rem}.detail-source{font-size:.72rem;color:var(--gallery-accent)}.detail-facts-error{font-size:.68rem;color:var(--gallery-muted)}.detail-facts-error button{color:var(--gallery-accent)}.detail-secondary-actions{display:flex;flex-wrap:wrap;gap:.15rem;padding-top:.8rem}.detail-related{margin:2rem 0;max-width:100%}
@media(max-width:1150px){.detail-layout,.detail-loading{grid-template-columns:minmax(0,1fr) 16rem}.detail-information{padding-left:.8rem}.detail-main-actions{gap:.25rem}}@media(max-width:800px){.detail-layout,.detail-loading{grid-template-columns:minmax(0,1fr)}.detail-information{position:static;max-height:none;padding:0;border-left:0}.detail-image-stage{height:65dvh}.detail-toolbar{gap:.3rem}.detail-choices{grid-template-columns:repeat(4,minmax(0,1fr))}.detail-related{margin:1.5rem 0}}
</style>
