<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  CalendarClock,
  KeyRound,
  RefreshCw,
  Trash2,
  UserPlus,
  Users,
} from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseInput from '../components/ui/BaseInput.vue'
import BaseModal from '../components/ui/BaseModal.vue'
import BaseSearchSelect from '../components/ui/BaseSearchSelect.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { apiErrorMessage } from '../lib/axios'
import { participantService, scheduleService } from '../services/schedule.service'
import { studentService } from '../services/student.service'
import { masterDataService } from '../services/master-data.service'
import type { ExamSchedule } from '../types/api'
import { useUiStore } from '../stores/ui'

const ui = useUiStore()
const route = useRoute()
const router = useRouter()

const examId = route.params.id as string

const loading = ref(true)
const scheduleId = ref<string | null>(null)
const token = ref('')
const startTime = ref('')
const endTime = ref('')
const savingSchedule = ref(false)

const participants = ref<{ id: string; student_id: string; nis: string; name: string; class_name: string | null; assigned_via: string }[]>([])
const classes = ref<{ id: string; name: string }[]>([])
const selectedClasses = ref<Record<string, boolean>>({})
const assigningClass = ref(false)
const students = ref<{ id: string; name: string; nis: string }[]>([])
const selectedStudent = ref('')
const assigningIndividual = ref(false)
const confirmDeleteSchedule = ref(false)
const removeParticipantTarget = ref<{ id: string; name: string } | null>(null)
const removingParticipant = ref(false)

function toLocalInput(iso: string | null | undefined): string {
  if (!iso) return ''
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function toISO(local: string): string {
  return new Date(local).toISOString()
}

onMounted(async () => {
  try {
    const [sched, parts, cls, stu] = await Promise.all([
      scheduleService.getByExam(examId),
      participantService.list(examId),
      masterDataService.classes.list({ page: 1, limit: 100 }),
      studentService.list({ page: 1, limit: 500 }),
    ])
    participants.value = (parts as unknown as typeof participants.value) ?? []
    classes.value = cls.data.map((c) => ({ id: c.id, name: c.name }))
    students.value = stu.data.map((s) => ({
      id: s.id,
      name: s.user?.name ?? s.nis,
      nis: s.nis,
    }))
    if (sched) {
      scheduleId.value = sched.id
      token.value = sched.token
      startTime.value = toLocalInput(sched.start_time)
      endTime.value = toLocalInput(sched.end_time)
    }
  } catch {
    ui.toastError('Gagal memuat jadwal & peserta.')
  } finally {
    loading.value = false
  }
})

const assignedStudentIds = computed(() => new Set(participants.value.map((p) => p.student_id)))
const availableStudents = computed(() =>
  students.value.filter((s) => !assignedStudentIds.value.has(s.id)),
)

const ui2 = ui
void ui2

async function saveSchedule() {
  if (!startTime.value || !endTime.value) {
    ui.toastWarning('Isi waktu mulai dan waktu selesai.')
    return
  }
  savingSchedule.value = true
  try {
    const payload = {
      start_time: toISO(startTime.value),
      end_time: toISO(endTime.value),
      token: token.value,
    }
    let saved: ExamSchedule
    if (scheduleId.value) {
      saved = await scheduleService.update(scheduleId.value, payload)
    } else {
      saved = await scheduleService.create(examId, payload)
      scheduleId.value = saved.id
    }
    token.value = saved.token
    ui.toastSuccess('Jadwal ujian disimpan.')
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal menyimpan jadwal.'))
  } finally {
    savingSchedule.value = false
  }
}

async function generateToken() {
  try {
    if (!scheduleId.value) {
      ui.toastWarning('Simpan jadwal terlebih dahulu.')
      return
    }
    token.value = await scheduleService.generateToken(scheduleId.value)
    ui.toastSuccess(`Token baru: ${token.value}`)
  } catch {
    ui.toastError('Gagal generate token.')
  }
}

async function deleteSchedule() {
  if (!scheduleId.value) return
  try {
    await scheduleService.remove(scheduleId.value)
    scheduleId.value = null
    token.value = ''
    startTime.value = ''
    endTime.value = ''
    ui.toastSuccess('Jadwal dihapus.')
  } catch {
    ui.toastError('Gagal menghapus jadwal.')
  }
}

async function assignClass() {
  const classIds = Object.entries(selectedClasses.value)
    .filter(([, v]) => v)
    .map(([k]) => k)
  if (classIds.length === 0) {
    ui.toastWarning('Pilih minimal satu kelas.')
    return
  }
  assigningClass.value = true
  try {
    const res = await participantService.assignClass(examId, classIds)
    ui.toastSuccess(`${res.assigned} peserta ditambahkan, ${res.skipped} dilewati.`)
    participants.value = await participantService.list(examId)
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal assign kelas.'))
  } finally {
    assigningClass.value = false
  }
}

async function assignIndividual() {
  if (!selectedStudent.value) {
    ui.toastWarning('Pilih siswa terlebih dahulu.')
    return
  }
  assigningIndividual.value = true
  try {
    const res = await participantService.assignIndividual(examId, [selectedStudent.value])
    ui.toastSuccess(`${res.assigned} peserta ditambahkan.`)
    selectedStudent.value = ''
    participants.value = await participantService.list(examId)
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal menambahkan peserta.'))
  } finally {
    assigningIndividual.value = false
  }
}

function askRemoveParticipant(p: { id: string; name: string }) {
  removeParticipantTarget.value = p
}

async function confirmRemoveParticipant() {
  if (!removeParticipantTarget.value) return
  removingParticipant.value = true
  try {
    await participantService.remove(examId, removeParticipantTarget.value.id)
    ui.toastSuccess(`${removeParticipantTarget.value.name} dikeluarkan dari ujian.`)
    participants.value = participants.value.filter(
      (x) => x.id !== removeParticipantTarget.value?.id,
    )
    removeParticipantTarget.value = null
  } catch {
    ui.toastError('Gagal menghapus peserta.')
  } finally {
    removingParticipant.value = false
  }
}
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-4xl space-y-6">
      <div>
        <button class="mb-1 text-xs font-medium text-muted-foreground hover:text-foreground" @click="router.push('/exams')">
          ‹ Ujian
        </button>
        <h1 class="text-2xl font-bold tracking-tight">Jadwal & Peserta Ujian</h1>
      </div>

      <LoadingState v-if="loading" />

      <template v-else>
        <!-- JADWAL -->
        <section class="rounded-xl border border-border bg-card p-5 shadow-sm">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="flex items-center gap-2 font-semibold">
              <CalendarClock class="size-5 text-primary" />
              Jadwal Ujian
            </h2>
            <BaseBadge v-if="token" tone="primary">
              <KeyRound class="mr-1 size-3" /> {{ token }}
            </BaseBadge>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <BaseInput v-model="startTime" label="Waktu Mulai" type="datetime-local" />
            <BaseInput v-model="endTime" label="Waktu Selesai" type="datetime-local" />
          </div>

          <div class="mt-4 flex flex-wrap items-center gap-2">
            <BaseButton @click="saveSchedule" :loading="savingSchedule">Simpan Jadwal</BaseButton>
            <BaseButton variant="outline" @click="generateToken">
              <RefreshCw /> Generate / Reset Token
            </BaseButton>
            <BaseButton v-if="scheduleId" variant="ghost" class="text-destructive" @click="confirmDeleteSchedule = true">
              <Trash2 /> Hapus Jadwal
            </BaseButton>
          </div>
          <p class="mt-3 text-xs text-muted-foreground">
            Token dipakai peserta untuk masuk ujian (Token Protection). Generate ulang membuat token lama tidak berlaku.
          </p>
        </section>

        <!-- PESERTA -->
        <section class="rounded-xl border border-border bg-card p-5 shadow-sm">
          <h2 class="mb-4 flex items-center gap-2 font-semibold">
            <Users class="size-5 text-primary" />
            Peserta ({{ participants.length }})
          </h2>

          <div class="space-y-4">
            <div>
              <p class="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Assign per Kelas</p>
              <div class="flex flex-wrap items-center gap-2">
                <label
                  v-for="c in classes"
                  :key="c.id"
                  class="flex cursor-pointer items-center gap-2 rounded-lg border border-border px-3 py-1.5 text-sm transition-colors hover:bg-accent/50"
                >
                  <input v-model="selectedClasses[c.id]" type="checkbox" class="accent-blue-600" />
                  {{ c.name }}
                </label>
                <BaseButton variant="secondary" size="sm" :loading="assigningClass" @click="assignClass">
                  <UserPlus /> Assign
                </BaseButton>
              </div>
            </div>

            <div class="flex flex-wrap items-end gap-2 border-t border-border pt-4">
              <div class="w-72">
                <BaseSearchSelect
                  v-model="selectedStudent"
                  label="Assign Individual"
                  placeholder="Pilih siswa…"
                  search-placeholder="Cari nama / NIS…"
                  :options="availableStudents.map((s) => ({ value: s.id, label: s.name, hint: s.nis }))"
                  empty-text="Semua siswa sudah jadi peserta"
                />
              </div>
              <BaseButton variant="secondary" :disabled="!selectedStudent" :loading="assigningIndividual" @click="assignIndividual">
                <UserPlus /> Tambah
              </BaseButton>
            </div>
          </div>

          <div class="mt-5 border-t border-border pt-4">
            <EmptyState v-if="participants.length === 0" title="Belum ada peserta" message="Assign lewat kelas atau individual di atas." />
            <ul v-else class="space-y-1.5">
              <li
                v-for="p in participants"
                :key="p.id"
                class="flex items-center justify-between gap-3 rounded-lg border border-border px-3 py-2 text-sm"
              >
                <span class="flex min-w-0 items-center gap-2">
                  <span class="font-mono text-xs text-muted-foreground">{{ p.nis }}</span>
                  <span class="truncate font-medium">{{ p.name }}</span>
                  <BaseBadge v-if="p.class_name" tone="outline">{{ p.class_name }}</BaseBadge>
                </span>
                <span class="flex items-center gap-2">
                  <BaseBadge :tone="p.assigned_via === 'class' ? 'info' : 'neutral'">{{ p.assigned_via }}</BaseBadge>
                  <BaseButton variant="ghost" size="icon" aria-label="Hapus" @click="askRemoveParticipant(p)">
                    <Trash2 class="text-destructive" />
                  </BaseButton>
                </span>
              </li>
            </ul>
          </div>
        </section>
      </template>

      <!-- KONFIRMASI HAPUS JADWAL -->
      <BaseModal :open="confirmDeleteSchedule" title="Hapus Jadwal Ujian?" @close="confirmDeleteSchedule = false">
        <p class="text-sm text-muted-foreground">
          Jadwal ujian akan dihapus (token jadwal ini juga tidak berlaku). Data peserta tidak ikut terhapus.
        </p>
        <template #footer>
          <BaseButton variant="outline" @click="confirmDeleteSchedule = false">Batal</BaseButton>
          <BaseButton variant="destructive" @click="deleteSchedule(); confirmDeleteSchedule = false">
            <Trash2 /> Ya, Hapus Jadwal
          </BaseButton>
        </template>
      </BaseModal>

      <!-- KONFIRMASI KELUARKAN PESERTA -->
      <BaseModal
        :open="!!removeParticipantTarget"
        title="Keluarkan Peserta?"
        @close="removeParticipantTarget = null"
      >
        <p class="text-sm leading-relaxed text-muted-foreground">
          Keluarkan
          <span class="font-semibold text-foreground">{{ removeParticipantTarget?.name }}</span>
          dari ujian ini? Seluruh riwayat attempt & jawabannya pada ujian ini juga akan dihapus.
        </p>
        <template #footer>
          <BaseButton variant="outline" @click="removeParticipantTarget = null">Batal</BaseButton>
          <BaseButton variant="destructive" :loading="removingParticipant" @click="confirmRemoveParticipant">
            <Trash2 /> Ya, Keluarkan
          </BaseButton>
        </template>
      </BaseModal>
    </div>
  </AppShell>
</template>
