import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { authService } from '../services/auth.service'
import { extractFieldErrors } from '../lib/axios'
import type { User } from '../types/api'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const loading = ref(false)

  const isAuthenticated = computed(() => user.value !== null)

  async function bootstrap(): Promise<void> {
    if (user.value) return
    try {
      user.value = await authService.me()
    } catch {
      user.value = null
    }
  }

  async function login(username: string, password: string): Promise<LoginResult> {
    loading.value = true
    try {
      await authService.login(username, password)
      user.value = await authService.me()
      return { ok: true }
    } catch (err) {
      return { ok: false, fieldErrors: extractFieldErrors(err) }
    } finally {
      loading.value = false
    }
  }

  async function logout(): Promise<void> {
    try {
      await authService.logout()
    } finally {
      forceLogout()
    }
  }

  function forceLogout(): void {
    user.value = null
  }

  return {
    user,
    loading,
    isAuthenticated,
    bootstrap,
    login,
    logout,
    forceLogout,
  }
})

interface LoginResult {
  ok: boolean
  fieldErrors?: Record<string, string>
}
