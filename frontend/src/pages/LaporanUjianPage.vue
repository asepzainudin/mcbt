<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import {
  Search,
  Calendar,
  ChevronLeft,
  ChevronRight,
} from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseSelect from '../components/ui/BaseSelect.vue'
import BaseTable from '../components/ui/BaseTable.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { apiErrorMessage } from '../lib/axios'
import { resultService, type ExamReportRow } from '../services/result.service'
import { examService } from '../services/exam.service'
import { masterDataService } from '../services/master-data.service'
import { useUiStore } from '../stores/ui'

const ui = useUiStore()

const loading = ref(true)
const rows = ref<ExamReportRow[]>([])
const total = ref(0)
const totalPages = ref(1)
const currentPage = ref(1)
const limit = ref(20)

const exams = ref<{ id: string; title: string }[]>([])
const subjects = ref<{ id: string; name: string }[]>([])
const classes = ref<{ id: string; name: string }[]>([])
const academicYears = ref<{ id: string; year: string; semester: string }[]>([])

const examFilter = ref('')
const subjectFilter = ref('')
const classFilter = ref('')
const ayFilter = ref('')
const dateFrom = ref('')
const dateTo = ref('')

onMounted(async () => {
  try {
    const [examList, subjectList, classList, ayList] = await Promise.all([
      examService.list({ page: 1, limit: 200 }),
      masterDataService.subjects.list({ page: 1, limit: 200 }),
      masterDataService.classes.list({ page: 1, limit: 200 }),
      masterDataService.academicYears.list({ page: 1, limit: 200 }),
    ])
    exams.value = examList.data.map((e) => ({ id: e.id, title: e.title }))
    subjects.value = subjectList.data.map((s) => ({ id: s.id, name: s.name }))
    classes.value = classList.data.map((c) => ({ id: c.id, name: c.name }))
    academicYears.value = ayList.data.map((a) => ({
      id: a.id,
      year: a.year,
      semester: a.semester === 'ODD' ? 'Ganjil' : 'Genap',
    }))
    await fetchReport()
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memuat data filter.'))
  } finally {
    loading.value = false
  }
})

async function fetchReport() {
  loading.value = true
  try {
    const res = await resultService.examReport({
      page: currentPage.value,
      limit: limit.value,
      exam_id: examFilter.value || undefined,
      subject_id: subjectFilter.value || undefined,
      class_id: classFilter.value || undefined,
      academic_year_id: ayFilter.value || undefined,
      date_from: dateFrom.value || undefined,
      date_to: dateTo.value || undefined,
    })
    rows.value = res.items
    total.value = res.total
    totalPages.value = res.total_pages
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memuat laporan ujian.'))
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  currentPage.value = 1
  fetchReport()
}

function resetFilters() {
  examFilter.value = ''
  subjectFilter.value = ''
  classFilter.value = ''
  ayFilter.value = ''
  dateFrom.value = ''
  dateTo.value = ''
  currentPage.value = 1
  fetchReport()
}

function fmtDate(raw: string | null | undefined): string {
  if (!raw) return '—'
  return new Date(raw).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function fmtDateTime(raw: string | null | undefined): string {
  if (!raw) return '—'
  return new Date(raw).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-7xl space-y-6">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Laporan Ujian</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          Rekap seluruh hasil ujian siswa dari berbagai ujian.
        </p>
      </div>

      <!-- FILTERS -->
      <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
          <BaseSelect
            v-model="examFilter"
            label="Ujian"
            :options="[{ value: '', label: 'Semua ujian' }, ...exams.map((e) => ({ value: e.id, label: e.title }))]"
          />
          <BaseSelect
            v-model="subjectFilter"
            label="Mata Pelajaran"
            :options="[{ value: '', label: 'Semua mapel' }, ...subjects.map((s) => ({ value: s.id, label: s.name }))]"
          />
          <BaseSelect
            v-model="classFilter"
            label="Kelas"
            :options="[{ value: '', label: 'Semua kelas' }, ...classes.map((c) => ({ value: c.id, label: c.name }))]"
          />
          <BaseSelect
            v-model="ayFilter"
            label="Tahun Ajaran"
            :options="[{ value: '', label: 'Semua tahun' }, ...academicYears.map((a) => ({ value: a.id, label: `${a.year} ${a.semester}` }))]"
          />
          <div class="space-y-1.5">
            <label class="text-sm font-medium leading-none text-foreground">Dari Tanggal</label>
            <input
              v-model="dateFrom"
              type="date"
              class="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
            />
          </div>
          <div class="space-y-1.5">
            <label class="text-sm font-medium leading-none text-foreground">Sampai Tanggal</label>
            <input
              v-model="dateTo"
              type="date"
              class="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
            />
          </div>
        </div>
        <div class="mt-4 flex gap-2">
          <BaseButton size="sm" @click="applyFilters">
            <Search class="size-4" /> Terapkan Filter
          </BaseButton>
          <BaseButton variant="outline" size="sm" @click="resetFilters">
            Reset
          </BaseButton>
        </div>
      </div>

      <!-- TABLE -->
      <LoadingState v-if="loading" />

      <template v-else>
        <div v-if="!rows || rows.length === 0">
          <EmptyState title="Belum ada data laporan ujian" />
        </div>

        <template v-else>
          <BaseTable>
            <template #head>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">No</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Nama Siswa</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">NIS</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Kelas</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Ujian</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Mata Pelajaran</th>
              <th class="px-4 py-3 text-center text-xs font-semibold uppercase tracking-wide text-muted-foreground">Nilai</th>
              <th class="px-4 py-3 text-center text-xs font-semibold uppercase tracking-wide text-muted-foreground">KKM</th>
              <th class="px-4 py-3 text-center text-xs font-semibold uppercase tracking-wide text-muted-foreground">Status</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Tanggal Selesai</th>
            </template>
            <tr
              v-for="(r, idx) in rows"
              :key="r.attempt_id"
              class="border-b border-border transition-colors last:border-0 hover:bg-accent/50"
            >
              <td class="px-4 py-3 font-mono text-sm text-muted-foreground">
                {{ (currentPage - 1) * limit + idx + 1 }}
              </td>
              <td class="px-4 py-3 font-medium">{{ r.student_name }}</td>
              <td class="px-4 py-3 font-mono text-xs text-muted-foreground">{{ r.nis }}</td>
              <td class="px-4 py-3 text-muted-foreground">{{ r.class_name ?? '—' }}</td>
              <td class="px-4 py-3">{{ r.exam_title }}</td>
              <td class="px-4 py-3 text-muted-foreground">{{ r.subject_name }}</td>
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
                  {{ r.passed ? 'Lulus' : 'Tidak Lulus' }}
                </BaseBadge>
              </td>
              <td class="px-4 py-3 text-sm text-muted-foreground">{{ fmtDateTime(r.submitted_at) }}</td>
            </tr>
          </BaseTable>

          <!-- PAGINATION -->
          <div v-if="totalPages && totalPages > 1" class="flex items-center justify-between">
            <p class="text-sm text-muted-foreground">
              Total <span class="font-semibold text-foreground">{{ total }}</span> data
            </p>
            <div class="flex items-center gap-2">
              <BaseButton
                variant="outline"
                size="sm"
                :disabled="currentPage <= 1"
                @click="currentPage--; fetchReport()"
              >
                <ChevronLeft class="size-4" />
              </BaseButton>
              <span class="text-sm text-muted-foreground">
                {{ currentPage }} / {{ totalPages }}
              </span>
              <BaseButton
                variant="outline"
                size="sm"
                :disabled="currentPage >= totalPages"
                @click="currentPage++; fetchReport()"
              >
                <ChevronRight class="size-4" />
              </BaseButton>
            </div>
          </div>
        </template>
      </template>
    </div>
  </AppShell>
</template>
