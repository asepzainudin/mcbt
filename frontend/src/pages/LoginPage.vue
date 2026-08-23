<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { GraduationCap, ShieldCheck, Zap, MonitorSmartphone } from 'lucide-vue-next'

import BaseButton from '../components/ui/BaseButton.vue'
import BaseInput from '../components/ui/BaseInput.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const username = ref('')
const password = ref('')
const errors = ref<Record<string, string>>({})
const formError = ref('')

async function onSubmit() {
  errors.value = {}
  formError.value = ''

  if (!username.value.trim()) errors.value.username = 'Username wajib diisi'
  if (!password.value) errors.value.password = 'Password wajib diisi'
  if (Object.keys(errors.value).length > 0) return

  const result = await auth.login(username.value.trim(), password.value)

  if (result.ok) {
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    router.push(redirect)
    return
  }

  for (const [field, message] of Object.entries(result.fieldErrors ?? {})) {
    if (['username', 'password'].includes(field)) {
      errors.value[field] = message
    } else {
      formError.value = message
    }
  }
}

const highlights = [
  { icon: ShieldCheck, title: 'Aman', desc: 'Cookie HttpOnly & CSRF protection' },
  { icon: Zap, title: 'Cepat', desc: 'Auto refresh token tanpa putus sesi' },
  { icon: MonitorSmartphone, title: 'Responsif', desc: 'Nyaman di desktop maupun mobile' },
]
</script>

<template>
  <div class="flex min-h-screen bg-background">
    <div
      class="relative hidden w-1/2 flex-col justify-between overflow-hidden bg-primary p-12 text-primary-foreground lg:flex"
    >
      <div
        class="pointer-events-none absolute -left-24 -top-24 size-96 rounded-full bg-white/10 blur-2xl"
      />
      <div
        class="pointer-events-none absolute -bottom-32 -right-16 size-[28rem] rounded-full bg-black/10 blur-2xl"
      />

      <div class="relative flex items-center gap-3">
        <span class="flex size-10 items-center justify-center rounded-xl bg-primary-foreground/15 backdrop-blur">
          <GraduationCap class="size-6" />
        </span>
        <span class="text-2xl font-bold tracking-tight">MCBT</span>
      </div>

      <div class="relative space-y-4">
        <h1 class="text-4xl font-bold leading-tight">
          Computer Based Test<br />yang modern & andal.
        </h1>
        <p class="max-w-md text-primary-foreground/80">
          Platform ujian berbasis komputer dengan manajemen bank soal, kelas,
          dan tahun akademik dalam satu tempat.
        </p>
      </div>

      <div class="relative grid grid-cols-3 gap-4">
        <div
          v-for="h in highlights"
          :key="h.title"
          class="rounded-xl border border-primary-foreground/15 bg-primary-foreground/10 p-4 backdrop-blur"
        >
          <component :is="h.icon" class="mb-2 size-5" />
          <p class="text-sm font-semibold">{{ h.title }}</p>
          <p class="mt-0.5 text-xs leading-relaxed text-primary-foreground/70">{{ h.desc }}</p>
        </div>
      </div>
    </div>

    <div class="flex w-full items-center justify-center px-6 lg:w-1/2">
      <div class="w-full max-w-sm">
        <div class="mb-8 lg:hidden">
          <span class="inline-flex items-center gap-2 rounded-lg bg-primary px-3 py-2 text-primary-foreground">
            <GraduationCap class="size-5" />
            <span class="font-bold">MCBT</span>
          </span>
        </div>

        <h2 class="text-2xl font-bold tracking-tight">Selamat datang 👋</h2>
        <p class="mt-1 text-sm text-muted-foreground">Masuk untuk mengakses dashboard.</p>

        <form
          class="mt-8 space-y-5 rounded-2xl border border-border bg-card p-6 shadow-sm"
          novalidate
          @submit.prevent="onSubmit"
        >
          <p
            v-if="formError"
            class="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
          >
            {{ formError }}
          </p>

        <BaseInput
          v-model="username"
          label="Username"
          placeholder="username atau email"
          autocomplete="username"
          required
          :error="errors.username"
        />

        <BaseInput
          v-model="password"
          type="password"
          label="Password"
          placeholder="••••••••"
          autocomplete="current-password"
          required
          :error="errors.password"
        />

          <BaseButton type="submit" block :loading="auth.loading" size="lg">Masuk</BaseButton>
        </form>

        <div class="mt-4 rounded-lg border border-dashed border-border p-3 text-center text-xs text-muted-foreground">
          Demo: <code class="font-mono text-foreground">admin123</code> /
          <code class="font-mono text-foreground">Admin@123</code>
        </div>
      </div>
    </div>
  </div>
</template>
