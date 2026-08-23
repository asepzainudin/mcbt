<script setup lang="ts">
import { Pencil, Plus, Search, Trash2 } from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseInput from '../components/ui/BaseInput.vue'
import BaseModal from '../components/ui/BaseModal.vue'
import BasePagination from '../components/ui/BasePagination.vue'
import BaseTable from '../components/ui/BaseTable.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { useCrudTable } from '../composables/useCrudTable'
import { masterDataService } from '../services/master-data.service'
import type { Subject } from '../types/api'

const crud = useCrudTable<Subject>({
  resource: 'subjects',
  itemLabel: 'Mapel',
  createFn: (p) => masterDataService.subjects.create(p as never),
  updateFn: (id, p) => masterDataService.subjects.update(id, p as never),
  removeFn: (id) => masterDataService.subjects.remove(id),
  toPayload: (f) => ({
    code: f.code,
    name: f.name,
    description: f.description || undefined,
  }),
  fromItem: (s) => ({
    code: s.code,
    name: s.name,
    description: s.description ?? '',
  }),
})
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-5xl space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Mata Pelajaran</h1>
          <p class="mt-1 text-sm text-muted-foreground">Kelola daftar mapel dan kodenya.</p>
        </div>
        <div class="flex items-end gap-2">
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <input
              v-model="crud.searchInput.value"
              placeholder="Cari kode / nama…"
              class="h-9 w-52 rounded-lg border border-input bg-transparent pl-9 pr-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
          <BaseButton @click="crud.openCreate()">
            <Plus /> Tambah
          </BaseButton>
        </div>
      </div>

      <LoadingState v-if="crud.loading.value" />

      <template v-else>
        <EmptyState
          v-if="crud.items.value.length === 0"
          title="Belum ada mapel"
          message="Tambahkan mata pelajaran pertama."
        />

        <template v-else>
          <BaseTable>
            <template #head>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Kode</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Nama</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Deskripsi</th>
              <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-muted-foreground">Aksi</th>
            </template>

            <tr v-for="s in crud.items.value" :key="s.id" class="border-b border-border transition-colors last:border-0 hover:bg-accent/50">
              <td class="px-4 py-3">
                <span class="rounded-md bg-primary/10 px-2 py-0.5 font-mono text-xs font-semibold text-primary">
                  {{ s.code }}
                </span>
              </td>
              <td class="px-4 py-3 font-medium">{{ s.name }}</td>
              <td class="max-w-xs truncate px-4 py-3 text-muted-foreground">{{ s.description ?? '—' }}</td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
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

      <BaseModal :open="crud.formOpen.value" :title="crud.isEditing.value ? 'Edit Mapel' : 'Tambah Mapel'" @close="crud.closeForm()">
        <form class="space-y-4" @submit.prevent="crud.submit()">
          <BaseInput v-model="crud.form.value.code" label="Kode" placeholder="MTK-X" :error="crud.fieldErrors.value.code" />
          <BaseInput v-model="crud.form.value.name" label="Nama Mapel" placeholder="Matematika" :error="crud.fieldErrors.value.name" />
          <BaseInput
            v-model="crud.form.value.description"
            label="Deskripsi (opsional)"
            placeholder="Keterangan singkat…"
            :error="crud.fieldErrors.value.description"
          />
          <div class="flex justify-end gap-2 pt-2">
            <BaseButton variant="outline" type="button" @click="crud.closeForm()">Batal</BaseButton>
            <BaseButton type="submit" :loading="crud.saving.value">Simpan</BaseButton>
          </div>
        </form>
      </BaseModal>

      <BaseModal :open="!!crud.deleteTarget.value" title="Konfirmasi Hapus" @close="crud.cancelDelete()">
        <p class="text-sm text-muted-foreground">
          Hapus mapel
          <span class="font-semibold text-foreground">{{ crud.deleteTarget.value?.name }} ({{ crud.deleteTarget.value?.code }})</span>?
          Mapel yang dipakai bank soal tidak dapat dihapus.
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
