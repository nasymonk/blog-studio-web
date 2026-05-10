<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { UploadIcon, SaveIcon, HomeIcon, UserIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

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
  } catch (e: any) { notify.error(e) }
  finally { loading.value = false }
})

async function save() {
  saving.value = true
  try {
    await api.saveSite({ description: description.value, profileImage: profileImage.value })
    notify.success(t.value.saveHomepage)
  } catch (e: any) { notify.error(e) }
  finally { saving.value = false }
}

async function onAvatarSelect(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  uploadingAvatar.value = true
  try {
    const result = await api.uploadAvatar(file)
    profileImage.value = result.path
    notify.success(t.value.uploadAvatar)
  } catch (e: any) { notify.error(e) }
  finally { uploadingAvatar.value = false }
}
</script>

<template>
  <div class="max-w-xl space-y-5">
    <div>
      <h1 class="flex items-center gap-2 font-serif text-lg font-semibold">
        <HomeIcon class="h-4 w-4" /> {{ t.home }}
      </h1>
      <p class="text-xs text-muted-foreground mt-0.5">{{ t.homeSubtitle }}</p>
    </div>

    <div class="stagger space-y-5">
      <!-- Avatar -->
      <div class="bg-muted/30 rounded-lg p-4 space-y-3 animate-fade-up">
        <p class="text-sm font-medium">{{ t.avatar }}</p>
        <div class="flex items-center gap-5">
          <div class="h-20 w-20 rounded-full bg-muted flex items-center justify-center overflow-hidden shrink-0">
            <img v-if="profileImage" :src="profileImage" class="h-full w-full object-cover" alt="avatar" />
            <UserIcon v-else class="h-8 w-8 text-muted-foreground/40" />
          </div>
          <div class="space-y-1.5">
            <label class="cursor-pointer inline-block">
              <input ref="avatarFileInput" type="file" accept="image/*" class="hidden" @change="onAvatarSelect" />
              <Button variant="ghost" size="sm" class="text-xs h-7" as-child>
                <span><UploadIcon class="h-3 w-3 mr-1" />{{ uploadingAvatar ? t.loading : t.uploadAvatar }}</span>
              </Button>
            </label>
            <p class="text-[11px] text-muted-foreground/60">{{ t.avatarFormats }}</p>
          </div>
        </div>
      </div>

      <!-- Tagline -->
      <div class="bg-muted/30 rounded-lg p-4 space-y-3 animate-fade-up">
        <p class="text-sm font-medium">{{ t.tagline }}</p>
        <div class="grid gap-1.5">
          <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70">{{ t.tagline }}</Label>
          <Textarea
            v-model="description"
            class="font-deco min-h-[80px] bg-transparent"
            rows="3"
            placeholder="野有蔓草…"
          />
        </div>
        <Button size="sm" class="rounded-full px-4" :disabled="saving || loading" @click="save">
          <SaveIcon class="h-3.5 w-3.5 mr-1" />{{ saving ? t.saving : t.saveHomepage }}
        </Button>
      </div>
    </div>
  </div>
</template>
