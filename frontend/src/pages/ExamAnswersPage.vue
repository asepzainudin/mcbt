<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Check, ChevronDown, ChevronRight } from 'lucide-vue-next'

const typeLabel: Record<string, string> = {
  MULTIPLE_CHOICE: 'PG',
  TRUE_FALSE: 'B/S',
  MULTIPLE_ANSWER: 'Multi',
  ESSAY: 'Esai',
  SHORT_ANSWER: 'Isian',
}

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import BaseCard from '../components/ui/BaseCard.vue'
import BaseInput from '../components/ui/BaseInput.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { apiErrorMessage } from '../lib/axios'
import { examService } from '../services/exam.service'
import type { Exam } from '../types/api'
import { useUiStore } from '../stores/ui'

const ui = useUiStore()
const route = useRoute()
const router = useRouter()

const examId = route.params.id as string
const exam = ref<Exam | null>(null)
const loading = ref(true)
const search = ref('')

interface SheetAnswer {
  question_id: string
  type: string
  text: string
  score_weight: number
  answer_value: string
  correct_keys: string[]
  option_texts: { option_key: string; text: string }[]
  score: number | null
  is_correct: boolean | null
  feedback: string | null
  graded_via: string | null
}

interface SheetStudent {
  attempt_id: string
  student_name: string
  nis: string
  status: string
  score: number | null
  submitted_at: string | null
  answers: SheetAnswer[]
}

const students = ref<SheetStudent[]>([])
const expanded = ref<Set<string>>(new Set())

const filtered = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return students.value
  return students.value.filter(
    (s) => s.student_name.toLowerCase().includes(q) || s.nis.includes(q),
  )
})

onMounted(async () => {
  try {
    exam.value = await examService.get(examId)
    const res = await fetch(`/api/v1/exams/${examId}/grading`, { credentials: 'include' })
    if (!res.ok) throw new Error('fetch failed')
    const json = await res.json()
    students.value = json.data ?? []
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memuat jawaban siswa.'))
  } finally {
    loading.value = false
  }
})

function toggleExpand(id: string) {
  if (expanded.value.has(id)) expanded.value.delete(id)
  else expanded.value.add(id)
}

</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-5xl space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <button class="mb-1 flex items-center gap-1 text-xs font-medium text-muted-foreground hover:text-foreground" @click="router.push('/exams')">
            <ArrowLeft class="size-3" /> Ujian
          </button>
          <h1 class="text-2xl font-bold tracking-tight">Jawaban Siswa</h1>
          <p class="mt-1 text-sm text-muted-foreground">{{ exam?.title ?? '' }}</p>
        </div>
        <div class="w-56">
          <BaseInput v-model="search" placeholder="Cari nama / NIS…" />
        </div>
      </div>

      <LoadingState v-if="loading" />

      <template v-else>
        <EmptyState v-if="filtered.length === 0" title="Belum ada jawaban" message="Belum ada siswa yang mengumpulkan ujian ini." />

        <div v-else class="space-y-4">
          <BaseCard v-for="s in filtered" :key="s.attempt_id" class="overflow-hidden">
            <button
              class="flex w-full items-center justify-between gap-3 p-4 text-left transition-colors hover:bg-accent/50"
              @click="toggleExpand(s.attempt_id)"
            >
              <div class="flex items-center gap-3">
                <span class="flex size-9 items-center justify-center rounded-full bg-primary/10 text-xs font-bold text-primary">
                  {{ s.student_name.split(' ').map((w) => w[0]).slice(0, 2).join('').toUpperCase() }}
                </span>
                <div>
                  <p class="font-semibold">{{ s.student_name }}</p>
                  <p class="text-xs text-muted-foreground">NIS {{ s.nis }}</p>
                </div>
              </div>
              <ChevronDown v-if="!expanded.has(s.attempt_id)" class="size-4 text-muted-foreground" />
              <ChevronRight v-else class="size-4 text-muted-foreground" />
            </button>

            <div v-if="expanded.has(s.attempt_id)" class="space-y-4 border-t border-border p-4">
              <div
                v-for="(a, ai) in s.answers"
                :key="a.question_id"
                class="rounded-lg border border-border p-4"
              >
                <div class="mb-2 flex items-center gap-2">
                  <span class="font-mono text-sm font-bold text-primary">{{ ai + 1 }}.</span>
                  <BaseBadge tone="outline">{{ typeLabel[a.type] ?? a.type }}</BaseBadge>
                </div>
                <p class="text-sm" v-html="a.text" />

                <!-- PILIHAN -->
                <div v-if="(a.option_texts?.length ?? 0) > 0" class="mt-3 space-y-1.5">
                  <div
                    v-for="opt in (a.option_texts ?? [])"
                    :key="opt.option_key"
                    :class="[
                      'flex items-center gap-2 rounded-md px-2.5 py-1.5 text-sm',
                      a.answer_value.includes(opt.option_key) ? 'bg-primary/10 font-medium' : '',
                      (a.correct_keys ?? []).includes(opt.option_key) ? 'text-success' : '',
                    ]"
                  >
                    <span class="font-mono font-bold">{{ opt.option_key }}.</span>
                    <span>{{ opt.text }}</span>
                    <Check v-if="a.answer_value.includes(opt.option_key)" class="ml-auto size-4 text-primary" />
                  </div>
                </div>

                <!-- ESAI / ISIAN -->
                <div v-else class="mt-2 rounded-lg bg-muted/40 p-3 text-sm whitespace-pre-wrap">
                  {{ a.answer_value || '\u2014' }}
                </div>
              </div>
            </div>
          </BaseCard>
        </div>
      </template>
    </div>
  </AppShell>
</template>
