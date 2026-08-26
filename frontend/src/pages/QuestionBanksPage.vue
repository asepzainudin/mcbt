<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  Copy,
  FileQuestion,
  Pencil,
  Plus,
  Search,
  Send,
  Archive,
  Trash2,
} from 'lucide-vue-next'

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
import { bankService } from '../services/question.service'
import { masterDataService } from '../services/master-data.service'
import type { BankStatus, QuestionBank } from '../types/api'
import { useUiStore } from '../stores/ui'

const ui = useUiStore()
const router = useRouter()

const subjects = ref<{ id: string; name: string; code: string }[]>([])
const subjectFilter = ref('')

onMounted(async () => {
  const res = await masterDataService.subjects.list({ page: 1, limit: 100 })
  subjects.value = res.data
})

const subjectOptions = computed(() => [
  { value: '', label: 'Semua mapel' },
  ...subjects.value.map((s) => ({ value: s.id, label: `${s.code} — ${s.name}` })),
])

const formSubjectOptions = computed(() => [
  { value: '', label: '— Pilih mapel —', disabled: true },
  ...subjects.value.map((s) => ({ value: s.id, label: `${s.code} — ${s.name}` })),
])

const subjectName = (id: string) => {
  const s = subjects.value.find((x) => x.id === id)
  return s ? `${s.code} — ${s.name}` : id
}

const extraParams = computed(() => ({ subject_id: subjectFilter.value || undefined }))

const crud = useCrudTable<QuestionBank>({
  itemLabel: 'Bank soal',
  defaultLimit: 10,
  extraParams,
  listFn: (p) => bankService.list(p),
  createFn: (p) => bankService.create(p as never),
  updateFn: (id, p) => bankService.update(id, p as never),
  removeFn: (id) => bankService.remove(id),
  toPayload: (f) => ({
    code: f.code,
    title: f.title,
    subject_id: f.subject_id,
    description: f.description || null,
  }),
  fromItem: (b) => ({
    code: b.code,
    title: b.title,
    subject_id: b.subject_id,
    description: b.description ?? '',
  }),
  validate: (f) => {
    const e: Record<string, string> = {}
    if (!f.code.trim()) e.code = 'Code wajib diisi'
    if (!f.title.trim()) e.title = 'Judul wajib diisi'
    if (!f.subject_id) e.subject_id = 'Pilih mata pelajaran'
    return e
  },
})

watch(subjectFilter, () => {
  crud.page.value = 1
  crud.fetchList()
})

const statusTone = (s: BankStatus) =>
  s === 'published' ? 'success' : s === 'archived' ? 'warning' : 'neutral'
const statusLabel = (s: BankStatus) =>
  ({ draft: 'Draft', published: 'Published', archived: 'Archived' })[s] ?? s

const fmtDate = (d?: string | null) =>
  d ? new Date(d).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' }) : '—'

// ---- clone / publish / archive ----
const busyId = ref<string | null>(null)

async function withBusy(id: string, fn: () => Promise<unknown>, okMsg: string) {
  busyId.value = id
  try {
    await fn()
    ui.toastSuccess(okMsg)
    await crud.fetchList()
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Aksi gagal.'))
  } finally {
    busyId.value = null
  }
}

import { apiErrorMessage } from '../lib/axios'
import { watch } from 'vue'
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-5xl space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">Bank Soal</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Kelola bank soal: clone, publish, dan arsipkan.
          </p>
        </div>
        <div class="flex flex-wrap items-end gap-2">
          <div class="w-52">
            <BaseSelect v-model="subjectFilter" :options="subjectOptions" />
          </div>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <input
              v-model="crud.searchInput.value"
              placeholder="Cari judul / code…"
              class="h-9 w-48 rounded-lg border border-input bg-transparent pl-9 pr-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
          <BaseButton @click="crud.openCreate()"><Plus /> Tambah</BaseButton>
        </div>
      </div>

      <LoadingState v-if="crud.loading.value" />

      <template v-else>
        <EmptyState v-if="crud.items.value.length === 0" title="Belum ada bank soal" message="Buat bank soal untuk mulai menambahkan soal." />

        <template v-else>
          <BaseTable>
            <template #head>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Bank</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Mapel</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Status</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Dibuat</th>
              <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-muted-foreground">Aksi</th>
            </template>

            <tr v-for="b in crud.items.value" :key="b.id" class="border-b border-border transition-colors last:border-0 hover:bg-accent/50">
              <td class="px-4 py-3">
                <router-link :to="`/question-banks/${b.id}`" class="flex items-center gap-2 font-medium text-primary hover:underline">
                  <FileQuestion class="size-4 shrink-0" />
                  {{ b.title }}
                </router-link>
                <p class="mt-0.5 font-mono text-xs text-muted-foreground">{{ b.code }}</p>
              </td>
              <td class="px-4 py-3">
                <BaseBadge tone="info">{{ b.subject?.code ?? subjectName(b.subject_id) }}</BaseBadge>
              </td>
              <td class="px-4 py-3">
                <BaseBadge :tone="statusTone(b.status)">{{ statusLabel(b.status) }}</BaseBadge>
              </td>
              <td class="px-4 py-3 text-sm text-muted-foreground">{{ fmtDate(b.created_at) }}</td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <BaseButton
                    variant="ghost"
                    size="icon"
                    title="Kelola soal"
                    @click="router.push(`/question-banks/${b.id}`)"
                  >
                    <FileQuestion />
                  </BaseButton>
                  <BaseButton
                    v-if="b.status === 'draft'"
                    variant="ghost"
                    size="icon"
                    title="Publish"
                    :disabled="busyId === b.id"
                    @click="withBusy(b.id, () => bankService.publish(b.id), 'Bank soal dipublikasikan.')"
                  >
                    <Send />
                  </BaseButton>
                  <BaseButton
                    variant="ghost"
                    size="icon"
                    title="Clone"
                    :disabled="busyId === b.id"
                    @click="withBusy(b.id, () => bankService.clone(b.id), 'Bank soal berhasil dikloning.')"
                  >
                    <Copy />
                  </BaseButton>
                  <BaseButton
                    v-if="b.status !== 'archived'"
                    variant="ghost"
                    size="icon"
                    title="Arsipkan"
                    :disabled="busyId === b.id"
                    @click="withBusy(b.id, () => bankService.archive(b.id), 'Bank soal diarsipkan.')"
                  >
                    <Archive />
                  </BaseButton>
                  <BaseButton variant="ghost" size="icon" aria-label="Edit" @click="crud.openEdit(b)">
                    <Pencil />
                  </BaseButton>
                  <BaseButton variant="ghost" size="icon" aria-label="Hapus" @click="crud.askDelete(b)">
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

      <BaseModal :open="crud.formOpen.value" :title="crud.isEditing.value ? 'Edit Bank Soal' : 'Tambah Bank Soal'" @close="crud.closeForm()">
        <form class="space-y-4" @submit.prevent="crud.submit()">
          <p v-if="subjects.length === 0" class="rounded-lg border border-dashed border-border p-3 text-sm text-muted-foreground">
            Belum ada mata pelajaran. Buat dulu di menu
            <router-link to="/subjects" class="font-medium text-primary hover:underline">Mata Pelajaran</router-link>.
          </p>
          <BaseInput v-model="crud.form.value.code" label="Code" placeholder="BANK-MTK" required :error="crud.fieldErrors.value.code" />
          <BaseInput v-model="crud.form.value.title" label="Judul" placeholder="Bank MTK XII" required :error="crud.fieldErrors.value.title" />
          <BaseSelect v-model="crud.form.value.subject_id" label="Mata Pelajaran" :options="formSubjectOptions" required :error="crud.fieldErrors.value.subject_id" />
          <BaseInput v-model="crud.form.value.description" label="Deskripsi (opsional)" :error="crud.fieldErrors.value.description" />
          <div class="flex justify-end gap-2 pt-2">
            <BaseButton variant="outline" type="button" @click="crud.closeForm()">Batal</BaseButton>
            <BaseButton type="submit" :loading="crud.saving.value">Simpan</BaseButton>
          </div>
        </form>
      </BaseModal>

      <BaseModal :open="!!crud.deleteTarget.value" title="Konfirmasi Hapus" @close="crud.cancelDelete()">
        <p class="text-sm text-muted-foreground">
          Hapus bank soal <span class="font-semibold text-foreground">{{ crud.deleteTarget.value?.title }}</span>?
          Semua soal di dalamnya juga akan terhapus.
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
