<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowLeft,
  Send,
  FileSpreadsheet,
  FileText,
} from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseModal from '../components/ui/BaseModal.vue'
import BaseSelect from '../components/ui/BaseSelect.vue'
import BaseTable from '../components/ui/BaseTable.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { api, apiErrorMessage } from '../lib/axios'
import { resultService } from '../services/result.service'
import { examService } from '../services/exam.service'
import { masterDataService } from '../services/master-data.service'
import type { Exam } from '../types/api'
import { useUiStore } from '../stores/ui'

const ui = useUiStore()
const route = useRoute()
const router = useRouter()

const examId = route.params.id as string
const exam = ref<Exam | null>(null)
const loading = ref(true)
const results = ref<
  {
    rank: number
    attempt_id: string
    student_name: string
    nis: string
    class_name: string | null
    score: number | null
    passing_grade: number
    passed: boolean
  }[]
>([])
const classes = ref<{ id: string; name: string }[]>([])
const classFilter = ref('')

const subjectName = computed(() => exam.value?.subject?.name ?? '')

onMounted(async () => {
  try {
    const [examData, cls] = await Promise.all([
      examService.get(examId),
      masterDataService.classes.list({ page: 1, limit: 100 }),
    ])
    exam.value = examData
    classes.value = cls.data.map((c) => ({ id: c.id, name: c.name }))
    await fetchResults()
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memuat data.'))
  } finally {
    loading.value = false
  }
})

async function fetchResults() {
  loading.value = true
  try {
    const res = await resultService.examResults(examId, classFilter.value || undefined)
    results.value = res
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memuat rekap nilai.'))
  } finally {
    loading.value = false
  }
}

const exporting = ref(false)

async function downloadResults(format: 'xlsx' | 'pdf') {
  exporting.value = true
  try {
    const res = await api.get(`/exams/${examId}/export?format=${format}`, { responseType: 'blob' })
    const dispo = (res.headers?.['content-disposition'] as string | undefined) ?? ''
    const match = dispo.match(/filename="([^"]+)"/)
    const name = match?.[1] ?? `hasil-ujian.${format}`
    const url = URL.createObjectURL(res.data as Blob)
    const a = document.createElement('a')
    a.href = url
    a.download = name
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(url)
    ui.toastSuccess(`Berhasil mengunduh ${name}`)
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal mengunduh file.'))
  } finally {
    exporting.value = false
  }
}

const rankIcon = (rank: number) => {
  if (rank === 1) return '🥇'
  if (rank === 2) return '🥈'
  if (rank === 3) return '🥉'
  return `${rank}`
}

const publishModal = ref(false)
const publishing = ref(false)

async function doPublish() {
  publishing.value = true
  try {
    await resultService.publishResults(examId, true)
    ui.toastSuccess('Hasil ujian dipublikasikan ke peserta.')
    publishModal.value = false
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal mempublikasikan hasil.'))
  } finally {
    publishing.value = false
  }
}
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-4xl space-y-6">
      <div>
        <button class="mb-1 flex items-center gap-1 text-xs font-medium text-muted-foreground hover:text-foreground" @click="router.push('/exams')">
          <ArrowLeft class="size-3" /> Ujian
        </button>
        <h1 class="text-2xl font-bold tracking-tight">Rekap Nilai &amp; Ranking</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          {{ exam?.title }} — {{ subjectName }}
        </p>
      </div>

      <LoadingState v-if="loading" />

      <template v-else>
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div class="w-56">
            <BaseSelect
              v-model="classFilter"
              label="Filter Kelas"
              :options="[{ value: '', label: 'Semua kelas' }, ...classes.map((c: { id: string; name: string }) => ({ value: c.id, label: c.name }))]"
              @update:model-value="fetchResults()"
            />
          </div>
          <div class="flex flex-wrap gap-2">
            <BaseButton variant="outline" :disabled="exporting" @click="downloadResults('xlsx')">
              <FileSpreadsheet class="size-4" /> Excel
            </BaseButton>
            <BaseButton variant="outline" :disabled="exporting" @click="downloadResults('pdf')">
              <FileText class="size-4" /> PDF
            </BaseButton>
            <BaseButton variant="outline" @click="publishModal = true">
              <Send /> Publikasikan Hasil
            </BaseButton>
          </div>
        </div>

        <EmptyState v-if="results.length === 0" title="Belum ada peserta yang mengumpulkan" />

        <BaseTable v-else>
          <template #head>
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Rank</th>
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Nama</th>
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">NIS</th>
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Kelas</th>
            <th class="px-4 py-3 text-center text-xs font-semibold uppercase tracking-wide text-muted-foreground">Nilai</th>
            <th class="px-4 py-3 text-center text-xs font-semibold uppercase tracking-wide text-muted-foreground">KKM</th>
            <th class="px-4 py-3 text-center text-xs font-semibold uppercase tracking-wide text-muted-foreground">Status</th>
          </template>
          <tr v-for="r in results" :key="r.attempt_id" class="border-b border-border transition-colors last:border-0 hover:bg-accent/50">
            <td class="px-4 py-3 font-mono text-sm">{{ rankIcon(r.rank) }}</td>
            <td class="px-4 py-3 font-medium">{{ r.student_name }}</td>
            <td class="px-4 py-3 font-mono text-xs text-muted-foreground">{{ r.nis }}</td>
            <td class="px-4 py-3 text-muted-foreground">{{ r.class_name ?? '—' }}</td>
            <td class="px-4 py-3 text-center">
              <span
                :class="[
                  'font-mono text-lg font-bold',
                  r.score !== null && r.score >= r.passing_grade ? 'text-success' : 'text-destructive',
                ]"
              >
                {{ r.score ?? '—' }}
              </span>
            </td>
            <td class="px-4 py-3 text-center text-muted-foreground">{{ r.passing_grade }}</td>
            <td class="px-4 py-3 text-center">
              <BaseBadge :tone="r.passed ? 'success' : 'destructive'">
                {{ r.passed ? 'Lulus' : 'Belum Lulus' }}
              </BaseBadge>
            </td>
          </tr>
        </BaseTable>
      </template>

      <!-- MODAL PUBLISH -->
      <BaseModal :open="publishModal" title="Publikasikan Hasil Ujian?" @close="publishModal = false">
        <p class="text-sm leading-relaxed text-muted-foreground">
          Nilai akan terlihat oleh seluruh peserta yang telah mengumpulkan ujian.
          Pastikan seluruh esai sudah dinilai sebelum mempublikasikan.
        </p>
        <template #footer>
          <BaseButton variant="outline" @click="publishModal = false">Batal</BaseButton>
          <BaseButton :loading="publishing" @click="doPublish">
            <Send /> Ya, Publikasikan
          </BaseButton>
        </template>
      </BaseModal>
    </div>
  </AppShell>
</template>
