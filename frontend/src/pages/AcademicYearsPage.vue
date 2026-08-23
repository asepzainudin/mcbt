<script setup lang="ts">
import { Pencil, Plus, Search, Trash2 } from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
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
import type { AcademicYear } from '../types/api'
import { useUiStore } from '../stores/ui'

const ui = useUiStore()

const crud = useCrudTable<AcademicYear>({
  itemLabel: 'Tahun ajaran',
  listFn: (p) => masterDataService.academicYears.list(p),
  createFn: (p) => masterDataService.academicYears.create(p as never),
  updateFn: (id, p) => masterDataService.academicYears.update(id, p as never),
  removeFn: (id) => masterDataService.academicYears.remove(id),
  toPayload: (f) => ({ year: f.year, semester: f.semester }),
  fromItem: (ay) => ({ year: ay.year, semester: ay.semester }),
})

const semesterOptions = [
  { value: 'ODD', label: 'Ganjil (ODD)' },
  { value: 'EVEN', label: 'Genap (EVEN)' },
]

async function activate(ay: AcademicYear) {
  try {
    await masterDataService.academicYears.activate(ay.id)
    ui.toastSuccess(`${ay.year} ${ay.semester} kini aktif.`)
    await crud.fetchList()
  } catch {
    ui.toastError('Gagal mengaktifkan tahun ajaran.')
  }
}
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-5xl space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Tahun Ajaran</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Kelola tahun ajaran & semester. Hanya satu yang boleh aktif.
          </p>
        </div>
        <div class="flex items-end gap-2">
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <input
              v-model="crud.searchInput.value"
              placeholder="Cari tahun…"
              class="h-9 w-48 rounded-lg border border-input bg-transparent pl-9 pr-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
          <BaseButton @click="crud.openCreate({ semester: 'ODD' })">
            <Plus /> Tambah
          </BaseButton>
        </div>
      </div>

      <LoadingState v-if="crud.loading.value" />

      <template v-else>
        <EmptyState
          v-if="crud.items.value.length === 0"
          title="Belum ada tahun ajaran"
          message="Tambahkan tahun ajaran pertama untuk mulai."
        />

        <template v-else>
          <BaseTable>
            <template #head>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Tahun</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Semester</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Status</th>
              <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-muted-foreground">Aksi</th>
            </template>

            <tr v-for="ay in crud.items.value" :key="ay.id" class="border-b border-border transition-colors last:border-0 hover:bg-accent/50">
              <td class="px-4 py-3 font-medium">{{ ay.year }}</td>
              <td class="px-4 py-3">
                <BaseBadge tone="outline">{{ ay.semester === 'ODD' ? 'Ganjil' : 'Genap' }}</BaseBadge>
              </td>
              <td class="px-4 py-3">
                <button v-if="!ay.is_active" class="text-xs font-medium text-primary hover:underline" @click="activate(ay)">
                  Aktifkan
                </button>
                <BaseBadge v-else tone="success" uppercase>aktif</BaseBadge>
              </td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <BaseButton variant="ghost" size="icon" aria-label="Edit" @click="crud.openEdit(ay)">
                    <Pencil />
                  </BaseButton>
                  <BaseButton variant="ghost" size="icon" aria-label="Hapus" :disabled="ay.is_active" @click="crud.askDelete(ay)">
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

      <BaseModal :open="crud.formOpen.value" :title="crud.isEditing.value ? 'Edit Tahun Ajaran' : 'Tambah Tahun Ajaran'" @close="crud.closeForm()">
        <form class="space-y-4" @submit.prevent="crud.submit()">
          <BaseInput v-model="crud.form.value.year" label="Tahun (YYYY/YYYY)" placeholder="2025/2026" required :error="crud.fieldErrors.value.year" />
          <BaseSelect v-model="crud.form.value.semester" label="Semester" :options="semesterOptions" required :error="crud.fieldErrors.value.semester" />
          <div class="flex justify-end gap-2 pt-2">
            <BaseButton variant="outline" type="button" @click="crud.closeForm()">Batal</BaseButton>
            <BaseButton type="submit" :loading="crud.saving.value">Simpan</BaseButton>
          </div>
        </form>
      </BaseModal>

      <BaseModal :open="!!crud.deleteTarget.value" title="Konfirmasi Hapus" @close="crud.cancelDelete()">
        <p class="text-sm text-muted-foreground">
          Hapus tahun ajaran
          <span class="font-semibold text-foreground">{{ crud.deleteTarget.value?.year }} ({{ crud.deleteTarget.value?.semester }})</span>?
          Tindakan ini tidak dapat dibatalkan.
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
