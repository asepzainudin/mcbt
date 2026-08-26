import { api } from '../lib/axios'
import type {
  ApiResponse,
  ImportResult,
  PaginationMeta,
  SchoolClass,
  Student,
} from '../types/api'

export interface StudentListParams {
  page?: number
  limit?: number
  search?: string
  class_id?: string
}

export interface StudentPayload {
  username?: string
  name: string
  email: string
  nis: string
  class_id?: string | null
  phone?: string | null
  address?: string | null
}

function clean(p: StudentListParams): Record<string, string | number> {
  const out: Record<string, string | number> = {}
  for (const [k, v] of Object.entries(p)) {
    if (v !== undefined && v !== '') out[k] = v as string | number
  }
  return out
}

export const studentService = {
  async list(
    params: StudentListParams,
  ): Promise<{ data: Student[]; meta: PaginationMeta | null }> {
    const res = await api.get<ApiResponse<Student[]>>('/students', { params: clean(params) })
    return { data: res.data.data, meta: res.data.meta ?? null }
  },

  async create(payload: StudentPayload): Promise<Student> {
    return (await api.post<ApiResponse<Student>>('/students', payload)).data.data
  },

  async update(id: string, payload: StudentPayload): Promise<Student> {
    return (await api.put<ApiResponse<Student>>(`/students/${id}`, payload)).data.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/students/${id}`)
  },

  async changeClass(id: string, targetClassId: string): Promise<Student> {
    return (
      await api.post<ApiResponse<Student>>(`/students/${id}/change-class`, {
        target_class_id: targetClassId,
      })
    ).data.data
  },

  async resetPassword(id: string, newPassword?: string): Promise<{ new_password: string }> {
    const body = newPassword ? { new_password: newPassword } : {}
    return (
      await api.post<ApiResponse<{ new_password: string }>>(`/students/${id}/reset-password`, body)
    ).data.data
  },

  async import(file: File): Promise<ImportResult> {
    const form = new FormData()
    form.append('file', file)
    const res = await api.post<ApiResponse<ImportResult>>('/students/import', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return res.data.data
  },

  templateUrl: '/api/v1/students/import/template',
}

export type ClassOption = Pick<SchoolClass, 'id' | 'name'>
