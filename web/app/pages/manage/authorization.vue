<script setup lang="ts">
import { ManagePage, TabbedSurface, AuthorizationUser, AuthorizationApplication, AuthorizationGrantBadge } from "@yueli/ui/admin";

import { assetMediaUrl } from "@yueli/asset-nuxt/media";

interface RoleView {
  key: string;
  displayName: string;
  kind: string;
  protected: boolean;
  capabilities: string[];
  assignmentSources: string[];
}
interface ApplicationView {
  createdAt: string;
  id: string;
  subject: string;
  role: string;
  reason: string;
}
interface GrantView {
  id: string;
  subject: string;
  role: string;
  source: string;
}
interface ConsoleView {
  activeRevision: number;
  policy: { number: number; state: string };
  roles: RoleView[];
  automaticRules: { key: string; enabled: boolean }[];
  applications: ApplicationView[];
  grants: GrantView[];
  capabilities: { key: string; displayName: string }[];
}

definePageMeta({ layout: "manage", middleware: ["auth", "admin"] });
useSeoMeta({ title: "权限与申请" });

const { call } = useApi("gallery-authorization");
const { isAdministrator } = useGalleryMe();
const route = useRoute();
const router = useRouter();
const hydrated = useClientHydrated();
const toast = useToast();
const busy = ref(false);
const createRoleOpen = ref(false);
const roleForm = reactive({
  key: "",
  displayName: "",
  capabilities: [] as string[],
  assignmentSources: ["application", "invitation", "direct"] as string[],
});
const grantForm = reactive({ subject: "", role: "content_operator" });
const { data, pending, error, refresh } = await useAsyncData(
  "gallery-authorization-console",
  () => call<ConsoleView>("/manage/console"),
  { server: false },
);
const state = computed(() => data.value);
const draft = computed(() => state.value?.policy.state === "draft");
const subjects = computed(() => [...(state.value?.applications.map(item => item.subject) || []), ...(state.value?.grants.map(item => item.subject) || [])]);
const directory = usePublicUserDirectory(subjects);
const accountOrigin = String(useRuntimeConfig().public.accountUrl).replace(/\/$/, "");
function userInfo(subject: string) {
  const user = directory.users.value[subject];
  return {
    subject, name: user?.displayName, handle: user?.handle,
    avatarUrl: user?.avatar ? assetMediaUrl(user.avatar, "thumbnail") : undefined,
    profileUrl: `${accountOrigin}/u/${encodeURIComponent(subject)}`,
    loading: directory.pending.value,
  };
}

type AuthorizationTab = "applications" | "permissions" | "users";
const authorizationTabs: readonly AuthorizationTab[] = [
  "applications",
  "permissions",
  "users",
];
const activeTab = computed<AuthorizationTab>({
  get: () => {
    const value = String(route.query.tab || "applications");
    return authorizationTabs.includes(value as AuthorizationTab)
      ? (value as AuthorizationTab)
      : "applications";
  },
  set: (value) => {
    void router.replace({
      query: {
        ...route.query,
        tab: value === "applications" ? undefined : value,
      },
    });
  },
});
const tabItems = computed(() => [
  {
    label: "申请",
    value: "applications",
    icon: "i-tabler-inbox",
    badge: state.value?.applications.length || undefined,
  },
  { label: "权限", value: "permissions", icon: "i-tabler-shield-lock" },
  {
    label: "用户管理",
    value: "users",
    icon: "i-tabler-users",
    badge: state.value?.grants.length || undefined,
  },
]);

function roleName(roleKey: string) {
  return (
    state.value?.roles.find((role) => role.key === roleKey)?.displayName ||
    roleKey
  );
}

async function mutate(task: () => Promise<unknown>, _success: string) {
  if (busy.value) return;
  busy.value = true;
  try {
    const result = await task();
    if (result === false) return;
    await refresh();
  } catch (failure) {
    const message = galleryFailureMessage(failure,"请刷新后重试。");
    toast.add({
      title: "操作失败",
      description: message || "请刷新后重试。",
      color: "error",
      icon: "i-tabler-alert-circle",
    });
  } finally {
    busy.value = false;
  }
}

function createDraft() {
  if (!state.value?.activeRevision) return;
  return mutate(
    () =>
      call("/manage/policies/drafts", {
        method: "POST",
        body: { expectedActiveRevision: state.value!.activeRevision },
      }),
    "策略草稿已创建",
  );
}

function toggleRoleCapability(role: RoleView, capability: string) {
  if (!draft.value || role.protected) return;
  const capabilities = role.capabilities.includes(capability)
    ? role.capabilities.filter((item) => item !== capability)
    : [...role.capabilities, capability];
  return mutate(
    () =>
      call(
        `/manage/policies/${state.value!.policy.number}/roles/${role.key}/capabilities`,
        { method: "PUT", body: { capabilities } },
      ),
    "角色能力已更新到草稿",
  );
}

function toggleAutomatic(enabled: boolean) {
  const rule = state.value?.automaticRules[0];
  if (!draft.value || !rule) return;
  return mutate(
    () =>
      call(
        `/manage/policies/${state.value!.policy.number}/automatic/${rule.key}`,
        { method: "PUT", body: { enabled } },
      ),
    enabled ? "已启用注册自动授权" : "已关闭注册自动授权",
  );
}

async function validateAndActivate() {
  const current = state.value;
  if (!draft.value || !current) return;
  await mutate(async () => {
    const validation = await call<{ valid: boolean; violations: string[] }>(
      `/manage/policies/${current.policy.number}/validate`,
      { method: "POST" },
    );
    if (!validation.valid) throw new Error(validation.violations.join("；"));
    const impact = await call<{ removedBindings: number }>(
      `/manage/policies/${current.policy.number}/preview`,
      { method: "POST" },
    );
    if (
      impact.removedBindings > 0 &&
      !window.confirm(
        `本次发布会移除 ${impact.removedBindings} 项能力绑定，是否继续？`,
      )
    )
      return false;
    await call(
      `/manage/policies/${current.policy.number}/activate`,
      {
        method: "POST",
        body: { expectedActiveRevision: current.activeRevision },
      },
    );
  }, "权限策略已发布");
}

function review(application: ApplicationView, decision: "approve" | "reject") {
  return mutate(
    () =>
      call(
        `/manage/applications/${application.id}/review`,
        {
          method: "POST",
          body: {
            decision,
            reason: decision === "approve" ? "管理员批准" : "管理员拒绝",
          },
        },
      ),
    decision === "approve" ? "申请已批准" : "申请已拒绝",
  );
}

function createRole() {
  if (!draft.value || !state.value) return;
  return mutate(async () => {
    await call(
      `/manage/policies/${state.value!.policy.number}/roles`,
      { method: "POST", body: roleForm },
    );
    createRoleOpen.value = false;
    Object.assign(roleForm, {
      key: "",
      displayName: "",
      capabilities: [],
      assignmentSources: ["application", "invitation", "direct"],
    });
  }, "自定义角色已创建");
}

function openCreateRole() {
  createRoleOpen.value = true;
}

function closeCreateRole() {
  createRoleOpen.value = false;
}

function grantRole() {
  if (!grantForm.subject.trim() || !grantForm.role) return;
  return mutate(async () => {
    await call("/manage/grants", {
      method: "POST",
      body: { subject: grantForm.subject.trim(), role: grantForm.role },
    });
    grantForm.subject = "";
  }, "角色已直接授予");
}

function revokeGrant(grant: GrantView) {
  if (!window.confirm(`确定撤销 ${grant.subject} 的 ${grant.role} 角色吗？`))
    return;
  return mutate(
    () =>
      call(`/manage/grants/${grant.id}`, {
        method: "DELETE",
      }),
    "角色授权已撤销",
  );
}
</script>

<template>
  <ManagePage id="authorization" title="权限与申请" icon="i-tabler-shield-lock">
    <template #actions>
      <UButton
        v-if="activeTab === 'permissions' && state && !draft"
        label="创建策略草稿"
        icon="i-tabler-file-plus"
        :loading="busy"
        @click="createDraft"
      />
      <UButton
        v-else-if="activeTab === 'permissions' && draft"
        label="验证并发布"
        icon="i-tabler-rocket"
        :loading="busy"
        @click="validateAndActivate"
      />
    </template>

    <div v-if="!hydrated || pending" class="yueli-card p-4 sm:p-5">
      <SkeletonList :rows="8" />
    </div>
    <UAlert
      v-else-if="!isAdministrator"
      color="error"
      icon="i-tabler-lock"
      title="只有管理员可以管理本站权限"
      description="图库权限独立存储，不继承用户中心或其他站点的管理员角色。"
    />
    <UAlert
      v-else-if="error || !state"
      color="error"
      icon="i-tabler-alert-circle"
      title="权限配置加载失败"
      description="请检查 Gallery API 与本站数据库状态。"
    />

    <TabbedSurface
      v-else
      v-model="activeTab"
      :items="tabItems"
      navigation-label="权限管理"
      data-manage-surface="authorization"
    >
      <UAlert v-if="directory.error.value" title="用户资料加载失败" description="请重试加载用户昵称与头像。" color="error" class="m-4">
        <template #actions><UButton label="重试" color="neutral" variant="outline" @click="directory.refresh()" /></template>
      </UAlert>
      <section
        v-if="activeTab === 'applications'"
        aria-labelledby="authorization-applications-title"
      >
        <div
          class="flex items-center justify-between gap-3 border-b border-default px-4 py-3 sm:px-5"
        >
          <h2
            id="authorization-applications-title"
            class="text-sm font-semibold text-highlighted"
          >
            待处理申请
          </h2>
          <UBadge
            :label="String(state.applications.length)"
            color="neutral"
            variant="soft"
          />
        </div>
        <div v-if="state.applications.length" class="divide-y divide-default">
          <AuthorizationApplication
            v-for="application in state.applications"
            :key="application.id"
            :user="userInfo(application.subject)"
            :role="roleName(application.role)"
            :reason="application.reason"
            :created-at="application.createdAt"
            :busy="busy"
            @review="review(application, $event)"
          />
        </div>
        <ManageEmpty
          v-else
          class="m-4 sm:m-5"
          icon="i-tabler-user-check"
          text="当前没有待处理申请"
        />
      </section>

      <section
        v-else-if="activeTab === 'permissions'"
        aria-labelledby="authorization-permissions-title"
      >
        <div
          class="flex flex-wrap items-center justify-between gap-3 border-b border-default px-4 py-3 sm:px-5"
        >
          <div class="min-w-0">
            <h2
              id="authorization-permissions-title"
              class="text-sm font-semibold text-highlighted"
            >
              角色与能力
            </h2>
            <p class="mt-0.5 text-xs text-muted">
              生效修订 {{ state.activeRevision }} · 当前修订
              {{ state.policy.number }}
            </p>
          </div>
          <div class="flex items-center gap-2">
            <UBadge
              :label="draft ? '草稿' : '已生效'"
              :color="draft ? 'warning' : 'success'"
              variant="subtle"
            />
            <UButton
              v-if="draft"
              label="新建角色"
              icon="i-tabler-user-plus"
              color="neutral"
              variant="soft"
              size="sm"
              @click="openCreateRole"
            />
          </div>
        </div>

        <div class="grid gap-3 p-4 sm:p-5 lg:grid-cols-2">
          <article
            v-for="role in state.roles"
            :key="role.key"
            class="rounded-xl border border-default p-4"
          >
            <div class="flex items-center justify-between gap-3">
              <div class="min-w-0">
                <p class="truncate font-medium text-highlighted">
                  {{ role.displayName }}
                </p>
                <p class="mt-0.5 truncate text-xs text-muted">
                  {{ role.key }}
                </p>
              </div>
              <UBadge
                :label="
                  role.protected
                    ? '受保护'
                    : role.kind === 'custom'
                      ? '自定义'
                      : '内置'
                "
                color="neutral"
                variant="soft"
              />
            </div>
            <div class="mt-4 grid gap-2 sm:grid-cols-2">
              <UCheckbox
                v-for="capability in state.capabilities"
                :key="capability.key"
                :model-value="role.capabilities.includes(capability.key)"
                :label="capability.displayName"
                :disabled="role.protected || !draft || busy"
                @update:model-value="toggleRoleCapability(role, capability.key)"
              />
            </div>
          </article>
        </div>

        <div
          class="flex items-start justify-between gap-4 border-t border-default px-4 py-4 sm:px-5"
        >
          <div>
            <h3 class="text-sm font-semibold text-highlighted">
              注册用户自动成为内容运营者
            </h3>
            <p class="mt-1 max-w-2xl text-xs leading-5 text-muted">
              关闭后，新用户需要申请或由管理员直接授权。
            </p>
          </div>
          <USwitch
            :model-value="state.automaticRules[0]?.enabled ?? false"
            :disabled="!draft || busy"
            aria-label="注册用户自动成为内容运营者"
            @update:model-value="toggleAutomatic(Boolean($event))"
          />
        </div>
      </section>

      <section v-else aria-labelledby="authorization-users-title">
        <div
          class="grid gap-3 border-b border-default p-4 sm:grid-cols-[minmax(0,1fr)_14rem_auto] sm:p-5"
        >
          <UFormField label="用户标识">
            <UInput
              v-model="grantForm.subject"
              placeholder="Identity subject"
              class="w-full"
            />
          </UFormField>
          <UFormField label="角色">
            <USelect
              v-model="grantForm.role"
              value-key="value"
              :items="
                state.roles
                  .filter(
                    (role) =>
                      role.kind !== 'custom' || role.capabilities.length,
                  )
                  .map((role) => ({
                    label: role.displayName,
                    value: role.key,
                  }))
              "
              class="w-full"
            />
          </UFormField>
          <div class="flex items-end">
            <UButton
              label="直接授予"
              icon="i-tabler-user-plus"
              :disabled="!grantForm.subject.trim() || busy"
              @click="grantRole"
            />
          </div>
        </div>

        <div
          class="flex items-center justify-between gap-3 border-b border-default px-4 py-3 sm:px-5"
        >
          <h2
            id="authorization-users-title"
            class="text-sm font-semibold text-highlighted"
          >
            已授权用户
          </h2>
          <UBadge
            :label="String(state.grants.length)"
            color="neutral"
            variant="soft"
          />
        </div>
        <div v-if="state.grants.length" class="divide-y divide-default">
          <article
            v-for="grant in state.grants"
            :key="grant.id"
            class="grid gap-3 px-4 py-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center sm:px-5"
          >
            <div class="min-w-0">
              <AuthorizationUser v-bind="userInfo(grant.subject)" />
              <AuthorizationGrantBadge class="mt-1" :role="roleName(grant.role)" :source="grant.source" />
            </div>
            <UButton
              label="撤销"
              color="error"
              variant="ghost"
              :disabled="busy"
              @click="revokeGrant(grant)"
            />
          </article>
        </div>
        <ManageEmpty
          v-else
          class="m-4 sm:m-5"
          icon="i-tabler-users"
          text="当前没有角色授权"
        />
      </section>
    </TabbedSurface>

    <UModal
      v-model:open="createRoleOpen"
      title="新建自定义角色"
      description="角色能力写入当前策略草稿，发布后才生效。"
    >
      <template #body>
        <div class="space-y-4">
          <UFormField label="角色标识" required>
            <UInput
              v-model="roleForm.key"
              placeholder="editor"
              class="w-full"
            />
          </UFormField>
          <UFormField label="显示名称" required>
            <UInput
              v-model="roleForm.displayName"
              placeholder="编辑"
              class="w-full"
            />
          </UFormField>
          <UFormField label="能力">
            <div class="grid gap-2 sm:grid-cols-2">
              <UCheckbox
                v-for="capability in state?.capabilities ?? []"
                :key="capability.key"
                :model-value="roleForm.capabilities.includes(capability.key)"
                :label="capability.displayName"
                @update:model-value="
                  roleForm.capabilities = $event
                    ? [...roleForm.capabilities, capability.key]
                    : roleForm.capabilities.filter(
                        (item) => item !== capability.key,
                      )
                "
              />
            </div>
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton
            label="取消"
            color="neutral"
            variant="outline"
            @click="closeCreateRole"
          />
          <UButton
            label="创建角色"
            :disabled="!roleForm.key.trim() || !roleForm.displayName.trim()"
            :loading="busy"
            @click="createRole"
          />
        </div>
      </template>
    </UModal>
  </ManagePage>
</template>
