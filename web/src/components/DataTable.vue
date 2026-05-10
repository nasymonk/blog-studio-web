<script setup lang="ts">
import { ref, computed } from 'vue'
import { ChevronUpIcon, ChevronDownIcon } from 'lucide-vue-next'

export interface Column {
  key: string
  label: string
  sortable?: boolean
  width?: string
}

const props = defineProps<{
  columns: Column[]
  data: Record<string, unknown>[]
  emptyMessage?: string
}>()

const emit = defineEmits<{
  rowClick: [row: Record<string, unknown>]
}>()

const sortKey = ref('')
const sortDir = ref<'asc' | 'desc'>('asc')

function toggleSort(key: string) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'asc'
  }
}

const sorted = computed(() => {
  if (!sortKey.value) return props.data
  return [...props.data].sort((a, b) => {
    const av = a[sortKey.value]
    const bv = b[sortKey.value]
    const cmp = String(av ?? '').localeCompare(String(bv ?? ''))
    return sortDir.value === 'asc' ? cmp : -cmp
  })
})
</script>

<template>
  <div class="border rounded-lg overflow-hidden">
    <table class="w-full text-sm">
      <thead>
        <tr class="bg-muted/50 border-b border-border">
          <th
            v-for="col in columns"
            :key="col.key"
            class="text-left px-3 py-2 font-medium text-muted-foreground text-xs"
            :class="{ 'cursor-pointer hover:text-foreground select-none': col.sortable }"
            :style="{ width: col.width }"
            @click="col.sortable && toggleSort(col.key)"
          >
            <div class="flex items-center gap-1">
              {{ col.label }}
              <template v-if="col.sortable && sortKey === col.key">
                <ChevronUpIcon v-if="sortDir === 'asc'" class="h-3 w-3" />
                <ChevronDownIcon v-else class="h-3 w-3" />
              </template>
            </div>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="data.length === 0">
          <td :colspan="columns.length" class="text-center py-8 text-muted-foreground text-sm">
            {{ emptyMessage ?? '暂无数据' }}
          </td>
        </tr>
        <tr
          v-for="(row, i) in sorted"
          :key="i"
          class="border-b border-border last:border-0 hover:bg-muted/30 transition-colors cursor-pointer"
          @click="emit('rowClick', row)"
        >
          <td v-for="col in columns" :key="col.key" class="px-3 py-2">
            <slot :name="col.key" :row="row" :value="row[col.key]">
              {{ row[col.key] }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
