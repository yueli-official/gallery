<script setup lang="ts">
import { createGalleryNotifier } from "~/utils/feedback";

const props = defineProps<{
  imageId: string;
  title: string;
}>();
const { loggedIn, login } = useAuth();
const { call } = useApi();
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
      `/api/v1/gallery/me/favorites/${encodeURIComponent(props.imageId)}`,
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
    class="gallery-favorite-button"
    :class="{ 'is-saved': saved, 'is-pending': pending }"
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
