<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Clock } from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import { candidateService } from '../services/candidate.service'
import type { StartAttemptResult } from '../types/api'
import { useUiStore } from '../stores/ui'

const route = useRoute()
const ui = useUiStore()

const attemptId = route.params.id as string
const attempt = ref<StartAttemptResult | null>(null)
const loading = ref(true)

const now = ref(Date.now())
let clockTimer: ReturnType<typeof setInterval> | undefined

onMounted(async () => {
  // attempt aktif diambil dari list candidate (endpoint detail attempt menyusul)
  try {
    const exams = await candidateService.listExams()
    const active = exams.find((e) => e.active_attempt_id === attemptId)
    if (active?.active_expires_at) {
      attempt.value = {
        attempt_id: attemptId,
        started_at: '',
        expires_at: active.active_expires_at,
        attempt_no: active.attempts_used,
      }
    } else {
      ui.toastWarning('Attempt tidak aktif atau sudah berakhir.')
    }
  } catch {
    ui.toastError('Gagal memuat attempt.')
  } finally {
    loading.value = false
  }
  clockTimer = setInterval(() => (now.value = Date.now()), 1000)
})
onUnmounted(() => clearInterval(clockTimer))

const remaining = computed(() => {
  if (!attempt.value) return '--:--:--'
  const diff = new Date(attempt.value.expires_at).getTime() - now.value
  if (diff <= 0) return '00:00:00'
  const h = Math.floor(diff / 3_600_000)
  const m = Math.floor((diff % 3_600_000) / 60_000)
  const s = Math.floor((diff % 60_000) / 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(h)}:${pad(m)}:${pad(s)}`
})
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-3xl space-y-6">
      <div class="flex items-center justify-between rounded-xl border-2 border-destructive/40 bg-card p-5 shadow-sm">
        <div>
          <p class="text-xs uppercase tracking-wide text-muted-foreground">Sisa Waktu</p>
          <p class="font-mono text-4xl font-bold text-destructive">{{ remaining }}</p>
        </div>
        <BaseBadge tone="destructive">Attempt #{{ attempt?.attempt_no ?? '-' }}</BaseBadge>
      </div>

      <LoadingState v-if="loading" />

      <div v-else-if="attempt" class="rounded-xl border border-dashed border-border p-10 text-center">
        <Clock class="mx-auto mb-3 size-10 text-muted-foreground" />
        <h2 class="text-lg font-semibold">Halaman Pengerjaan Soal</h2>
        <p class="mt-1 text-sm text-muted-foreground">
          Attempt #{{ attempt.attempt_no }} aktif — pengerjaan & jawaban soal akan dibangun pada step berikutnya.
        </p>
      </div>
    </div>
  </AppShell>
</template>
