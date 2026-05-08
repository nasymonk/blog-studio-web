<script setup lang="ts">
import { ref, watch, nextTick, onMounted } from 'vue'
import { RefreshCwIcon, XIcon } from 'lucide-vue-next'
import { api } from '@/services/api'
import type { AuditEntry } from '@/services/api'
import { useI18n } from '@/i18n'
import { useNotify } from '@/composables/useNotify'

const { t } = useI18n()
const notify = useNotify()
const entries = ref<AuditEntry[]>([])
const loading = ref(false)
const selected = ref<AuditEntry | null>(null)
const drawerRef = ref<HTMLElement | null>(null)
let triggerEl: HTMLElement | null = null

const FOCUSABLE = 'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'

function trapFocus(e: KeyboardEvent) {
  if (!drawerRef.value) return
  const els = Array.from(drawerRef.value.querySelectorAll<HTMLElement>(FOCUSABLE)).filter(el => !(el as HTMLButtonElement).disabled)
  if (!els.length) return
  const first = els[0], last = els[els.length - 1]
  if (e.key === 'Tab') {
    if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last.focus() }
    else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first.focus() }
  }
  if (e.key === 'Escape') closeDrawer()
}

function openDrawer(entry: AuditEntry) {
  triggerEl = document.activeElement as HTMLElement
  selected.value = entry
}

function closeDrawer() {
  selected.value = null
  nextTick(() => triggerEl?.focus())
}

watch(selected, async (val) => {
  if (val) {
    await nextTick()
    const first = drawerRef.value?.querySelector<HTMLElement>(FOCUSABLE)
    first?.focus()
  }
})

function fmtTime(iso: string) {
  if (!iso) return ''
  try { return new Date(iso).toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }) }
  catch { return iso }
}

async function load() {
  loading.value = true
  try { entries.value = await api.audit(100) }
  catch (e: any) { notify.error(e) }
  finally { loading.value = false }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="flex items-center gap-2 mb-4">
      <h2 style="margin:0">{{ t.auditLog }}</h2>
      <button class="btn btn-ghost btn-icon btn-sm" :disabled="loading" @click="load">
        <RefreshCwIcon :size="14" :class="{ spin: loading }" />
      </button>
    </div>

    <div v-if="loading" class="skeleton" style="height:300px"></div>
    <div v-else-if="entries.length === 0" class="empty-state"><p>暂无审计记录</p></div>
    <table v-else class="audit-table">
      <thead>
        <tr>
          <th>{{ t.time }}</th>
          <th>Slug</th>
          <th>{{ t.op }}</th>
          <th>{{ t.result }}</th>
          <th>Build</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="e in entries" :key="e.auditId" tabindex="0" style="cursor:pointer" @click="openDrawer(e)" @keydown.enter="openDrawer(e)">
          <td>{{ fmtTime(e.timestamp) }}</td>
          <td class="truncate" style="max-width:160px">{{ e.slug }}</td>
          <td>{{ e.operation }}</td>
          <td>
            <span class="badge" :class="e.result === 'success' ? 'badge-ok' : 'badge-error'">{{ e.result }}</span>
          </td>
          <td>
            <span v-if="e.buildResult" class="badge" :class="e.buildResult.success ? 'badge-ok' : 'badge-error'">
              {{ e.buildResult.exitCode }}
            </span>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Drawer -->
    <Teleport to="body">
      <template v-if="selected">
        <div class="drawer-overlay" aria-hidden="true" @click="closeDrawer"></div>
        <div
          ref="drawerRef"
          class="drawer"
          role="dialog"
          aria-modal="true"
          aria-labelledby="drawer-title"
          @keydown="trapFocus"
        >
          <div class="drawer-header">
            <h3 id="drawer-title">{{ selected.slug }} · {{ selected.operation }}</h3>
            <button class="btn btn-ghost btn-icon btn-sm" aria-label="关闭" @click="closeDrawer"><XIcon :size="16" aria-hidden="true" /></button>
          </div>
          <div class="drawer-body">
            <dl style="display:grid;gap:10px;grid-template-columns:130px 1fr">
              <dt class="text-muted text-sm">时间</dt><dd>{{ fmtTime(selected.timestamp) }}</dd>
              <dt class="text-muted text-sm">审计 ID</dt><dd style="font-family:monospace;font-size:12px">{{ selected.auditId }}</dd>
              <dt class="text-muted text-sm">操作</dt><dd>{{ selected.operation }}</dd>
              <dt class="text-muted text-sm">结果</dt><dd><span class="badge" :class="selected.result === 'success' ? 'badge-ok' : 'badge-error'">{{ selected.result }}</span></dd>
              <dt class="text-muted text-sm">备份 ID</dt><dd>{{ selected.backupId || '—' }}</dd>
              <dt class="text-muted text-sm">错误码</dt><dd>{{ selected.errorCode || '—' }}</dd>
              <dt class="text-muted text-sm">错误信息</dt><dd>{{ selected.errorBrief || '—' }}</dd>
            </dl>
            <div v-if="selected.buildResult" style="margin-top:16px">
              <p class="field-label">构建日志</p>
              <pre class="pre" style="background:var(--entry-strong);padding:12px;border-radius:6px;max-height:300px;overflow:auto">{{ selected.buildResult.stderr || selected.buildResult.stdout || '（空）' }}</pre>
            </div>
          </div>
        </div>
      </template>
    </Teleport>
  </div>
</template>
