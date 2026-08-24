<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  Pencil,
  CalendarClock,
  ListTree,
  Plus,
  Search,
  Settings2,
  Trash2,
} from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseInput from '../components/ui/BaseInput.vue'
import BaseModal from '../components/ui/BaseModal.vue'
import BasePagination from '../components/ui/BasePagination.vue'
import BaseSelect from '../components/ui/BaseSelect.vue'
import BaseSwitch from '../components/ui/BaseSwitch.vue'
import BaseTable from '../components/ui/BaseTable.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { useCrudTable } from '../composables/useCrudTable'
import { examService } from '../services/exam.service'
import { masterDataService } from '../services/master-data.service'
import type { Exam, ExamSettingsPayload } from '../types/api'
import { useUiStore } from '../stores/ui'
import { apiErrorMessage } from '../lib/axios'

const router = useRouter()
const ui = useUiStore()

const subjects = ref<{ id: string; name: string; code: string }[]>([])
const academicYears = ref<{ id: string; year: string; semester: string }[]>([])

onMounted(async () => {
  const [s, ay] = await Promise.all([
    masterDataService.subjects.list({ page: 1, limit: 100 }),
    masterDataService.academicYears.list({ page: 1, limit: 100 }),
  ])
  subjects.value = s.data
  academicYears.value = ay.data
})

const subjectOptions = computed(() => [
  { value: '', label: '— Pilih mapel —', disabled: true },
  ...subjects.value.map((s) => ({ value: s.id, label: `${s.code} — ${s.name}` })),
])

const filterSubject = ref('')
const filterStatus = ref('')

const statusOptions = [
  { value: '', label: 'Semua status' },
  { value: 'draft', label: 'Draft' },
  { value: 'published', label: 'Published' },
  { value: 'closed', label: 'Closed' },
]

const extraParams = computed(() => ({
  subject_id: filterSubject.value || undefined,
  status: filterStatus.value || undefined,
}))

const crud = useCrudTable<Exam>({
  itemLabel: 'Ujian',
  extraParams,
  listFn: (p) => examService.list(p),
  createFn: (p) => examService.create(p as never),
  updateFn: (id, p) => examService.update(id, p as never),
  removeFn: (id) => examService.remove(id),
  toPayload: (f) => ({
    title: f.title,
    subject_id: f.subject_id,
    academic_year_id: f.academic_year_id || null,
    question_bank_id: f.question_bank_id || null,
    description: f.description || null,
  }),
  fromItem: (e) => ({
    title: e.title,
    subject_id: e.subject_id,
    academic_year_id: e.academic_year_id ?? '',
    question_bank_id: e.question_bank_id ?? '',
    description: e.description ?? '',
  }),
  validate: (f) => {
    const e: Record<string, string> = {}
    if (!f.title.trim()) e.title = 'Judul wajib diisi'
    if (!f.subject_id) e.subject_id = 'Pilih mata pelajaran'
    return e
  },
})

watch([filterSubject, filterStatus], () => {
  crud.page.value = 1
  crud.fetchList()
})

const statusTone = (s: string) =>
  s === 'published' ? 'success' : s === 'closed' ? 'warning' : 'neutral'

// ---- modal settings ----
const settingsOpen = ref(false)
const settingsExam = ref<Exam | null>(null)
const savingSettings = ref(false)
const settingsErrors = ref<Record<string, string>>({})

const st = ref({
  duration_minutes: '60',
  max_attempts: '1',
  passing_grade: '75',
  randomize_questions: false,
  randomize_options: false,
  allow_backtrack: true,
  auto_submit: true,
  show_result_immediately: false,
  negative_marking: false,
  negative_value: '0',
  token_enabled: false,
})

function openSettings(e: Exam) {
  settingsExam.value = e
  st.value = {
    duration_minutes: String(e.duration_minutes),
    max_attempts: String(e.max_attempts),
    passing_grade: String(e.passing_grade),
    randomize_questions: e.randomize_questions,
    randomize_options: e.randomize_options,
    allow_backtrack: e.allow_backtrack,
    auto_submit: e.auto_submit,
    show_result_immediately: e.show_result_immediately,
    negative_marking: e.negative_marking,
    negative_value: String(e.negative_value),
    token_enabled: e.token_enabled,
  }
  settingsErrors.value = {}
  settingsOpen.value = true
}

async function saveSettings() {
  if (!settingsExam.value) return
  settingsErrors.value = {}
  savingSettings.value = true
  try {
    const payload: ExamSettingsPayload = {
      ...st.value,
      duration_minutes: Number(st.value.duration_minutes),
      max_attempts: Number(st.value.max_attempts),
      passing_grade: Number(st.value.passing_grade),
      negative_value: Number(st.value.negative_value),
    }
    const updated = await examService.updateSettings(settingsExam.value.id, payload)
    ui.toastSuccess('Pengaturan ujian disimpan.')
    if (updated.exam_token) {
      ui.toastInfo(`Token ujian: ${updated.exam_token}`)
    }
    settingsOpen.value = false
    await crud.fetchList()
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal menyimpan pengaturan.'))
  } finally {
    savingSettings.value = false
  }
}
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-5xl space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Ujian</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Buat ujian dan atur durasi, attempt, randomisasi, hingga token akses.
          </p>
        </div>
        <div class="flex flex-wrap items-end gap-2">
          <div class="w-44">
            <BaseSelect v-model="filterSubject" :options="[{ value: '', label: 'Semua mapel' }, ...subjects.map((s) => ({ value: s.id, label: s.code }))]" />
          </div>
          <div class="w-36">
            <BaseSelect v-model="filterStatus" :options="statusOptions" />
          </div>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <input
              v-model="crud.searchInput.value"
              placeholder="Cari judul…"
              class="h-9 w-44 rounded-lg border border-input bg-transparent pl-9 pr-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
          <BaseButton @click="crud.openCreate()"><Plus /> Tambah</BaseButton>
        </div>
      </div>

      <LoadingState v-if="crud.loading.value" />

      <template v-else>
        <EmptyState v-if="crud.items.value.length === 0" title="Belum ada ujian" message="Buat ujian pertama Anda." />

        <template v-else>
          <BaseTable>
            <template #head>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Ujian</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Status</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Durasi</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">KKM</th>
              <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-muted-foreground">Aksi</th>
            </template>

            <tr v-for="e in crud.items.value" :key="e.id" class="border-b border-border transition-colors last:border-0 hover:bg-accent/50">
              <td class="px-4 py-3">
                <p class="font-medium">{{ e.title }}</p>
                <p class="mt-0.5 text-xs text-muted-foreground">
                  {{ e.subject?.code }} · {{ e.question_bank?.title ?? 'tanpa bank' }}
                </p>
              </td>
              <td class="px-4 py-3">
                <BaseBadge :tone="statusTone(e.status)">{{ e.status }}</BaseBadge>
              </td>
              <td class="px-4 py-3 text-sm text-muted-foreground">{{ e.duration_minutes }} mnt</td>
              <td class="px-4 py-3 text-sm text-muted-foreground">{{ e.passing_grade }}</td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <BaseButton variant="ghost" size="icon" title="Jadwal & Peserta" @click="router.push(`/exams/${e.id}/schedule`)">
                    <CalendarClock />
                  </BaseButton>
                  <BaseButton variant="ghost" size="icon" title="Sections" @click="router.push(`/exams/${e.id}/sections`)">
                    <ListTree />
                  </BaseButton>
                  <BaseButton variant="ghost" size="icon" title="Pengaturan" @click="openSettings(e)">
                    <Settings2 />
                  </BaseButton>
                  <BaseButton variant="ghost" size="icon" title="Edit" @click="crud.openEdit(e)">
                    <Pencil />
                  </BaseButton>
                  <BaseButton variant="ghost" size="icon" title="Hapus" @click="crud.askDelete(e)">
                    <Trash2 class="text-destructive" />
                  </BaseButton>
                </div>
              </td>
            </tr>
          </BaseTable>

          <div class="flex flex-col items-center gap-3 sm:flex-row sm:justify-between">
            <p class="text-sm text-muted-foreground">{{ crud.rangeText.value }}</p>
            <BasePagination v-if="crud.meta.value" :page="crud.meta.value.page" :total-pages="crud.totalPages.value" @change="crud.goTo" />
          </div>
        </template>
      </template>

      <!-- FORM CORE -->
      <BaseModal :open="crud.formOpen.value" :title="crud.isEditing.value ? 'Edit Ujian' : 'Tambah Ujian'" @close="crud.closeForm()">
        <form class="space-y-4" @submit.prevent="crud.submit()">
          <BaseInput v-model="crud.form.value.title" label="Judul Ujian" placeholder="UAS Matematika" required :error="crud.fieldErrors.value.title" />
          <BaseSelect v-model="crud.form.value.subject_id" label="Mata Pelajaran" :options="subjectOptions" required :error="crud.fieldErrors.value.subject_id" />
          <BaseSelect
            v-model="crud.form.value.academic_year_id"
            label="Tahun Ajaran (opsional)"
            :options="[{ value: '', label: '—' }, ...academicYears.map((ay) => ({ value: ay.id, label: `${ay.year} ${ay.semester === 'ODD' ? 'Ganjil' : 'Genap'}` }))]"
            :error="crud.fieldErrors.value.academic_year_id"
          />
          <BaseInput v-model="crud.form.value.description" label="Deskripsi (opsional)" :error="crud.fieldErrors.value.description" />
          <div class="flex justify-end gap-2 pt-2">
            <BaseButton variant="outline" type="button" @click="crud.closeForm()">Batal</BaseButton>
            <BaseButton type="submit" :loading="crud.saving.value">Simpan</BaseButton>
          </div>
        </form>
      </BaseModal>

      <!-- SETTINGS -->
      <BaseModal :open="settingsOpen" title="Pengaturan Ujian" @close="settingsOpen = false">
        <form class="max-h-[70vh] space-y-5 overflow-y-auto pr-1" @submit.prevent="saveSettings">
          <section class="grid grid-cols-3 gap-3">
            <BaseInput v-model="st.duration_minutes" label="Durasi (menit)" type="number" min="1" max="600" :error="settingsErrors.duration_minutes" />
            <BaseInput v-model="st.max_attempts" label="Max Attempt" type="number" min="1" max="10" :error="settingsErrors.max_attempts" />
            <BaseInput v-model="st.passing_grade" label="KKM (0–100)" type="number" step="0.1" :error="settingsErrors.passing_grade" />
          </section>

          <div class="space-y-3 rounded-xl border border-border p-4">
            <BaseSwitch v-model="st.randomize_questions" label="Acak Soal" description="Urutan soal dirandom untuk tiap peserta" />
            <BaseSwitch v-model="st.randomize_options" label="Acak Pilihan Jawaban" description="Urutan opsi dirandom untuk tiap peserta" />
            <BaseSwitch v-model="st.allow_backtrack" label="Izinkan Backtrack" description="Peserta dapat kembali ke soal sebelumnya" />
            <BaseSwitch v-model="st.auto_submit" label="Auto Submit" description="Jawaban dikirim otomatis saat waktu habis" />
            <BaseSwitch v-model="st.show_result_immediately" label="Tampilkan Hasil Langsung" description="Nilai tampil segera setelah submit" />
          </div>

          <div class="space-y-3 rounded-xl border border-border p-4">
            <BaseSwitch v-model="st.negative_marking" label="Negative Marking" description="Salah jawaban mengurangi nilai" />
            <BaseInput
              v-if="st.negative_marking"
              v-model="st.negative_value"
              label="Pengurangan per jawaban salah"
              type="number"
              step="0.25"
              min="0"
              max="100"
            />
          </div>

          <div class="space-y-3 rounded-xl border border-border p-4">
            <BaseSwitch v-model="st.token_enabled" label="Token Protection" description="Peserta wajib memasukkan token untuk memulai ujian" />
            <p v-if="st.token_enabled && settingsExam?.exam_token" class="rounded-lg bg-primary/5 px-3 py-2 text-sm">
              Token saat ini:
              <code class="ml-1 font-mono text-base font-bold tracking-widest text-primary">{{ settingsExam.exam_token }}</code>
              <span class="ml-2 text-xs text-muted-foreground">(disimpan ulang bila tetap aktif)</span>
            </p>
          </div>

          <div class="flex justify-end gap-2 pt-2">
            <BaseButton variant="outline" type="button" @click="settingsOpen = false">Batal</BaseButton>
            <BaseButton type="submit" :loading="savingSettings">Simpan Pengaturan</BaseButton>
          </div>
        </form>
      </BaseModal>

      <BaseModal :open="!!crud.deleteTarget.value" title="Konfirmasi Hapus" @close="crud.cancelDelete()">
        <p class="text-sm text-muted-foreground">
          Hapus ujian <span class="font-semibold text-foreground">{{ crud.deleteTarget.value?.title }}</span>?
        </p>
        <template #footer>
          <BaseButton variant="outline" @click="crud.cancelDelete()">Batal</BaseButton>
          <BaseButton variant="destructive" :loading="crud.deleting.value" @click="crud.confirmDelete()">
            <Trash2 /> Hapus
          </BaseButton>
        </template>
      </BaseModal>
    </div>
  </AppShell>
</template>
