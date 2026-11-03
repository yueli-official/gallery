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
const zoom = ref<number | null>(null);
const fullscreen = ref(false);
const fullscreenSupported = ref(false);
const viewerElement = ref<HTMLElement>();
const viewportElement = ref<HTMLElement>();
const viewport = ref({ width: 0, height: 0 });
const pan = ref({ x: 0, y: 0 });
const drag = ref<{ id: number; x: number; y: number; panX: number; panY: number; panning: boolean }>();
const viewerError = ref("");
let qualifiedViewTimer: ReturnType<typeof setTimeout> | undefined;
const { favoritePending, toggleFavorite, shareImage, track } =
  useGalleryImageActions(image);

const fitScale = computed(() => {
  if (!image.value || !viewport.value.width || !viewport.value.height) return 1;
  return Math.min(viewport.value.width / image.value.width, viewport.value.height / image.value.height, 1);
});
const imageScale = computed(() => zoom.value ?? fitScale.value);
const zoomPercent = computed(() => Math.round(imageScale.value * 100));
const canZoomOut = computed(() => imageScale.value > .01);
const canZoomIn = computed(() => imageScale.value < 4);
const canPan = computed(() => image.value && (image.value.width * imageScale.value > viewport.value.width || image.value.height * imageScale.value > viewport.value.height));
const viewerLabel = computed(() =>
  props.image ? `图片详情：${props.image.title}` : "图片详情",
);
const mediaStyle = computed(() => ({
  width: `${(image.value?.width || 0) * imageScale.value}px`,
  height: `${(image.value?.height || 0) * imageScale.value}px`,
  transform: `translate3d(${pan.value.x}px, ${pan.value.y}px, 0)`,
}));

function clampPan(x = pan.value.x, y = pan.value.y): void {
  const maxX = Math.max(0, ((image.value?.width || 0) * imageScale.value - viewport.value.width) / 2);
  const maxY = Math.max(0, ((image.value?.height || 0) * imageScale.value - viewport.value.height) / 2);
  pan.value = { x: Math.max(-maxX, Math.min(maxX, x)), y: Math.max(-maxY, Math.min(maxY, y)) };
}

watch(viewportElement, (element, _, onCleanup) => {
  if (!import.meta.client || !element) return;
  const observer = new ResizeObserver(([entry]) => {
    if (!entry) return;
    viewport.value = { width: entry.contentRect.width, height: entry.contentRect.height };
    clampPan();
  });
  observer.observe(element);
  onCleanup(() => observer.disconnect());
}, { flush: "post" });

function resetMedia(): void {
  zoom.value = null;
  pan.value = { x: 0, y: 0 };
}

function changeZoom(delta: number): void {
  setZoom(imageScale.value + delta);
}

function setZoom(value: number): void {
  zoom.value = Math.min(4, Math.max(.01, value));
  clampPan();
}

function onZoomInput(event: Event): void {
  setZoom(Number((event.target as HTMLInputElement).value) / 100);
}

async function toggleFullscreen(): Promise<void> {
  if (!import.meta.client || !viewerElement.value) return;
  viewerError.value = "";
  try {
    if (document.fullscreenElement === viewerElement.value) await document.exitFullscreen();
    else await viewerElement.value.requestFullscreen();
  } catch { viewerError.value = "暂时无法切换全屏，请重试。"; }
}

function onFullscreenChange(): void {
  fullscreen.value = document.fullscreenElement === viewerElement.value;
}

function navigate(target?: GalleryImageCard): void {
  if (!target) return;
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

function onKeydown(event: KeyboardEvent): void {
  if (event.defaultPrevented || (event.target instanceof Element && event.target.closest("input,textarea,select,[contenteditable=true],[role=menu],[role=dialog]"))) return;
  if (reportOpen.value) return;
  if (event.key === "ArrowLeft" && props.previous) {
    event.preventDefault();
    navigate(props.previous);
  } else if (event.key === "ArrowRight" && props.next) {
    event.preventDefault();
    navigate(props.next);
  } else if ((event.key === "+" || event.key === "=") && canZoomIn.value) {
    event.preventDefault();
    changeZoom(0.25);
  } else if (event.key === "-" && canZoomOut.value) {
    event.preventDefault();
    changeZoom(-0.25);
  } else if (event.key === "0") {
    event.preventDefault();
    resetMedia();
  }
}

function onPointerDown(event: PointerEvent): void {
  if (event.button !== 0 || !event.isPrimary) return;
  drag.value = { id: event.pointerId, x: event.clientX, y: event.clientY, panX: pan.value.x, panY: pan.value.y, panning: Boolean(canPan.value) };
  if (canPan.value) {
    event.preventDefault();
    viewportElement.value?.setPointerCapture(event.pointerId);
  }
}

function onPointerMove(event: PointerEvent): void {
  const start = drag.value;
  if (!start?.panning || start.id !== event.pointerId) return;
  clampPan(start.panX + event.clientX - start.x, start.panY + event.clientY - start.y);
}

function onPointerUp(event: PointerEvent): void {
  const start = drag.value;
  if (!start || start.id !== event.pointerId) return;
  drag.value = undefined;
  if (viewportElement.value?.hasPointerCapture(event.pointerId)) viewportElement.value.releasePointerCapture(event.pointerId);
  if (start.panning || event.pointerType === "mouse") return;
  const deltaX = event.clientX - start.x;
  const deltaY = event.clientY - start.y;
  if (Math.abs(deltaX) < 70 || Math.abs(deltaX) < Math.abs(deltaY) * 1.25)
    return;
  navigate(deltaX > 0 ? props.previous : props.next);
}

function onWheel(event: WheelEvent): void {
  if (event.ctrlKey || event.metaKey) {
    event.preventDefault();
    changeZoom(event.deltaY < 0 ? .1 : -.1);
  } else if (canPan.value) {
    event.preventDefault();
    clampPan(pan.value.x - event.deltaX, pan.value.y - event.deltaY);
  }
}

function beginQualifiedView(): void {
  if (qualifiedViewTimer) clearTimeout(qualifiedViewTimer);
  if (!props.image) return;
  qualifiedViewTimer = setTimeout(() => void track("qualified_view"), 2200);
}

watch(
  () => props.image?.id,
  () => {
    pan.value = { x: 0, y: 0 };
    drag.value = undefined;
    beginQualifiedView();
  },
);

onMounted(() => {
  window.addEventListener("keydown", onKeydown);
  document.addEventListener("fullscreenchange", onFullscreenChange);
  fullscreenSupported.value = Boolean(document.fullscreenEnabled && document.documentElement.requestFullscreen);
  onFullscreenChange();
  beginQualifiedView();
});
onBeforeUnmount(() => {
  window.removeEventListener("keydown", onKeydown);
  document.removeEventListener("fullscreenchange", onFullscreenChange);
  if (qualifiedViewTimer) clearTimeout(qualifiedViewTimer);
});

const router = useRouter();
const { data: classificationOptions } = await useFetch<import("~/types/gallery").GallerySubmissionOptions>("/api/gallery/submission-options");
const assignedGroups = computed(() => (classificationOptions.value?.facets || []).map(group => ({ group, values: group.values.filter(node => image.value?.facets.some(item => item.facetId === group.id && item.valueId === node.id)) })).filter(item => item.values.length));
const category = computed(() => classificationOptions.value?.categories.find(node => node.slug === image.value?.primaryCategorySlug));
const relatedBrowse = computed(() => {
  for (const slug of ["work"]) {
    const assignment = assignedGroups.value.find(item => item.group.slug === slug);
    const value = assignment?.values.find(node => node.parentId) || assignment?.values[0];
    if (assignment && value) return { label: `更多${value.name}图片`, to: { path: "/images", query: { facets: `${assignment.group.slug}:${value.slug}` } } };
  }
  if (category.value) return { label: `更多${category.value.name}图片`, to: { path: "/images", query: { categories: category.value.slug } } };
  return undefined;
});
const fileFacts = ref<{url:string;mime:string;sizeBytes:number}>();
const factsFailed = ref(false);
const downloadPending = ref(false);
const downloadError = ref("");
const downloadItems = computed(() => [
  { label: "WebP", icon: "i-tabler-photo", disabled: downloadPending.value, onSelect: () => { void download("webp"); } },
  { label: "JPG", icon: "i-tabler-file-type-jpg", disabled: downloadPending.value, onSelect: () => { void download("jpg"); } },
]);
const { can } = useGalleryMe();
async function loadFacts(){if(!image.value)return;factsFailed.value=false;try{fileFacts.value=await $fetch(`/api/gallery/images/${encodeURIComponent(image.value.id)}/download`);}catch{factsFailed.value=true;}}
onMounted(()=>void loadFacts());
watch(()=>image.value?.id,()=>{fileFacts.value=undefined;downloadError.value="";if(import.meta.client)void loadFacts();});
function facetLink(slug:string,value:string){void router.push({path:"/images",query:{facets:`${slug}:${value}`}});}
function gcd(a:number,b:number):number {while(b){const n=a%b;a=b;b=n;}return a||1;}
const ratio = computed(()=>{if(!image.value)return "";const divisor=gcd(image.value.width,image.value.height);return `${image.value.width/divisor}:${image.value.height/divisor}`;});
async function download(format: "webp" | "jpg" = "webp"){
 if(!image.value||downloadPending.value)return;downloadPending.value=true;downloadError.value="";
 try{
  const facts=await $fetch<{url:string;mime:string;sizeBytes:number}>(`/api/gallery/images/${encodeURIComponent(image.value.id)}/download`);fileFacts.value=facts;
  const response=await fetch(assetUploadURL(facts.url));if(!response.ok)throw new Error("image_download_failed");
  let blob=await response.blob();
  if(format==="jpg"){
   const bitmap=await createImageBitmap(blob);
   try {
    const canvas=document.createElement("canvas");canvas.width=bitmap.width;canvas.height=bitmap.height;
    const context=canvas.getContext("2d");if(!context)throw new Error("image_conversion_failed");
    context.fillStyle="#fff";context.fillRect(0,0,canvas.width,canvas.height);context.drawImage(bitmap,0,0);
    blob=await new Promise<Blob>((resolve,reject)=>canvas.toBlob(value=>value?resolve(value):reject(new Error("image_conversion_failed")),"image/jpeg",.95));
   } finally { bitmap.close(); }
  }
  const url=URL.createObjectURL(blob);const anchor=document.createElement("a");
  const extension=format;
  anchor.href=url;anchor.download=`${image.value.title.replace(/[\\/:*?"<>|]/g,"_")}.${extension}`;document.body.appendChild(anchor);anchor.click();anchor.remove();
  // The browser starts reading the blob asynchronously after the click.
  setTimeout(() => URL.revokeObjectURL(url), 1000);
 }catch(reason){downloadError.value=galleryFailureMessage(reason,"图片下载失败，请稍后重试");}finally{downloadPending.value=false;}
}
</script>
<template>
 <article class="gallery-detail" :aria-label="viewerLabel">
  <div class="detail-toolbar"><button type="button" @click="emit('close')"><UIcon name="i-tabler-arrow-left" />返回图片库</button></div>
  <div v-if="!image && (status==='pending' || status==='idle')" class="detail-loading"><USkeleton class="h-[65vh] rounded-md" /><USkeleton class="h-64 rounded-md" /></div>
  <UAlert v-else-if="failed || !image" color="error" variant="subtle" title="图片暂时无法查看"><template #actions><UButton label="重试" @click="emit('retry')" /><UButton to="/images" label="返回图库" color="neutral" variant="outline" /></template></UAlert>
  <template v-else>
   <div class="detail-layout">
    <section ref="viewerElement" class="detail-viewer-column" :aria-label="`图片查看器：${image.title}`">
     <div class="detail-image-stage" :aria-busy="status==='pending'">
      <div ref="viewportElement" class="detail-image-viewport" :class="{ 'can-pan':canPan, dragging:drag?.panning }" @pointerdown="onPointerDown" @pointermove="onPointerMove" @pointerup="onPointerUp" @pointercancel="drag=undefined" @wheel="onWheel">
       <img v-bind="galleryImageSources(image,'display',true)" :alt="image.altText || image.title" :width="image.width" :height="image.height" :style="mediaStyle" :draggable="false" />
      </div>
      <UButton v-if="fullscreenSupported && !fullscreen" class="detail-immersive-action" icon="i-tabler-slideshow" label="沉浸浏览" size="md" @click="toggleFullscreen" />
      <UButton class="detail-image-nav is-previous" icon="i-tabler-chevron-left" color="neutral" variant="soft" aria-label="上一张图片" :disabled="!previous || status==='pending'" @click="navigate(previous)" />
      <UButton class="detail-image-nav is-next" icon="i-tabler-chevron-right" color="neutral" variant="soft" aria-label="下一张图片" :disabled="!next || status==='pending'" @click="navigate(next)" />
      <span v-if="status==='pending'" class="detail-image-loading" role="status"><UIcon name="i-tabler-loader-2" class="animate-spin" />正在加载</span>
     </div>
     <div class="detail-viewer-tools" role="toolbar" aria-label="图片工具栏">
      <span class="detail-viewer-title">{{ image.title }}</span>
      <div class="detail-zoom-controls">
       <UButton icon="i-tabler-minus" color="neutral" variant="ghost" aria-label="缩小图片" :disabled="!canZoomOut" @click="changeZoom(-.25)" />
       <output class="detail-zoom-value" aria-label="当前图片缩放比例">{{ zoomPercent }}%</output>
       <UButton icon="i-tabler-plus" color="neutral" variant="ghost" aria-label="放大图片" :disabled="!canZoomIn" @click="changeZoom(.25)" />
       <input class="detail-zoom-slider" type="range" min="1" max="400" step="1" :value="zoomPercent" aria-label="图片缩放比例" @input="onZoomInput" />
      </div>
      <div class="detail-fit-controls">
       <UButton icon="i-tabler-arrows-minimize" label="适应窗口" color="neutral" variant="ghost" size="sm" :aria-pressed="zoom===null" @click="resetMedia" />
      </div>
      <UButton v-if="fullscreen" icon="i-tabler-x" label="退出浏览" color="neutral" variant="ghost" size="sm" @click="toggleFullscreen" />
     </div>
     <p v-if="viewerError" class="detail-viewer-error" role="alert">{{ viewerError }}</p>
    </section>
    <aside class="detail-information">
     <header><h1>{{ image.title }}</h1><p v-if="image.description">{{ image.description }}</p></header>
     <div class="detail-main-actions"><UDropdownMenu :items="downloadItems" :content="{ align:'start' }"><UButton icon="i-tabler-download" label="下载" class="detail-download-button" :loading="downloadPending" :disabled="downloadPending" aria-label="选择图片下载格式" /></UDropdownMenu><UTooltip :text="image.favorited ? '取消收藏' : '收藏图片'"><UButton :icon="image.favorited ? 'i-tabler-heart-filled' : 'i-tabler-heart'" color="neutral" variant="ghost" :aria-label="image.favorited ? '取消收藏图片' : '收藏图片'" :aria-pressed="Boolean(image.favorited)" :loading="favoritePending" @click="toggleFavorite" /></UTooltip><UTooltip text="分享图片"><UButton icon="i-tabler-share-3" color="neutral" variant="ghost" aria-label="分享图片" @click="share" /></UTooltip></div>
     <p v-if="downloadError" class="text-xs text-error" role="alert">{{ downloadError }}</p>
     <section v-if="category || assignedGroups.length" class="detail-taxonomy" aria-label="分类、作品与角色">
      <GalleryVisualChoice v-if="category" :node="category" kind="category" compact :aria-label="`分类：${category.name}`" :title="`分类：${category.name}`" @select="void router.push({path:'/images',query:{categories:category.slug}})" />
      <template v-for="item in assignedGroups" :key="item.group.id"><GalleryVisualChoice v-for="node in item.values" :key="node.id" :node="node" compact :aria-label="`${item.group.name}：${node.name}`" :title="`${item.group.name}：${node.name}`" @select="facetLink(item.group.slug,node.slug)" /></template>
     </section>
     <section class="detail-metadata-group"><h2>文件信息</h2><dl><div><dt>分辨率</dt><dd>{{ image.width }} × {{ image.height }}</dd></div><div><dt>比例</dt><dd>{{ ratio }}</dd></div><template v-if="fileFacts"><div><dt>格式</dt><dd>{{ fileFacts.mime.replace('image/','').toUpperCase() }}</dd></div><div><dt>文件大小</dt><dd>{{ (fileFacts.sizeBytes/1024/1024).toFixed(2) }} MiB</dd></div></template></dl><p v-if="factsFailed" class="detail-facts-error">文件信息未能读取 <button type="button" @click="loadFacts">重试</button></p></section>
     <section class="detail-metadata-group"><h2>标签</h2><div v-if="image.tags.length" class="detail-tags mb-2"><NuxtLink v-for="tag in image.tags" :key="tag" :to="{path:'/images',query:{tag}}">{{ tag }}</NuxtLink></div><GalleryTagContribution :image-id="image.id" :tags="image.tags" /></section>
     <section class="detail-metadata-group"><h2>来源</h2><a v-if="image.sourceUrl" :href="image.sourceUrl" target="_blank" rel="noopener noreferrer nofollow" referrerpolicy="no-referrer" class="detail-source">查看原始页面 <UIcon name="i-tabler-arrow-up-right" /></a><p v-else class="detail-facts-error">未填写</p></section>
     <div class="detail-secondary-actions"><UButton v-if="can('gallery.image.update')" :to="`/manage/images?q=${encodeURIComponent(image.title)}`" icon="i-tabler-edit" label="编辑信息" color="neutral" variant="ghost" size="xs" /><UButton icon="i-tabler-flag" label="报告问题" color="neutral" variant="ghost" size="xs" @click="reportKind='report';reportOpen=true" /><UButton icon="i-tabler-link" label="补充来源" color="neutral" variant="ghost" size="xs" @click="reportKind='source_correction';reportOpen=true" /></div>
    </aside>
   </div>
   <section v-if="related.length" class="detail-related"><div class="gallery-section-heading"><h2>相关图片</h2><NuxtLink v-if="relatedBrowse" :to="relatedBrowse.to" class="text-xs text-muted">{{ relatedBrowse.label }} →</NuxtLink></div><GalleryImageGrid :items="related.slice(0,8)" heading-level="h3" view="related" /></section>
   <GalleryCommentSection :image-id="image.id" />
   <GalleryReportDialog v-model:open="reportOpen" :image-id="image.id" :kind="reportKind" />
  </template>
 </article>
</template>
<style scoped>
.detail-toolbar{display:flex;align-items:center;justify-content:space-between;gap:.75rem;margin-bottom:1rem}.detail-toolbar>button{display:flex;align-items:center;gap:.5rem;font-size:.75rem;color:var(--gallery-muted)}
.detail-layout,.detail-loading{display:grid;grid-template-columns:minmax(0,1fr) 19rem;gap:1.3rem;align-items:start}
.detail-viewer-column{min-width:0;overflow:hidden;border:1px solid var(--gallery-line);border-radius:.5rem;background:var(--gallery-viewer-canvas)}
.detail-image-stage{position:relative;height:min(68dvh,55rem);min-height:15rem}
.detail-immersive-action{position:absolute;top:.85rem;right:.85rem;z-index:3;gap:.45rem;padding:.65rem 1rem;border-radius:.6rem;font-size:.85rem;font-weight:600;box-shadow:0 .25rem 1rem #0003}
.detail-image-viewport{position:absolute;inset:0;display:flex;align-items:center;justify-content:center;overflow:hidden;touch-action:pan-y}
.detail-image-viewport img{display:block;flex:none;max-width:none;max-height:none;object-fit:contain;user-select:none;-webkit-user-drag:none}
.detail-image-viewport.can-pan{cursor:grab;touch-action:none}.detail-image-viewport.dragging{cursor:grabbing}
.detail-image-nav{position:absolute;top:50%;transform:translateY(-50%);z-index:2;border:1px solid #ffffff26;border-radius:.5rem;background:#111720b3;color:#fff;backdrop-filter:blur(8px);padding:.7rem .4rem}
.detail-image-nav:hover:not(:disabled){background:#111720ed}.detail-image-nav:disabled{opacity:.2}.detail-image-nav.is-previous{left:.65rem}.detail-image-nav.is-next{right:.65rem}
.detail-image-loading{position:absolute;top:1rem;left:50%;transform:translateX(-50%);display:flex;align-items:center;gap:.4rem;padding:.35rem .65rem;border-radius:2rem;background:#111720cc;color:white;font-size:.7rem}
.detail-viewer-tools{display:flex;align-items:center;justify-content:center;flex-wrap:wrap;gap:.65rem;padding:.5rem .65rem;border-top:1px solid var(--gallery-line);background:var(--gallery-panel)}
.detail-viewer-title{display:none;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:.75rem;color:var(--gallery-muted)}
.detail-zoom-controls,.detail-fit-controls{display:flex;align-items:center;gap:.15rem}.detail-fit-controls{padding-left:.35rem}
.detail-zoom-value{min-width:3.1rem;text-align:center;font-size:.7rem;font-variant-numeric:tabular-nums}.detail-zoom-slider{width:clamp(4rem,9vw,8rem);accent-color:var(--gallery-accent);cursor:pointer}
.detail-viewer-error{padding:.5rem .8rem;font-size:.75rem;color:var(--ui-error)}
.detail-viewer-column:fullscreen{display:flex;flex-direction:column;width:100%;height:100dvh;max-width:none;border:0;border-radius:0;background:var(--gallery-viewer-canvas);color:var(--gallery-ink)}
.detail-viewer-column:fullscreen .detail-image-stage{flex:1;min-height:0;height:auto}.detail-viewer-column:fullscreen .detail-viewer-tools{flex:none;justify-content:flex-end;padding-bottom:max(.5rem,env(safe-area-inset-bottom))}.detail-viewer-column:fullscreen .detail-viewer-title{display:block;flex:1}.detail-viewer-column:fullscreen .detail-zoom-slider{width:8rem}
.detail-information{position:sticky;top:5rem;max-height:calc(100dvh - 6rem);overflow:auto;scrollbar-width:thin;padding-left:1rem;border-left:1px solid var(--gallery-line)}
.detail-information h1{font-size:1.2rem;font-weight:650;line-height:1.5}.detail-information header p{font-size:.73rem;line-height:1.8;color:var(--gallery-muted);margin-top:.6rem;white-space:pre-line;overflow-wrap:anywhere}
.detail-main-actions{display:flex;align-items:center;gap:.5rem;margin:1.2rem 0}.detail-download-button{flex:1;justify-content:center}
.detail-metadata-group{padding:1rem 0;border-bottom:1px solid var(--gallery-line)}.detail-metadata-group h2{font-size:.78rem;font-weight:600;margin-bottom:.65rem}.detail-taxonomy{display:flex;align-items:center;flex-wrap:wrap;gap:.4rem;padding:.2rem 0 1rem;border-bottom:1px solid var(--gallery-line)}.detail-metadata-group dl{display:grid;gap:.5rem;font-size:.7rem}.detail-metadata-group dl>div{display:flex;justify-content:space-between;gap:1rem}.detail-metadata-group dt{color:var(--gallery-muted)}
.detail-tags{display:flex;flex-wrap:wrap;gap:.35rem}.detail-tags>a{border:1px solid var(--gallery-line);background:var(--gallery-panel);border-radius:.3rem;padding:.25rem .5rem;font-size:.65rem}.detail-source{font-size:.72rem;color:var(--gallery-accent)}.detail-facts-error{font-size:.68rem;color:var(--gallery-muted)}.detail-facts-error button{color:var(--gallery-accent)}.detail-secondary-actions{display:flex;flex-wrap:wrap;gap:.15rem;padding-top:.8rem}.detail-related{container:gallery-related / inline-size;margin:2rem 0;min-width:0}
@media(max-width:1150px){.detail-layout,.detail-loading{grid-template-columns:minmax(0,1fr) 16rem}.detail-information{padding-left:.8rem}.detail-viewer-tools{gap:.25rem}.detail-zoom-slider{width:4rem}.detail-fit-controls{padding-left:.25rem}}
@media(max-width:800px){.detail-layout,.detail-loading{grid-template-columns:minmax(0,1fr)}.detail-information{position:static;max-height:none;padding:0;border-left:0}.detail-image-stage{height:65dvh}.detail-related{margin:1.5rem 0}.detail-viewer-column:fullscreen .detail-viewer-title{display:none}}
@media(max-width:480px){.detail-zoom-slider,.detail-viewer-column:fullscreen .detail-zoom-slider{display:none}.detail-image-nav.is-previous{left:.35rem}.detail-image-nav.is-next{right:.35rem}.detail-viewer-tools,.detail-viewer-column:fullscreen .detail-viewer-tools{justify-content:center;gap:.15rem}.detail-fit-controls{padding-left:.15rem}}
</style>
