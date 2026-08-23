<script setup lang="ts">
import { ref } from 'vue'
import { Download, FileSpreadsheet, Upload } from 'lucide-vue-next'

import BaseButton from '../ui/BaseButton.vue'
import BaseModal from '../ui/BaseModal.vue'
import { api, apiErrorMessage } from '../../lib/axios'
import type { ImportResult } from '../../types/api'

interface Props {
  open: boolean
  resource: 'teachers' | 'students'
  title: string
  templateUrl: string
}

const props = defineProps<Props>()
const emit = defineEmits<{ (e: 'close'): void; (e: 'imported'): void }>()

const file = ref<File | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)
const uploading = ref(false)
const result = ref<ImportResult | null>(null)
const error = ref('')

function onFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  file.value = target.files?.[0] ?? null
  result.value = null
  error.value = ''
}

async function upload() {
  if (!file.value) return
  uploading.value = true
  error.value = ''
  result.value = null

  try {
    const form = new FormData()
    form.append('file', file.value)
    const res = await api.post(`/api/v1/${props.resource}/import`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    result.value = res.data.data as ImportResult
    emit('imported')
  } catch (err) {
    error.value = apiErrorMessage(err, 'Gagal mengunggah file.')
  } finally {
    uploading.value = false
  }
}

function close() {
  file.value = null
  result.value = null
  error.value = ''
  if (inputRef.value) inputRef.value.value = ''
  emit('close')
}
</script>

<template>
  <BaseModal :open="open" :title="title" @close="close()">
    <div class="space-y-4">
      <a
        :href="templateUrl"
        class="flex items-center gap-2 rounded-lg border border-dashed border-primary/40 bg-primary/5 px-3 py-2.5 text-sm font-medium text-primary transition-colors hover:bg-primary/10"
      >
        <Download class="size-4" />
        Unduh template Excel (.xlsx)
      </a>

      <label
        class="flex cursor-pointer items-center gap-3 rounded-lg border border-border bg-muted/40 px-3 py-3 text-sm transition-colors hover:bg-muted"
      >
        <FileSpreadsheet class="size-5 shrink-0 text-emerald-600" />
        <span class="truncate">
          {{ file ? file.name : 'Pilih file Excel…' }}
        </span>
        <input ref="inputRef" type="file" accept=".xlsx,.xls" class="hidden" @change="onFileChange" />
      </label>

      <p v-if="error" class="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
        {{ error }}
      </p>

      <div v-if="result" class="space-y-2 rounded-lg border border-border p-3 text-sm">
        <p class="font-medium text-success">{{ result.imported_count }} baris berhasil diimpor.</p>
        <div v-if="result.skipped.length > 0" class="space-y-1">
          <p class="font-medium text-warning">{{ result.skipped.length }} baris dilewati:</p>
          <ul class="max-h-32 space-y-0.5 overflow-y-auto text-xs text-muted-foreground">
            <li v-for="(s, i) in result.skipped" :key="i">
              Baris {{ s.row }} — <span class="font-mono">{{ s.field }}</span>: {{ s.reason }}
            </li>
          </ul>
        </div>
      </div>
    </div>

    <template #footer>
      <BaseButton variant="outline" @click="close()">Tutup</BaseButton>
      <BaseButton :disabled="!file || !!result" :loading="uploading" @click="upload()">
        <Upload /> Impor
      </BaseButton>
    </template>
  </BaseModal>
</template>
