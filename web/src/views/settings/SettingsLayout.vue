<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { useI18n } from '@/i18n'

const { t } = useI18n()

const tabs = [
  { to: '/settings/general', label: () => t.value.generalSettings },
  { to: '/settings/writing', label: () => t.value.writingSettings },
  { to: '/settings/security', label: () => t.value.securitySettings },
  { to: '/settings/audit', label: () => t.value.auditLog },
]
</script>

<template>
  <div class="settings-page">
    <nav class="settings-sidebar">
      <RouterLink v-for="tab in tabs" :key="tab.to" :to="tab.to" class="nav-btn">
        {{ tab.label() }}
      </RouterLink>
    </nav>
    <div class="settings-content">
      <RouterView />
    </div>
  </div>
</template>
