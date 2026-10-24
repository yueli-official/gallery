<script setup lang="ts">
definePageMeta({ layout: false, middleware: ["auth", "admin"] });
useSeoMeta({ title: "初始化图库" });
useHead({ bodyAttrs: { class: "gallery-public-active" } });
const { user } = useAuth();
const { call } = useApi("gallery-authorization");
const { status, refresh } = useGalleryAdministratorClaim();
const { refreshMe } = useGalleryMe();
const pending = ref(false);
const complete = ref(false);
const failure = ref("");
async function claim() {
  if (pending.value || complete.value) return;
  pending.value = true;
  failure.value = "";
  try {
    await call("/setup/claim", { method: "POST" });
    complete.value = true;
    await refresh();
    await refreshMe();
    await navigateTo("/manage");
  } catch (reason) {
    await refresh().catch(() => null);
    failure.value = status.value?.claimed ? "本站已完成初始化，当前账号没有获得管理员权限。" : galleryFailureMessage(reason, "初始化未完成，请稍后重试。");
  } finally { pending.value = false; }
}
</script>
<template>
  <div class="setup-page">
    <header><NuxtLink to="/" class="setup-brand"><img src="/favicon.svg?v=20261024-brand" width="32" height="32" alt="" /><span>月离图库</span></NuxtLink><ConsumerAccountControl /></header>
    <main>
      <UIcon name="i-tabler-user-shield" class="setup-icon" />
      <h1>初始化图库</h1>
      <p>图库尚未设置管理员。确认后，当前账号将获得本站管理权限。此入口在初始化完成后关闭。</p>
      <div class="setup-account"><span>当前账号</span><strong>{{ user?.name || user?.email || '已登录用户' }}</strong></div>
      <UAlert v-if="failure" color="error" variant="subtle" :title="failure" />
      <div class="setup-actions"><UButton to="/" color="neutral" variant="ghost" label="返回图库" /><UButton :label="complete ? '已完成' : pending ? '初始化中' : '初始化站点并成为管理员'" :loading="pending" :disabled="complete || status?.claimed" @click="claim" /></div>
    </main>
  </div>
</template>
<style scoped>
.setup-page{min-height:100dvh;background:var(--gallery-panel);color:var(--gallery-ink)}.setup-page>header{display:flex;align-items:center;justify-content:space-between;padding:1rem 1.5rem;border-bottom:1px solid var(--gallery-line);background:var(--gallery-canvas)}.setup-brand{display:flex;align-items:center;gap:.65rem;font-size:.85rem;font-weight:600}.setup-page main{max-width:30rem;margin:clamp(2rem,10vh,6rem) auto;padding:1.5rem;border:1px solid var(--gallery-line);border-radius:.6rem;background:var(--gallery-canvas)}.setup-icon{font-size:1.8rem;color:var(--gallery-accent)}.setup-page h1{font-size:1.3rem;font-weight:650;margin:.8rem 0 .6rem}.setup-page main>p{font-size:.78rem;line-height:1.8;color:var(--gallery-muted)}.setup-account{display:flex;flex-direction:column;gap:.4rem;margin:1.25rem 0;padding:1rem;border-radius:.4rem;background:var(--gallery-panel);overflow-wrap:anywhere}.setup-account>span{font-size:.68rem;color:var(--gallery-muted)}.setup-account>strong{font-size:.85rem;font-weight:550}.setup-actions{display:flex;justify-content:flex-end;gap:.5rem;flex-wrap:wrap;margin-top:1.25rem}@media(max-width:550px){.setup-page main{margin:2rem .75rem;padding:1.2rem}.setup-page>header{padding:.75rem}}
</style>
