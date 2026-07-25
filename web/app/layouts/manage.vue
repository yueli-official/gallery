<script setup lang="ts">
import type { DropdownMenuItem } from "@nuxt/ui";
import type {
  AdminNavigationItem,
  AdminSearchGroup,
  AdminShellMessages,
} from "@yueli/ui/admin";

const route = useRoute();
const { brand } = useSiteRuntime();
const { can, isAdministrator } = useGalleryMe();
const sidebarOpen = ref(false);

const messages: AdminShellMessages = {
  skipToContent: "跳到主要内容",
  search: "搜索图库后台",
  searchPlaceholder: "搜索页面与常用操作",
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
          label: "今日",
          icon: "i-tabler-sun-high",
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
          label: "专题策展",
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
  ...(can("gallery.discovery.read")
    ? [
        {
          label: "发现策略",
          icon: "i-tabler-sparkles",
          to: "/manage/discovery",
          active: active("/manage/discovery"),
          onSelect: closeSidebar,
        },
      ]
    : []),
  ...(can("gallery.asset_settings.manage")
    ? [
        {
          label: "资源设置",
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

const workspaceMenuItems = computed<DropdownMenuItem[][]>(() => [
  [{ type: "label", label: brand.value }],
  [
    {
      label: "图库运营",
      icon: "i-tabler-photo",
      type: "checkbox",
      checked: true,
      onSelect: (event: Event) => event.preventDefault(),
    },
  ],
  [
    {
      label: "打开图库",
      icon: "i-tabler-external-link",
      to: "/",
      onSelect: closeSidebar,
    },
  ],
]);
</script>

<template>
  <ClientOnly>
    <YAdminShell
      v-model:open="sidebarOpen"
      :navigation="navigation"
      :search-groups="searchGroups"
      :messages="messages"
      storage-key="gallery-manage"
      main-id="manage-main"
      :default-size="16"
      :min-size="14"
      :max-size="20"
    >
      <template #brand="{ collapsed }">
        <UDropdownMenu
          :items="workspaceMenuItems"
          :content="{ align: 'center', collisionPadding: 12 }"
          :ui="{
            content: collapsed
              ? 'w-56'
              : 'w-(--reka-dropdown-menu-trigger-width)',
          }"
        >
          <UButton
            type="button"
            color="neutral"
            variant="ghost"
            :block="!collapsed"
            :square="collapsed"
            :aria-label="`打开${brand}站点菜单`"
            :class="[
              'min-h-11 gap-2 px-1.5 data-[state=open]:bg-elevated',
              !collapsed && 'w-full justify-start',
              collapsed && 'aspect-square justify-center px-0',
            ]"
          >
            <span
              class="grid size-7 shrink-0 place-items-center rounded-md bg-primary/10 text-primary"
            >
              <UIcon name="i-tabler-photo" class="size-4" />
            </span>
            <span
              v-if="!collapsed"
              class="min-w-0 truncate text-sm font-semibold text-highlighted"
            >
              {{ brand }}
            </span>
            <UIcon
              v-if="!collapsed"
              name="i-tabler-chevrons-up-down"
              class="ms-auto size-3.5 text-dimmed"
            />
          </UButton>
        </UDropdownMenu>
      </template>

      <template #sidebar-footer="{ collapsed }">
        <ConsumerManageAccountControl
          home-to=""
          show-appearance
          :trigger-mode="collapsed ? 'collapsed' : 'sidebar'"
        />
      </template>

      <main
        id="manage-main"
        tabindex="-1"
        class="min-w-0 flex-1 overflow-y-auto p-4 outline-none sm:p-6"
      >
        <slot />
      </main>
      <YBackToTop
        target-id="manage-main"
        scroll-container-id="manage-main"
        avoid-selector="[data-manage-dock], [data-back-to-top-avoid]"
        label="返回顶部"
      />
    </YAdminShell>

    <template #fallback>
      <div
        class="fixed inset-0 grid place-items-center bg-default text-sm text-muted"
        role="status"
      >
        正在打开控制台
      </div>
    </template>
  </ClientOnly>
</template>
