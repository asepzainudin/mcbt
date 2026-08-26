<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ListTree,
  Pencil,
  Plus,
  Trash2,
  Unlink,
} from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import BaseBadge from '../components/ui/BaseBadge.vue'
import BaseButton from '../components/ui/BaseButton.vue'
import BaseInput from '../components/ui/BaseInput.vue'
import BaseModal from '../components/ui/BaseModal.vue'
import EmptyState from '../components/ui/EmptyState.vue'
import LoadingState from '../components/ui/LoadingState.vue'
import { apiErrorMessage } from '../lib/axios'
import { sectionService } from '../services/section.service'
import { bankService } from '../services/question.service'
import type { ExamSection, QuestionBank, SectionQuestion } from '../types/api'
import { useUiStore } from '../stores/ui'

const ui = useUiStore()
const route = useRoute()
const router = useRouter()

const examId = route.params.id as string

const sections = ref<ExamSection[]>([])
const loading = ref(true)

const banks = ref<QuestionBank[]>([])
const selectedBanks = ref<Record<string, boolean>>({})
const totalRandom = ref('')
const bankSearch = ref('')

const filteredBanks = computed(() => {
  const q = bankSearch.value.trim().toLowerCase()
  if (!q) return banks.value
  return banks.value.filter(
    (b) =>
      b.title.toLowerCase().includes(q) ||
      b.code.toLowerCase().includes(q) ||
      (b.subject?.name ?? '').toLowerCase().includes(q) ||
      (b.subject?.code ?? '').toLowerCase().includes(q),
  )
})

const selectedBankIds = computed(() =>
  Object.entries(selectedBanks.value)
    .filter(([, v]) => v)
    .map(([k]) => k),
)

function toggleBank(id: string) {
  selectedBanks.value[id] = !selectedBanks.value[id]
}

function clearSelectedBanks() {
  selectedBanks.value = {}
}

const typeLabel: Record<string, string> = {
  MULTIPLE_CHOICE: 'PG',
  TRUE_FALSE: 'B/S',
  MULTIPLE_ANSWER: 'Multi',
  ESSAY: 'Esai',
  SHORT_ANSWER: 'Isian',
}

onMounted(async () => {
  try {
    const [b, bankList] = await Promise.all([
      sectionService.listByExam(examId),
      bankService.list({ page: 1, limit: 100 }),
    ])
    sections.value = b
    banks.value = bankList.data
  } catch {
    ui.toastError('Gagal memuat data ujian.')
  } finally {
    loading.value = false
  }
})

const sortedSections = computed(() =>
  [...sections.value].sort((a, b) => a.sequence - b.sequence),
)

async function refresh() {
  sections.value = await sectionService.listByExam(examId)
}

// ---- form section ----
const formOpen = ref(false)
const editingSection = ref<ExamSection | null>(null)
const saving = ref(false)
const formName = ref('')
const formSequence = ref(String(nextSequence()))
const fieldErrors = ref<Record<string, string>>({})

function nextSequence(): number {
  return sections.value.reduce((max, s) => Math.max(max, s.sequence), 0) + 1
}

function openCreate() {
  editingSection.value = null
  formName.value = ''
  formSequence.value = String(nextSequence())
  fieldErrors.value = {}
  formOpen.value = true
}

function openEdit(s: ExamSection) {
  editingSection.value = s
  formName.value = s.name
  formSequence.value = String(s.sequence)
  fieldErrors.value = {}
  formOpen.value = true
}

async function submitSection() {
  fieldErrors.value = {}
  if (!formName.value.trim()) {
    fieldErrors.value.name = 'Nama section wajib diisi'
    return
  }
  const seq = Number(formSequence.value)
  if (!seq || seq < 1) {
    fieldErrors.value.sequence = 'Sequence minimal 1'
    return
  }

  saving.value = true
  try {
    if (editingSection.value) {
      await sectionService.update(editingSection.value.id, {
        name: formName.value,
        sequence: seq,
      })
      ui.toastSuccess('Section diperbarui.')
    } else {
      await sectionService.create(examId, { name: formName.value, sequence: seq })
      ui.toastSuccess('Section ditambahkan.')
    }
    formOpen.value = false
    await refresh()
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal menyimpan section.'))
  } finally {
    saving.value = false
  }
}

async function deleteSection(s: ExamSection) {
  try {
    await sectionService.remove(s.id)
    ui.toastSuccess('Section dihapus.')
    await refresh()
  } catch {
    ui.toastError('Gagal menghapus section.')
  }
}

// ---- konfirmasi hapus ----
const deleteSectionTarget = ref<ExamSection | null>(null)
const deletingSection = ref(false)

function askDeleteSection(s: ExamSection) {
  deleteSectionTarget.value = s
}

async function confirmDeleteSection() {
  if (!deleteSectionTarget.value) return
  deletingSection.value = true
  try {
    await deleteSection(deleteSectionTarget.value)
    deleteSectionTarget.value = null
  } finally {
    deletingSection.value = false
  }
}

const unmapTarget = ref<SectionQuestion | null>(null)

function askUnmap(q: SectionQuestion) {
  unmapTarget.value = q
}

async function confirmUnmap() {
  if (!unmapTarget.value || !mapTarget.value) return
  await unmapQuestion(unmapTarget.value)
  unmapTarget.value = null
}

// ---- mapping soal ----
const mapTarget = ref<ExamSection | null>(null)
const mapping = ref(false)
const mappedQuestions = ref<SectionQuestion[]>([])
const loadingMapped = ref(false)

async function openMap(s: ExamSection) {
  mapTarget.value = s
  selectedBanks.value = {}
  bankSearch.value = ''
  totalRandom.value = ''
  await loadMapped(s.id)
  // auto-centang bank yang sudah memiliki soal termapping di section ini
  for (const q of mappedQuestions.value) {
    if (q.question_bank_id) selectedBanks.value[q.question_bank_id] = true
  }
}

async function loadMapped(sectionId: string) {
  loadingMapped.value = true
  try {
    mappedQuestions.value = await sectionService.listQuestions(sectionId)
  } catch (err) {
    mappedQuestions.value = []
    ui.toastError(apiErrorMessage(err, 'Gagal memuat soal termapping.'))
  } finally {
    loadingMapped.value = false
  }
}

async function submitMap() {
  if (!mapTarget.value) return
  const bankIds = selectedBankIds.value
  if (bankIds.length === 0) {
    fieldErrors.value = { banks: 'Pilih minimal satu bank soal' }
    return
  }

  mapping.value = true
  try {
    const res = await sectionService.mapQuestions(mapTarget.value.id, {
      question_bank_ids: bankIds,
      total_random_questions: Number(totalRandom.value) || 0,
    })
    ui.toastSuccess(`${res.mapped_count} soal dimapping.`)
    await loadMapped(mapTarget.value.id)
    await refresh()
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal mapping soal.'))
  } finally {
    mapping.value = false
  }
}

async function unmapQuestion(q: SectionQuestion) {
  if (!mapTarget.value) return
  try {
    await sectionService.removeQuestion(mapTarget.value.id, q.id)
    mappedQuestions.value = mappedQuestions.value.filter((x) => x.id !== q.id)
    ui.toastSuccess('Soal dikeluarkan dari section.')
    await refresh()
  } catch {
    ui.toastError('Gagal mengeluarkan soal.')
  }
}
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-5xl space-y-6">
      <div>
        <button class="mb-1 text-xs font-medium text-muted-foreground hover:text-foreground" @click="router.push('/exams')">
          ‹ Ujian
        </button>
        <h1 class="text-2xl font-bold tracking-tight">Section Ujian</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          Bagi ujian menjadi section (mis. PG & Esai) dan mapping soal dari bank soal.
        </p>
      </div>

      <LoadingState v-if="loading" />

      <template v-else>
        <div class="flex justify-end">
          <BaseButton @click="openCreate"><Plus /> Tambah Section</BaseButton>
        </div>

        <EmptyState v-if="sections.length === 0" title="Belum ada section" message="Buat section untuk mulai mapping soal." />

        <div v-else class="space-y-4">
          <div
            v-for="s in sortedSections"
            :key="s.id"
            class="rounded-xl border border-border bg-card p-4 shadow-sm"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="flex items-start gap-3">
                <span class="flex size-9 items-center justify-center rounded-lg bg-primary/10 font-bold text-primary">
                  {{ s.sequence }}
                </span>
                <div>
                  <p class="font-semibold">{{ s.name }}</p>
                  <p class="mt-0.5 text-xs text-muted-foreground">{{ s.question_count ?? 0 }} soal termapping</p>
                </div>
              </div>
              <div class="flex gap-1">
                <BaseButton variant="outline" size="sm" @click="openMap(s)">
                  <ListTree /> Mapping Soal
                </BaseButton>
                <BaseButton variant="ghost" size="icon" aria-label="Edit" @click="openEdit(s)">
                  <Pencil />
                </BaseButton>
                <BaseButton variant="ghost" size="icon" aria-label="Hapus" @click="askDeleteSection(s)">
                  <Trash2 class="text-destructive" />
                </BaseButton>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- FORM SECTION -->
      <BaseModal :open="formOpen" :title="editingSection ? 'Edit Section' : 'Tambah Section'" @close="formOpen = false">
        <form class="space-y-4" @submit.prevent="submitSection">
          <BaseInput v-model="formName" label="Nama Section" placeholder="Section A - PG" required :error="fieldErrors.name" />
          <BaseInput v-model="formSequence" label="Urutan" type="number" min="1" required :error="fieldErrors.sequence" />
          <div class="flex justify-end gap-2 pt-2">
            <BaseButton variant="outline" type="button" @click="formOpen = false">Batal</BaseButton>
            <BaseButton type="submit" :loading="saving">Simpan</BaseButton>
          </div>
        </form>
      </BaseModal>

      <!-- KONFIRMASI HAPUS SECTION -->
      <BaseModal :open="!!deleteSectionTarget" title="Hapus Section?" @close="deleteSectionTarget = null">
        <p class="text-sm leading-relaxed text-muted-foreground">
          Hapus section
          <span class="font-semibold text-foreground">{{ deleteSectionTarget?.name }}</span>?
          Seluruh mapping soal di section ini juga akan terhapus.
        </p>
        <template #footer>
          <BaseButton variant="outline" @click="deleteSectionTarget = null">Batal</BaseButton>
          <BaseButton variant="destructive" :loading="deletingSection" @click="confirmDeleteSection">
            <Trash2 /> Ya, Hapus Section
          </BaseButton>
        </template>
      </BaseModal>

      <!-- KONFIRMASI LEPAS MAPPING -->
      <BaseModal :open="!!unmapTarget" title="Lepas Soal dari Section?" @close="unmapTarget = null">
        <p class="text-sm leading-relaxed text-muted-foreground">
          Keluarkan soal
          <span class="font-semibold text-foreground">"{{ unmapTarget?.text }}"</span>
          dari section ini?
        </p>
        <template #footer>
          <BaseButton variant="outline" @click="unmapTarget = null">Batal</BaseButton>
          <BaseButton variant="destructive" @click="confirmUnmap">
            <Unlink /> Ya, Keluarkan
          </BaseButton>
        </template>
      </BaseModal>

      <!-- MAPPING MODAL -->
      <BaseModal :open="!!mapTarget" :title="`Mapping Soal — ${mapTarget?.name ?? ''}`" @close="mapTarget = null">
        <div class="max-h-[70vh] space-y-5 overflow-y-auto pr-1">
          <div>
            <div class="mb-2 flex items-center justify-between gap-2">
              <p class="text-sm font-medium">
                Pilih Bank Soal
                <span v-if="selectedBankIds.length" class="ml-1 text-xs font-semibold text-primary">
                  ({{ selectedBankIds.length }} dipilih)
                </span>
              </p>
              <button
                v-if="selectedBankIds.length"
                type="button"
                class="text-xs font-medium text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
                @click="clearSelectedBanks"
              >
                Kosongkan
              </button>
            </div>

            <input
              v-model="bankSearch"
              placeholder="Cari bank soal / mata pelajaran…"
              class="h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />

            <div class="mt-2 flex max-h-40 flex-wrap gap-1.5 overflow-y-auto rounded-lg border border-border p-2">
              <button
                v-for="b in filteredBanks"
                :key="b.id"
                type="button"
                :class="
                  selectedBanks[b.id]
                    ? 'border-primary bg-primary text-primary-foreground'
                    : 'bg-background hover:border-primary/40 hover:bg-accent/60'
                "
                class="rounded-full border border-border px-3 py-1 text-xs font-medium transition-colors"
                :title="b.title"
                @click="toggleBank(b.id)"
              >
                {{ b.code }}
                <span class="opacity-70">· {{ b.title }}</span>
              </button>
              <p v-if="filteredBanks.length === 0" class="px-1 py-0.5 text-xs text-muted-foreground">
                Tidak ada bank yang cocok dengan pencarian.
              </p>
            </div>
            <p v-if="fieldErrors.banks" class="mt-1.5 text-xs text-destructive">{{ fieldErrors.banks }}</p>
          </div>

          <div class="w-48">
            <BaseInput
              v-model="totalRandom"
              label="Total Soal Acak (opsional)"
              type="number"
              min="0"
              placeholder="kosong = semua soal"
            />
          </div>

          <div class="flex justify-end">
            <BaseButton :disabled="mapping" :loading="mapping" @click="submitMap">
              <ListTree /> Mapping Sekarang
            </BaseButton>
          </div>

          <div class="border-t border-border pt-4">
            <p class="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Soal termapping ({{ mappedQuestions.length }})
            </p>
            <LoadingState v-if="loadingMapped" />
            <p v-else-if="mappedQuestions.length === 0" class="text-sm text-muted-foreground">Belum ada soal.</p>
            <ul v-else class="space-y-1.5">
              <li
                v-for="q in mappedQuestions"
                :key="q.id"
                class="flex items-center justify-between gap-3 rounded-lg border border-border px-3 py-2 text-sm"
              >
                <span class="flex min-w-0 items-center gap-2">
                  <BaseBadge tone="outline">{{ typeLabel[q.type] ?? q.type }}</BaseBadge>
                  <span class="truncate">{{ q.text }}</span>
                </span>
                <BaseButton variant="ghost" size="icon" aria-label="Keluarkan" @click="askUnmap(q)">
                  <Unlink class="text-destructive" />
                </BaseButton>
              </li>
            </ul>
          </div>
        </div>
        <template #footer>
          <BaseButton variant="outline" @click="mapTarget = null">Selesai</BaseButton>
        </template>
      </BaseModal>
    </div>
  </AppShell>
</template>
