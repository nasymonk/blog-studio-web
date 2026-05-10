<script setup lang="ts">
import { ChevronRightIcon } from 'lucide-vue-next'
import { RouterLink } from 'vue-router'

export interface BreadcrumbItem {
  label: string
  to?: string
}

defineProps<{ items: BreadcrumbItem[] }>()
</script>

<template>
  <nav class="flex items-center gap-1 text-sm text-muted-foreground mb-4" aria-label="面包屑导航">
    <template v-for="(item, i) in items" :key="i">
      <ChevronRightIcon v-if="i > 0" class="h-3 w-3 shrink-0" />
      <RouterLink
        v-if="item.to"
        :to="item.to"
        class="hover:text-foreground transition-colors link-slide"
      >
        {{ item.label }}
      </RouterLink>
      <span v-else class="text-foreground font-medium">{{ item.label }}</span>
    </template>
  </nav>
</template>
