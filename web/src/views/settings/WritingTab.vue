<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { SaveIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import type { Config } from '@/services/api'
import { useStore } from '@/store'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'

const store = useStore()
const { t } = useI18n()
const notify = useNotify()
const cfg = ref<Config | null>(null)
const dirty = ref(false)
const saving = ref(false)

onMounted(async () => {
  try {
    cfg.value = store.config ?? await api.config()
    store.config = cfg.value
  } catch (e: any) { notify.error(e) }
})

async function save() {
  if (!cfg.value) return
  saving.value = true
  try {
    const saved = await api.saveConfig(cfg.value)
    store.config = saved; cfg.value = saved; dirty.value = false
    notify.success(t.value.saveSettings)
  } catch (e: any) { notify.error(e) }
  finally { saving.value = false }
}
</script>

<template>
  <div v-if="cfg" class="space-y-6">
    <Card>
      <CardHeader><CardTitle class="text-sm font-medium">{{ t.writingSettings }}</CardTitle></CardHeader>
      <CardContent class="grid gap-4 sm:grid-cols-2">
        <div class="grid gap-1.5 sm:col-span-2">
          <Label class="text-xs uppercase tracking-wider text-muted-foreground">{{ t.blogRoot }}</Label>
          <Input v-model="cfg.site.blogRoot" class="h-9" @input="dirty=true" />
          <p class="text-xs text-muted-foreground/60">{{ t.blogRootHint }}</p>
        </div>
        <div class="grid gap-1.5">
          <Label class="text-xs uppercase tracking-wider text-muted-foreground">{{ t.postSection }}</Label>
          <Input v-model="cfg.site.postSection" class="h-9" @input="dirty=true" />
          <p class="text-xs text-muted-foreground/60">{{ t.postSectionHint }}</p>
        </div>
        <div class="grid gap-1.5">
          <Label class="text-xs uppercase tracking-wider text-muted-foreground">{{ t.buildCommand }}</Label>
          <Select v-model="cfg.site.buildCommand" @update:model-value="dirty=true">
            <SelectTrigger class="h-9"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="hugo --minify">hugo --minify</SelectItem>
              <SelectItem value="hugo">hugo</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardContent>
      <CardFooter>
        <Button class="rounded-full px-5" :disabled="saving || !dirty" @click="save">
          <SaveIcon class="h-4 w-4 mr-1.5" />{{ saving ? t.saving : t.saveSettings }}
        </Button>
      </CardFooter>
    </Card>
  </div>
  <Skeleton v-else class="h-[200px] max-w-3xl" />
</template>
