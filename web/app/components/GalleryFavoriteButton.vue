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
    notifier.add({ title: "已加入我的收藏", color: "success" });
  } catch (reason: any) {
    notifier.add({
      title: "收藏没有保存",
      description: reason?.data?.message || reason?.message || "请稍后重试",
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
    class="gallery-favorite-button absolute right-[0.55rem] top-[0.55rem] z-2 grid size-[2.35rem] -translate-y-1 place-items-center rounded-[0.7rem] border border-white/70 bg-white/90 text-neutral-900 opacity-0 transition-[opacity,transform] group-hover:translate-y-0 group-hover:opacity-100 group-focus-within:translate-y-0 group-focus-within:opacity-100 [@media(hover:none)]:translate-y-0 [@media(hover:none)]:opacity-100"
    :class="{
      'border-primary bg-primary text-white': saved,
      'cursor-wait opacity-65': pending,
    }"
    :disabled="pending"
    :aria-label="saved ? `已收藏：${title}` : `收藏：${title}`"
    @click="save"
  >
    <svg
      viewBox="0 0 24 24"
      width="18"
      height="18"
      aria-hidden="true"
    >
      <path
        d="M6 4.75A1.75 1.75 0 0 1 7.75 3h8.5A1.75 1.75 0 0 1 18 4.75V21l-6-3.6L6 21V4.75Z"
        :fill="saved ? 'currentColor' : 'none'"
        stroke="currentColor"
        stroke-width="1.8"
        stroke-linejoin="round"
      />
    </svg>
  </button>
</template>
