<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useI18n } from '@/i18n'

const { t } = useI18n()
const route = useRoute()

const tabs = [
  { to: '/settings/general', label: () => t.value.generalSettings },
  { to: '/settings/writing', label: () => t.value.writingSettings },
  { to: '/settings/security', label: () => t.value.securitySettings },
  { to: '/settings/audit', label: () => t.value.auditLog },
]
</script>

<template>
  <div class="flex flex-col gap-6">
    <nav class="flex gap-1 border-b border-border" role="tablist">
      <RouterLink
        v-for="tab in tabs"
        :key="tab.to"
        :to="tab.to"
        role="tab"
        class="px-3 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors -mb-px border-b-2 border-transparent"
        :class="{ 'border-primary text-foreground': route.path === tab.to }"
      >
        {{ tab.label() }}
      </RouterLink>
    </nav>
    <RouterView />
  </div>
</template>
