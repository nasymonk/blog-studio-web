<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useI18n } from '@/i18n'
import Breadcrumb from '@/components/Breadcrumb.vue'

const { t } = useI18n()
const route = useRoute()

const tabs = [
  { to: '/settings/general', label: () => t.value.generalSettings },
  { to: '/settings/writing', label: () => t.value.writingSettings },
  { to: '/settings/security', label: () => t.value.securitySettings },
  { to: '/settings/audit', label: () => t.value.auditLog },
]

const breadcrumbItems = computed(() => {
  const currentTab = tabs.find(tab => route.path === tab.to)
  return [
    { label: t.value.settings },
    ...(currentTab ? [{ label: currentTab.label() }] : []),
  ]
})
</script>

<template>
  <div class="flex flex-col gap-6 max-w-3xl">
    <Breadcrumb :items="breadcrumbItems" />
    <nav class="flex gap-1 border-b border-border/60" role="tablist">
      <RouterLink
        v-for="tab in tabs"
        :key="tab.to"
        :to="tab.to"
        role="tab"
        :aria-selected="route.path === tab.to"
        class="relative px-3 py-2 text-sm text-muted-foreground transition-colors -mb-px"
        :class="route.path === tab.to ? 'text-foreground font-medium' : 'hover:text-foreground'"
      >
        {{ tab.label() }}
        <div
          v-if="route.path === tab.to"
          class="absolute bottom-0 left-2 right-2 h-0.5 bg-accent rounded-full animate-fade-in"
        />
      </RouterLink>
    </nav>
    <div role="tabpanel" class="pt-4">
      <RouterView />
    </div>
  </div>
</template>
