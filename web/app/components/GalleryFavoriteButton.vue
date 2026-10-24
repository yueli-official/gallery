<script setup lang="ts">
import { createGalleryNotifier } from "~/utils/feedback";

const props = defineProps<{
  imageId: string;
  title: string;
}>();
const { loggedIn, login } = useAuth();
const { call } = useGalleryApi();
const notifier = createGalleryNotifier(useToast());
const pending = ref(false);
const saved = ref(false);

async function save(): Promise<void> {
  if (pending.value || saved.value) return;
  if (!loggedIn.value) {
    await login();
    return;
  }
  pending.value = true;
  try {
    await call(
      `/me/favorites/${encodeURIComponent(props.imageId)}`,
      { method: "PUT", body: { version: 0 } },
    );
    saved.value = true;
  } catch (reason: any) {
    notifier.add({
      title: "收藏没有保存",
      description: galleryFailureMessage(reason, "请稍后重试"),
      color: "error",
    });
  } finally {
    pending.value = false;
  }
}
</script>

<template>
  <button
    type="button"
    class="gallery-favorite-button"
    :class="{
      'is-saved': saved,
      'cursor-wait opacity-65': pending,
    }"
    :disabled="pending"
    :aria-label="saved ? `已收藏：${title}` : `收藏：${title}`"
    @click="save"
  >
    <UIcon :name="saved ? 'i-tabler-heart-filled' : 'i-tabler-heart'" aria-hidden="true" />
  </button>
</template>

<style scoped>
.gallery-favorite-button{position:absolute;right:.45rem;top:.45rem;z-index:2;display:grid;place-items:center;width:1.9rem;height:1.9rem;border:1px solid #ffffff40;background:#10172080;color:#f7f4ed;border-radius:50%;cursor:pointer}.gallery-favorite-button.is-saved{background:var(--gallery-accent);color:var(--gallery-on-accent)}.gallery-favorite-button:disabled{opacity:.65}.gallery-favorite-button:hover:not(.is-saved){border-color:var(--gallery-accent);color:var(--gallery-accent)}
@media(pointer:coarse){.gallery-favorite-button{width:2.75rem;height:2.75rem}}
</style>
