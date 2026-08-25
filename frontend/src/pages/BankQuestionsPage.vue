<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowDown,
  ArrowUp,
  Check,
  Eye,
  ImagePlus,
  Pencil,
  Plus,
  Search,
  Send,
  Trash2,
  Upload,
} from 'lucide-vue-next'

import AppShell from '../components/layout/AppShell.vue'
import RichTextEditor from '../components/ui/RichTextEditor.vue'
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
import { apiErrorMessage } from '../lib/axios'
import { uploadMedia } from '../services/media.service'
import { bankService, questionService } from '../services/question.service'
import type { Question, QuestionType, QuestionPayload } from '../types/api'
import { useUiStore } from '../stores/ui'

const ui = useUiStore()
const route = useRoute()
const router = useRouter()

const bankId = route.params.id as string
const bankTitle = ref(route.query.title ? String(route.query.title) : 'Bank Soal')
const bankStatus = ref<string>(String(route.query.status ?? 'draft'))

const questions = ref<Question[]>([])
const meta = ref<{ page: number; total_pages: number; total_items: number } | null>(null)
const page = ref(1)
const search = ref('')
const loading = ref(false)

let debounceTimer: ReturnType<typeof setTimeout> | undefined
function onSearchInput(v: string) {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    search.value = v.trim()
    page.value = 1
    fetchQuestions()
  }, 300)
}

async function fetchQuestions() {
  loading.value = true
  try {
    const res = await questionService.list({
      bank_id: bankId,
      page: page.value,
      limit: 20,
      search: search.value || undefined,
    })
    questions.value = res.data
    meta.value = res.meta
  } catch {
    ui.toastError('Gagal memuat soal.')
  } finally {
    loading.value = false
  }
}
onMounted(fetchQuestions)

const typeLabel: Record<string, string> = {
  MULTIPLE_CHOICE: 'Pilihan Ganda',
  TRUE_FALSE: 'Benar / Salah',
  MULTIPLE_ANSWER: 'Multiple Answer',
  ESSAY: 'Esai',
  SHORT_ANSWER: 'Isian Singkat',
}
const typeTone: Record<string, 'primary' | 'success' | 'info' | 'warning' | 'neutral'> = {
  MULTIPLE_CHOICE: 'primary',
  TRUE_FALSE: 'success',
  MULTIPLE_ANSWER: 'info',
  ESSAY: 'warning',
  SHORT_ANSWER: 'neutral',
}

const typeSelectOptions = Object.entries(typeLabel).map(([value, label]) => ({ value, label }))

// ---- form ----
const formOpen = ref(false)
const editingQuestion = ref<Question | null>(null)
const saving = ref(false)
const fieldErrors = ref<Record<string, string>>({})

const formType = ref<QuestionType>('MULTIPLE_CHOICE')
const formText = ref('')
const formScore = ref('1.0')
const formExplanation = ref('')
const formMediaId = ref<string | null>(null)
const formMediaUrl = ref<string | null>(null)
const formMediaPosition = ref<'before' | 'after'>('after')
const showImport = ref(false)
const formAnswerKeys = ref('')

interface OptionRow {
  option_key: string
  text: string
  is_correct: boolean
  media_id: string | null
  media_url: string | null
  uploading: boolean
}

const optionRows = ref<OptionRow[]>([])

const needsOptions = computed(() => formType.value !== 'ESSAY' && formType.value !== 'SHORT_ANSWER')
const multiCorrect = computed(() => formType.value === 'MULTIPLE_ANSWER')

watch(formType, (t) => {
  if (t === 'TRUE_FALSE' && optionRows.value.length !== 2) {
    optionRows.value = [
      { option_key: 'A', text: 'BENAR', is_correct: true, media_id: null, media_url: null, uploading: false },
      { option_key: 'B', text: 'SALAH', is_correct: false, media_id: null, media_url: null, uploading: false },
    ]
  }
})

function addOption() {
  if (!needsOptions.value || optionRows.value.length >= 5) return
  const key = String.fromCharCode(65 + optionRows.value.length)
  optionRows.value.push({
    option_key: key,
    text: '',
    is_correct: false,
    media_id: null,
    media_url: null,
    uploading: false,
  })
}

function removeOption(i: number) {
  optionRows.value.splice(i, 1)
  optionRows.value.forEach((row, idx) => {
    row.option_key = String.fromCharCode(65 + idx)
  })
}

function moveOption(i: number, dir: -1 | 1) {
  const j = i + dir
  if (j < 0 || j >= optionRows.value.length) return
  const tmp = optionRows.value[i]
  optionRows.value[i] = optionRows.value[j]
  optionRows.value[j] = tmp
  optionRows.value.forEach((row, idx) => {
    row.option_key = String.fromCharCode(65 + idx)
  })
}

function onCorrectChange(i: number) {
  const row = optionRows.value[i]
  row.is_correct = !row.is_correct
  if (!multiCorrect.value && row.is_correct) {
    optionRows.value.forEach((r, idx) => {
      if (idx !== i) r.is_correct = false
    })
  }
}

function openCreate() {
  editingQuestion.value = null
  formType.value = 'MULTIPLE_CHOICE'
  formText.value = ''
  formScore.value = '1.0'
  formExplanation.value = ''
  formMediaId.value = null
  formMediaUrl.value = null
  formMediaPosition.value = 'after'
  formAnswerKeys.value = ''
  optionRows.value = [
    { option_key: 'A', text: '', is_correct: true, media_id: null, media_url: null, uploading: false },
    { option_key: 'B', text: '', is_correct: false, media_id: null, media_url: null, uploading: false },
  ]
  fieldErrors.value = {}
  formOpen.value = true
}

function openEdit(q: Question) {
  editingQuestion.value = q
  formType.value = q.type
  formText.value = q.text
  formScore.value = String(q.score_weight)
  formExplanation.value = q.explanation ?? ''
  formMediaId.value = q.media_id
  formMediaUrl.value = q.media?.url ?? null
  formMediaPosition.value = (q.media_position as 'before' | 'after') ?? 'after'
  formAnswerKeys.value = (q.answer_keys ?? []).join('\n')
  optionRows.value = (q.options ?? []).map((o) => ({
    option_key: o.option_key,
    text: o.text,
    is_correct: o.is_correct,
    media_id: o.media_id,
    media_url: o.media?.url ?? null,
    uploading: false,
  }))
  fieldErrors.value = {}
  formOpen.value = true
}

function clientValidate(): Record<string, string> {
  const e: Record<string, string> = {}
  if (!formText.value.trim()) e.text = 'Teks soal wajib diisi'
  if (needsOptions.value) {
    const filled = optionRows.value.filter((o) => o.text.trim() !== '')
    if (filled.length < 2) e.options = 'Minimal 2 opsi terisi'
    else if (optionRows.value.every((o) => !o.is_correct)) e.options = 'Tandai minimal satu jawaban benar'
  }
  if (formType.value === 'SHORT_ANSWER' && !formAnswerKeys.value.trim())
    e.answer_keys = 'Minimal satu jawaban diterima (satu per baris)'
  return e
}

async function onUploadImage(event: Event, target: 'question' | number) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    if (target === 'question') {
      const media = await uploadMedia(file, 'QUESTION_IMAGE')
      formMediaId.value = media.id
      formMediaUrl.value = media.url
      ui.toastSuccess('Gambar soal terunggah.')
    } else {
      const row = optionRows.value[target]
      row.uploading = true
      const media = await uploadMedia(file, 'OPTION_IMAGE')
      row.media_id = media.id
      row.media_url = media.url
      ui.toastSuccess('Gambar opsi terunggah.')
    }
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal mengunggah gambar.'))
  } finally {
    if (target !== 'question') optionRows.value[target].uploading = false
    input.value = ''
  }
}

async function submit() {
  fieldErrors.value = {}
  const clientErrors = clientValidate()
  if (Object.keys(clientErrors).length > 0) {
    fieldErrors.value = clientErrors
    return
  }

  const payload: QuestionPayload = {
    type: formType.value,
    text: formText.value,
    score_weight: Number(formScore.value) || 1,
    explanation: formExplanation.value || null,
    options: needsOptions.value
      ? optionRows.value.map((o) => ({
          option_key: o.option_key,
          text: o.text,
          is_correct: o.is_correct,
          media_id: o.media_id,
        }))
      : [],
    media_id: formMediaId.value,
    media_position: formMediaPosition.value,
    answer_keys: formType.value === 'SHORT_ANSWER'
      ? formAnswerKeys.value.split('\n').map((s) => s.trim()).filter(Boolean)
      : undefined,
  }

  saving.value = true
  try {
    if (editingQuestion.value) {
      await questionService.update(editingQuestion.value.id, payload)
      ui.toastSuccess('Soal diperbarui.')
    } else {
      await questionService.createInBank(bankId, payload)
      ui.toastSuccess('Soal ditambahkan.')
    }
    formOpen.value = false
    await fetchQuestions()
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal menyimpan soal.'))
  } finally {
    saving.value = false
  }
}

// ---- reorder & tandai benar pada soal tersimpan ----
async function shiftSavedOption(q: Question, index: number, dir: -1 | 1) {
  const ids = (q.options ?? []).map((o) => o.id)
  const j = index + dir
  if (j < 0 || j >= ids.length) return
  ;[ids[index], ids[j]] = [ids[j], ids[index]]
  try {
    const updated = await questionService.reorderOptions(q.id, ids)
    Object.assign(q, updated)
  } catch {
    ui.toastError('Gagal mengubah urutan opsi.')
  }
}

async function markCorrect(q: Question, optionId: string) {
  try {
    const updated = await questionService.setCorrectOption(q.id, optionId)
    Object.assign(q, updated)
    ui.toastSuccess('Jawaban benar diperbarui.')
  } catch {
    ui.toastError('Gagal menandai jawaban benar.')
  }
}

// ---- preview ----
const previewQuestion = ref<Question | null>(null)
const previewData = ref<Awaited<ReturnType<typeof questionService.preview>> | null>(null)
const loadingPreview = ref(false)

async function openPreview(q: Question) {
  previewQuestion.value = q
  previewData.value = null
  loadingPreview.value = true
  try {
    previewData.value = await questionService.preview(q.id)
  } catch {
    ui.toastError('Gagal memuat preview.')
    previewQuestion.value = null
  } finally {
    loadingPreview.value = false
  }
}

// ---- hapus ----
const deleteTarget = ref<Question | null>(null)
const deleting = ref(false)
async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await questionService.remove(deleteTarget.value.id)
    ui.toastSuccess('Soal dihapus.')
    deleteTarget.value = null
    await fetchQuestions()
  } catch {
    ui.toastError('Gagal menghapus soal.')
  } finally {
    deleting.value = false
  }
}

async function publishBank() {
  try {
    const updated = await bankService.publish(bankId)
    bankStatus.value = updated.status
    ui.toastSuccess('Bank soal dipublikasikan.')
  } catch (err) {
    ui.toastError(apiErrorMessage(err, 'Gagal publish.'))
  }
}
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-5xl space-y-6">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <div>
          <button class="mb-1 text-xs font-medium text-muted-foreground hover:text-foreground" @click="router.push('/question-banks')">
            ‹ Bank Soal
          </button>
          <div class="flex items-center gap-2">
            <h1 class="text-2xl font-bold tracking-tight">{{ bankTitle }}</h1>
            <BaseBadge :tone="bankStatus === 'published' ? 'success' : bankStatus === 'archived' ? 'warning' : 'neutral'">
              {{ bankStatus }}
            </BaseBadge>
          </div>
        </div>
        <div class="flex flex-wrap items-end gap-2">
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <input
              :value="search"
              placeholder="Cari soal…"
              class="h-9 w-44 rounded-lg border border-input bg-transparent pl-9 pr-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              @input="onSearchInput(($event.target as HTMLInputElement).value)"
            />
          </div>
          <BaseButton v-if="bankStatus === 'draft'" variant="outline" @click="publishBank">
            <Send /> Publish
          </BaseButton>
          <BaseButton variant="outline" @click="showImport = true"><Upload /> Impor</BaseButton>
          <BaseButton @click="openCreate()"><Plus /> Tambah Soal</BaseButton>
        </div>
      </div>

      <LoadingState v-if="loading" />

      <template v-else>
        <EmptyState v-if="questions.length === 0" title="Belum ada soal" message="Tambah manual lewat tombol Tambah Soal." />

        <template v-else>
          <BaseTable>
            <template #head>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">#</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Soal</th>
              <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Tipe</th>
              <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-muted-foreground">Aksi</th>
            </template>

            <tr v-for="(q, qi) in questions" :key="q.id" class="border-b border-border align-top transition-colors last:border-0 hover:bg-accent/50">
              <td class="px-4 py-3 text-muted-foreground">{{ ((meta?.page ?? 1) - 1) * 20 + qi + 1 }}</td>
              <td class="px-4 py-3">
                <img
                  v-if="q.media && q.media_position === 'before'"
                  :src="q.media.url"
                  class="mb-2 max-h-24 rounded-lg border border-border"
                  alt=""
                />
                <div class="prose-sm max-w-lg [&_ol]:list-decimal [&_ol]:pl-5 [&_p]:my-1 [&_ul]:list-disc [&_ul]:pl-5" v-html="q.text" />
                <img
                  v-if="q.media && q.media_position !== 'before'"
                  :src="q.media.url"
                  class="mt-2 max-h-24 rounded-lg border border-border"
                  alt=""
                />
                <ul v-if="q.options && q.options.length" class="mt-2 space-y-1 text-xs text-muted-foreground">
                  <li v-for="(opt, oi) in q.options" :key="opt.id" class="flex items-center gap-1.5">
                    <span class="w-12 shrink-0 space-x-0.5">
                      <BaseButton variant="ghost" size="icon" class="!size-5 !rounded p-0" aria-label="Naik" @click="shiftSavedOption(q, oi, -1)">
                        <ArrowUp class="!size-3" />
                      </BaseButton>
                      <BaseButton variant="ghost" size="icon" class="!size-5 !rounded p-0" aria-label="Turun" @click="shiftSavedOption(q, oi, 1)">
                        <ArrowDown class="!size-3" />
                      </BaseButton>
                    </span>
                    <span class="font-mono font-semibold">{{ opt.option_key }}.</span>
                    <span :class="opt.is_correct ? 'font-semibold text-success' : ''">{{ opt.text }}</span>
                    <Check v-if="opt.is_correct" class="size-3.5 text-success" />
                    <button
                      v-else-if="q.type === 'MULTIPLE_CHOICE' || q.type === 'TRUE_FALSE'"
                      class="text-[10px] uppercase text-primary hover:underline"
                      @click="markCorrect(q, opt.id)"
                    >
                      tandai benar
                    </button>
                  </li>
                </ul>
                <p v-if="(q.answer_keys ?? []).length" class="mt-1 text-xs text-muted-foreground">
                  Kunci: <span class="font-mono">{{ (q.answer_keys ?? []).join(', ') }}</span>
                </p>
              </td>
              <td class="px-4 py-3">
                <BaseBadge :tone="typeTone[q.type] ?? 'neutral'">{{ typeLabel[q.type] ?? q.type }}</BaseBadge>
                <BaseBadge v-if="q.is_used" tone="warning" class="ml-1">digunakan ujian</BaseBadge>
                <p class="mt-1 text-xs text-muted-foreground">bobot {{ q.score_weight }}</p>
              </td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <BaseButton variant="ghost" size="icon" title="Preview" @click="openPreview(q)">
                    <Eye />
                  </BaseButton>
                  <template v-if="!q.is_used">
                    <BaseButton variant="ghost" size="icon" title="Edit" @click="openEdit(q)">
                      <Pencil />
                    </BaseButton>
                    <BaseButton variant="ghost" size="icon" title="Hapus" @click="deleteTarget = q">
                      <Trash2 class="text-destructive" />
                    </BaseButton>
                  </template>
                </div>
              </td>
            </tr>
          </BaseTable>

          <div class="flex flex-col items-center gap-3 sm:flex-row sm:justify-between">
            <p class="text-sm text-muted-foreground">
              Menampilkan {{ questions.length }} dari {{ meta?.total_items ?? 0 }} soal
            </p>
            <BasePagination v-if="meta" :page="meta.page" :total-pages="meta.total_pages" @change="(p) => { page = p; fetchQuestions() }" />
          </div>
        </template>
      </template>

      <ImportModal
        :open="showImport"
        title="Impor Soal dari Excel"
        upload-path="/questions/import/validate"
        process-path="/questions/import/process"
        :extra-fields="{ question_bank_id: bankId }"
        template-url="/api/v1/questions/import/template"
        @close="showImport = false"
        @imported="fetchQuestions()"
      />

      <!-- FORM SOAL -->
      <BaseModal :open="formOpen" :title="editingQuestion ? 'Edit Soal' : 'Tambah Soal'" @close="formOpen = false">
        <form class="max-h-[70vh] space-y-4 overflow-y-auto pr-1" @submit.prevent="submit">
          <BaseSelect v-model="formType" label="Tipe Soal" required :options="typeSelectOptions" />

          <div>
            <p class="mb-1.5 text-sm font-medium leading-none text-foreground">
              Teks Soal <span class="text-destructive" aria-hidden="true">*</span>
            </p>
            <RichTextEditor v-model="formText" />
            <p v-if="fieldErrors.text" class="mt-1 text-xs font-medium text-destructive">{{ fieldErrors.text }}</p>
          </div>

          <div class="flex items-center gap-3">
            <label class="flex cursor-pointer items-center gap-2 rounded-lg border border-dashed border-border px-3 py-2 text-xs text-muted-foreground transition-colors hover:bg-muted">
              <ImagePlus class="size-4" />
              {{ formMediaId ? 'Ganti gambar soal' : 'Lampirkan gambar soal' }}
              <input type="file" accept=".png,.jpg,.jpeg,.gif,.webp" class="hidden" @change="onUploadImage($event, 'question')" />
            </label>
            <img v-if="formMediaUrl" :src="formMediaUrl" class="h-10 rounded border border-border" alt="" />
            <BaseBadge v-if="formMediaId" tone="success">siap</BaseBadge>
          </div>

          <div v-if="formMediaId" class="w-56">
            <BaseSelect
              v-model="formMediaPosition"
              label="Posisi Gambar"
              :options="[
                { value: 'after', label: 'Setelah teks soal' },
                { value: 'before', label: 'Sebelum teks soal' },
              ]"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <BaseInput v-model="formScore" label="Score Weight" type="number" step="0.1" min="0.1" />
            <BaseInput v-model="formExplanation" label="Pembahasan (opsional)" />
          </div>

          <div v-if="needsOptions" class="space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Opsi —
              <span v-if="multiCorrect">centang semua jawaban benar</span>
              <span v-else>pilih satu jawaban benar</span>
              <span class="text-destructive">*</span>
            </p>
            <div v-for="(row, i) in optionRows" :key="i" class="flex items-start gap-2 rounded-lg border border-border p-2">
              <span class="pt-2 font-mono text-sm font-bold">{{ row.option_key }}</span>
              <input
                type="checkbox"
                class="mt-2.5 accent-blue-600"
                :checked="row.is_correct"
                :aria-label="`Tandai opsi ${row.option_key} benar`"
                @change="onCorrectChange(i)"
              />
              <div class="flex-1 space-y-1">
                <BaseInput v-model="row.text" placeholder="Isi opsi…" />
                <div v-if="row.media_url" class="flex items-center gap-2">
                  <img :src="row.media_url" class="h-8 rounded border border-border" alt="" />
                  <span class="text-[10px] text-success">gambar siap</span>
                </div>
              </div>
              <label class="cursor-pointer pt-1.5 text-muted-foreground transition-colors hover:text-primary" :title="`Unggah gambar opsi ${row.option_key}`">
                <ImagePlus v-if="!row.uploading" class="size-4" />
                <span v-else class="text-xs">…</span>
                <input type="file" accept=".png,.jpg,.jpeg,.gif,.webp" class="hidden" @change="onUploadImage($event, i)" />
              </label>
              <div class="flex flex-col pt-1">
                <BaseButton variant="ghost" size="icon" class="!size-6" aria-label="Naik" @click="moveOption(i, -1)">
                  <ArrowUp class="!size-3" />
                </BaseButton>
                <BaseButton variant="ghost" size="icon" class="!size-6" aria-label="Turun" @click="moveOption(i, 1)">
                  <ArrowDown class="!size-3" />
                </BaseButton>
              </div>
              <BaseButton variant="ghost" size="icon" class="mt-1" aria-label="Hapus opsi" @click="removeOption(i)">
                <Trash2 class="size-4 text-destructive" />
              </BaseButton>
            </div>
            <BaseButton v-if="formType !== 'TRUE_FALSE' && optionRows.length < 5" variant="outline" size="sm" @click="addOption()">
              <Plus /> Tambah Opsi
            </BaseButton>
            <p v-if="fieldErrors.options" class="text-xs font-medium text-destructive">{{ fieldErrors.options }}</p>
          </div>

          <BaseInput
            v-if="formType === 'SHORT_ANSWER'"
            v-model="formAnswerKeys"
            label="Jawaban Diterima (satu per baris)"
            :error="fieldErrors.answer_keys"
          />

          <div class="flex justify-end gap-2 pt-2">
            <BaseButton variant="outline" type="button" @click="formOpen = false">Batal</BaseButton>
            <BaseButton type="submit" :loading="saving">Simpan Soal</BaseButton>
          </div>
        </form>
      </BaseModal>

      <!-- PREVIEW -->
      <BaseModal :open="!!previewQuestion" title="Preview Soal" @close="previewQuestion = null">
        <div v-if="loadingPreview"><LoadingState /></div>
        <div v-else-if="previewData" class="space-y-4">
          <div class="flex items-center gap-2">
            <BaseBadge :tone="typeTone[previewData.type] ?? 'neutral'">{{ typeLabel[previewData.type] ?? previewData.type }}</BaseBadge>
            <span class="text-xs text-muted-foreground">bobot {{ previewData.score_weight }}</span>
          </div>
          <img
            v-if="previewData.image && previewData.media_position === 'before'"
            :src="previewData.image.url"
            class="mb-3 max-h-40 rounded-lg border border-border"
            alt=""
          />
          <div class="prose-sm [&_ol]:list-decimal [&_ol]:pl-5 [&_p]:my-1 [&_ul]:list-disc [&_ul]:pl-5" v-html="previewData.content_html" />
          <img
            v-if="previewData.image && previewData.media_position !== 'before'"
            :src="previewData.image.url"
            class="mt-3 max-h-40 rounded-lg border border-border"
            alt=""
          />

          <ul v-if="previewData.options.length" class="space-y-2 text-sm">
            <li
              v-for="opt in previewData.options"
              :key="opt.option_key"
              :class="[
                'flex items-start gap-2 rounded-lg border p-2.5',
                previewData.correct_keys.includes(opt.option_key)
                  ? 'border-success/40 bg-success/5'
                  : 'border-border',
              ]"
            >
              <span class="font-mono font-bold">{{ opt.option_key }}.</span>
              <span v-html="opt.text" />
              <Check v-if="previewData.correct_keys.includes(opt.option_key)" class="ml-auto size-4 shrink-0 text-success" />
            </li>
          </ul>

          <div v-if="previewData.type === 'SHORT_ANSWER'" class="rounded-lg border border-border p-3 text-sm">
            <p class="mb-1 text-xs font-semibold uppercase text-muted-foreground">Jawaban diterima</p>
            <p class="font-mono text-xs">{{ previewData.answer_keys.join(', ') }}</p>
          </div>

          <div v-if="previewData.explanation" class="rounded-lg bg-muted/60 p-3 text-sm">
            <p class="mb-1 text-xs font-semibold uppercase text-muted-foreground">Pembahasan</p>
            <p>{{ previewData.explanation }}</p>
          </div>
        </div>
        <template #footer>
          <BaseButton variant="outline" @click="previewQuestion = null">Tutup</BaseButton>
        </template>
      </BaseModal>

      <BaseModal :open="!!deleteTarget" title="Konfirmasi Hapus" @close="deleteTarget = null">
        <p class="text-sm text-muted-foreground">Hapus soal ini beserta opsinya?</p>
        <template #footer>
          <BaseButton variant="outline" @click="deleteTarget = null">Batal</BaseButton>
          <BaseButton variant="destructive" :loading="deleting" @click="confirmDelete()">
            <Trash2 /> Hapus
          </BaseButton>
        </template>
      </BaseModal>
    </div>
  </AppShell>
</template>
