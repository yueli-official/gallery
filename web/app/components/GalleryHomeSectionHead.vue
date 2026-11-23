<script setup lang="ts">
withDefaults(
  defineProps<{
    index: string;
    title: string;
    to?: string;
    actionLabel?: string;
  }>(),
  {
    to: undefined,
    actionLabel: "查看全部",
  },
);
</script>

<template>
  <header class="home-section-head">
    <div class="home-section-heading">
      <span class="home-section-index font-numeric" aria-hidden="true">{{ index }}</span>
      <h2 class="home-section-title">{{ title }}</h2>
    </div>
    <div class="home-section-action">
      <slot name="action">
        <NuxtLink v-if="to" :to="to" class="home-text-link">
          {{ actionLabel }}
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path
              d="M4 12h15m-5-5 5 5-5 5"
              fill="none"
              stroke="currentColor"
              stroke-width="1.7"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </NuxtLink>
      </slot>
    </div>
  </header>
</template>

<style scoped>
.home-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.15rem;
}

.home-section-heading {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  min-width: 0;
}

.home-section-index {
  display: inline-grid;
  place-items: center;
  height: 1.55rem;
  min-width: 1.95rem;
  padding: 0 0.45rem;
  border-radius: 999px;
  background: var(--gallery-accent);
  font-size: 0.6rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  color: var(--gallery-on-accent);
}

.home-section-title {
  font-size: clamp(1.35rem, 2vw, 1.7rem);
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -0.045em;
  color: var(--gallery-ink);
  text-wrap: balance;
}

.home-section-action {
  flex: none;
}

.home-text-link {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  height: 2rem;
  padding: 0 0.8rem;
  border-radius: 999px;
  background: var(--gallery-panel);
  font-size: 0.7rem;
  font-weight: 500;
  color: var(--gallery-muted);
  transition:
    background 160ms ease,
    color 160ms ease;
}

.home-text-link:hover {
  background: var(--gallery-accent);
  color: var(--gallery-on-accent);
}

.home-text-link svg {
  width: 0.8rem;
  height: 0.8rem;
}

.home-text-link:focus-visible {
  outline: 2px solid var(--gallery-accent);
  outline-offset: 2px;
}

@media (max-width: 640px) {
  .home-section-title {
    font-size: 1.25rem;
  }

  .home-section-index {
    height: 1.4rem;
    min-width: 1.75rem;
  }

  .home-text-link {
    height: 1.85rem;
    padding: 0 0.65rem;
    font-size: 0.66rem;
  }
}
</style>
