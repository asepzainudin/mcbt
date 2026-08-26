<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { KeyRound, Loader2, Save, UserRound } from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge, { type BadgeTone } from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseCard from '../components/ui/BaseCard.vue'
import BaseInput from '../components/ui/BaseInput.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { api, apiErrorMessage } from '../lib/axios'
import { useAuthStore } from '../stores/auth'
import { useUiStore } from '../stores/ui'

const auth = useAuthStore()
const ui = useUiStore()

interface ProfileData {
  id: string
  username: string
  name: string
  email: string
  nis: string | null
  class_name: string | null
  nip: string | null
  phone: string | null
}

const loading = ref(true)
const profile = ref<ProfileData | null>(null)

const roleToneMap: Record<string, BadgeTone> = {
  admin: 'primary',
  teacher: 'info',
  student: 'success',
}
const roleTone = (role: string): BadgeTone => roleToneMap[role] ?? 'neutral'

// ---- form profil ----
const savingProfile = ref(false)
const form = reactive({ name: '', phone: '' })

async function loadProfile() {
  loading.value = true
  try {
    const res = await api.get('/auth/profile')
    profile.value = res.data?.data
    form.name = profile.value?.name ?? ''
    form.phone = profile.value?.phone ?? ''
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memuat profil.'))
  } finally {
    loading.value = false
  }
}

onMounted(loadProfile)

async function saveProfile() {
  if (!form.name.trim() || form.name.trim().length < 2) {
    ui.toastError('Nama minimal 2 karakter.')
    return
  }
  savingProfile.value = true
  try {
    const res = await api.put('/auth/profile', {
      name: form.name.trim(),
      phone: form.phone.trim() || null,
    })
    profile.value = res.data?.data
    await auth.bootstrap()
    ui.toastSuccess('Profil berhasil diperbarui.')
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal memperbarui profil.'))
  } finally {
    savingProfile.value = false
  }
}

// ---- form password ----
const savingPassword = ref(false)
const pwForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const pwMismatch = computed(
  () => pwForm.confirm_password.length > 0 && pwForm.new_password !== pwForm.confirm_password,
)

async function savePassword() {
  if (pwForm.new_password.length < 8) {
    ui.toastError('Password baru minimal 8 karakter.')
    return
  }
  if (pwMismatch.value) {
    ui.toastError('Konfirmasi password tidak sama.')
    return
  }
  savingPassword.value = true
  try {
    await api.put('/auth/change-password', { ...pwForm })
    pwForm.old_password = ''
    pwForm.new_password = ''
    pwForm.confirm_password = ''
    ui.toastSuccess('Password berhasil diperbarui.')
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal mengubah password.'))
  } finally {
    savingPassword.value = false
  }
}

const initials = computed(() => {
  const name = profile.value?.name || auth.user?.name || ''
  return name.split(' ').map((w) => w[0]).slice(0, 2).join('').toUpperCase()
})

// kolom telepon tersimpan di profil siswa/guru — sembunyikan untuk admin murni
const hasPhoneField = computed(() => !!profile.value?.nis || !!profile.value?.nip)
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Profil Saya</h1>
        <p class="mt-1 text-sm text-muted-foreground">Kelola informasi akun dan keamanan.</p>
      </div>

      <LoadingState v-if="loading" message="Memuat profil…" />

      <template v-else-if="profile">
        <!-- ringkasan -->
        <BaseCard class="p-6">
          <div class="flex items-start justify-between gap-4">
            <div class="flex items-center gap-4">
              <span
                class="flex size-14 items-center justify-center rounded-2xl bg-primary text-lg font-bold text-primary-foreground shadow-sm"
              >
                {{ initials }}
              </span>
              <div>
                <p class="text-lg font-semibold">{{ profile.name }}</p>
                <p class="text-sm text-muted-foreground">{{ profile.email }}</p>
              </div>
            </div>
          </div>

          <dl class="mt-5 grid gap-x-6 gap-y-2 border-t border-border pt-4 text-sm sm:grid-cols-2">
            <div class="flex justify-between gap-2">
              <dt class="text-muted-foreground">Username</dt>
              <dd class="font-medium">@{{ profile.username }}</dd>
            </div>
            <div v-if="profile.nis" class="flex justify-between gap-2">
              <dt class="text-muted-foreground">NIS</dt>
              <dd class="font-mono font-medium">{{ profile.nis }}</dd>
            </div>
            <div v-if="profile.class_name" class="flex justify-between gap-2">
              <dt class="text-muted-foreground">Kelas</dt>
              <dd class="font-medium">{{ profile.class_name }}</dd>
            </div>
            <div v-if="profile.nip" class="flex justify-between gap-2">
              <dt class="text-muted-foreground">NIP</dt>
              <dd class="font-mono font-medium">{{ profile.nip }}</dd>
            </div>
          </dl>

          <div class="mt-4 flex flex-wrap gap-1.5">
            <BaseBadge v-for="role in auth.user?.roles ?? []" :key="role" :tone="roleTone(role)">
              {{ role }}
            </BaseBadge>
          </div>
        </BaseCard>

        <!-- edit profil -->
        <BaseCard class="p-6">
          <div class="flex items-center gap-2">
            <UserRound class="size-4 text-muted-foreground" />
            <h2 class="text-base font-semibold">Informasi Profil</h2>
          </div>

          <form class="mt-4 grid gap-4 sm:grid-cols-2" @submit.prevent="saveProfile">
            <BaseInput v-model="form.name" label="Nama Lengkap" required />
            <BaseInput v-if="hasPhoneField" v-model="form.phone" label="No. Telepon" placeholder="08xx…" />

            <div class="sm:col-span-2">
              <BaseButton type="submit" :disabled="savingProfile">
                <Loader2 v-if="savingProfile" class="size-4 animate-spin" />
                <Save v-else class="size-4" />
                Simpan Perubahan
              </BaseButton>
            </div>
          </form>
        </BaseCard>

        <!-- ubah password -->
        <BaseCard class="p-6">
          <div class="flex items-center gap-2">
            <KeyRound class="size-4 text-muted-foreground" />
            <h2 class="text-base font-semibold">Ubah Password</h2>
          </div>

          <form class="mt-4 space-y-4" @submit.prevent="savePassword">
            <BaseInput
              v-model="pwForm.old_password"
              type="password"
              label="Password Saat Ini"
              autocomplete="current-password"
              required
            />
            <div class="grid gap-4 sm:grid-cols-2">
              <BaseInput
                v-model="pwForm.new_password"
                type="password"
                label="Password Baru"
                autocomplete="new-password"
                minlength="8"
                required
              />
              <div>
                <BaseInput
                  v-model="pwForm.confirm_password"
                  type="password"
                  label="Konfirmasi Password Baru"
                  autocomplete="new-password"
                  required
                />
                <p v-if="pwMismatch" class="mt-1 text-xs font-medium text-destructive">
                  Konfirmasi tidak sama dengan password baru.
                </p>
              </div>
            </div>

            <BaseButton type="submit" :disabled="savingPassword || pwMismatch">
              <Loader2 v-if="savingPassword" class="size-4 animate-spin" />
              <KeyRound v-else class="size-4" />
              Perbarui Password
            </BaseButton>
          </form>
        </BaseCard>
      </template>
    </div>
  </AppShell>
</template>
