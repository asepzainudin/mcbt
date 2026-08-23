<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Pencil, Plus, Search, Trash2 } from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
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
import type { AcademicYear, SchoolClass } from '../types/api'

const academicYears = ref<AcademicYear[]>([])
const yearFilter = ref('')

onMounted(async () => {
  const res = await masterDataService.academicYears.list({ page: 1, limit: 100 })
  academicYears.value = res.data
})

const yearOptions = computed(() => [
  { value: '', label: 'Semua tahun ajaran' },
  ...academicYears.value.map((ay) => ({
    value: ay.id,
    label: `${ay.year} — ${ay.semester === 'ODD' ? 'Ganjil' : 'Genap'}${ay.is_active ? ' (aktif)' : ''}`,
  })),
])

const extraParams = computed(() => ({
  academic_year_id: yearFilter.value || undefined,
}))

const crud = useCrudTable<SchoolClass>({
  resource: 'classes',
  itemLabel: 'Kelas',
  extraParams,
  createFn: (p) => masterDataService.classes.create(p as never),
  updateFn: (id, p) => masterDataService.classes.update(id, p as never),
  removeFn: (id) => masterDataService.classes.remove(id),
  toPayload: (f) => ({ name: f.name, academic_year_id: f.academic_year_id }),
  fromItem: (c) => ({
    name: c.name,
    academic_year_id: c.academic_year_id,
  }),
})

watch(yearFilter, () => {
  crud.page.value = 1
  crud.fetchList()
})
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-5xl space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Kelas</h1>
          <p class="mt-1 text-sm text-muted-foreground">Kelola kelas per tahun ajaran.</p>
        </div>
        <div class="flex flex-wrap items-end gap-2">
          <div class="w-56">
            <BaseSelect v-model="yearFilter" :options="yearOptions" />
          </div>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <input
              v-model="crud.searchInput.value"
              placeholder="Cari kelas…"
              class="h-9 w-44 rounded-lg border border-input bg-transparent pl-9 pr-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
          <BaseButton
            :disabled="academicYears.length === 0"
            @click="crud.openCreate({ academic_year_id: yearFilter || academicYears[0]?.id || '' })"
          >
            <Plus /> Tambah
          </BaseButton>
        </div>
      </div>

      <p v-if="academicYears.length === 0" class="rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
        Buat tahun ajaran terlebih dahulu di menu Tahun Ajaran.
      </p>

      <LoadingState v-if="crud.loading.value" />

      <template v-else>
        <EmptyState
          v-if="crud.items.value.length === 0"
          title="Belum ada kelas"
          message="Tambahkan kelas untuk tahun ajaran yang dipilih."
        />

        <template v-else>
          <BaseTable>
            <template #head>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Nama</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Tahun Ajaran</th>
              <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-muted-foreground">Aksi</th>
            </template>

            <tr v-for="c in crud.items.value" :key="c.id" class="border-b border-border transition-colors last:border-0 hover:bg-accent/50">
              <td class="px-4 py-3 font-medium">{{ c.name }}</td>
              <td class="px-4 py-3 text-muted-foreground">
                {{ c.academic_year?.year }} · {{ c.academic_year?.semester === 'ODD' ? 'Ganjil' : 'Genap' }}
              </td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <BaseButton variant="ghost" size="icon" aria-label="Edit" @click="crud.openEdit(c)">
                    <Pencil />
                  </BaseButton>
                  <BaseButton variant="ghost" size="icon" aria-label="Hapus" @click="crud.askDelete(c)">
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

      <BaseModal :open="crud.formOpen.value" :title="crud.isEditing.value ? 'Edit Kelas' : 'Tambah Kelas'" @close="crud.closeForm()">
        <form class="space-y-4" @submit.prevent="crud.submit()">
          <BaseInput v-model="crud.form.value.name" label="Nama Kelas" placeholder="XII IPA 1" :error="crud.fieldErrors.value.name" />
          <BaseSelect
            v-model="crud.form.value.academic_year_id"
            label="Tahun Ajaran"
            :options="yearOptions.slice(1)"
            :error="crud.fieldErrors.value.academic_year_id"
          />
          <div class="flex justify-end gap-2 pt-2">
            <BaseButton variant="outline" type="button" @click="crud.closeForm()">Batal</BaseButton>
            <BaseButton type="submit" :loading="crud.saving.value">Simpan</BaseButton>
          </div>
        </form>
      </BaseModal>

      <BaseModal :open="!!crud.deleteTarget.value" title="Konfirmasi Hapus" @close="crud.cancelDelete()">
        <p class="text-sm text-muted-foreground">
          Hapus kelas
          <span class="font-semibold text-foreground">{{ crud.deleteTarget.value?.name }}</span>?
          Siswa di kelas ini akan terlepas dari kelas.
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
