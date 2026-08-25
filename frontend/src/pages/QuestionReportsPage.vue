<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Check, Flag } from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseModal from '../components/ui/BaseModal.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { useUiStore } from '../stores/ui'

const ui = useUiStore()

interface ReportRow {
  id: string
  exam_title: string
  student_name: string
  nis: string
  question_text: string
  reason: string
  status: string
  resolution?: string | null
  created_at: string
}

const reports = ref<ReportRow[]>([])
const loading = ref(true)
const statusFilter = ref('pending')

const statusOptions = [
  { value: '', label: 'Semua' },
  { value: 'pending', label: 'Pending' },
  { value: 'reviewing', label: 'Reviewing' },
  { value: 'resolved', label: 'Resolved' },
  { value: 'rejected', label: 'Rejected' },
]

async function fetchReports() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (statusFilter.value) params.status = statusFilter.value
    const qs = Object.entries(params).map(([k, v]) => `${k}=${v}`).join('&')
    const res = await fetch(`/api/v1/question-reports?${qs}`, { credentials: 'include' })
    const json = await res.json()
    reports.value = json.data ?? []
  } catch {
    ui.toastError('Gagal memuat laporan.')
  } finally {
    loading.value = false
  }
}
onMounted(fetchReports)

const statusTone = (s: string) =>
  s === 'resolved' ? 'success' : s === 'rejected' ? 'destructive' : s === 'reviewing' ? 'info' : 'neutral'

const resolveModal = ref(false)
const resolveTarget = ref<ReportRow | null>(null)
const resolutionText = ref('')
const resolving = ref(false)

function openResolve(r: ReportRow) {
  resolveTarget.value = r
  resolutionText.value = ''
  resolveModal.value = true
}

async function doResolve(status: string) {
  if (!resolveTarget.value) return
  resolving.value = true
  try {
    await fetch(`/api/v1/question-reports/${resolveTarget.value.id}/resolve`, {
      method: 'PATCH',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status, resolution: resolutionText.value || undefined }),
    })
    ui.toastSuccess('Laporan ditangani.')
    resolveModal.value = false
    await fetchReports()
  } catch {
    ui.toastError('Gagal menangani laporan.')
  } finally {
    resolving.value = false
  }
}
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-4xl space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 class="flex items-center gap-2 text-2xl font-bold tracking-tight">
            <Flag class="size-6 text-amber-500" /> Laporan Soal
          </h1>
          <p class="mt-1 text-sm text-muted-foreground">Kelola laporan soal bermasalah dari siswa.</p>
        </div>
        <div class="w-40">
          <select
            v-model="statusFilter"
            class="h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            @change="fetchReports()"
          >
            <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
          </select>
        </div>
      </div>

      <LoadingState v-if="loading" />

      <template v-else>
        <EmptyState v-if="reports.length === 0" title="Tidak ada laporan" />

        <div v-else class="space-y-4">
          <div
            v-for="r in reports"
            :key="r.id"
            class="rounded-xl border border-border bg-card p-5 shadow-sm"
          >
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div class="min-w-0 flex-1">
                <p class="font-medium">{{ r.question_text }}</p>
                <p class="mt-1 text-xs text-muted-foreground">
                  {{ r.student_name }} ({{ r.nis }}) · {{ r.exam_title }}
                </p>
                <p class="mt-2 text-sm leading-relaxed">
                  <span class="font-medium">Alasan:</span> {{ r.reason }}
                </p>
              </div>
              <BaseBadge :tone="statusTone(r.status)">{{ r.status }}</BaseBadge>
            </div>

            <p v-if="r.resolution" class="mt-2 rounded-lg bg-muted/60 px-3 py-2 text-xs">
              <span class="font-medium">Resolusi:</span> {{ r.resolution }}
            </p>

            <div class="mt-3 flex justify-end gap-1" v-if="r.status === 'pending' || r.status === 'reviewing'">
              <BaseButton variant="outline" size="sm" @click="openResolve(r)">
                <Check /> Tangani
              </BaseButton>
            </div>
          </div>
        </div>
      </template>

      <!-- MODAL RESOLVE -->
      <BaseModal :open="resolveModal" title="Tangani Laporan" @close="resolveModal = false">
        <div class="space-y-4">
          <div class="rounded-lg bg-muted/60 p-3 text-sm">
            <p class="font-medium">{{ resolveTarget?.question_text }}</p>
            <p class="mt-1 text-xs text-muted-foreground">{{ resolveTarget?.student_name }} ({{ resolveTarget?.nis }})</p>
            <p class="mt-1 text-xs">Alasan: {{ resolveTarget?.reason }}</p>
          </div>
          <textarea
            v-model="resolutionText"
            rows="3"
            placeholder="Tulis resolusi / tindak lanjut…"
            class="w-full rounded-lg border border-input bg-transparent p-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>
        <template #footer>
          <BaseButton variant="outline" @click="resolveModal = false">Batal</BaseButton>
          <BaseButton :disabled="!resolveTarget" :loading="resolving" @click="doResolve('resolved')">
            <Check /> Selesaikan
          </BaseButton>
        </template>
      </BaseModal>
    </div>
  </AppShell>
</template>
