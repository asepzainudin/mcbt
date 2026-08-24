<script setup lang="ts">
import { Pencil, Plus, Search, Trash2, Upload } from 'lucide-vue-next'
import { ref } from 'vue'

import AppShell from '../components/layout/AppShell.vue'
import ImportModal from '../components/admin/ImportModal.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseInput from '../components/ui/BaseInput.vue'
import BaseModal from '../components/ui/BaseModal.vue'
import BasePagination from '../components/ui/BasePagination.vue'
import BaseTable from '../components/ui/BaseTable.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { useCrudTable } from '../composables/useCrudTable'
import { teacherService } from '../services/teacher.service'
import type { Teacher } from '../types/api'

const showImport = ref(false)

const crud = useCrudTable<Teacher>({
  itemLabel: 'Guru',
  defaultLimit: 20,
  listFn: (p) => teacherService.list(p),
  createFn: (p) => teacherService.create(p as never),
  updateFn: (id, p) => teacherService.update(id, p as never),
  removeFn: (id) => teacherService.remove(id),
  toPayload: (f) => ({
    username: f.username,
    name: f.name,
    email: f.email,
    nip: f.nip || null,
    phone: f.phone || null,
    address: f.address || null,
  }),
  fromItem: (t) => ({
    username: t.user?.username ?? '',
    name: t.user?.name ?? '',
    email: t.user?.email ?? '',
    nip: t.nip ?? '',
    phone: t.phone ?? '',
    address: t.address ?? '',
  }),
})
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-5xl space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Data Guru</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Akun guru dibuat dengan password default <code class="font-mono">McBT@1234</code>.
          </p>
        </div>
        <div class="flex flex-wrap items-end gap-2">
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <input
              v-model="crud.searchInput.value"
              placeholder="Cari nama/NIP…"
              class="h-9 w-48 rounded-lg border border-input bg-transparent pl-9 pr-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
          <BaseButton variant="outline" @click="showImport = true">
            <Upload /> Impor
          </BaseButton>
          <BaseButton @click="crud.openCreate()">
            <Plus /> Tambah
          </BaseButton>
        </div>
      </div>

      <LoadingState v-if="crud.loading.value" />

      <template v-else>
        <EmptyState v-if="crud.items.value.length === 0" title="Belum ada guru" message="Tambahkan manual atau impor dari Excel." />

        <template v-else>
          <BaseTable>
            <template #head>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Nama</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Username</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">NIP</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Telepon</th>
              <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-muted-foreground">Aksi</th>
            </template>

            <tr v-for="t in crud.items.value" :key="t.id" class="border-b border-border transition-colors last:border-0 hover:bg-accent/50">
              <td class="px-4 py-3">
                <p class="font-medium">{{ t.user?.name }}</p>
                <p class="text-xs text-muted-foreground">{{ t.user?.email }}</p>
              </td>
              <td class="px-4 py-3 font-mono text-xs">{{ t.user?.username }}</td>
              <td class="px-4 py-3 font-mono text-xs text-muted-foreground">{{ t.nip ?? '—' }}</td>
              <td class="px-4 py-3 text-muted-foreground">{{ t.phone ?? '—' }}</td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <BaseButton variant="ghost" size="icon" aria-label="Edit" @click="crud.openEdit(t)">
                    <Pencil />
                  </BaseButton>
                  <BaseButton variant="ghost" size="icon" aria-label="Hapus" @click="crud.askDelete(t)">
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

      <BaseModal :open="crud.formOpen.value" :title="crud.isEditing.value ? 'Edit Guru' : 'Tambah Guru'" @close="crud.closeForm()">
        <form class="space-y-4" @submit.prevent="crud.submit()">
          <BaseInput v-model="crud.form.value.username" label="Username" :disabled="crud.isEditing.value" required :error="crud.fieldErrors.value.username" />
          <BaseInput v-model="crud.form.value.name" label="Nama Lengkap" required :error="crud.fieldErrors.value.name" />
          <BaseInput v-model="crud.form.value.email" type="email" label="Email" required :error="crud.fieldErrors.value.email" />
          <BaseInput v-model="crud.form.value.nip" label="NIP (opsional)" :error="crud.fieldErrors.value.nip" />
          <BaseInput v-model="crud.form.value.phone" label="Telepon (opsional)" :error="crud.fieldErrors.value.phone" />
          <div class="flex justify-end gap-2 pt-2">
            <BaseButton variant="outline" type="button" @click="crud.closeForm()">Batal</BaseButton>
            <BaseButton type="submit" :loading="crud.saving.value">Simpan</BaseButton>
          </div>
        </form>
      </BaseModal>

      <ImportModal
        :open="showImport"
        upload-path="/teachers/import"
        title="Impor Data Guru"
        :template-url="teacherService.templateUrl"
        @close="showImport = false"
        @imported="crud.fetchList()"
      />

      <BaseModal :open="!!crud.deleteTarget.value" title="Konfirmasi Hapus" @close="crud.cancelDelete()">
        <p class="text-sm text-muted-foreground">
          Hapus guru <span class="font-semibold text-foreground">{{ crud.deleteTarget.value?.user?.name }}</span>?
          Akun login guru juga akan dihapus.
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
