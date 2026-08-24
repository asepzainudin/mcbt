import { computed, onMounted, ref, watch, type ComputedRef } from 'vue'
import type { AxiosError } from 'axios'

import { apiErrorMessage, extractFieldErrors } from '../lib/axios'
import { useUiStore } from '../stores/ui'
import type { PaginationMeta } from '../types/api'

export interface CrudListParams {
  page?: number
  limit?: number
  search?: string
}

export interface CrudConfig<T> {
  itemLabel: string
  defaultLimit?: number
  extraParams?: ComputedRef<Record<string, string | number | undefined>>
  listFn: (
    params: CrudListParams,
  ) => Promise<{ data: T[]; meta: PaginationMeta | null }>
  createFn: (payload: Record<string, unknown>) => Promise<T>
  updateFn: (id: string, payload: Record<string, unknown>) => Promise<T>
  removeFn: (id: string) => Promise<void>
  toPayload: (form: Record<string, string>) => Record<string, unknown>
  fromItem?: (item: T) => Record<string, string>
  /** validasi sisi klien sebelum submit; return map field -> pesan */
  validate?: (form: Record<string, string>, isEditing: boolean) => Record<string, string>
}

export function useCrudTable<T extends { id: string }>(config: CrudConfig<T>) {
  const ui = useUiStore()

  const items = ref<T[]>([])
  const meta = ref<PaginationMeta | null>(null)
  const loading = ref(false)

  const searchInput = ref('')
  const search = ref('')
  const page = ref(1)
  const limit = ref(config.defaultLimit ?? 10)

  let debounceTimer: ReturnType<typeof setTimeout> | undefined
  watch(searchInput, (val) => {
    clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => {
      search.value = val.trim()
      page.value = 1
    }, 300)
  })

  // ---- list state machine ----
  const fetchList = async () => {
    loading.value = true
    try {
      const params: CrudListParams = {
        page: page.value,
        limit: limit.value,
        search: search.value,
        ...(config.extraParams?.value ?? {}),
      }
      const result = await config.listFn(params)
      items.value = result.data
      meta.value = result.meta
    } catch {
      ui.toastError(`Gagal memuat ${config.itemLabel}.`)
    } finally {
      loading.value = false
    }
  }

  watch([page, limit], fetchList)
  watch(search, fetchList)
  onMounted(fetchList)

  function goTo(p: number) {
    page.value = p
  }

  // ---- form modal state ----
  const formOpen = ref(false)
  const editingId = ref<string | null>(null)
  const saving = ref(false)
  const form = ref<Record<string, string>>({})
  const fieldErrors = ref<Record<string, string>>({})

  const isEditing = computed(() => editingId.value !== null)

  function openCreate(defaults: Record<string, string> = {}) {
    editingId.value = null
    form.value = { ...defaults }
    fieldErrors.value = {}
    formOpen.value = true
  }

  function openEdit(item: T) {
    editingId.value = item.id
    form.value = config.fromItem
      ? config.fromItem(item)
      : Object.fromEntries(
          Object.entries(item as unknown as Record<string, unknown>).map(([k, v]) => [
            k,
            v === null || v === undefined ? '' : String(v),
          ]),
        )
    fieldErrors.value = {}
    formOpen.value = true
  }

  function closeForm() {
    formOpen.value = false
  }

  async function submit() {
    fieldErrors.value = {}

    if (config.validate) {
      const clientErrors = config.validate(form.value, editingId.value !== null)
      if (Object.keys(clientErrors).length > 0) {
        fieldErrors.value = clientErrors
        return
      }
    }

    saving.value = true
    try {
      const payload = config.toPayload(form.value)
      if (editingId.value) {
        await config.updateFn(editingId.value, payload)
        ui.toastSuccess(`${config.itemLabel} diperbarui.`)
      } else {
        await config.createFn(payload)
        ui.toastSuccess(`${config.itemLabel} ditambahkan.`)
      }
      formOpen.value = false
      await fetchList()
    } catch (err) {
      const axiosErr = err as AxiosError<{ errors?: Record<string, string> }>
      if (axiosErr.response?.status === 422) {
        fieldErrors.value = extractFieldErrors(err)
      } else if (axiosErr.response?.status === 409) {
        ui.toastError(apiErrorMessage(err))
      } else {
        ui.toastError(apiErrorMessage(err, `Gagal menyimpan ${config.itemLabel}.`))
      }
    } finally {
      saving.value = false
    }
  }

  // ---- delete confirm state ----
  const deleteTarget = ref<T | null>(null)
  const deleting = ref(false)

  function askDelete(item: T) {
    deleteTarget.value = item
  }

  function cancelDelete() {
    deleteTarget.value = null
  }

  async function confirmDelete() {
    if (!deleteTarget.value) return
    deleting.value = true
    try {
      await config.removeFn(deleteTarget.value.id)
      ui.toastSuccess(`${config.itemLabel} dihapus.`)
      deleteTarget.value = null

      if (items.value.length === 1 && page.value > 1) {
        page.value -= 1
      } else {
        await fetchList()
      }
    } catch (err) {
      ui.toastError(apiErrorMessage(err, `Gagal menghapus ${config.itemLabel}.`))
    } finally {
      deleting.value = false
    }
  }

  return {
    items,
    meta,
    loading,
    searchInput,
    page,
    limit,
    totalPages: computed(() => meta.value?.total_pages ?? 0),
    rangeText: computed(() => {
      if (!meta.value || meta.value.total_items === 0) return ''
      const start = (meta.value.page - 1) * meta.value.limit + 1
      const end = Math.min(meta.value.page * meta.value.limit, meta.value.total_items)
      return `Menampilkan ${start}–${end} dari ${meta.value.total_items} data`
    }),
    fetchList,
    goTo,
    formOpen,
    isEditing,
    saving,
    form,
    fieldErrors,
    openCreate,
    openEdit,
    closeForm,
    submit,
    deleteTarget,
    deleting,
    askDelete,
    cancelDelete,
    confirmDelete,
  }
}
