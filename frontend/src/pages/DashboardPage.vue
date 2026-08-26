<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  ShieldCheck,
  Activity,
  UserRound,
  GraduationCap,
  Users,
  BookOpenCheck,
  ClipboardList,
  Trophy,
  CheckCircle2,
  Target,
  TrendingUp,
} from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge, { type BadgeTone } from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseCard from '../components/ui/BaseCard.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { api } from '../lib/axios'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()

type HealthStatus = 'checking' | 'up' | 'down'
const health = ref<HealthStatus>('checking')

async function checkHealth() {
  health.value = 'checking'
  try {
    const res = await api.get('/health')
    health.value = res.data?.data?.status === 'UP' ? 'up' : 'down'
  } catch {
    health.value = 'down'
  }
}

onMounted(checkHealth)

const roleToneMap: Record<string, BadgeTone> = {
  admin: 'primary',
  teacher: 'info',
  student: 'success',
}

const roleTone = (role: string): BadgeTone => roleToneMap[role] ?? 'neutral'

// ---- dashboard per role ----
interface StatItem {
  label: string
  value: string | number
  icon: unknown
}

const loadingStats = ref(false)
const adminStats = ref<Record<string, number> | null>(null)
const teacherStats = ref<Record<string, number> | null>(null)
const studentStats = ref<{
  assigned_exams: number
  completed_exams: number
  passed_exams: number
  average_score: number | null
  best_score: number | null
} | null>(null)

const primaryRole = computed(() => {
  const roles = auth.user?.roles ?? []
  if (roles.includes('admin')) return 'admin'
  if (roles.includes('teacher')) return 'teacher'
  if (roles.includes('student')) return 'student'
  return null
})

function fmtScore(v: number | null | undefined): string {
  return v == null ? '-' : Number(v).toFixed(1)
}

const adminCards = computed<StatItem[]>(() => [
  { label: 'Total Siswa', value: adminStats.value?.total_students ?? '-', icon: Users },
  { label: 'Total Guru', value: adminStats.value?.total_teachers ?? '-', icon: GraduationCap },
  { label: 'Bank Soal', value: adminStats.value?.total_question_banks ?? '-', icon: BookOpenCheck },
  { label: 'Ujian Aktif', value: adminStats.value?.published_exams ?? '-', icon: ClipboardList },
  { label: 'Ujian Berlangsung', value: adminStats.value?.ongoing_exams ?? '-', icon: Activity },
  { label: 'Total Percobaan', value: adminStats.value?.total_attempts ?? '-', icon: Trophy },
])

const teacherCards = computed<StatItem[]>(() => [
  { label: 'Bank Soal Saya', value: teacherStats.value?.total_banks ?? '-', icon: BookOpenCheck },
  { label: 'Bank Dipublikasi', value: teacherStats.value?.published_banks ?? '-', icon: CheckCircle2 },
  { label: 'Total Soal', value: teacherStats.value?.total_questions ?? '-', icon: ClipboardList },
  { label: 'Ujian Dibuat', value: teacherStats.value?.total_exams ?? '-', icon: Activity },
  { label: 'Ujian Aktif', value: teacherStats.value?.published_exams ?? '-', icon: Trophy },
  { label: 'Total Siswa', value: teacherStats.value?.total_students ?? '-', icon: Users },
])

const studentCards = computed<StatItem[]>(() => [
  { label: 'Ujian Ditugaskan', value: studentStats.value?.assigned_exams ?? '-', icon: ClipboardList },
  { label: 'Sudah Dikerjakan', value: studentStats.value?.completed_exams ?? '-', icon: CheckCircle2 },
  { label: 'Lulus', value: studentStats.value?.passed_exams ?? '-', icon: Trophy },
  { label: 'Rata-rata Skor', value: fmtScore(studentStats.value?.average_score), icon: TrendingUp },
  { label: 'Skor Terbaik', value: fmtScore(studentStats.value?.best_score), icon: Target },
])

async function loadStats() {
  if (!primaryRole.value) return
  loadingStats.value = true
  try {
    const res = await api.get(`/dashboard/${primaryRole.value}`)
    const data = res.data?.data
    if (primaryRole.value === 'admin') adminStats.value = data
    else if (primaryRole.value === 'teacher') teacherStats.value = data
    else studentStats.value = data
  } catch {
    // dashboard stat gagal dimuat: biarkan kartu tetap tampil dengan '-'
  } finally {
    loadingStats.value = false
  }
}

onMounted(loadStats)

</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-5xl space-y-6">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Dashboard</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          Selamat datang kembali, {{ auth.user?.name }}.
        </p>
      </div>

      <LoadingState v-if="!auth.user" message="Memuat profil…" />

      <template v-else>
        <!-- statistik sesuai role -->
        <LoadingState v-if="loadingStats" message="Memuat ringkasan…" />

        <template v-else>
          <div
            v-if="primaryRole === 'admin' && adminStats"
            class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3"
          >
            <BaseCard v-for="card in adminCards" :key="card.label" class="p-5">
              <div class="flex items-center gap-2 text-muted-foreground">
                <component :is="card.icon" class="size-4" />
                <p class="text-xs font-semibold uppercase tracking-wide">{{ card.label }}</p>
              </div>
              <p class="mt-3 text-3xl font-bold tracking-tight">{{ card.value }}</p>
            </BaseCard>

            <!-- ekspor dipindah ke menu masing-masing -->
          </div>

          <div
            v-else-if="primaryRole === 'teacher' && teacherStats"
            class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3"
          >
            <BaseCard v-for="card in teacherCards" :key="card.label" class="p-5">
              <div class="flex items-center gap-2 text-muted-foreground">
                <component :is="card.icon" class="size-4" />
                <p class="text-xs font-semibold uppercase tracking-wide">{{ card.label }}</p>
              </div>
              <p class="mt-3 text-3xl font-bold tracking-tight">{{ card.value }}</p>
            </BaseCard>
          </div>

          <div
            v-else-if="primaryRole === 'student' && studentStats"
            class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3"
          >
            <BaseCard v-for="card in studentCards" :key="card.label" class="p-5">
              <div class="flex items-center gap-2 text-muted-foreground">
                <component :is="card.icon" class="size-4" />
                <p class="text-xs font-semibold uppercase tracking-wide">{{ card.label }}</p>
              </div>
              <p class="mt-3 text-3xl font-bold tracking-tight">{{ card.value }}</p>
            </BaseCard>
          </div>
        </template>

        <BaseCard class="relative overflow-hidden p-6">
          <div
            class="pointer-events-none absolute -right-10 -top-10 size-40 rounded-full bg-primary/5 blur-2xl"
          />
          <div class="flex items-start justify-between gap-4">
            <div class="flex items-center gap-4">
              <span
                class="flex size-14 items-center justify-center rounded-2xl bg-primary text-lg font-bold text-primary-foreground shadow-sm"
              >
                {{ auth.user.name.split(' ').map((w) => w[0]).slice(0, 2).join('') }}
              </span>
              <div>
                <p class="text-lg font-semibold">{{ auth.user.name }}</p>
                <p class="text-sm text-muted-foreground">@{{ auth.user.username }}</p>
              </div>
            </div>
            <BaseBadge tone="outline">ID: {{ auth.user.id.slice(0, 8) }}…</BaseBadge>
          </div>

          <div class="mt-5 border-t border-border pt-4">
            <p class="mb-2 flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide text-muted-foreground">
              <ShieldCheck class="size-3.5" />
              Roles
            </p>
            <div class="flex flex-wrap gap-1.5">
              <BaseBadge v-for="role in auth.user.roles" :key="role" :tone="roleTone(role)">
                {{ role }}
              </BaseBadge>
            </div>
          </div>
        </BaseCard>

        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <BaseCard class="p-5">
            <div class="flex items-center gap-2 text-muted-foreground">
              <Activity class="size-4" />
              <p class="text-xs font-semibold uppercase tracking-wide">System Status</p>
            </div>
            <div class="mt-3 flex items-center gap-2">
              <span
                :class="[
                  'size-2.5 rounded-full',
                  health === 'up'
                    ? 'bg-success animate-pulse'
                    : health === 'down'
                      ? 'bg-destructive'
                      : 'bg-muted-foreground animate-pulse',
                ]"
              />
              <p class="font-semibold">
                {{ health === 'up' ? 'Operational' : health === 'down' ? 'Down' : 'Checking…' }}
              </p>
              <BaseBadge v-if="health === 'up'" tone="success" uppercase>UP</BaseBadge>
            </div>
            <BaseButton variant="ghost" size="sm" class="mt-3 px-0" @click="checkHealth">
              Periksa ulang
            </BaseButton>
          </BaseCard>

          <BaseCard class="p-5">
            <div class="flex items-center gap-2 text-muted-foreground">
              <UserRound class="size-4" />
              <p class="text-xs font-semibold uppercase tracking-wide">Profil</p>
            </div>
            <dl class="mt-3 space-y-1.5 text-sm">
              <div class="flex justify-between gap-2">
                <dt class="text-muted-foreground">Username</dt>
                <dd class="font-medium">{{ auth.user.username }}</dd>
              </div>
              <div class="flex justify-between gap-2">
                <dt class="text-muted-foreground">Nama</dt>
                <dd class="max-w-[55%] truncate text-right font-medium">{{ auth.user.name }}</dd>
              </div>
            </dl>
          </BaseCard>
        </div>
      </template>
    </div>
  </AppShell>
</template>
