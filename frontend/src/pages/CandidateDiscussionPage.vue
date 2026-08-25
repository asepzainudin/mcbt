<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Check, X } from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { apiErrorMessage } from '../lib/axios'
import { attemptService } from '../services/candidate.service'
import { useUiStore } from '../stores/ui'

const route = useRoute()
const ui = useUiStore()

const attemptId = route.params.id as string
const loading = ref(true)
const questions = ref<
  {
    question_id: string
    section_name: string
    type: string
    text: string
    score_weight: number
    media?: { url: string } | null
    media_position: string
    options: { option_key: string; text: string; media_url?: string }[]
    correct_keys: string[]
    explanation: string | null
    answer_value: string
    is_correct: boolean | null
    score: number | null
    feedback: string | null
    is_flagged: boolean
  }[]
>([])
const submitted = ref(false)

onMounted(async () => {
  try {
    const res = await attemptService.getDiscussion(attemptId)
    submitted.value = true
    questions.value = res as never[]
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memuat pembahasan.'))
  } finally {
    loading.value = false
  }
})

const typeLabel: Record<string, string> = {
  MULTIPLE_CHOICE: 'Pilihan Ganda',
  TRUE_FALSE: 'Benar / Salah',
  MULTIPLE_ANSWER: 'Jawaban Ganda',
  ESSAY: 'Esai',
  SHORT_ANSWER: 'Isian Singkat',
}

const needsOptions = (q: { type: string }) =>
  q.type === 'MULTIPLE_CHOICE' || q.type === 'TRUE_FALSE' || q.type === 'MULTIPLE_ANSWER'
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-4xl space-y-6">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Pembahasan Soal</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          Pembahasan tampil setelah ujian dikumpulkan.
        </p>
      </div>

      <LoadingState v-if="loading" />

      <EmptyState v-else-if="questions.length === 0" title="Tidak ada soal" />

      <div v-else class="space-y-4">
        <div
          v-for="(q, i) in questions"
          :key="q.question_id"
          class="rounded-xl border border-border bg-card p-6 shadow-sm"
        >
          <div class="mb-3 flex flex-wrap items-center gap-2">
            <span class="flex size-8 items-center justify-center rounded-lg bg-primary font-bold text-primary-foreground">
              {{ i + 1 }}
            </span>
            <BaseBadge tone="outline">{{ q.section_name }}</BaseBadge>
            <BaseBadge tone="info">{{ typeLabel[q.type] ?? q.type }}</BaseBadge>
            <BaseBadge v-if="q.is_correct === true" tone="success">Benar</BaseBadge>
            <BaseBadge v-else-if="q.is_correct === false" tone="destructive">Salah</BaseBadge>
          </div>

          <img
            v-if="q.media && q.media_position === 'before'"
            :src="q.media.url"
            class="mb-3 max-h-40 rounded-lg border border-border"
            alt=""
          />

          <div class="prose-sm max-w-none [&_ol]:list-decimal [&_ol]:pl-5 [&_p]:my-1 [&_ul]:list-disc [&_ul]:pl-5" v-html="q.text" />

          <img
            v-if="q.media && q.media_position !== 'before'"
            :src="q.media.url"
            class="mt-3 max-h-40 rounded-lg border border-border"
            alt=""
          />

          <!-- OPSI -->
          <div v-if="needsOptions(q)" class="mt-4 space-y-2">
            <div
              v-for="opt in q.options"
              :key="opt.option_key"
              :class="[
                'flex items-start gap-3 rounded-lg border p-3 text-sm transition-colors',
                q.correct_keys.includes(opt.option_key)
                  ? 'border-success/50 bg-success/5'
                  : 'border-border',
              ]"
            >
              <span
                :class="[
                  'flex size-7 shrink-0 items-center justify-center rounded-full border font-bold',
                  q.correct_keys.includes(opt.option_key)
                    ? 'border-success bg-success text-white'
                    : 'border-border text-muted-foreground',
                ]"
              >
                {{ opt.option_key }}
              </span>
              <span class="min-w-0 flex-1">
                <span class="block">{{ opt.text }}</span>
                <img v-if="opt.media_url" :src="opt.media_url" class="mt-1.5 max-h-16 rounded border border-border" alt="" />
              </span>
              <span v-if="q.correct_keys.includes(opt.option_key)" class="mt-0.5 shrink-0">
                <Check class="size-4 text-success" />
              </span>
              <span v-else-if="q.answer_value.includes(opt.option_key)" class="mt-0.5 shrink-0">
                <X class="size-4 text-destructive" />
              </span>
            </div>
            <p class="mt-1 text-xs text-muted-foreground">
              Jawaban Anda:
              <span class="font-mono font-semibold">{{ q.answer_value || '—' }}</span>
            </p>
          </div>

          <!-- ESAI / ISIAN -->
          <div v-else-if="q.type === 'ESSAY' || q.type === 'SHORT_ANSWER'" class="mt-4">
            <p class="mb-1 text-xs font-semibold uppercase text-muted-foreground">Jawaban Anda</p>
            <div class="rounded-lg border border-border bg-muted/40 p-3 text-sm whitespace-pre-wrap">
              {{ q.answer_value || '—' }}
            </div>
          </div>

          <!-- PEMBAHASAN -->
          <div v-if="q.explanation" class="mt-4 rounded-lg border border-primary/30 bg-primary/5 p-4">
            <p class="mb-1 text-xs font-semibold uppercase text-primary">Pembahasan</p>
            <p class="text-sm leading-relaxed text-foreground">{{ q.explanation }}</p>
          </div>

          <div v-if="q.feedback" class="mt-2 rounded-lg border border-border bg-muted/40 p-3 text-sm">
            <p class="mb-0.5 text-xs font-semibold uppercase text-muted-foreground">Feedback Guru</p>
            <p>{{ q.feedback }}</p>
          </div>
        </div>
      </div>
    </div>
  </AppShell>
</template>
