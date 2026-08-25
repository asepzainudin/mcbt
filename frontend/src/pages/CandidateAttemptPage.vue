<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  AlarmClock,
  Check,
  ChevronLeft,
  ChevronRight,
  Flag,
  Loader2,
  Send,
} from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseModal from '../components/ui/BaseModal.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { apiErrorMessage } from '../lib/axios'
import {
  attemptService,
  type AttemptQuestion,
} from '../services/candidate.service'
import { useUiStore } from '../stores/ui'

const route = useRoute()
const ui = useUiStore()

const attemptId = route.params.id as string

const loading = ref(true)
const attemptNo = ref(1)
const expiresAt = ref<string>('')
const submittedAt = ref<string | null>(null)
const submitted = ref(false)
const expired = ref(false)

// sisa waktu: sumber kebenaran = server (heartbeat), dikurangi lokal per detik
const remainingSeconds = ref(0)

const questions = ref<AttemptQuestion[]>([])
const current = ref(0)
const answers = ref<Record<string, { value: string; flagged: boolean }>>({})

const saveState = ref<'idle' | 'saving' | 'saved'>('idle')
const dirty = ref<Record<string, string>>({})

let heartbeatTimer: ReturnType<typeof setInterval> | undefined
let autosaveTimer: ReturnType<typeof setInterval> | undefined
let tickTimer: ReturnType<typeof setInterval> | undefined
let saveTimer: ReturnType<typeof setTimeout> | undefined

const submitting = ref(false)
const submitModalOpen = ref(false)

const nowTick = ref(Date.now())

onMounted(async () => {
  try {
    await reloadSheet()
    startTimers()
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memuat lembar soal.'))
  } finally {
    loading.value = false
  }
})
onUnmounted(stopTimers)

function stopTimers() {
  clearInterval(heartbeatTimer)
  clearInterval(autosaveTimer)
  clearInterval(tickTimer)
  clearTimeout(saveTimer)
}

function startTimers() {
  heartbeatTimer = setInterval(heartbeat, 15_000) // sinkron server 15s
  autosaveTimer = setInterval(flushDirty, 20_000) // autosave batch 20s
  tickTimer = setInterval(tick, 1000) // turunkan counter lokal 1s
}

async function reloadSheet() {
  const sheet = await attemptService.getQuestions(attemptId)
  attemptNo.value = sheet.attempt.attempt_no
  expiresAt.value = sheet.attempt.expires_at
  submittedAt.value = sheet.attempt.submitted_at ?? null
  submitted.value = sheet.attempt.status === 'submitted'
  if (submitted.value) stopTimers()
  questions.value = sheet.questions
  for (const q of sheet.questions) {
    answers.value[q.question_id] = { value: q.answer_value, flagged: q.is_flagged }
  }
  await heartbeat()
}

async function heartbeat() {
  if (submitted.value) return
  try {
    const hb = await attemptService.heartbeat(attemptId)
    remainingSeconds.value = hb.remaining_seconds
    if (hb.attempt_status === 'submitted') {
      submitted.value = true
      submittedAt.value = hb.submitted_at
      stopTimers()
      return
    }
    if (hb.is_expired) {
      expired.value = true
      void doSubmit(true)
    }
  } catch {
    // gagal jaringan sesaat: counter lokal tetap jalan
  }
}

function tick() {
  nowTick.value = Date.now()
  if (submitted.value) return
  if (remainingSeconds.value > 0) remainingSeconds.value--
  if (remainingSeconds.value <= 0 && expiresAt.value) {
    remainingSeconds.value = 0
    expired.value = true
    void doSubmit(true)
  }
}

// ---- autosave ----
function markDirty(questionId: string, value: string) {
  dirty.value[questionId] = value
}

async function flushDirty() {
  const entries = Object.entries(dirty.value)
  if (entries.length === 0 || submitted.value) return
  const batch = entries.map(([question_id, value]) => ({ question_id, value }))
  dirty.value = {}
  saveState.value = 'saving'
  try {
    await attemptService.autosave(attemptId, batch)
    saveState.value = 'saved'
    setTimeout(() => {
      if (saveState.value === 'saved') saveState.value = 'idle'
    }, 2000)
  } catch (err) {
    for (const item of batch) {
      dirty.value[item.question_id] = item.value
    }
    saveState.value = 'idle'
    ui.toastError(apiErrorMessage(err, 'Gagal menyimpan jawaban.'))
  }
}

// ---- interaksi jawaban ----
const currentQuestion = computed(() => questions.value[current.value] ?? null)
const currentAnswer = computed(
  () => answers.value[currentQuestion.value?.question_id ?? ''] ?? { value: '', flagged: false },
)

const answeredCount = computed(
  () => Object.values(answers.value).filter((a) => a.value !== '').length,
)
const flaggedCount = computed(
  () => Object.values(answers.value).filter((a) => a.flagged).length,
)

const locked = computed(
  () => submitted.value || expired.value || remainingSeconds.value <= 0,
)

function navState(i: number): string {
  const q = questions.value[i]
  const a = answers.value[q?.question_id ?? '']
  const classes = ['relative h-10 w-10 rounded-lg text-sm font-semibold transition-colors']
  if (i === current.value) {
    classes.push('ring-2 ring-primary ring-offset-2 ring-offset-background')
  }
  if (a?.flagged) classes.push('bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300')
  else if (a && a.value !== '') classes.push('bg-primary text-primary-foreground')
  else classes.push('bg-muted text-muted-foreground hover:bg-accent')
  return classes.join(' ')
}

function goTo(i: number) {
  if (i >= 0 && i < questions.value.length) {
    void flushDirty()
    current.value = i
  }
}

function setChoice(optionKey: string) {
  if (locked.value || !currentQuestion.value) return
  const q = currentQuestion.value
  const a = answers.value[q.question_id]
  if (q.type === 'MULTIPLE_ANSWER') {
    const keys = (a.value ? a.value.split(',') : []).filter(Boolean)
    const idx = keys.indexOf(optionKey)
    if (idx >= 0) keys.splice(idx, 1)
    else keys.push(optionKey)
    a.value = keys.sort().join(',')
  } else {
    a.value = optionKey
  }
  markDirty(q.question_id, a.value)
  void flushDirty()
  saveState.value = 'saving'
}

function onTextAnswer(event: Event) {
  if (locked.value || !currentQuestion.value) return
  const a = answers.value[currentQuestion.value.question_id]
  a.value = (event.target as HTMLTextAreaElement).value
  markDirty(currentQuestion.value.question_id, a.value)
  clearTimeout(saveTimer)
  saveTimer = setTimeout(() => void flushDirty(), 800)
}

function isSelected(optionKey: string): boolean {
  const a = currentAnswer.value
  if (!a.value) return false
  return a.value.split(',').includes(optionKey)
}

async function toggleFlag() {
  const q = currentQuestion.value
  if (!q || locked.value) return
  const a = answers.value[q.question_id]
  try {
    const res = await attemptService.setFlag(attemptId, q.question_id, !a.flagged)
    a.flagged = res.is_flagged
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal mengubah tanda.'))
  }
}

// ---- submit ----

const fmtSubmitted = computed(() =>
  submittedAt.value ? new Date(submittedAt.value).toLocaleString('id-ID') : '',
)

async function doSubmit(auto: boolean) {
  if (submitting.value || submitted.value) return
  submitting.value = true
  try {
    await flushDirty()
    const result = await attemptService.submit(attemptId, true)
    submitted.value = true
    submittedAt.value = result.submitted_at
    submitModalOpen.value = false
    stopTimers()
    ui.toastSuccess(
      auto ? 'Waktu habis — jawaban dikumpulkan otomatis.' : 'Ujian berhasil dikumpulkan.',
    )
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal mengumpulkan ujian.'))
  } finally {
    submitting.value = false
  }
}

function requestSubmit() {
  if (remainingSeconds.value > 0) submitModalOpen.value = true
  else void doSubmit(true)
}

// ---- tampilan ----
function needsOptions(q: AttemptQuestion): boolean {
  return q.type === 'MULTIPLE_CHOICE' || q.type === 'TRUE_FALSE' || q.type === 'MULTIPLE_ANSWER'
}

const typeLabel: Record<string, string> = {
  MULTIPLE_CHOICE: 'Pilihan Ganda',
  TRUE_FALSE: 'Benar / Salah',
  MULTIPLE_ANSWER: 'Jawaban Ganda',
  ESSAY: 'Esai',
  SHORT_ANSWER: 'Isian Singkat',
}

function pad(n: number): string {
  return String(n).padStart(2, '0')
}

const clockDisplay = computed(() => {
  const t = Math.max(0, remainingSeconds.value)
  return `${pad(Math.floor(t / 3600))}:${pad(Math.floor((t % 3600) / 60))}:${pad(t % 60)}`
})
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-6xl space-y-4">
      <!-- SELESAI -->
      <div v-if="submitted" class="rounded-xl border border-success/40 bg-card p-10 text-center shadow-sm">
        <span class="mx-auto mb-4 flex size-14 items-center justify-center rounded-full bg-success/15">
          <Check class="size-7 text-success" />
        </span>
        <h1 class="text-2xl font-bold">Ujian Selesai</h1>
        <p class="mt-2 text-sm text-muted-foreground">
          Jawaban Anda telah dikumpulkan{{ fmtSubmitted ? ` pada ${fmtSubmitted}` : '' }}.
        </p>
        <BaseBadge tone="success" class="mt-4">Attempt #{{ attemptNo }} — submitted</BaseBadge>
      </div>

      <template v-else>
        <!-- BAR ATAS -->
        <div class="flex flex-wrap items-center justify-between gap-3 rounded-xl border-2 border-destructive/40 bg-card p-4 shadow-sm">
          <div class="flex items-center gap-3">
            <span class="flex size-10 items-center justify-center rounded-lg bg-destructive/10 text-destructive">
              <AlarmClock class="size-5" />
            </span>
            <div>
              <p class="text-xs uppercase tracking-wide text-muted-foreground">Sisa Waktu (waktu server)</p>
              <p
                :class="[
                  'font-mono text-2xl font-bold',
                  remainingSeconds <= 60 ? 'animate-pulse text-destructive' : 'text-destructive',
                ]"
              >
                {{ clockDisplay }}
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <span class="inline-flex items-center gap-1 text-xs text-muted-foreground">
              <Loader2 v-if="saveState === 'saving'" class="size-3.5 animate-spin" />
              <Check v-else-if="saveState === 'saved'" class="size-3.5 text-success" />
              {{ saveState === 'saving' ? 'Menyimpan…' : saveState === 'saved' ? 'Tersimpan' : 'Autosave aktif' }}
            </span>
            <BaseButton variant="destructive" size="sm" :disabled="locked || submitting" @click="requestSubmit">
              <Send /> Kumpulkan
            </BaseButton>
          </div>
        </div>

        <LoadingState v-if="loading" />

        <div v-else-if="questions.length === 0" class="rounded-xl border border-dashed border-border p-10 text-center text-muted-foreground">
          Ujian ini belum memiliki soal.
        </div>

        <div v-else class="grid gap-4 lg:grid-cols-[1fr_260px]">
          <!-- LEMBAR SOAL -->
          <div class="space-y-4">
            <div v-if="currentQuestion" class="rounded-xl border border-border bg-card p-6 shadow-sm">
              <div class="mb-3 flex flex-wrap items-center gap-2">
                <span class="flex size-8 items-center justify-center rounded-lg bg-primary font-bold text-primary-foreground">
                  {{ current + 1 }}
                </span>
                <BaseBadge tone="outline">{{ currentQuestion.section_name }}</BaseBadge>
                <BaseBadge tone="info">{{ typeLabel[currentQuestion.type] ?? currentQuestion.type }}</BaseBadge>
                <BaseBadge v-if="currentAnswer.flagged" tone="warning">
                  <Flag class="mr-1 size-3" /> Ragu-ragu
                </BaseBadge>
              </div>

              <img
                v-if="currentQuestion.media && currentQuestion.media_position === 'before'"
                :src="currentQuestion.media.url"
                class="mb-3 max-h-48 rounded-lg border border-border"
                alt=""
              />

              <div class="prose-sm max-w-none [&_ol]:list-decimal [&_ol]:pl-5 [&_p]:my-1 [&_ul]:list-disc [&_ul]:pl-5" v-html="currentQuestion.text" />

              <img
                v-if="currentQuestion.media && currentQuestion.media_position !== 'before'"
                :src="currentQuestion.media.url"
                class="mt-3 max-h-48 rounded-lg border border-border"
                alt=""
              />

              <div v-if="needsOptions(currentQuestion)" class="mt-5 space-y-2">
                <button
                  v-for="opt in currentQuestion.options"
                  :key="opt.option_key"
                  type="button"
                  :disabled="locked"
                  :class="[
                    'flex w-full items-start gap-3 rounded-xl border p-3 text-left text-sm transition-all',
                    'disabled:cursor-not-allowed disabled:opacity-60',
                    isSelected(opt.option_key)
                      ? 'border-primary bg-primary/5 ring-1 ring-primary'
                      : 'border-border hover:border-primary/40 hover:bg-accent/40',
                  ]"
                  @click="setChoice(opt.option_key)"
                >
                  <span
                    :class="[
                      'flex size-7 shrink-0 items-center justify-center rounded-full border font-bold',
                      isSelected(opt.option_key)
                        ? 'border-primary bg-primary text-primary-foreground'
                        : 'border-border text-muted-foreground',
                    ]"
                  >
                    {{ opt.option_key }}
                  </span>
                  <span class="min-w-0 flex-1 space-y-1.5">
                    <span class="block" v-html="opt.text" />
                    <img v-if="opt.media" :src="opt.media.url" class="max-h-20 rounded border border-border" alt="" />
                  </span>
                  <Check v-if="isSelected(opt.option_key)" class="mt-1 size-4 shrink-0 text-primary" />
                </button>
                <p class="pt-1 text-[11px] text-muted-foreground">
                  {{ currentQuestion.type === 'MULTIPLE_ANSWER' ? 'Centang semua jawaban yang benar.' : 'Pilih satu jawaban.' }}
                  Jawaban tersimpan otomatis ke server.
                </p>
              </div>

              <div v-else-if="currentQuestion.type === 'ESSAY' || currentQuestion.type === 'SHORT_ANSWER'" class="mt-5">
                <textarea
                  :value="currentAnswer.value"
                  rows="6"
                  :disabled="locked"
                  placeholder="Ketik jawaban Anda di sini… (tersimpan otomatis)"
                  class="w-full rounded-lg border border-input bg-transparent p-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-60"
                  @input="onTextAnswer"
                />
              </div>

              <div class="mt-5 flex items-center justify-between border-t border-border pt-4">
                <BaseButton
                  :variant="currentAnswer.flagged ? 'secondary' : 'outline'"
                  :disabled="locked"
                  @click="toggleFlag()"
                >
                  <Flag /> {{ currentAnswer.flagged ? 'Lepas Ragu-ragu' : 'Ragu-ragu' }}
                </BaseButton>
                <div class="flex gap-2">
                  <BaseButton variant="outline" :disabled="current === 0" @click="goTo(current - 1)">
                    <ChevronLeft /> Sebelumnya
                  </BaseButton>
                  <BaseButton :disabled="current >= questions.length - 1" @click="goTo(current + 1)">
                    Selanjutnya <ChevronRight />
                  </BaseButton>
                </div>
              </div>
            </div>
          </div>

          <!-- NAVIGASI -->
          <aside class="h-fit space-y-4 rounded-xl border border-border bg-card p-4 shadow-sm lg:sticky lg:top-20">
            <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Navigasi Soal ({{ questions.length }})
            </p>
            <div class="grid grid-cols-5 gap-2">
              <button
                v-for="(q, i) in questions"
                :key="q.question_id"
                :class="navState(i)"
                @click="goTo(i)"
              >
                {{ i + 1 }}
                <span
                  v-if="answers[q.question_id]?.flagged"
                  class="absolute -right-0.5 -top-0.5 size-2.5 rounded-full bg-amber-500"
                />
              </button>
            </div>

            <div class="space-y-1 border-t border-border pt-3 text-xs text-muted-foreground">
              <p class="flex items-center gap-2">
                <span class="size-3 rounded bg-primary" /> Dijawab ({{ answeredCount }})
              </p>
              <p class="flex items-center gap-2">
                <span class="size-3 rounded bg-amber-400" /> Ragu-ragu ({{ flaggedCount }})
              </p>
              <p class="flex items-center gap-2">
                <span class="size-3 rounded bg-muted" /> Belum dijawab ({{ questions.length - answeredCount }})
              </p>
            </div>

            <BaseButton block variant="destructive" :disabled="locked || submitting" @click="requestSubmit">
              <Send /> Kumpulkan Jawaban
            </BaseButton>
          </aside>
        </div>
      </template>

      <!-- KONFIRMASI SUBMIT -->
      <BaseModal :open="submitModalOpen" title="Kumpulkan Ujian?" @close="submitModalOpen = false">
        <p class="text-sm text-muted-foreground">
          Jawaban akan dikumpulkan dan <span class="font-semibold text-foreground">tidak dapat diubah lagi</span>.
          Terjawab: <span class="font-semibold text-foreground">{{ answeredCount }}/{{ questions.length }}</span>.
        </p>
        <template #footer>
          <BaseButton variant="outline" @click="submitModalOpen = false">Kembali</BaseButton>
          <BaseButton variant="destructive" :loading="submitting" @click="doSubmit(false)">
            <Send /> Ya, Kumpulkan
          </BaseButton>
        </template>
      </BaseModal>
    </div>
  </AppShell>
</template>
