<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  CalendarClock,
  Clock,
  FileEdit,
  PlayCircle,
  RefreshCw,
  Flag,
} from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseInput from '../components/ui/BaseInput.vue'
import BaseModal from '../components/ui/BaseModal.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { apiErrorMessage } from '../lib/axios'
import { candidateService } from '../services/candidate.service'
import type { CandidateExam } from '../types/api'
import { useUiStore } from '../stores/ui'

const ui = useUiStore()
const router = useRouter()

const exams = ref<CandidateExam[]>([])
const loading = ref(true)

const now = ref(Date.now())
let clockTimer: ReturnType<typeof setInterval> | undefined
onMounted(() => {
  clockTimer = setInterval(() => (now.value = Date.now()), 1000)
})

import { onUnmounted } from 'vue'
onUnmounted(() => clearInterval(clockTimer))

async function fetchExams() {
  loading.value = true
  try {
    exams.value = await candidateService.listExams()
  } catch {
    ui.toastError('Gagal memuat daftar ujian.')
  } finally {
    loading.value = false
  }
}
onMounted(fetchExams)

function fmt(dt: string) {
  return new Date(dt).toLocaleString('id-ID', {
    day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit',
  })
}

const countdown = (expiresAt: string) => {
  const diff = new Date(expiresAt).getTime() - now.value
  if (diff <= 0) return '00:00:00'
  const h = Math.floor(diff / 3_600_000)
  const m = Math.floor((diff % 3_600_000) / 60_000)
  const s = Math.floor((diff % 60_000) / 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(h)}:${pad(m)}:${pad(s)}`
}

function hasPendingEssay(e: CandidateExam): boolean {
  return e.has_essay === true && e.essay_ungraded === true
}

const stateOf = (e: CandidateExam): 'active' | 'available' | 'done' | 'upcoming' => {
  if (e.active_attempt_id) return 'active'
  if (e.last_status === 'submitted') return 'done'
  if (e.attempts_used >= e.max_attempts) return 'done'
  if (new Date(e.start_time).getTime() > Date.now()) return 'upcoming'
  return 'available'
}

// ---- modal token ----
const tokenModalOpen = ref(false)
const tokenExam = ref<CandidateExam | null>(null)
const tokenInput = ref('')
const validating = ref(false)
const fieldError = ref('')

function openTokenModal(e: CandidateExam) {
  tokenExam.value = e
  tokenInput.value = ''
  fieldError.value = ''
  tokenModalOpen.value = true
}

async function validateThenStart() {
  if (!tokenExam.value) return
  validating.value = true
  fieldError.value = ''
  try {
    await candidateService.validateToken(tokenExam.value.exam_id, tokenInput.value)
    const attempt = await candidateService.start(tokenExam.value.exam_id, tokenInput.value)
    ui.toastSuccess(`Attempt #${attempt.attempt_no} dimulai — selamat mengerjakan!`)
    router.push(`/candidate/attempts/${attempt.attempt_id}`)
  } catch (err) {
    fieldError.value = apiErrorMessage(err, 'Gagal memulai ujian.')
  } finally {
    validating.value = false
  }
}

// tanpa token protection → langsung start
async function startWithoutToken(e: CandidateExam) {
  try {
    const attempt = await candidateService.start(e.exam_id)
    ui.toastSuccess(`Attempt #${attempt.attempt_no} dimulai — selamat mengerjakan!`)
    router.push(`/candidate/attempts/${attempt.attempt_id}`)
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memulai ujian.'))
  }
}

const grouped = computed(() => ({
  active: exams.value.filter((e) => stateOf(e) === 'active'),
  available: exams.value.filter((e) => stateOf(e) === 'available'),
  rest: exams.value.filter((s) => ['done', 'upcoming'].includes(stateOf(s))),
}))
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-4xl space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Ujian Saya</h1>
          <p class="mt-1 text-sm text-muted-foreground">Daftar ujian yang diikuti & jadwalnya.</p>
        </div>
        <BaseButton variant="outline" size="icon" aria-label="Refresh" @click="fetchExams">
          <RefreshCw />
        </BaseButton>
      </div>

      <LoadingState v-if="loading" />

      <template v-else>
        <EmptyState
          v-if="exams.length === 0"
          title="Belum ada ujian"
          message="Ujian akan muncul di sini setelah Anda ditambahkan sebagai peserta."
        />

        <template v-else>
          <!-- SEDANG MENGERJAKAN -->
          <div v-if="grouped.active.length" class="space-y-3">
            <p class="text-xs font-semibold uppercase tracking-wide text-destructive">Sedang Berlangsung</p>
            <div
              v-for="e in grouped.active"
              :key="e.exam_id"
              class="rounded-xl border-2 border-destructive/40 bg-card p-5 shadow-sm"
            >
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <h3 class="font-semibold">{{ e.title }}</h3>
                  <p class="mt-0.5 text-xs text-muted-foreground">{{ e.subject_code }} — {{ e.subject_name }}</p>
                </div>
                <div class="text-right">
                  <p class="font-mono text-2xl font-bold text-destructive">{{ countdown(e.active_expires_at!) }}</p>
                  <p class="text-[10px] uppercase tracking-wide text-muted-foreground">sisa waktu</p>
                </div>
              </div>
              <BaseButton class="mt-4 w-full sm:w-auto" @click="router.push(`/candidate/attempts/${e.active_attempt_id}`)">
                <PlayCircle /> Lanjutkan Mengerjakan
              </BaseButton>
            </div>
          </div>

          <!-- TERSEDIA -->
          <div v-if="grouped.available.length" class="space-y-3">
            <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Tersedia</p>
            <div
              v-for="e in grouped.available"
              :key="e.exam_id"
              class="rounded-xl border border-border bg-card p-5 shadow-sm"
            >
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div class="min-w-0">
                  <h3 class="font-semibold">{{ e.title }}</h3>
                  <p class="mt-0.5 text-xs text-muted-foreground">{{ e.subject_code }} — {{ e.subject_name }}</p>
                  <div class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                    <span class="flex items-center gap-1"><Clock class="size-3.5" /> {{ e.duration_minutes }} menit</span>
                    <span class="flex items-center gap-1"><CalendarClock class="size-3.5" /> {{ fmt(e.start_time) }} → {{ fmt(e.end_time) }}</span>
                    <span>Attempt: {{ e.attempts_used }}/{{ e.max_attempts }}</span>
                    <span>KKM: {{ e.passing_grade }}</span>
                  </div>
                </div>
                <BaseButton
                  v-if="e.token_enabled"
                  @click="openTokenModal(e)"
                >
                  <PlayCircle /> Masukkan Token
                </BaseButton>
                <BaseButton v-else @click="startWithoutToken(e)">
                  <PlayCircle /> Mulai
                </BaseButton>
              </div>
            </div>
          </div>

          <!-- LAINNYA -->
          <div v-if="grouped.rest.length" class="space-y-3">
            <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Lainnya</p>
            <div
              v-for="e in grouped.rest"
              :key="e.exam_id"
              class="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border bg-card p-4 shadow-sm opacity-80"
            >
              <div>
                <h3 class="font-medium">{{ e.title }}</h3>
                <p class="mt-0.5 text-xs text-muted-foreground">
                  {{ fmt(e.start_time) }} → {{ fmt(e.end_time) }} · attempt {{ e.attempts_used }}/{{ e.max_attempts }}
                </p>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <template v-if="e.show_result_immediately !== false && e.score !== null && e.last_status === 'submitted'">
                  <span class="text-sm text-muted-foreground">Nilai:</span>
                  <span
                    :class="[
                      'font-mono text-lg font-bold',
                      e.score >= e.passing_grade ? 'text-success' : 'text-destructive',
                    ]"
                  >{{ e.score }}</span>
                  <BaseBadge :tone="e.score >= e.passing_grade ? 'success' : 'destructive'">
                    {{ e.score >= e.passing_grade ? 'Lulus' : 'Belum Lulus' }}
                  </BaseBadge>
                </template>
                <BaseBadge :tone="stateOf(e) === 'done' ? 'success' : 'neutral'">
                  {{ stateOf(e) === 'done' ? 'Selesai' : 'Belum dibuka' }}
                </BaseBadge>
              </div>
              <p
                v-if="hasPendingEssay(e)"
                class="mt-2 flex items-center gap-1.5 rounded-lg bg-amber-50 px-3 py-1.5 text-xs text-amber-700 dark:bg-amber-900/20 dark:text-amber-300"
              >
                <Flag class="size-3.5" />
                Nilai belum termasuk soal esai — menunggu koreksi manual guru.
              </p>
            </div>
          </div>
        </template>
      </template>

      <!-- MODAL TOKEN -->
      <BaseModal :open="tokenModalOpen" title="Masukkan Token Ujian" @close="tokenModalOpen = false">
        <div class="space-y-4">
          <p class="text-sm text-muted-foreground">
            Masukkan token dari pengawas untuk membuka ujian
            <span class="font-medium text-foreground">{{ tokenExam?.title }}</span>.
          </p>
          <BaseInput
            v-model="tokenInput"
            label="Token"
            placeholder="Contoh: MTK88"
            required
            :error="fieldError"
            @keyup.enter="validateThenStart"
          />
        </div>
        <template #footer>
          <BaseButton variant="outline" @click="tokenModalOpen = false">Batal</BaseButton>
          <BaseButton :disabled="tokenInput.length < 3" :loading="validating" @click="validateThenStart">
            <FileEdit /> Validasi & Mulai
          </BaseButton>
        </template>
      </BaseModal>
    </div>
  </AppShell>
</template>
