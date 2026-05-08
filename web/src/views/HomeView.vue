<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { UploadIcon, SaveIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'

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
  <div class="home-page" style="max-width:640px;display:grid;gap:20px">
    <div class="card">
      <div class="card-header"><h2>{{ t.avatar }}</h2></div>
      <div class="card-body">
        <div class="avatar-area">
          <img v-if="profileImage" :src="profileImage" class="avatar-img" alt="avatar" />
          <div v-else class="avatar-placeholder">🌿</div>
          <label style="cursor:pointer">
            <input ref="avatarFileInput" type="file" accept="image/*" style="display:none" @change="onAvatarSelect" />
            <span class="btn btn-sm" :class="{ 'btn-ghost': !uploadingAvatar }">
              <UploadIcon :size="14" />{{ uploadingAvatar ? t.loading : t.uploadAvatar }}
            </span>
          </label>
          <p class="field-hint">{{ t.avatar }} · jpg、png、gif、webp</p>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-header"><h2>{{ t.tagline }}</h2></div>
      <div class="card-body">
        <div class="field">
          <label class="field-label">{{ t.tagline }}</label>
          <textarea v-model="description" class="textarea" rows="3" :placeholder="'野有蔓草…'" />
        </div>
      </div>
      <div class="card-footer">
        <button class="btn btn-primary" :disabled="saving || loading" @click="save">
          <SaveIcon :size="14" />{{ saving ? t.saving : t.saveHomepage }}
        </button>
      </div>
    </div>
  </div>
</template>
