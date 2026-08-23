<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ShieldCheck, Activity, KeyRound, UserRound } from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge, { type BadgeTone } from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseCard from '../components/ui/BaseCard.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { api } from '../lib/axios'
import { useAuthStore } from '../stores/auth'
import { useUiStore } from '../stores/ui'

const auth = useAuthStore()
const ui = useUiStore()

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

function copyTokenHint() {
  ui.toastInfo('Sesi dikelola otomatis lewat cookie — tidak perlu token manual.')
}

const errorSample = () => {
  ui.openErrorModal('Ini contoh modal error global yang muncul saat server mengembalikan status 500.')
}
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

          <BaseCard class="flex flex-col p-5">
            <div class="flex items-center gap-2 text-muted-foreground">
              <KeyRound class="size-4" />
              <p class="text-xs font-semibold uppercase tracking-wide">Sesi & Keamanan</p>
            </div>
            <p class="mt-3 flex-1 text-sm leading-relaxed text-muted-foreground">
              Token disimpan di HttpOnly cookie dan diperbarui otomatis oleh interceptor.
            </p>
            <div class="mt-4 flex flex-wrap gap-2">
              <BaseButton variant="outline" size="sm" @click="copyTokenHint">Detail Sesi</BaseButton>
              <BaseButton variant="secondary" size="sm" @click="errorSample">Test Modal 500</BaseButton>
            </div>
          </BaseCard>
        </div>
      </template>
    </div>
  </AppShell>
</template>
