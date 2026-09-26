<script setup lang="ts">
import type {
  AdminNavigationItem,
  AdminSearchGroup,
  AdminShellMessages,
} from "@yueli/ui/admin";

const route = useRoute();
const { brand } = useSiteRuntime();
const { can, isAdministrator } = useGalleryMe();
const sidebarOpen = ref(false);
const immersive = computed(() =>
  /^\/manage\/collections\/[^/]+$/.test(route.path),
);

const currentLabel = computed(() => {
  if (route.path === "/manage") return "控制台";
  if (route.path.startsWith("/manage/images")) return "图片";
  if (route.path.startsWith("/manage/comments")) return "评论";
  if (route.path.startsWith("/manage/submissions")) return "投稿审核";
  if (route.path.startsWith("/manage/collections")) return "专题";
  if (route.path.startsWith("/manage/classification")) return "分类与维度";
  if (route.path.startsWith("/manage/cases")) return "处理单";
  if (route.path.startsWith("/manage/discovery")) return "站点设置";
  if (route.path.startsWith("/manage/assets")) return "资源策略";
  if (route.path.startsWith("/manage/authorization")) return "权限与申请";
  return "控制台";
});

useHead({ bodyAttrs: { class: "gallery-manage-active" } });

const messages: AdminShellMessages = {
  skipToContent: "跳到主要内容",
  search: "搜索图库后台",
  searchPlaceholder: "搜索页面与常用操作",
  currentLocation: "当前位置",
};

function closeSidebar() {
  sidebarOpen.value = false;
}

function active(path: string, exact = false) {
  return exact ? route.path === path : route.path.startsWith(path);
}

const navigation = computed<readonly AdminNavigationItem[]>(() => [
  ...(can("gallery.dashboard.read")
    ? [
        {
          label: "控制台",
          icon: "i-tabler-dashboard",
          to: "/manage",
          active: active("/manage", true),
          onSelect: closeSidebar,
        },
      ]
    : []),
  ...(can("gallery.image.read") ||
  can("gallery.image.update") ||
  can("gallery.image.hide")
    ? [
        {
          label: "图片",
          icon: "i-tabler-photo",
          to: "/manage/images",
          active: active("/manage/images"),
          onSelect: closeSidebar,
        },
      ]
    : []),
  ...(can("gallery.comment.read") ||
  can("gallery.comment.moderate") ||
  can("gallery.comment.delete") ||
  isAdministrator.value
    ? [
        {
          label: "评论",
          icon: "i-tabler-messages",
          to: "/manage/comments",
          active: active("/manage/comments"),
          onSelect: closeSidebar,
        },
      ]
    : []),
  ...(can("gallery.submission.read") || can("gallery.submission.review")
    ? [
        {
          label: "投稿审核",
          icon: "i-tabler-photo-check",
          to: "/manage/submissions",
          active: active("/manage/submissions"),
          onSelect: closeSidebar,
        },
      ]
    : []),
  ...(can("gallery.collection.read") || can("gallery.collection.manage")
    ? [
        {
          label: "专题",
          icon: "i-tabler-folders",
          to: "/manage/collections",
          active: active("/manage/collections"),
          onSelect: closeSidebar,
        },
      ]
    : []),
  ...(can("gallery.classification.read") ||
  can("gallery.classification.proposal_review") ||
  can("gallery.classification.govern")
    ? [
        {
          label: "分类与维度",
          icon: "i-tabler-category",
          to: "/manage/classification",
          active: active("/manage/classification"),
          onSelect: closeSidebar,
        },
      ]
    : []),
  ...(can("gallery.case.read") || can("gallery.case.resolve")
    ? [
        {
          label: "处理单",
          icon: "i-tabler-shield-check",
          to: "/manage/cases",
          active: active("/manage/cases"),
          onSelect: closeSidebar,
        },
      ]
    : []),
  ...(can("gallery.discovery.read") || can("gallery.discovery.manage")
    ? [
        {
          label: "站点设置",
          icon: "i-tabler-settings",
          to: "/manage/discovery",
          active: active("/manage/discovery"),
          onSelect: closeSidebar,
        },
      ]
    : []),
  ...(can("gallery.asset_settings.manage")
    ? [
        {
          label: "资源策略",
          icon: "i-tabler-database-cog",
          to: "/manage/assets",
          active: active("/manage/assets"),
          onSelect: closeSidebar,
        },
      ]
    : []),
  ...(isAdministrator.value
    ? [
        {
          label: "权限与申请",
          icon: "i-tabler-shield-lock",
          to: "/manage/authorization",
          active: active("/manage/authorization"),
          onSelect: closeSidebar,
        },
      ]
    : []),
]);

const searchGroups = computed<readonly AdminSearchGroup[]>(() => {
  const pages = navigation.value.map((item, index) => ({
    id: `gallery-page-${index}`,
    label: item.label,
    icon: item.icon,
    to: item.to,
  }));
  const actions = [
    ...(can("gallery.submission.review")
      ? [
          {
            id: "review-submissions",
            label: "处理待审核投稿",
            icon: "i-tabler-photo-check",
            to: "/manage/submissions?reviewState=pending&outcome=pending",
          },
        ]
      : []),
    ...(can("gallery.classification.proposal_review")
      ? [
          {
            id: "review-tag-proposals",
            label: "处理标签提案",
            icon: "i-tabler-tags",
            to: "/manage/classification",
          },
        ]
      : []),
    ...(isAdministrator.value
      ? [
          {
            id: "authorization",
            label: "处理角色申请",
            icon: "i-tabler-user-check",
            to: "/manage/authorization",
          },
        ]
      : []),
  ];
  return [
    { id: "gallery-pages", label: "管理页面", items: pages },
    ...(actions.length
      ? [{ id: "gallery-actions", label: "常用操作", items: actions }]
      : []),
  ];
});
</script>

<template>
  <YAdminConsoleLayout
    :class="{ 'yueli-admin-branded': !immersive }"
    :navigation="navigation"
    :search-groups="searchGroups"
    :messages="messages"
    storage-key="gallery-manage"
    main-id="manage-main"
    :brand-label="brand"
    brand-icon="i-tabler-photo"
    brand-to="/"
    :context-label="brand"
    :current-label="currentLabel"
    :immersive="immersive"
    back-to-top-label="返回顶部"
    data-gallery-manage-shell
    class="max-sm:[&_a]:min-h-11 max-sm:[&_summary]:min-h-11 max-sm:[&_button]:min-h-11 max-sm:[&_a[aria-label]]:min-w-11 max-sm:[&_button[aria-label]]:min-w-11 max-sm:[&_button[role='checkbox']]:relative max-sm:[&_button[role='checkbox']]:min-h-4 max-sm:[&_button[role='checkbox']]:min-w-4 max-sm:[&_button[role='checkbox']]:after:absolute max-sm:[&_button[role='checkbox']]:after:-inset-3.5 max-sm:[&_button[role='checkbox']]:after:content-[''] max-sm:[&_button[role='switch']]:relative max-sm:[&_button[role='switch']]:min-h-5 max-sm:[&_button[role='switch']]:after:absolute max-sm:[&_button[role='switch']]:after:-inset-3.5 max-sm:[&_button[role='switch']]:after:content-['']"
  >
    <template #topbar-right>
      <ConsumerManageAccountControl
        home-to=""
        show-appearance
        trigger-mode="inline"
      />
    </template>
    <slot />
  </YAdminConsoleLayout>
</template>
