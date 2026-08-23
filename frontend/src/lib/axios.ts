import axios, {
  AxiosError,
  AxiosHeaders,
  type AxiosRequestConfig,
} from 'axios'
import { getCookie } from './cookies'
import { useUiStore } from '../stores/ui'
import { useAuthStore } from '../stores/auth'
import type { ApiError } from '../types/api'

const UNSAFE_METHODS = new Set(['post', 'put', 'patch', 'delete'])

export const api = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
  headers: { Accept: 'application/json' },
})

api.interceptors.request.use((config) => {
  if (config.method && UNSAFE_METHODS.has(config.method)) {
    const csrf = getCookie('csrf_token')
    if (csrf) {
      ;(config.headers as AxiosHeaders).set('X-CSRF-Token', csrf)
    }
  }
  return config
})

let refreshInflight: Promise<boolean> | null = null

async function refreshSession(): Promise<boolean> {
  if (!refreshInflight) {
    refreshInflight = api
      .post('/auth/refresh-token')
      .then(() => true)
      .catch(() => false)
      .finally(() => {
        refreshInflight = null
      })
  }
  return refreshInflight
}

function redirectToLogin() {
  const auth = useAuthStore()
  auth.forceLogout()
  if (window.location.pathname !== '/login') {
    window.location.href = '/login'
  }
}

type RetriableConfig = AxiosRequestConfig & { _retry?: boolean }

async function handle401(error: AxiosError): Promise<unknown> {
  const original = error.config as RetriableConfig | undefined
  const url = original?.url ?? ''

  if (!original || original._retry || url.includes('/auth/login') || url.includes('/auth/refresh-token')) {
    redirectToLogin()
    return Promise.reject(error)
  }

  const refreshed = await refreshSession()
  if (!refreshed) {
    redirectToLogin()
    return Promise.reject(error)
  }

  original._retry = true
  const csrf = getCookie('csrf_token')
  if (csrf) {
    ;(original.headers as AxiosHeaders)?.set('X-CSRF-Token', csrf)
  }
  return api.request(original)
}

export function extractFieldErrors(error: unknown): Record<string, string> {
  if (axios.isAxiosError<ApiError>(error)) {
    const fields = error.response?.data?.errors
    if (fields && Object.keys(fields).length > 0) {
      return fields
    }
    return { _form: error.response?.data?.message ?? 'Validasi gagal' }
  }
  return { _form: 'Terjadi kesalahan tak terduga' }
}

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiError>) => {
    const ui = useUiStore()
    const status = error.response?.status

    switch (status) {
      case 401:
        return handle401(error)

      case 403: {
        const message = error.response?.data?.message ?? 'Anda tidak memiliki akses'
        ui.toastError(message)
        break
      }

      case 422: {
        const message = error.response?.data?.message ?? 'Validasi gagal'
        ui.toastWarning(message)
        break
      }

      case 429:
        ui.toastWarning('Terlalu banyak permintaan. Coba lagi nanti.')
        break

      default: {
        if (status && status >= 500) {
          ui.openErrorModal(
            'Terjadi kesalahan pada server. Tim kami sedang menangani masalah ini.',
          )
        }
        break
      }
    }

    return Promise.reject(error)
  },
)

export function apiErrorMessage(error: unknown, fallback = 'Terjadi kesalahan'): string {
  if (axios.isAxiosError<ApiError>(error)) {
    return error.response?.data?.message ?? fallback
  }
  return fallback
}
