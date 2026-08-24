import { api } from '../lib/axios'
import type {
  ApiResponse,
  Exam,
  ExamPayload,
  ExamSettingsPayload,
  ExamStatus,
  PaginationMeta,
} from '../types/api'

export interface ExamListParams {
  page?: number
  limit?: number
  search?: string
  subject_id?: string
  academic_year_id?: string
  status?: ExamStatus
}

function clean(p: object): Record<string, string | number> {
  const out: Record<string, string | number> = {}
  for (const [k, v] of Object.entries(p)) {
    if (v !== undefined && v !== '') out[k] = v as string | number
  }
  return out
}

export const examService = {
  async list(params: ExamListParams) {
    const res = await api.get<ApiResponse<Exam[]>>('/exams', { params: clean(params) })
    return {
      data: res.data.data,
      meta: res.data.meta as PaginationMeta | null,
    }
  },

  async get(id: string): Promise<Exam> {
    return (await api.get<ApiResponse<Exam>>(`/exams/${id}`)).data.data
  },

  async create(payload: ExamPayload): Promise<Exam> {
    return (await api.post<ApiResponse<Exam>>('/exams', payload)).data.data
  },

  async update(id: string, payload: Partial<ExamPayload>): Promise<Exam> {
    return (await api.put<ApiResponse<Exam>>(`/exams/${id}`, payload)).data.data
  },

  async updateSettings(id: string, payload: ExamSettingsPayload): Promise<Exam> {
    return (
      await api.put<ApiResponse<Exam>>(`/exams/${id}/settings`, payload)
    ).data.data
  },

  async publish(id: string): Promise<Exam> {
    return (await api.patch<ApiResponse<Exam>>(`/exams/${id}/publish`)).data.data
  },

  async close(id: string): Promise<Exam> {
    return (await api.patch<ApiResponse<Exam>>(`/exams/${id}/close`)).data.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/exams/${id}`)
  },
}
