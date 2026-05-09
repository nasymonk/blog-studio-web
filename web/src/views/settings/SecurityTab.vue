<script setup lang="ts">
import { ref } from 'vue'
import { LockIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const { t } = useI18n()
const notify = useNotify()
const currentPassword = ref('')
const newPassword = ref('')
const saving = ref(false)
const error = ref('')

async function changePassword() {
  error.value = ''
  if (newPassword.value.length < 6) { error.value = t.value.passwordMinLength; return }
  saving.value = true
  try {
    await api.changePassword(currentPassword.value, newPassword.value)
    notify.success(t.value.passwordChanged)
    currentPassword.value = ''
    newPassword.value = ''
  } catch (e: any) { error.value = e?.message || t.value.passwordChangeFailed }
  finally { saving.value = false }
}
</script>

<template>
  <div class="max-w-md space-y-4">
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2 font-serif text-base">
          <LockIcon class="h-4 w-4 text-accent" /> {{ t.securitySettings }}
        </CardTitle>
        <p class="text-xs text-muted-foreground">{{ t.passwordDesc }}</p>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="grid gap-1.5">
          <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70">{{ t.currentPassword }}</Label>
          <Input v-model="currentPassword" type="password" autocomplete="current-password" />
        </div>
        <div class="grid gap-1.5">
          <Label class="text-[10px] uppercase tracking-wider text-muted-foreground/70">{{ t.newPassword }}</Label>
          <Input v-model="newPassword" type="password" autocomplete="new-password" />
          <p v-if="error" class="text-xs text-destructive">{{ error }}</p>
        </div>
        <Button class="rounded-full px-5" :disabled="saving || !currentPassword || !newPassword" @click="changePassword">
          {{ saving ? t.saving : t.changePassword }}
        </Button>
      </CardContent>
    </Card>
  </div>
</template>
