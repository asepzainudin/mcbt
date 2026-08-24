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
} from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { apiErrorMessage } from '../lib/axios'
import { attemptService, type AttemptQuestion } from '../services/candidate.service'
import { useUiStore } from '../stores/ui'

const route = useRoute()
const ui = useUiStore()

const attemptId = route.params.id as string

const loading = ref(true)
const expiresAt = ref<string>('')
const attemptNo = ref(1)
const expired = ref(false)

const questions = ref<AttemptQuestion[]>([])
const current = ref(0)
const answers = ref<Record<string, { value: string; flagged: boolean }>>({})

const saveState = ref<'idle' | 'saving' | 'saved'>('idle')
let saveTimer: ReturnType<typeof setTimeout> | undefined
let pendingSave: { questionId: string; value: string } | null = null

const now = ref(Date.now())
let clockTimer: ReturnType<typeof setInterval> | undefined

onMounted(async () => {
  try {
    const sheet = await attemptService.getQuestions(attemptId)
    expiresAt.value = sheet.attempt.expires_at
    attemptNo.value = sheet.attempt.attempt_no
    if (sheet.attempt.status !== 'in_progress') expired.value = true
    questions.value = sheet.questions
    for (const q of sheet.questions) {
      answers.value[q.question_id] = { value: q.answer_value, flagged: q.is_flagged }
    }
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memuat lembar soal.'))
  } finally {
    loading.value = false
  }
  clockTimer = setInterval(() => {
    now.value = Date.now()
    if (!expired.value && expiresAt.value && new Date(expiresAt.value).getTime() <= Date.now()) {
      expired.value = true
      ui.toastWarning('Waktu pengerjaan habis.')
    }
  }, 1000)
})
onUnmounted(() => clearInterval(clockTimer))

const ui2 = ui
void ui2

const remaining = computed(() => {
  if (!expiresAt.value) return '--:--:--'
  const diff = new Date(expiresAt.value).getTime() - now.value
  if (diff <= 0) return '00:00:00'
  const h = Math.floor(diff / 3_600_000)
  const m = Math.floor((diff % 3_600_000) / 60_000)
  const s = Math.floor((diff % 60_000) / 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(h)}:${pad(m)}:${pad(s)}`
})

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
    flushSave()
    current.value = i
  }
}

// ---- penyimpanan real-time (debounce 800ms) ----
function queueSave(questionId: string, value: string) {
  pendingSave = { questionId, value }
  saveState.value = 'saving'
  clearTimeout(saveTimer)
  saveTimer = setTimeout(flushSave, 800)
}

async function flushSave() {
  clearTimeout(saveTimer)
  if (!pendingSave) return
  const { questionId, value } = pendingSave
  pendingSave = null
  try {
    await attemptService.saveAnswer(attemptId, questionId, value)
    saveState.value = 'saved'
    setTimeout(() => {
      if (saveState.value === 'saved') saveState.value = 'idle'
    }, 2000)
  } catch (err) {
    saveState.value = 'idle'
    ui.toastError(apiErrorMessage(err, 'Gagal menyimpan jawaban.'))
  }
}

function setChoice(optionKey: string) {
  if (expired.value || !currentQuestion.value) return
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
  queueSave(q.question_id, a.value)
}

function onTextAnswer(event: Event) {
  if (expired.value || !currentQuestion.value) return
  const a = answers.value[currentQuestion.value.question_id]
  a.value = (event.target as HTMLTextAreaElement).value
  queueSave(currentQuestion.value.question_id, a.value)
}

function isSelected(optionKey: string): boolean {
  const a = currentAnswer.value
  if (!a.value) return false
  return a.value.split(',').includes(optionKey)
}

async function toggleFlag() {
  const q = currentQuestion.value
  if (!q || expired.value) return
  const a = answers.value[q.question_id]
  try {
    const res = await attemptService.setFlag(attemptId, q.question_id, !a.flagged)
    a.flagged = res.is_flagged
    ui.toastSuccess(a.flagged ? 'Ditandai ragu-ragu.' : 'Tanda ragu-ragu dilepas.')
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal mengubah tanda.'))
  }
}

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
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-6xl space-y-4">
      <!-- BAR ATAS -->
      <div class="flex flex-wrap items-center justify-between gap-3 rounded-xl border-2 border-destructive/40 bg-card p-4 shadow-sm">
        <div class="flex items-center gap-3">
          <span class="flex size-10 items-center justify-center rounded-lg bg-destructive/10 text-destructive">
            <AlarmClock class="size-5" />
          </span>
          <div>
            <p class="text-xs uppercase tracking-wide text-muted-foreground">Sisa Waktu</p>
            <p class="font-mono text-2xl font-bold" :class="expired ? 'text-muted-foreground' : 'text-destructive'">
              {{ expired ? 'WAKTU HABIS' : remaining }}
            </p>
          </div>
        </div>
        <div class="flex items-center gap-2 text-xs text-muted-foreground">
          <span>Attempt #{{ attemptNo }}</span>
          <span class="inline-flex items-center gap-1">
            <template v-if="saveState === 'saving'">
              <Loader2 class="size-3.5 animate-spin" /> Menyimpan…
            </template>
            <template v-else-if="saveState === 'saved'">
              <Check class="size-3.5 text-success" /> Tersimpan
            </template>
          </span>
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

            <!-- OPSI PILIHAN -->
            <div v-if="needsOptions(currentQuestion)" class="mt-5 space-y-2">
              <button
                v-for="opt in currentQuestion.options"
                :key="opt.option_key"
                type="button"
                :disabled="expired"
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
                Jawaban tersimpan otomatis.
              </p>
            </div>

            <!-- ESAI / ISIAN -->
            <div v-else-if="currentQuestion.type === 'ESSAY' || currentQuestion.type === 'SHORT_ANSWER'" class="mt-5">
              <textarea
                :value="currentAnswer.value"
                rows="6"
                :disabled="expired"
                placeholder="Ketik jawaban Anda di sini… (tersimpan otomatis)"
                class="w-full rounded-lg border border-input bg-transparent p-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-60"
                @input="onTextAnswer"
              />
            </div>

            <!-- AKSI SOAL -->
            <div class="mt-5 flex items-center justify-between border-t border-border pt-4">
              <BaseButton
                :variant="currentAnswer.flagged ? 'secondary' : 'outline'"
                :disabled="expired"
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
              <span class="size-3 rounded bg-muted" /> Belum dijawab
                ({{ questions.length - answeredCount }})
            </p>
          </div>
        </aside>
      </div>
    </div>
  </AppShell>
</template>
