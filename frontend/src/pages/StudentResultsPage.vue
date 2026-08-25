<script setup lang="ts">
import { onMounted, ref } from 'vue'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import BaseCard from '../components/ui/BaseCard.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { apiErrorMessage } from '../lib/axios'
import { resultService } from '../services/result.service'
import type { StudentResultRow } from '../types/api'
import { useAuthStore } from '../stores/auth'
import { useUiStore } from '../stores/ui'

const ui = useUiStore()
const auth = useAuthStore()

const results = ref<StudentResultRow[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    // resolve student_id dari user login
    const me = auth.user
    if (!me) return
    // endpoint /students/{id}/results membutuhkan student_id (bukan user_id)
    // siswa mengakses miliknya → backend cek ownership
    results.value = await resultService.myResults()
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memuat hasil ujian.'))
  } finally {
    loading.value = false
  }
})

const fmtDate = (d: string | null) =>
  d ? new Date(d).toLocaleString('id-ID', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' }) : '—'
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-4xl space-y-6">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Hasil Ujian Saya</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          Rekap nilai ujian yang telah Anda ikuti.
        </p>
      </div>

      <LoadingState v-if="loading" />

      <template v-else>
        <EmptyState v-if="results.length === 0" title="Belum ada hasil ujian" message="Ikuti ujian terlebih dahulu untuk melihat hasil." />

        <div v-else class="space-y-4">
          <BaseCard v-for="r in results" :key="r.exam_id" class="p-5">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div class="min-w-0">
                <h3 class="font-semibold">{{ r.exam_title }}</h3>
                <p class="mt-0.5 text-xs text-muted-foreground">{{ r.subject_name }}</p>
              </div>
              <div class="text-right">
                <p class="text-[10px] uppercase tracking-wide text-muted-foreground">Nilai</p>
                <p
                  :class="[
                    'font-mono text-2xl font-bold',
                    r.score !== null
                      ? r.score >= r.passing_grade ? 'text-success' : 'text-destructive'
                      : 'text-muted-foreground',
                  ]"
                >
                  {{ r.score !== null ? r.score : '—' }}
                </p>
              </div>
            </div>
            <div class="mt-3 flex flex-wrap gap-2 text-xs text-muted-foreground">
              <span>KKM: {{ r.passing_grade }}</span>
              <span>·</span>
              <span>Dikumpulkan: {{ fmtDate(r.submitted_at) }}</span>
            </div>
            <div v-if="r.score !== null" class="mt-3 flex flex-wrap items-center gap-2">
              <BaseBadge :tone="r.score >= r.passing_grade ? 'success' : 'destructive'">
                {{ r.score >= r.passing_grade ? 'Lulus' : 'Belum Lulus' }}
              </BaseBadge>
            </div>
          </BaseCard>
        </div>
      </template>
    </div>
  </AppShell>
</template>
