<script setup lang="ts">
const route = useRoute();
const items = [
  { label: "作品审核", icon: "i-tabler-shield-check", to: "/manage/reviews" },
  { label: "创作者准入", icon: "i-tabler-user-check", to: "/manage/creators" },
];
</script>

<template>
  <div class="flex h-full flex-col bg-elevated/30">
    <NuxtLink to="/manage/reviews" class="flex h-16 items-center gap-2 border-b border-default px-5 font-display font-semibold text-highlighted">
      <span class="grid size-8 place-items-center rounded-lg bg-primary/10 text-primary"><UIcon name="i-tabler-photo-shield" class="size-5" /></span>
      图库管理
    </NuxtLink>
    <nav aria-label="图库管理" class="flex-1 space-y-1 p-3">
      <NuxtLink
        v-for="item in items"
        :key="item.to"
        :to="item.to"
        class="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition"
        :class="route.path.startsWith(item.to) ? 'bg-primary/10 text-primary' : 'text-muted hover:bg-elevated hover:text-default'"
      >
        <UIcon :name="item.icon" class="size-5" />{{ item.label }}
      </NuxtLink>
    </nav>
  </div>
</template>
