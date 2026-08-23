import { api } from '../lib/axios'
import type {
  ApiResponse,
  ImportResult,
  PaginationMeta,
  Teacher,
} from '../types/api'

export interface TeacherListParams {
  page?: number
  limit?: number
  search?: string
}

export interface TeacherPayload {
  username: string
  name: string
  email: string
  nip?: string | null
  phone?: string | null
  address?: string | null
}

export interface TeacherUpdatePayload extends Omit<TeacherPayload, 'username'> {}

function clean(p: TeacherListParams): Record<string, string | number> {
  const out: Record<string, string | number> = {}
  for (const [k, v] of Object.entries(p)) {
    if (v !== undefined && v !== '') out[k] = v as string | number
  }
  return out
}

export const teacherService = {
  async list(
    params: TeacherListParams,
  ): Promise<{ data: Teacher[]; meta: PaginationMeta | null }> {
    const res = await api.get<ApiResponse<Teacher[]>>('/teachers', { params: clean(params) })
    return { data: res.data.data, meta: res.data.meta ?? null }
  },

  async create(payload: TeacherPayload): Promise<Teacher> {
    return (await api.post<ApiResponse<Teacher>>('/teachers', payload)).data.data
  },

  async update(id: string, payload: TeacherUpdatePayload): Promise<Teacher> {
    return (await api.put<ApiResponse<Teacher>>(`/teachers/${id}`, payload)).data.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/teachers/${id}`)
  },

  async import(file: File): Promise<ImportResult> {
    const form = new FormData()
    form.append('file', file)
    const res = await api.post<ApiResponse<ImportResult>>('/teachers/import', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return res.data.data
  },

  templateUrl: '/api/v1/teachers/import/template',
}
