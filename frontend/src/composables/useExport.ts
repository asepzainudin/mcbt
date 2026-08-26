import { ref } from 'vue'

import { api, apiErrorMessage } from '../lib/axios'
import { useUiStore } from '../stores/ui'

export function useExport() {
  const ui = useUiStore()
  const exporting = ref<string | null>(null)

  async function download(path: string) {
    exporting.value = path
    try {
      const res = await api.get(path, { responseType: 'blob' })
      const dispo = (res.headers?.['content-disposition'] as string | undefined) ?? ''
      const match = dispo.match(/filename="([^"]+)"/)
      const name = match?.[1] ?? 'unduhan'
      const url = URL.createObjectURL(res.data as Blob)
      const a = document.createElement('a')
      a.href = url
      a.download = name
      document.body.appendChild(a)
      a.click()
      a.remove()
      URL.revokeObjectURL(url)
      ui.toastSuccess(`Berhasil mengunduh ${name}`)
    } catch (e) {
      ui.toastError(apiErrorMessage(e))
    } finally {
      exporting.value = null
    }
  }

  return { exporting, download }
}
