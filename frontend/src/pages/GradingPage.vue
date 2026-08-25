<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowLeft,
  Calculator,
  Check,
  Save,
} from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseInput from '../components/ui/BaseInput.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { apiErrorMessage } from '../lib/axios'
import { examService } from '../services/exam.service'
import { useUiStore } from '../stores/ui'

interface UngradedEssay {
  attempt_id: string
  answer_id: string
  student_name: string
  nis: string
  question_id: string
  question_text: string
  score_weight: number
  answer_value: string
}

const ui = useUiStore()
const route = useRoute()
const router = useRouter()

const examId = route.params.id as string

const loading = ref(true)
const calculating = ref(false)
const essays = ref<UngradedEssay[]>([])

const scores = ref<Record<string, string>>({})
const feedbacks = ref<Record<string, string>>({})
const savingId = ref<string | null>(null)
const savedIds = ref<Set<string>>(new Set())
const fieldErrors = ref<Record<string, string>>({})

onMounted(async () => {
  await calculate()
})

async function calculate() {
  calculating.value = true
  try {
    const res = await examService.calculateGrades(examId)
    ui.toastSuccess(
      `Nilai otomatis diproses: ${res.attempts_graded} attempt, ${res.questions_graded} soal objektif.`,
    )
    essays.value = await examService.ungradedEssays(examId)
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memproses penilaian.'))
  } finally {
    loading.value = false
    calculating.value = false
  }
}

async function saveGrade(essay: UngradedEssay) {
  const key = essay.answer_id
  const score = Number(scores.value[key])
  fieldErrors.value = {}

  if (score < 0 || score > essay.score_weight) {
    fieldErrors.value[key] = `Nilai harus di antara 0 dan ${essay.score_weight}`
    return
  }

  savingId.value = key
  try {
    await examService.gradeEssay(
      essay.attempt_id,
      essay.question_id,
      score,
      feedbacks.value[key] || null,
    )
    savedIds.value.add(key)
    ui.toastSuccess(`Nilai ${essay.student_name} tersimpan.`)
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal menyimpan nilai.'))
  } finally {
    savingId.value = null
  }
}
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-4xl space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <button class="mb-1 flex items-center gap-1 text-xs font-medium text-muted-foreground hover:text-foreground" @click="router.push('/exams')">
            <ArrowLeft class="size-3" />
            ‹ Ujian
          </button>
          <h1 class="text-2xl font-bold tracking-tight">Penilaian Esai</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Hitung nilai otomatis untuk soal objektif, lalu koreksi esai secara manual.
          </p>
        </div>
        <BaseButton :loading="calculating" @click="calculate">
          <Calculator /> Hitung Nilai Otomatis
        </BaseButton>
      </div>

      <LoadingState v-if="loading || calculating" />

      <template v-else>
        <EmptyState
          v-if="essays.length === 0"
          title="Tidak ada esai menunggu koreksi"
          message="Semua jawaban esai sudah dinilai, atau belum ada siswa yang mengumpulkan."
        />

        <div v-else class="space-y-4">
          <div
            v-for="essay in essays"
            :key="essay.answer_id"
            class="rounded-xl border border-border bg-card p-5 shadow-sm"
          >
            <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
              <div>
                <p class="font-semibold">{{ essay.student_name }}</p>
                <p class="text-xs text-muted-foreground">NIS {{ essay.nis }}</p>
              </div>
              <BaseBadge tone="info">bobot {{ essay.score_weight }}</BaseBadge>
            </div>

            <p class="mb-2 text-sm font-medium">{{ essay.question_text }}</p>
            <div class="mb-4 rounded-lg bg-muted/60 p-3 text-sm">
              <p class="mb-1 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">Jawaban siswa</p>
              <p class="whitespace-pre-wrap">{{ essay.answer_value }}</p>
            </div>

            <div class="grid gap-3 sm:grid-cols-[140px_1fr_auto] sm:items-start">
              <BaseInput
                v-model="scores[essay.answer_id]"
                label="Nilai"
                type="number"
                step="0.1"
                :min="0"
                :max="essay.score_weight"
                :error="fieldErrors[essay.answer_id]"
              />
              <BaseInput
                v-model="feedbacks[essay.answer_id]"
                label="Feedback (opsional)"
                placeholder="Masukan untuk siswa…"
              />
              <BaseButton
                class="sm:mt-6"
                :loading="savingId === essay.answer_id"
                @click="saveGrade(essay)"
              >
                <Save /> Simpan
              </BaseButton>
            </div>
            <p v-if="savedIds.has(essay.answer_id)" class="mt-2 flex items-center gap-1 text-xs text-success">
              <Check class="size-3.5" /> Nilai tersimpan
            </p>
          </div>
        </div>
      </template>
    </div>
  </AppShell>
</template>
