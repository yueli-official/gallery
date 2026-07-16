<script setup lang="ts">
import type {
  GalleryClassificationCatalogFacet,
  GalleryClassificationCatalogNode,
  GalleryClassificationGovernanceCommand,
} from "~/types/gallery";

type IdentityKind = GalleryClassificationGovernanceCommand["kind"];
type ManagedIdentity =
  GalleryClassificationCatalogFacet | GalleryClassificationCatalogNode;
type Operation = "status" | "reparent" | "merge" | "delete";

defineProps<{ facets: GalleryClassificationCatalogFacet[] }>();
defineEmits<{
  action: [operation: Operation, kind: IdentityKind, item: ManagedIdentity];
}>();

const statusColor = (status: string) =>
  (status === "active"
    ? "success"
    : status === "replaced"
      ? "warning"
      : "neutral") as any;
</script>

<template>
  <section aria-labelledby="classification-facet-heading">
    <div class="mb-3">
      <h2
        id="classification-facet-heading"
        class="font-semibold text-highlighted"
      >
        Facet
      </h2>
      <p class="mt-1 text-sm text-muted">
        Facet 本身只启停或删除；需要合并轴时，先逐个治理 Value，再删除空轴。
      </p>
    </div>
    <div class="grid gap-4 xl:grid-cols-2">
      <article
        v-for="facet in facets"
        :key="facet.id"
        class="rounded-lg border border-default bg-default p-4 sm:p-5"
      >
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <div class="flex items-center gap-2">
              <h3 class="font-semibold text-highlighted">{{ facet.name }}</h3>
              <UBadge
                :color="statusColor(facet.status)"
                variant="soft"
                :label="facet.status"
              />
            </div>
            <p class="mt-1 text-xs text-muted">
              {{ facet.slug }} · {{ facet.id }}
            </p>
          </div>
          <div class="flex gap-2">
            <UButton
              v-if="facet.status !== 'replaced'"
              class="min-h-11"
              color="neutral"
              variant="ghost"
              size="sm"
              :label="facet.status === 'active' ? '停用' : '启用'"
              @click="$emit('action', 'status', 'facet', facet)"
            />
            <UButton
              class="min-h-11"
              color="error"
              variant="ghost"
              size="sm"
              label="删除"
              @click="$emit('action', 'delete', 'facet', facet)"
            />
          </div>
        </div>
        <div class="mt-4 divide-y divide-default">
          <div v-for="value in facet.values" :key="value.id" class="py-3">
            <div
              class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between"
            >
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-highlighted">{{
                    value.name
                  }}</span>
                  <UBadge
                    size="xs"
                    :color="statusColor(value.status)"
                    variant="soft"
                    :label="value.status"
                  />
                </div>
                <p class="mt-1 truncate text-xs text-dimmed">
                  {{ value.slug }} · {{ value.id }}
                </p>
              </div>
              <div class="flex flex-wrap gap-2">
                <UButton
                  v-if="value.status !== 'replaced'"
                  class="min-h-11"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  :label="value.status === 'active' ? '停用' : '启用'"
                  @click="$emit('action', 'status', 'facet_value', value)"
                />
                <UButton
                  v-if="value.status !== 'replaced'"
                  class="min-h-11"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  label="移动"
                  @click="$emit('action', 'reparent', 'facet_value', value)"
                />
                <UButton
                  v-if="value.status !== 'replaced'"
                  class="min-h-11"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  label="合并"
                  @click="$emit('action', 'merge', 'facet_value', value)"
                />
                <UButton
                  class="min-h-11"
                  color="error"
                  variant="ghost"
                  size="xs"
                  label="删除"
                  @click="$emit('action', 'delete', 'facet_value', value)"
                />
              </div>
            </div>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>
