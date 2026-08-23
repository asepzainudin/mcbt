import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ToastType = 'success' | 'error' | 'warning' | 'info'

export interface Toast {
  id: number
  type: ToastType
  message: string
}

let nextId = 1

export const useUiStore = defineStore('ui', () => {
  const toasts = ref<Toast[]>([])
  const errorModalOpen = ref(false)
  const errorModalMessage = ref('')

  function push(type: ToastType, message: string, duration = 4000) {
    const toast: Toast = { id: nextId++, type, message }
    toasts.value.push(toast)
    if (duration > 0) {
      setTimeout(() => dismiss(toast.id), duration)
    }
  }

  function dismiss(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  function toastSuccess(message: string) {
    push('success', message)
  }
  function toastError(message: string) {
    push('error', message, 6000)
  }
  function toastWarning(message: string) {
    push('warning', message, 5000)
  }
  function toastInfo(message: string) {
    push('info', message)
  }

  function openErrorModal(message: string) {
    errorModalMessage.value = message
    errorModalOpen.value = true
  }
  function closeErrorModal() {
    errorModalOpen.value = false
  }

  return {
    toasts,
    errorModalOpen,
    errorModalMessage,
    dismiss,
    toastSuccess,
    toastError,
    toastWarning,
    toastInfo,
    openErrorModal,
    closeErrorModal,
  }
})
