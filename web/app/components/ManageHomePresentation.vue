<script setup lang="ts">
import type { GalleryHomePresentation } from "~/types/gallery";

const { call } = useGalleryApi();
const form = reactive<GalleryHomePresentation>({ spotlight: [], categories: [] });
const loading = ref(true);
const saving = ref(false);
const saved = ref(false);
const error = ref("");

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const response = await call<{ presentation: GalleryHomePresentation }>(
      "/admin/home-presentation",
    );
    form.spotlight = response.presentation.spotlight;
    form.categories = response.presentation.categories;
  } catch (reason: any) {
    error.value = reason?.message || "无法读取首页角色数据";
  } finally {
    loading.value = false;
  }
}

async function save() {
  if (saving.value) return;
  saving.value = true;
  saved.value = false;
  error.value = "";
  try {
    const response = await call<{ presentation: GalleryHomePresentation }>(
      "/admin/home-presentation",
      {
        method: "PATCH",
        body: {
          spotlight: form.spotlight,
          categories: form.categories,
        },
      },
    );
    form.spotlight = response.presentation.spotlight;
    form.categories = response.presentation.categories;
    saved.value = true;
  } catch (reason: any) {
    error.value = reason?.message || "保存失败";
  } finally {
    saving.value = false;
  }
}

onMounted(load);
</script>

<template>
  <section class="home-presentation">
    <header class="home-presentation-head">
      <div>
        <h2>角色 Spotlight 与分类封面</h2>
        <p>这一组数据独立保存；首页首屏角色带与三张分类封面都来自这里。</p>
      </div>
      <div class="home-presentation-save">
        <span v-if="saved" class="home-presentation-state is-saved">已保存</span>
        <span v-else-if="error" class="home-presentation-state is-error">{{ error }}</span>
        <UButton
          type="button"
          :loading="saving"
          :label="saving ? '保存中' : '保存首页视觉'"
          @click="save"
        />
      </div>
    </header>

    <div v-if="loading" class="space-y-3">
      <USkeleton class="h-40 rounded-lg" />
      <USkeleton class="h-40 rounded-lg" />
    </div>

    <template v-else>
      <div class="home-presentation-group">
        <div class="home-presentation-group-head">
          <h3>角色 Spotlight</h3>
          <span>{{ form.spotlight.length }} 位角色</span>
        </div>
        <article
          v-for="(character, index) in form.spotlight"
          :key="character.id"
          class="home-presentation-item"
        >
          <div class="home-presentation-item-head">
            <span class="home-presentation-file font-numeric">
              FILE {{ String(index + 1).padStart(3, "0") }}
            </span>
            <span class="home-presentation-id">{{ character.id }}</span>
            <USwitch v-model="character.enabled" label="显示" />
          </div>
          <div class="home-presentation-grid">
            <UFormField label="角色名">
              <UInput v-model="character.name" maxlength="80" class="w-full" />
            </UFormField>
            <UFormField label="系列">
              <UInput v-model="character.series" maxlength="80" class="w-full" />
            </UFormField>
            <UFormField label="阵营">
              <UInput v-model="character.faction" maxlength="80" class="w-full" />
            </UFormField>
            <UFormField label="主图地址">
              <UInput v-model="character.imageUrl" class="w-full" />
            </UFormField>
            <UFormField label="缩略图地址">
              <UInput v-model="character.thumbUrl" class="w-full" />
            </UFormField>
            <UFormField label="强调色">
              <UInput v-model="character.accent" placeholder="#7fd8ff" class="w-full" />
            </UFormField>
            <UFormField label="舞台色场" class="md:col-span-2">
              <UTextarea v-model="character.wash" :rows="2" class="w-full" />
            </UFormField>
          </div>
        </article>
      </div>

      <div class="home-presentation-group">
        <div class="home-presentation-group-head">
          <h3>分类封面</h3>
          <span>{{ form.categories.length }} 个分类</span>
        </div>
        <article
          v-for="cover in form.categories"
          :key="cover.categorySlug"
          class="home-presentation-item"
        >
          <div class="home-presentation-item-head">
            <span class="home-presentation-file">{{ cover.categoryName }}</span>
            <span class="home-presentation-id">{{ cover.categorySlug }}</span>
          </div>
          <div class="home-presentation-grid">
            <UFormField label="封面地址">
              <UInput v-model="cover.imageUrl" class="w-full" />
            </UFormField>
            <UFormField label="缩略图地址">
              <UInput v-model="cover.thumbUrl" class="w-full" />
            </UFormField>
            <UFormField label="强调色">
              <UInput v-model="cover.accent" placeholder="#ffd277" class="w-full" />
            </UFormField>
          </div>
        </article>
      </div>
    </template>
  </section>
</template>

<style scoped>
.home-presentation {
  border: 1px solid var(--ui-border-muted);
  background: var(--ui-bg);
}

.home-presentation-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.1rem 1.2rem;
  border-bottom: 1px solid var(--ui-border-muted);
}

.home-presentation-head h2 {
  font-size: 1rem;
  font-weight: 650;
  color: var(--ui-text-highlighted);
}

.home-presentation-head p {
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: var(--ui-text-muted);
}

.home-presentation-save {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.6rem;
}

.home-presentation-state {
  font-size: 0.7rem;
}

.home-presentation-state.is-saved {
  color: #12805c;
}

.home-presentation-state.is-error {
  color: #b42318;
}

.home-presentation-group {
  padding: 1.1rem 1.2rem;
  border-bottom: 1px solid var(--ui-border-muted);
}

.home-presentation-group:last-child {
  border-bottom: 0;
}

.home-presentation-group-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.8rem;
}

.home-presentation-group-head h3 {
  font-size: 0.82rem;
  font-weight: 650;
  color: var(--ui-text-highlighted);
}

.home-presentation-group-head span {
  font-size: 0.7rem;
  color: var(--ui-text-muted);
}

.home-presentation-item {
  padding: 0.9rem 0;
  border-top: 1px solid var(--ui-border-muted);
}

.home-presentation-item:first-of-type {
  border-top: 0;
  padding-top: 0;
}

.home-presentation-item-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.6rem;
  margin-bottom: 0.7rem;
}

.home-presentation-file {
  padding: 0.1rem 0.35rem;
  background: var(--gallery-accent);
  font-size: 0.6rem;
  font-weight: 650;
  color: var(--gallery-on-accent);
}

.home-presentation-id {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  font-family: var(--font-mono);
  font-size: 0.62rem;
  color: var(--ui-text-dimmed);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-presentation-grid {
  display: grid;
  gap: 0.7rem;
}

@media (min-width: 768px) {
  .home-presentation-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>
