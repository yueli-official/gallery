<script setup lang="ts">
import {AssetDraftNotice} from '@yueli/asset-nuxt/components';
import type { GallerySubmission, GallerySubmissionOptions } from "~/types/gallery";
import type { GallerySubmissionEditor } from "~/types/submission-editor";
import { GALLERY_UPLOAD_ACCEPT, GALLERY_UPLOAD_BATCH_LIMIT, GALLERY_UPLOAD_FORMAT_LABEL, galleryUploadAnimationError, galleryUploadFileError } from "~/utils/galleryUpload";
import { applyGallerySubmissionDefaults } from "~/utils/gallerySubmissionBatch";
import { gallerySubmissionFailure } from "~/utils/gallerySubmissionFailure";
import { galleryDraftKey } from "~/utils/clientId";
type QueueStatus = "ready" | "uploading" | "submitting" | "completed" | "duplicate" | "failed";
interface QueueItem extends GallerySubmissionEditor { id:string; file:File; previewUrl:string; customized:boolean; status:QueueStatus;progress:number;error:string;assetId:string;submission?:GallerySubmission }
const { loggedIn, login } = useAuth();
const { call } = useGalleryApi();
const { upload,draft } = useGalleryAssetUpload('gallery-submission-queue');
const { data: options, error: optionsError, refresh: refreshOptions } = await useFetch<GallerySubmissionOptions>("/api/gallery/submission-options");
const queue = ref<QueueItem[]>([]);
const defaults = ref<GallerySubmissionEditor>({ title:"",description:"",sourceUrl:"",primaryCategoryId:"",sceneValueIds:[],tags:"",facetValues:{} });
const editingId = ref("");
const batch = ref(false);
const running = ref(false);
const errors = ref<string[]>([]);
const fileInput = ref<HTMLInputElement>();
const focused = computed(() => queue.value.find(item => item.id === editingId.value));
const focusedModel = computed<GallerySubmissionEditor>({ get: () => focused.value!, set: value => {if (focused.value) Object.assign(focused.value,value,{customized:true});} });
const counts = computed(() => ({ ready:queue.value.filter(item=>item.status==='ready').length,failed:queue.value.filter(item=>item.status==='failed').length,completed:queue.value.filter(item=>item.status==='completed').length }));
const statusLabels:Record<QueueStatus,string>={ready:"待提交",uploading:"上传中",submitting:"提交中",completed:"已提交",duplicate:"已存在",failed:"失败"};
function terminal(item:QueueItem){return item.status==='completed'||item.status==='duplicate';}
function valid(item:QueueItem){return Boolean(item.title.trim() && item.primaryCategoryId) && Boolean(options.value) && options.value!.facets.every(group=>{const n=(group.slug==='scene'?item.sceneValueIds:item.facetValues[group.id]||[]).length;return n>=group.minValues && n<=group.maxValues;});}
const validMetadata = computed(() => queue.value.filter(item=>!terminal(item)).every(valid));
async function addFiles(files:File[]) {
 if(running.value)return;errors.value=[];
 const capacity=GALLERY_UPLOAD_BATCH_LIMIT-queue.value.length;
 if(files.length>capacity)errors.value.push(`每批最多 ${GALLERY_UPLOAD_BATCH_LIMIT} 张，本次只加入前 ${Math.max(0,capacity)} 张。`);
 for(const file of files.slice(0,Math.max(0,capacity))) {
  if(queue.value.some(item=>item.file.name===file.name&&item.file.size===file.size&&item.file.lastModified===file.lastModified))continue;
  const failure=galleryUploadFileError(file)||await galleryUploadAnimationError(file);
  if(failure){errors.value.push(`${file.name}：${failure}`);continue;}
  const item:QueueItem={...JSON.parse(JSON.stringify(defaults.value)),id:galleryDraftKey(),title:defaults.value.title.trim()||file.name.replace(/\.[^.]+$/,""),file,previewUrl:URL.createObjectURL(file),customized:false,status:"ready",progress:0,error:"",assetId:""};
  queue.value.push(item);if(!editingId.value)editingId.value=item.id;
 }
}
async function chooseFiles(event:Event){const input=event.target as HTMLInputElement;await addFiles([...(input.files||[])]);input.value="";}
function drop(event:DragEvent){void addFiles([...(event.dataTransfer?.files||[])]);}
function removeItem(id:string){const index=queue.value.findIndex(item=>item.id===id);if(index<0||running.value)return;URL.revokeObjectURL(queue.value[index]!.previewUrl);queue.value.splice(index,1);if(editingId.value===id)editingId.value=queue.value[0]?.id||"";}
function applyDefaults(){for(const item of queue.value){if(terminal(item))continue;applyGallerySubmissionDefaults(item,defaults.value);item.facetValues=JSON.parse(JSON.stringify(defaults.value.facetValues));}batch.value=false;}
function body(item:QueueItem){const title=item.title.trim();return {assetId:item.assetId,title,description:item.description.trim(),sourceUrl:item.sourceUrl.trim(),altText:title,categoryIds:[item.primaryCategoryId],primaryCategoryId:item.primaryCategoryId,tags:item.tags.split(/[,，]/).map(value=>value.trim()).filter(Boolean),facets:(options.value?.facets||[]).map(group=>({facetId:group.id,valueIds:group.slug==='scene'?item.sceneValueIds:item.facetValues[group.id]||[]})).filter(group=>group.valueIds.length)};}
async function run(status:'ready'|'failed'){
 if(running.value||!validMetadata.value)return;running.value=true;
 try {for(const item of queue.value.filter(item=>item.status===status)){
  item.error="";
  try {
   if(!item.assetId){item.status="uploading";item.assetId=(await upload(item.file,value=>item.progress=value)).id;}
   item.status="submitting";item.submission=(await call<{submission:GallerySubmission}>("/submissions",{method:"POST",body:body(item)})).submission;item.status="completed";item.progress=100;
  }catch(reason){const failure=gallerySubmissionFailure(reason);item.status=failure.kind==='already-submitted'?"duplicate":"failed";item.error=failure.message;}
 }}finally{running.value=false;}
}
onBeforeUnmount(()=>queue.value.forEach(item=>URL.revokeObjectURL(item.previewUrl)));
useSeoMeta({title:"投稿图片 · 月离图库",robots:"noindex,follow"});
</script>
<template>
 <div class="submission-page">
  <header class="submission-heading"><div><h1>投稿图片</h1><p>选择图片，整理信息，再提交到图库。</p></div><span>{{ queue.length ? '② 整理信息' : '① 选择图片' }}</span></header>
  <UAlert v-if="optionsError" color="error" title="分类选项加载失败" class="mb-4"><template #actions><UButton label="重试" @click="refreshOptions()" /></template></UAlert>
  <p class="submission-help">上传后保存为高质量 WebP，最长边 4096 像素，不放大小图，不保留原文件。</p>
  <AssetDraftNotice :expires-at="draft.expiresAt.value" :reset-required="draft.resetRequired.value" :disabled="running" @restart="draft.restart" />
  <div class="submission-workspace">
   <section class="submission-queue">
    <div class="submission-drop" @dragover.prevent @drop.prevent="drop"><UIcon name="i-tabler-photo-plus" /><div><p>将图片拖到这里，或选择文件</p><small>{{ GALLERY_UPLOAD_FORMAT_LABEL }} · 每批最多 {{ GALLERY_UPLOAD_BATCH_LIMIT }} 张 · 单张不超过 20 MiB</small></div><UButton label="选择图片" icon="i-tabler-folder-open" variant="outline" :disabled="running||queue.length>=GALLERY_UPLOAD_BATCH_LIMIT" @click="fileInput?.click()" /><input ref="fileInput" type="file" multiple :accept="GALLERY_UPLOAD_ACCEPT" :disabled="running" class="sr-only" aria-label="选择投稿图片" @change="chooseFiles" /></div>
    <UAlert v-if="errors.length" color="warning" title="部分文件没有加入" class="mt-4"><template #description><ul><li v-for="message in errors" :key="message">{{ message }}</li></ul></template></UAlert>
    <div class="submission-queue-heading"><h2>文件队列 <span>· {{ queue.length }} 张</span></h2><UButton label="批量设置" icon="i-tabler-adjustments-horizontal" color="neutral" variant="outline" size="sm" :disabled="running||!queue.length" @click="void (batch = true)" /></div>
    <div v-if="!queue.length" class="submission-empty"><UIcon name="i-tabler-photo" /><p>把喜欢的图片收集到一起。</p></div>
    <div v-else class="submission-cards"><article v-for="item in queue" :key="item.id" :class="{editing:editingId===item.id&&!batch}" class="submission-card"><button type="button" class="submission-card-select" :aria-label="`编辑图片：${item.title}`" @click="editingId=item.id;batch=false"><img :src="item.previewUrl" :alt="item.title" /><div class="submission-card-copy"><h3>{{ item.title }}</h3><p>{{ statusLabels[item.status] }}<span v-if="item.status==='uploading'"> · {{ item.progress }}%</span></p></div></button><UButton icon="i-tabler-x" color="neutral" variant="ghost" size="xs" :aria-label="`移除文件：${item.file.name}`" :disabled="running" class="submission-remove" @click="removeItem(item.id)" /><p v-if="item.error" class="submission-error" role="alert">{{ item.error }}</p></article></div>
   </section>
   <aside class="submission-inspector"><div class="submission-inspector-heading"><h2>{{ batch ? '批量设置' : '图片信息' }}</h2><UButton v-if="batch" label="返回逐张编辑" color="neutral" variant="ghost" size="xs" @click="void (batch=false)" /><span v-else-if="focused">{{ queue.indexOf(focused)+1 }} / {{ queue.length }}</span></div><p v-if="batch" class="submission-help">点击“应用到待提交图片”后覆盖对应字段，已完成项目保持原样。</p><GallerySubmissionFields v-if="batch" v-model="defaults" :options="options" :disabled="running" batch /><GallerySubmissionFields v-else-if="focused" v-model="focusedModel" :options="options" :disabled="running||terminal(focused)" /><p v-else class="submission-help">选择图片后，在这里整理分类和维度。</p><UButton v-if="batch" label="应用到待提交图片" variant="outline" block class="mt-4" :disabled="running" @click="applyDefaults" /></aside>
  </div>
  <footer class="submission-footer"><div><p v-if="counts.completed">{{ counts.completed }} 张已提交<span v-if="counts.failed">，{{ counts.failed }} 张可重试</span></p><p v-else>提交后可在“我的投稿”查看处理与审核状态。</p><small v-if="queue.length&&!validMetadata">请为待提交图片填写标题、分类和必填维度。</small><small v-else-if="!loggedIn">当前使用临时投稿身份，<button type="button" @click="void login()">登录后管理更方便</button>。</small></div><div class="submission-submit-actions"><UButton v-if="counts.failed" label="重试失败项目" color="neutral" variant="outline" :loading="running" :disabled="!validMetadata" @click="run('failed')" /><UButton :label="running ? '正在提交' : `提交 ${counts.ready} 张投稿`" icon="i-tabler-upload" :loading="running" :disabled="!counts.ready||!validMetadata||Boolean(optionsError)" @click="run('ready')" /></div></footer>
 </div>
</template>
<style scoped>
.submission-page{padding-bottom:5rem}.submission-heading{display:flex;align-items:center;justify-content:space-between;gap:1rem;margin-bottom:1.3rem}.submission-heading h1{font-size:1.45rem;font-weight:650}.submission-heading p{color:var(--gallery-muted);font-size:.75rem;margin-top:.4rem}.submission-heading>span{font-size:.75rem;color:var(--gallery-accent)}.submission-workspace{display:grid;grid-template-columns:minmax(0,1fr) 23rem;gap:1.2rem;align-items:start}.submission-queue{min-width:0}.submission-drop{display:flex;align-items:center;justify-content:center;gap:1rem;padding:1.7rem 1rem;border:1px dashed var(--gallery-line);border-radius:.5rem;background:var(--gallery-panel)}.submission-drop>.iconify{font-size:2rem;color:var(--gallery-muted);flex-shrink:0}.submission-drop p{font-size:.8rem}.submission-drop small{display:block;font-size:.61rem;color:var(--gallery-muted);margin-top:.35rem}.submission-queue-heading{display:flex;justify-content:space-between;align-items:center;margin:1.3rem 0 1rem}.submission-queue-heading h2{font-size:.9rem;font-weight:600}.submission-queue-heading h2 span{color:var(--gallery-muted);font-weight:400}.submission-cards{display:grid;grid-template-columns:repeat(auto-fill,minmax(13.5rem,1fr));gap:.6rem;align-items:start}.submission-card{position:relative;overflow:hidden;border:1px solid var(--gallery-line);border-radius:.45rem;background:var(--gallery-panel)}.submission-card.editing{border-color:var(--gallery-accent)}.submission-card-select{display:block;text-align:left;width:100%}.submission-card-select img{display:block;width:100%;aspect-ratio:3/4;object-fit:contain;background:var(--gallery-viewer-canvas)}.submission-card-copy{padding:.65rem}.submission-card-copy h3{font-size:.72rem;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.submission-card-copy p{font-size:.65rem;color:var(--gallery-muted);margin-top:.3rem}.submission-remove{position:absolute;top:.35rem;right:.35rem;background:var(--gallery-panel-strong)}.submission-error{font-size:.7rem;padding:0 .65rem .65rem;color:var(--ui-error)}.submission-empty{display:grid;place-content:center;text-align:center;gap:1rem;min-height:19rem;color:var(--gallery-muted);font-size:.8rem}.submission-empty .iconify{font-size:2.5rem;margin:auto}.submission-inspector{position:sticky;top:5rem;max-height:calc(100dvh - 11rem);overflow:auto;scrollbar-width:thin;padding:1rem;background:var(--gallery-panel);border:1px solid var(--gallery-line);border-radius:.5rem}.submission-inspector-heading{display:flex;align-items:center;justify-content:space-between;margin-bottom:1rem}.submission-inspector-heading h2{font-size:.95rem;font-weight:600}.submission-inspector-heading>span,.submission-help{font-size:.7rem;color:var(--gallery-muted)}.submission-help{line-height:1.7;margin-bottom:1rem}.submission-footer{position:fixed;bottom:0;left:12.5rem;right:0;z-index:28;display:flex;justify-content:space-between;align-items:center;gap:1rem;padding:1rem max(1.25rem,calc((100% - 12.5rem - var(--gallery-content-max))/2 + 1.25rem));background:var(--gallery-canvas);border-top:1px solid var(--gallery-line)}.submission-footer p{font-size:.73rem;color:var(--gallery-muted)}.submission-footer small{display:block;font-size:.65rem;margin-top:.35rem;color:var(--gallery-accent)}.submission-footer button{cursor:pointer}.submission-submit-actions{display:flex;gap:.5rem;flex-shrink:0}
@media(max-width:1250px){.submission-workspace{grid-template-columns:minmax(0,1fr) 20rem}.submission-cards{grid-template-columns:repeat(2,minmax(0,1fr))}.submission-drop{flex-wrap:wrap}}@media(max-width:900px){.submission-footer{left:0;padding:.75rem}.submission-footer p{font-size:.65rem}}@media(max-width:700px){.submission-workspace{grid-template-columns:minmax(0,1fr)}.submission-inspector{position:static;max-height:none}.submission-heading>span{display:none}.submission-footer{align-items:flex-start}.submission-footer>div:first-child{max-width:48%}.submission-footer p,.submission-footer small{font-size:.58rem}.submission-submit-actions{flex-wrap:wrap;justify-content:flex-end}}
</style>
