<script setup lang="ts">
import { PublicCommentThread } from "@yueli/ui/comments";
import type {
  PublicCommentDraft,
  PublicCommentMessages,
  PublicCommentOrder,
  PublicCommentState,
} from "@yueli/ui/comments";
import type { GalleryComment, GalleryCommentPage } from "~/types/gallery";

const props = defineProps<{ imageId: string }>();
const { call } = useGalleryApi();
const { loggedIn, user, login } = useAuth();

const items = ref<GalleryComment[]>([]);
const total = ref(0);
const state = ref<PublicCommentState>("loading");
const order = ref<PublicCommentOrder>("asc");
const viewer = computed(() => ({
  authenticated: loggedIn.value,
  name: user.value?.name || user.value?.email || "",
  avatarUrl: user.value?.avatar || "",
}));
const messages: PublicCommentMessages = {
  count: (count) => `${count} 条评论`,
  replies: (count) => `${count} 条回复`,
  sort: "评论排序",
  loading: "正在加载评论",
  oldest: "最早",
  newest: "最新",
  reply: "回复",
  cancelReply: "取消回复",
  anonymous: "匿名用户",
  empty: "还没有评论，来说第一句吧",
  closed: "这张图片已关闭评论",
  loadError: "评论加载失败",
  retry: "重新加载",
  writeComment: "写下你的评论…",
  writeReply: "写下回复…",
  authorName: "昵称 *",
  authorEmail: "邮箱（选填，不公开）",
  anonymousHint: "匿名评论需要审核",
  login: "登录后直接发布",
  submit: "发表评论",
  submitReply: "回复",
  submitted: "评论已发布",
  pending: "评论已提交，待审核后显示",
  submitError: "评论没有发表，请稍后重试",
  nameRequired: "请填写昵称",
};

async function load() {
  state.value = "loading";
  try {
    const response = await call<GalleryCommentPage>(
      `/images/${encodeURIComponent(props.imageId)}/comments`,
      { query: { page: 1, size: 100, sortOrder: order.value } },
    );
    items.value = response.items;
    total.value = response.total;
    state.value = "ready";
  } catch {
    state.value = "error";
  }
}

async function submitComment(draft: PublicCommentDraft) {
  try {
    const response = await call<{ pending: boolean }>(
      `/images/${encodeURIComponent(props.imageId)}/comments`,
      {
        method: "POST",
        body: {
          content: draft.content,
          parentId: draft.parentId,
          authorName: draft.authorName,
          authorEmail: draft.authorEmail,
        },
      },
    );
    if (!response.pending) await load();
    return { pending: response.pending };
  } catch (error: any) {
    throw new Error(error?.data?.message || messages.submitError);
  }
}

function formatCommentTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "刚刚";
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

watch(() => props.imageId, load);
watch(order, load);
onMounted(load);
</script>

<template>
  <section class="gallery-detail-comments" data-gallery-comments>
    <PublicCommentThread
      v-model:order="order"
      :comments="items"
      :total="total"
      :state="state"
      :viewer="viewer"
      :messages="messages"
      :format-time="formatCommentTime"
      :submit="submitComment"
      :login="login"
      :retry="load"
      allow-anonymous
      input-position="bottom"
    />
  </section>
</template>
