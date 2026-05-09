<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { AlertCircleIcon, RefreshCwIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'

const route = useRoute()
const notify = useNotify()
const slug = decodeURIComponent(route.params.slug as string)
const url = ref('')
const loading = ref(true)
const error = ref(false)

async function loadPreview() {
  loading.value = true
  error.value = false
  url.value = ''
  try {
    const draft = await api.post(slug)
    const result = await api.preview(draft)
    if (result.error) { notify.error(result.error); error.value = true; return }
    url.value = result.url
  } catch (e: any) {
    notify.error(e)
    error.value = true
  } finally {
    loading.value = false
  }
}

onMounted(loadPreview)
</script>

<template>
  <div class="h-[calc(100vh-48px)] flex flex-col">
    <div v-if="loading" class="flex-1 flex flex-col gap-3 p-6 animate-fade-in">
      <Skeleton class="h-8 w-48" />
      <Skeleton class="h-4 w-full max-w-md" />
      <Skeleton class="flex-1 rounded" />
    </div>
    <div v-else-if="error" class="flex-1 flex flex-col items-center justify-center gap-3 animate-fade-in">
      <AlertCircleIcon class="h-10 w-10 text-destructive/40" />
      <p class="font-serif text-sm text-muted-foreground">预览生成失败</p>
      <Button variant="outline" class="rounded-full px-5" @click="loadPreview">
        <RefreshCwIcon class="h-3.5 w-3.5 mr-1.5" />重试
      </Button>
    </div>
    <iframe v-else-if="url" :src="url" class="flex-1 border-none bg-background animate-fade-in" />
  </div>
</template>
