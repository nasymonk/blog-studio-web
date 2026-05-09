<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { UploadIcon, SaveIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from '@/components/ui/card'

const { t } = useI18n()
const notify = useNotify()

const description = ref('')
const profileImage = ref('')
const loading = ref(false)
const saving = ref(false)
const uploadingAvatar = ref(false)
const avatarFileInput = ref<HTMLInputElement | null>(null)

onMounted(async () => {
  loading.value = true
  try {
    const data = await api.getSite()
    description.value = data.description || ''
    profileImage.value = data.profileImage || ''
  } catch (e: any) {
    notify.error(e)
  } finally {
    loading.value = false
  }
})

async function save() {
  saving.value = true
  try {
    await api.saveSite({ description: description.value, profileImage: profileImage.value })
    notify.success(t.value.saveHomepage)
  } catch (e: any) {
    notify.error(e)
  } finally {
    saving.value = false
  }
}

async function onAvatarSelect(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  uploadingAvatar.value = true
  try {
    const result = await api.uploadAvatar(file)
    profileImage.value = result.path
    notify.success(t.value.uploadAvatar)
  } catch (e: any) {
    notify.error(e)
  } finally {
    uploadingAvatar.value = false
  }
}
</script>

<template>
  <div class="max-w-xl space-y-5">
    <Card>
      <CardHeader><CardTitle class="font-serif">{{ t.avatar }}</CardTitle></CardHeader>
      <CardContent>
        <div class="flex items-center gap-4">
          <div class="h-16 w-16 rounded-full bg-muted flex items-center justify-center overflow-hidden shrink-0">
            <img v-if="profileImage" :src="profileImage" class="h-full w-full object-cover" alt="avatar" />
            <span v-else class="text-2xl">🌿</span>
          </div>
          <div class="space-y-1">
            <label class="cursor-pointer">
              <Input ref="avatarFileInput" type="file" accept="image/*" class="hidden" @change="onAvatarSelect" />
              <Button variant="outline" size="sm" as-child>
                <span><UploadIcon class="h-3.5 w-3.5 mr-1" />{{ uploadingAvatar ? t.loading : t.uploadAvatar }}</span>
              </Button>
            </label>
            <p class="text-xs text-muted-foreground">jpg、png、gif、webp</p>
          </div>
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader><CardTitle class="font-serif">{{ t.tagline }}</CardTitle></CardHeader>
      <CardContent>
        <div class="grid gap-1.5">
          <Label>{{ t.tagline }}</Label>
          <Textarea v-model="description" rows="3" placeholder="野有蔓草…" />
        </div>
      </CardContent>
      <CardFooter>
        <Button :disabled="saving || loading" @click="save">
          <SaveIcon class="h-4 w-4 mr-1" />{{ saving ? t.saving : t.saveHomepage }}
        </Button>
      </CardFooter>
    </Card>
  </div>
</template>
