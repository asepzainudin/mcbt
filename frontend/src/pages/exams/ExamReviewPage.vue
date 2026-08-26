<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { CheckCircle2, Circle, FileSearch, ListTree } from 'lucide-vue-next'

import AppShell from '../../components/layout/AppShell.vue'
import BaseBadge from '../../components/ui/BaseBadge.vue'
import BaseButton from '../../components/ui/BaseButton.vue'
import BaseCard from '../../components/ui/BaseCard.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import { apiErrorMessage } from '../../lib/axios'
import { examService } from '../../services/exam.service'
import { sectionService } from '../../services/section.service'
import type { ExamReview, ExamReviewSection } from '../../types/api'
import { useUiStore } from '../../stores/ui'

const ui = useUiStore()
const route = useRoute()
const router = useRouter()

const examId = route.params.id as string

const loading = ref(true)
const review = ref<ExamReview | null>(null)
const examTitle = ref('')
const filterType = ref('')

const typeLabel: Record<string, string> = {
  MULTIPLE_CHOICE: 'Pilihan Ganda',
  TRUE_FALSE: 'Benar/Salah',
  MULTIPLE_ANSWER: 'Jawaban Ganda',
  ESSAY: 'Esai',
  SHORT_ANSWER: 'Isian Singkat',
}

onMounted(async () => {
  try {
    const [r, e] = await Promise.all([
      sectionService.review(examId),
      examService.get(examId).catch(() => null),
    ])
    review.value = r
    examTitle.value = e?.title ?? ''
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memuat review soal.'))
  } finally {
    loading.value = false
  }
})

const typeOptions = computed(() => {
  const set = new Set<string>()
  for (const s of review.value?.sections ?? []) {
    for (const q of s.questions) set.add(q.type)
  }
  return [
    { value: '', label: 'Semua jenis' },
    ...[...set].map((t) => ({ value: t, label: typeLabel[t] ?? t })),
  ]
})

function matchFilter(text: string | null | undefined): boolean {
  return !filterType.value || text === filterType.value
}

const filteredSections = computed<ExamReviewSection[]>(() => {
  const secs = [...(review.value?.sections ?? [])].sort((a, b) => a.sequence - b.sequence)
  return secs.map((s) => ({
    ...s,
    questions: s.questions.filter((q) => matchFilter(q.type)),
  }))
})

const shownCount = computed(() =>
  filteredSections.value.reduce((n, s) => n + s.questions.length, 0),
)

const typeCounts = computed(() => {
  const counts: Record<string, number> = {}
  for (const s of review.value?.sections ?? []) {
    for (const q of s.questions) counts[q.type] = (counts[q.type] ?? 0) + 1
  }
  return counts
})

const optionText = (o: { text?: string; content?: string; label?: string }) =>
  o.text || o.content || o.label || '-'

const stripHtml = (s: string | null | undefined) =>
  (s ?? '').replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim()
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-4xl space-y-6">
      <div>
        <button
          class="mb-1 flex items-center gap-1 text-xs font-medium text-muted-foreground hover:text-foreground"
          @click="router.push('/exams')"
        >
          ‹ Ujian
        </button>
        <h1 class="flex items-center gap-2 text-2xl font-bold tracking-tight">
          <FileSearch class="size-6 text-primary" /> Review Soal Ujian
        </h1>
        <p class="mt-1 text-sm text-muted-foreground">
          {{ examTitle || 'Muatan soal ujian' }} — periksa kembali sebelum diujikan ke siswa.
        </p>
      </div>

      <LoadingState v-if="loading" message="Memuat soal…" />

      <template v-else-if="review">
        <!-- ringkasan -->
        <div class="grid gap-3 sm:grid-cols-3">
          <BaseCard class="p-4">
            <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Total Soal</p>
            <p class="mt-1 text-2xl font-bold">{{ review.total_questions }}</p>
          </BaseCard>
          <BaseCard class="p-4">
            <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Total Skor</p>
            <p class="mt-1 text-2xl font-bold">{{ review.total_score }}</p>
          </BaseCard>
          <BaseCard class="p-4">
            <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Per Jenis</p>
            <div class="mt-1.5 flex flex-wrap gap-1">
              <BaseBadge v-for="(n, t) in typeCounts" :key="t" tone="info">
                {{ typeLabel[t] ?? t }}: {{ n }}
              </BaseBadge>
              <span v-if="!Object.keys(typeCounts).length" class="text-sm text-muted-foreground">-</span>
            </div>
          </BaseCard>
        </div>

        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="w-52">
            <select
              v-model="filterType"
              class="h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              <option v-for="o in typeOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
            </select>
          </div>
          <p class="text-sm text-muted-foreground">Menampilkan {{ shownCount }} soal</p>
        </div>

        <EmptyState
          v-if="review.total_questions === 0"
          title="Belum ada soal"
          message="Mapping soal terlebih dahulu lewat menu Section Ujian."
        >
          <template #action>
            <BaseButton variant="outline" @click="router.push(`/exams/${examId}/sections`)">
              <ListTree /> Ke Section Ujian
            </BaseButton>
          </template>
        </EmptyState>

        <!-- daftar per section -->
        <div v-for="s in filteredSections" v-else :key="s.id" class="space-y-3">
          <div class="flex items-center justify-between rounded-xl bg-primary/10 px-4 py-2.5">
            <p class="font-semibold">
              <span class="mr-2 text-primary">#{{ s.sequence }}</span>{{ s.name }}
            </p>
            <p class="text-xs font-medium text-muted-foreground">
              {{ s.question_count }} soal · {{ s.total_score }} poin
            </p>
          </div>

          <BaseCard v-for="(q, qi) in s.questions" :key="q.id" class="p-5">
            <div class="flex items-start justify-between gap-3">
              <p class="flex items-start gap-3 font-medium leading-relaxed">
                <span class="font-mono text-sm text-muted-foreground">{{ qi + 1 }}.</span>
                <span>{{ stripHtml(q.text || q.content) || '(tanpa teks)' }}</span>
              </p>
              <BaseBadge tone="outline">{{ typeLabel[q.type] ?? q.type }}</BaseBadge>
            </div>

            <!-- opsi pilihan -->
            <ul v-if="['MULTIPLE_CHOICE', 'TRUE_FALSE', 'MULTIPLE_ANSWER'].includes(q.type)" class="mt-3 space-y-1.5 pl-8">
              <li
                v-for="o in q.options"
                :key="o.id"
                :class="
                  o.is_correct
                    ? 'border-success/40 bg-success/10 font-medium'
                    : 'border-border'
                "
                class="flex items-center gap-2 rounded-lg border px-3 py-1.5 text-sm"
              >
                <component
                  :is="o.is_correct ? CheckCircle2 : Circle"
                  :class="o.is_correct ? 'text-success' : 'text-muted-foreground'"
                  class="size-4 shrink-0"
                />
                <span class="font-mono text-xs uppercase">{{ o.option_key }}</span>
                <span>{{ stripHtml(optionText(o)) }}</span>
                <BaseBadge v-if="o.is_correct" tone="success" class="ml-auto">Kunci</BaseBadge>
              </li>
            </ul>

            <!-- kunci esai/isian -->
            <div v-else-if="q.answer_keys.length" class="mt-3 pl-8">
              <div class="rounded-lg border border-success/40 bg-success/10 px-3 py-2 text-sm">
                <span class="font-semibold text-success">Kunci Jawaban:</span>
                {{ stripHtml(q.answer_keys.join('\n')) }}
              </div>
            </div>

            <!-- pembahasan -->
            <div v-if="q.explanation" class="mt-2 pl-8">
              <div class="rounded-lg bg-accent/60 px-3 py-2 text-sm text-muted-foreground">
                <span class="font-semibold text-foreground">Pembahasan:</span>
                {{ stripHtml(q.explanation) }}
              </div>
            </div>

            <p class="mt-3 pl-8 text-xs text-muted-foreground">Bobot nilai: {{ q.score_weight }}</p>
          </BaseCard>
        </div>
      </template>
    </div>
  </AppShell>
</template>
