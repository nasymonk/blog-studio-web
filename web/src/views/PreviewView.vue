<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '@/services/api'
import { useNotify } from '@/composables/useNotify'

const route = useRoute()
const notify = useNotify()
const slug = decodeURIComponent(route.params.slug as string)
const url = ref('')
const loading = ref(true)

onMounted(async () => {
  try {
    const draft = await api.post(slug)
    const result = await api.preview(draft)
    if (result.error) { notify.error(result.error); return }
    url.value = result.url
  } catch (e: any) {
    notify.error(e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div style="height:calc(100vh - var(--topbar-height));display:flex;flex-direction:column">
    <div v-if="loading" style="display:grid;place-items:center;height:100%;color:var(--tertiary)">生成预览中…</div>
    <iframe v-else-if="url" :src="url" style="flex:1;border:none;background:white"></iframe>
    <div v-else style="display:grid;place-items:center;height:100%;color:var(--error)">预览生成失败</div>
  </div>
</template>
