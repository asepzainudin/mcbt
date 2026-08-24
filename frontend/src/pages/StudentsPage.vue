<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import {
  ArrowRightLeft,
  KeyRound,
  Pencil,
  Plus,
  Search,
  Trash2,
  Upload,
} from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import ImportModal from '../components/admin/ImportModal.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseInput from '../components/ui/BaseInput.vue'
import BaseModal from '../components/ui/BaseModal.vue'
import BasePagination from '../components/ui/BasePagination.vue'
import BaseSelect from '../components/ui/BaseSelect.vue'
import BaseTable from '../components/ui/BaseTable.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { useCrudTable } from '../composables/useCrudTable'
import { masterDataService } from '../services/master-data.service'
import { studentService } from '../services/student.service'
import type { SchoolClass, Student } from '../types/api'
import { useUiStore } from '../stores/ui'

const ui = useUiStore()

const classes = ref<SchoolClass[]>([])
const classFilter = ref('')
const showImport = ref(false)

onMounted(async () => {
  const res = await masterDataService.classes.list({ page: 1, limit: 100 })
  classes.value = res.data
})

const classOptions = computed(() => [
  { value: '', label: 'Semua kelas' },
  ...classes.value.map((c) => ({
    value: c.id,
    label: `${c.name} — ${c.academic_year?.year ?? ''}`,
  })),
])

const extraParams = computed(() => ({ class_id: classFilter.value || undefined }))

const crud = useCrudTable<Student>({
  itemLabel: 'Siswa',
  defaultLimit: 20,
  extraParams,
  listFn: (p) => studentService.list(p),
  createFn: (p) => studentService.create(p as never),
  updateFn: (id, p) => studentService.update(id, p as never),
  removeFn: (id) => studentService.remove(id),
  toPayload: (f) => ({
    username: f.username,
    name: f.name,
    email: f.email,
    nis: f.nis,
    class_id: f.class_id || null,
    phone: f.phone || null,
    address: f.address || null,
  }),
  fromItem: (s) => ({
    username: s.user?.username ?? '',
    name: s.user?.name ?? '',
    email: s.user?.email ?? '',
    nis: s.nis,
    class_id: s.class_id ?? '',
    phone: s.phone ?? '',
    address: s.address ?? '',
  }),
})

watch(classFilter, () => {
  crud.page.value = 1
  crud.fetchList()
})

// ---- pindah kelas ----
const moveTarget = ref<Student | null>(null)
const moveClassId = ref('')
const moving = ref(false)

function openMove(s: Student) {
  moveTarget.value = s
  moveClassId.value = s.class_id ?? ''
}

async function submitMove() {
  if (!moveTarget.value || !moveClassId.value) return
  moving.value = true
  try {
    await studentService.changeClass(moveTarget.value.id, moveClassId.value)
    ui.toastSuccess('Kelas siswa diperbarui.')
    moveTarget.value = null
    await crud.fetchList()
  } catch {
    ui.toastError('Gagal memindahkan kelas.')
  } finally {
    moving.value = false
  }
}

// ---- reset password ----
const resetTarget = ref<Student | null>(null)
const resetting = ref(false)
const newPassword = ref('')

function openReset(s: Student) {
  resetTarget.value = s
  newPassword.value = ''
}

async function submitReset() {
  if (!resetTarget.value) return
  resetting.value = true
  try {
    const res = await studentService.resetPassword(resetTarget.value.id)
    newPassword.value = res.new_password
    ui.toastSuccess('Password siswa direset.')
  } catch {
    ui.toastError('Gagal mereset password.')
  } finally {
    resetting.value = false
  }
}

async function copyPassword() {
  try {
    await navigator.clipboard.writeText(newPassword.value)
    ui.toastSuccess('Password disalin ke clipboard.')
  } catch {
    ui.toastWarning('Tidak dapat menyalin — salin manual.')
  }
}

function closeReset() {
  resetTarget.value = null
  newPassword.value = ''
}
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-5xl space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Data Siswa</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Password default <code class="font-mono">McBT@1234</code> — gunakan Reset Password bila perlu.
          </p>
        </div>
        <div class="flex flex-wrap items-end gap-2">
          <div class="w-52">
            <BaseSelect v-model="classFilter" :options="classOptions" />
          </div>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <input
              v-model="crud.searchInput.value"
              placeholder="Cari nama/NIS…"
              class="h-9 w-44 rounded-lg border border-input bg-transparent pl-9 pr-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
          <BaseButton variant="outline" @click="showImport = true"><Upload /> Impor</BaseButton>
          <BaseButton @click="crud.openCreate({ class_id: classFilter })"><Plus /> Tambah</BaseButton>
        </div>
      </div>

      <LoadingState v-if="crud.loading.value" />

      <template v-else>
        <EmptyState v-if="crud.items.value.length === 0" title="Belum ada siswa" message="Tambahkan manual atau impor dari Excel." />

        <template v-else>
          <BaseTable>
            <template #head>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">NIS</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Nama</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Kelas</th>
              <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-muted-foreground">Aksi</th>
            </template>

            <tr v-for="s in crud.items.value" :key="s.id" class="border-b border-border transition-colors last:border-0 hover:bg-accent/50">
              <td class="px-4 py-3 font-mono text-xs">{{ s.nis }}</td>
              <td class="px-4 py-3">
                <p class="font-medium">{{ s.user?.name }}</p>
                <p class="text-xs text-muted-foreground">@{{ s.user?.username }}</p>
              </td>
              <td class="px-4 py-3">
                <BaseBadge v-if="s.class" tone="outline">{{ s.class.name }}</BaseBadge>
                <span v-else class="text-xs text-muted-foreground">—</span>
              </td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <BaseButton variant="ghost" size="icon" aria-label="Pindah kelas" @click="openMove(s)">
                    <ArrowRightLeft />
                  </BaseButton>
                  <BaseButton variant="ghost" size="icon" aria-label="Reset password" @click="openReset(s)">
                    <KeyRound />
                  </BaseButton>
                  <BaseButton variant="ghost" size="icon" aria-label="Edit" @click="crud.openEdit(s)">
                    <Pencil />
                  </BaseButton>
                  <BaseButton variant="ghost" size="icon" aria-label="Hapus" @click="crud.askDelete(s)">
                    <Trash2 class="text-destructive" />
                  </BaseButton>
                </div>
              </td>
            </tr>
          </BaseTable>

          <div class="flex flex-col items-center gap-3 sm:flex-row sm:justify-between">
            <p class="text-sm text-muted-foreground">{{ crud.rangeText.value }}</p>
            <BasePagination v-if="crud.meta.value" :page="crud.meta.value.page" :total-pages="crud.totalPages.value" @change="crud.goTo" />
          </div>
        </template>
      </template>

      <ImportModal
        :open="showImport"
        upload-path="/students/import"
        title="Impor Data Siswa"
        :template-url="studentService.templateUrl"
        @close="showImport = false"
        @imported="crud.fetchList()"
      />

      <BaseModal :open="crud.formOpen.value" :title="crud.isEditing.value ? 'Edit Siswa' : 'Tambah Siswa'" @close="crud.closeForm()">
        <form class="space-y-4" @submit.prevent="crud.submit()">
          <BaseInput v-model="crud.form.value.username" label="Username" :disabled="crud.isEditing.value" required :error="crud.fieldErrors.value.username" />
          <BaseInput v-model="crud.form.value.name" label="Nama Lengkap" required :error="crud.fieldErrors.value.name" />
          <BaseInput v-model="crud.form.value.email" type="email" label="Email" required :error="crud.fieldErrors.value.email" />
          <BaseInput v-model="crud.form.value.nis" label="NIS" required :error="crud.fieldErrors.value.nis" />
          <BaseSelect v-model="crud.form.value.class_id" label="Kelas (opsional)" :options="classOptions.slice(1)" :error="crud.fieldErrors.value.class_id" />
          <BaseInput v-model="crud.form.value.phone" label="Telepon (opsional)" :error="crud.fieldErrors.value.phone" />
          <div class="flex justify-end gap-2 pt-2">
            <BaseButton variant="outline" type="button" @click="crud.closeForm()">Batal</BaseButton>
            <BaseButton type="submit" :loading="crud.saving.value">Simpan</BaseButton>
          </div>
        </form>
      </BaseModal>

      <BaseModal :open="!!moveTarget" title="Pindah Kelas" @close="moveTarget = null">
        <p class="mb-4 text-sm text-muted-foreground">
          Pindahkan <span class="font-medium text-foreground">{{ moveTarget?.user?.name }}</span> ke kelas:
        </p>
        <BaseSelect v-model="moveClassId" :options="classOptions.slice(1)" />
        <template #footer>
          <BaseButton variant="outline" @click="moveTarget = null">Batal</BaseButton>
          <BaseButton :disabled="!moveClassId" :loading="moving" @click="submitMove">Pindahkan</BaseButton>
        </template>
      </BaseModal>

      <BaseModal :open="!!resetTarget" title="Reset Password Siswa" @close="closeReset">
        <div v-if="!newPassword" class="space-y-4">
          <p class="text-sm text-muted-foreground">
            Reset password <span class="font-medium text-foreground">{{ resetTarget?.user?.name }}</span>?
            Password lama tidak berlaku dan sesi aktif akan dicabut.
          </p>
        </div>
        <div v-else class="space-y-3">
          <p class="text-sm text-muted-foreground">
            Password baru untuk <span class="font-medium text-foreground">{{ resetTarget?.user?.name }}</span>:
          </p>
          <div class="flex items-center justify-between rounded-lg border border-primary/40 bg-primary/5 px-4 py-3">
            <code class="font-mono text-lg font-bold tracking-wider text-primary">{{ newPassword }}</code>
            <BaseButton variant="outline" size="sm" @click="copyPassword">Salin</BaseButton>
          </div>
          <p class="text-xs text-muted-foreground">Catat dan bagikan kepada siswa — tidak akan ditampilkan lagi.</p>
        </div>

        <template #footer>
          <template v-if="!newPassword">
            <BaseButton variant="outline" @click="closeReset">Batal</BaseButton>
            <BaseButton :loading="resetting" @click="submitReset">Reset</BaseButton>
          </template>
          <BaseButton v-else @click="closeReset">Selesai</BaseButton>
        </template>
      </BaseModal>

      <BaseModal :open="!!crud.deleteTarget.value" title="Konfirmasi Hapus" @close="crud.cancelDelete()">
        <p class="text-sm text-muted-foreground">
          Hapus siswa <span class="font-semibold text-foreground">{{ crud.deleteTarget.value?.user?.name }} ({{ crud.deleteTarget.value?.nis }})</span>?
          Akun login siswa juga akan dihapus.
        </p>
        <template #footer>
          <BaseButton variant="outline" @click="crud.cancelDelete()">Batal</BaseButton>
          <BaseButton variant="destructive" :loading="crud.deleting.value" @click="crud.confirmDelete()">
            <Trash2 /> Hapus
          </BaseButton>
        </template>
      </BaseModal>
    </div>
  </AppShell>
</template>
