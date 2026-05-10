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
  const err = validate()
  if (err) { notify.warn(err); return }
  saving.value = true
  try {
    const saved = await api.saveConfig(cfg.value)
    store.config = saved; cfg.value = saved; dirty.value = false
    notify.success(t.value.saveSettings)
  } catch (e: any) { notify.error(e) }
  finally { saving.value = false }
}

function validate(): string {
  if (!cfg.value) return ''
  if (!cfg.value.basePath.startsWith('/')) return t.value.basePathError
  if (cfg.value.preview.ttlMinutes < 10) return t.value.previewTTLError
  return ''
}
</script>

<template>
  <div v-if="cfg" class="space-y-6">
    <Card>
      <CardHeader><CardTitle class="text-sm font-medium">{{ t.generalSettings }}</CardTitle></CardHeader>
      <CardContent class="grid gap-4 sm:grid-cols-2">
        <div class="grid gap-1.5">
          <Label class="text-xs uppercase tracking-wider text-muted-foreground">{{ t.siteId }}</Label>
          <Input v-model="cfg.site.id" class="h-9" @input="dirty=true" />
        </div>
        <div class="grid gap-1.5">
          <Label class="text-xs uppercase tracking-wider text-muted-foreground">{{ t.siteName }}</Label>
          <Input v-model="cfg.site.name" class="h-9" @input="dirty=true" />
        </div>
        <div class="grid gap-1.5 sm:col-span-2">
          <Label class="text-xs uppercase tracking-wider text-muted-foreground">{{ t.basePath }}</Label>
          <Input v-model="cfg.basePath" class="h-9" @input="dirty=true" />
          <p class="text-xs text-muted-foreground/60">{{ t.basePathExample }}</p>
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader><CardTitle class="text-sm font-medium">{{ t.previewSettings }}</CardTitle></CardHeader>
      <CardContent class="grid gap-4 sm:grid-cols-2">
        <div class="grid gap-1.5">
          <Label class="text-xs uppercase tracking-wider text-muted-foreground">{{ t.previewTTL }}</Label>
          <Input v-model.number="cfg.preview.ttlMinutes" type="number" min="10" class="h-9" @input="dirty=true" />
        </div>
        <div class="grid gap-1.5">
          <Label class="text-xs uppercase tracking-wider text-muted-foreground">{{ t.maxUploadMb }}</Label>
          <Input type="number" min="1" class="h-9" :model-value="Math.round(cfg.maxUploadBytes / 1024 / 1024)"
            @update:model-value="(v: string | number) => { cfg!.maxUploadBytes = Number(v) * 1024 * 1024; dirty=true }" />
        </div>
      </CardContent>
      <CardFooter>
        <Button class="rounded-full px-5" :disabled="saving || !dirty" @click="save">
          <SaveIcon class="h-4 w-4 mr-1.5" />{{ saving ? t.saving : t.saveSettings }}
        </Button>
      </CardFooter>
    </Card>
  </div>
  <Skeleton v-else class="h-[300px] max-w-3xl" />
</template>
