<script setup lang="ts">
import type { GalleryClassificationTagProposal } from "~/types/gallery";
import { getApiFailure } from "@yueli/http-runtime";

const props = defineProps<{ imageId: string; tags: string[] }>();
const { loggedIn } = useAuth();
const { call } = useGalleryApi();
const open = ref(false);
const menuOpen = ref(false);
const selected = ref<string[]>([]);
const search = ref("");
const suggestions = ref<string[]>([]);
const loading = ref(false);
const pending = ref(false);
const error = ref("");
const suggestionError = ref("");
const result = ref("");
let requestVersion = 0;
let timer: ReturnType<typeof setTimeout> | undefined;

const options = computed(() => suggestions.value
  .filter((tag) => !props.tags.includes(tag) && !selected.value.includes(tag)));

async function loadSuggestions() {
  const version = ++requestVersion;
  loading.value = true;
  suggestionError.value = "";
  try {
    const response = await call<{ items: { name: string }[] }>("/tag-suggestions", { query: { q: search.value } });
    if (version === requestVersion) suggestions.value = response.items.map((item) => item.name);
  } catch (cause) {
    if (version === requestVersion) suggestionError.value = galleryFailureMessage(cause, "标签建议未能加载，请重试");
  } finally {
    if (version === requestVersion) loading.value = false;
  }
}

watch(search, () => {
  clearTimeout(timer);
  // Invalidate an older search immediately, before the debounce expires.
  requestVersion++;
  suggestions.value = [];
  loading.value = false;
  if (open.value && search.value.length <= 64) timer = setTimeout(loadSuggestions, 250);
});
watch(open, (value) => {
  clearTimeout(timer);
  requestVersion++;
  menuOpen.value = false;
  loading.value = false;
  if (value) void loadSuggestions();
});
watch(() => props.imageId, () => {
  open.value = false;
  selected.value = [];
  search.value = "";
  result.value = "";
  error.value = "";
});
onBeforeUnmount(() => { clearTimeout(timer); requestVersion++; });

function start() {
  selected.value = [];
  search.value = "";
  error.value = "";
  open.value = true;
}

function close() { open.value = false; }

function selectionChanged() {
  menuOpen.value = false;
  search.value = "";
  error.value = "";
}

function add(value: string) {
  const word = value.trim().replace(/\s+/gu, " ");
  if (!word || [...word].length > 32) { error.value = "每个标签最多 32 个字"; return; }
  if (props.tags.includes(word)) { error.value = "这张图片已经有这个标签"; return; }
  if (selected.value.includes(word)) { search.value = ""; menuOpen.value = false; return; }
  if (selected.value.length >= 10) { error.value = "一次最多补充 10 个标签"; return; }
  selected.value.push(word);
  menuOpen.value = false;
  search.value = "";
  error.value = "";
}

async function submit() {
  if (pending.value) return;
  // Accept a final typed word even when the user submits without choosing the create row.
  if (search.value.trim()) add(search.value);
  if (error.value || !selected.value.length) return;
  if (selected.value.length > 10) { error.value = "一次最多补充 10 个标签"; return; }
  pending.value = true;
  error.value = "";
  const imageId = props.imageId;
  try {
    const response = await call<{ items: GalleryClassificationTagProposal[] }>(`/images/${encodeURIComponent(imageId)}/tag-proposals`, {
      method: "POST", body: { tags: selected.value },
    });
    if (props.imageId !== imageId) return;
    const waiting = response.items.filter((item) => item.status === "pending").length;
    result.value = waiting ? `${waiting} 个标签待审核，通过后会显示在这里` : "这些标签已处理，无需重复提交";
    open.value = false;
  } catch (cause) {
    if (props.imageId === imageId) error.value = !loggedIn.value && getApiFailure(cause)?.code === "gallery.rate_limited"
      ? "这次补充超出每分钟 10 个标签的额度，请减少数量或稍后再试"
      : galleryFailureMessage(cause, "标签未能提交，请稍后重试");
  } finally { pending.value = false; }
}
</script>

<template>
  <div>
    <UButton icon="i-tabler-plus" label="补充标签" color="neutral" variant="ghost" size="xs" @click="start" />
    <p v-if="result" class="mt-1 text-xs text-muted" role="status">{{ result }}</p>
    <UModal v-model:open="open" title="补充标签" description="选择已有标签，也可以输入新词。审核通过后会显示在图片上。" :dismissible="!pending" :ui="{ content: 'divide-y-0 ring-0 shadow-xl', header: 'pb-0', body: 'pt-4' }">
      <template #body>
        <form class="space-y-4" @submit.prevent="submit">
          <UFormField label="图片标签" :hint="loggedIn ? '一次最多 10 个' : '每分钟最多 10 个'" :error="error || undefined">
            <UInputMenu v-model="selected" v-model:search-term="search" v-model:open="menuOpen" :items="options" :content="{ side: 'top' }" :ui="{ content: 'ring-0 shadow-lg', trailing: 'hidden', base: 'pe-3' }" variant="soft" multiple create-item="always" ignore-filter :loading="loading" :disabled="pending" class="w-full" placeholder="例如白发、白丝、双马尾" aria-label="补充图片标签" @create="add" @update:model-value="selectionChanged">
              <template #create-item-label="{ item }">补充新词「{{ item }}」</template>
              <template #empty>没有匹配的标签，可以补充新词</template>
            </UInputMenu>
          </UFormField>
          <p v-if="suggestionError" class="text-xs text-error">{{ suggestionError }} <button type="button" class="underline" @click="loadSuggestions">重试</button></p>
          <div v-if="!search && suggestions.length" class="flex flex-wrap gap-1.5" aria-label="常用标签">
            <UButton v-for="tag in suggestions.filter((word) => !selected.includes(word) && !tags.includes(word)).slice(0, 12)" :key="tag" :label="tag" color="neutral" variant="soft" size="xs" :disabled="pending || selected.length >= 10" @click="add(tag)" />
          </div>
          <div class="flex justify-end gap-2">
            <UButton label="取消" color="neutral" variant="ghost" :disabled="pending" @click="close" />
            <UButton type="submit" label="提交审核" :loading="pending" :disabled="!selected.length && !search.trim()" />
          </div>
        </form>
      </template>
    </UModal>
  </div>
</template>
