<script setup lang="ts">
defineProps<{
  brandName?: string;
  brandTagline?: string;
}>();

const { isAdmin } = useAuth();
</script>

<template>
  <header
    class="platform-topbar sticky top-0 z-30 border-b border-default bg-default/90 backdrop-blur-xl"
  >
    <div
      class="mx-auto flex h-16 w-full max-w-screen-2xl items-center gap-4 px-4 sm:px-6 lg:px-8"
    >
      <NuxtLink
        to="/"
        class="flex min-w-0 items-center gap-3 rounded-lg focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-primary"
      >
        <span
          class="grid size-9 shrink-0 place-items-center rounded-xl bg-primary text-inverted"
        >
          <UIcon name="i-tabler-photo" class="size-5" />
        </span>
        <span class="min-w-0">
          <span class="block truncate font-display font-semibold text-highlighted">
            {{ brandName }}
          </span>
          <span class="hidden truncate text-xs text-muted sm:block">
            {{ brandTagline }}
          </span>
        </span>
      </NuxtLink>

      <nav class="ml-4 hidden items-center gap-1 md:flex" aria-label="主要导航">
        <UButton to="/" color="neutral" variant="ghost" label="发现" />
        <UButton to="/studio" color="neutral" variant="ghost" label="创作中心" />
      </nav>

      <div class="ml-auto flex items-center gap-1">
        <UButton to="/studio" color="neutral" variant="ghost" icon="i-tabler-brush" class="md:hidden" aria-label="创作中心" />
        <UTooltip text="切换颜色模式">
        <UColorModeButton
            color="neutral"
            variant="ghost"
            aria-label="切换颜色模式"
        />
        </UTooltip>
        <UButton
          v-if="isAdmin"
          to="/manage/reviews"
          color="neutral"
          variant="ghost"
          icon="i-tabler-shield-check"
          aria-label="内容审核"
        />
        <ConsumerAccountControl />
      </div>
    </div>
  </header>
</template>
