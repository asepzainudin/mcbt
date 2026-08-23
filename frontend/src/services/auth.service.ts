import { api } from '../lib/axios'
import type { ApiResponse, LoginResponseData, MeResponseData } from '../types/api'

export const authService = {
  async login(username: string, password: string): Promise<LoginResponseData> {
    const res = await api.post<ApiResponse<LoginResponseData>>('/auth/login', {
      username,
      password,
    })
    return res.data.data
  },

  async logout(): Promise<void> {
    await api.post('/auth/logout')
  },

  async me(): Promise<MeResponseData> {
    const res = await api.get<ApiResponse<MeResponseData>>('/auth/me')
    return res.data.data
  },
}
