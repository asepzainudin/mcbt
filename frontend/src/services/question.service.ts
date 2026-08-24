import { api } from '../lib/axios'
import type {
  ApiResponse,
  BankPayload,
  BankStatus,
  OptionPayload,
  PaginationMeta,
  Question,
  QuestionBank,
  QuestionPayload,
  QuestionType,
} from '../types/api'

export interface BankListParams {
  page?: number
  limit?: number
  search?: string
  subject_id?: string
}

export interface QuestionListParams {
  page?: number
  limit?: number
  search?: string
  bank_id?: string
  type?: QuestionType
}

function clean(p: object): Record<string, string | number> {
  const out: Record<string, string | number> = {}
  for (const [k, v] of Object.entries(p)) {
    if (v !== undefined && v !== '') out[k] = v as string | number
  }
  return out
}

export const bankService = {
  async list(params: BankListParams) {
    const res = await api.get<ApiResponse<QuestionBank[]>>('/question-banks', { params: clean(params) })
    return {
      data: res.data.data,
      meta: res.data.meta as PaginationMeta | null,
    }
  },

  async get(id: string): Promise<QuestionBank> {
    return (await api.get<ApiResponse<QuestionBank>>(`/question-banks/${id}`)).data.data
  },

  async create(payload: BankPayload): Promise<QuestionBank> {
    return (await api.post<ApiResponse<QuestionBank>>('/question-banks', payload)).data.data
  },

  async update(id: string, payload: Partial<BankPayload>): Promise<QuestionBank> {
    return (await api.put<ApiResponse<QuestionBank>>(`/question-banks/${id}`, payload)).data.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/question-banks/${id}`)
  },

  async clone(id: string): Promise<QuestionBank> {
    return (await api.post<ApiResponse<QuestionBank>>(`/question-banks/${id}/clone`)).data.data
  },

  async publish(id: string): Promise<QuestionBank> {
    return (await api.patch<ApiResponse<QuestionBank>>(`/question-banks/${id}/publish`)).data.data
  },

  async archive(id: string): Promise<QuestionBank> {
    return (await api.patch<ApiResponse<QuestionBank>>(`/question-banks/${id}/archive`)).data.data
  },
}

export const questionService = {
  async list(params: QuestionListParams) {
    const res = await api.get<ApiResponse<Question[]>>('/questions', { params: clean(params) })
    return {
      data: res.data.data,
      meta: res.data.meta as PaginationMeta | null,
    }
  },

  async get(id: string): Promise<Question> {
    return (await api.get<ApiResponse<Question>>(`/questions/${id}`)).data.data
  },

  async createInBank(bankId: string, payload: QuestionPayload): Promise<Question> {
    return (
      await api.post<ApiResponse<Question>>(`/question-banks/${bankId}/questions`, payload)
    ).data.data
  },

  async update(id: string, payload: QuestionPayload): Promise<Question> {
    return (await api.put<ApiResponse<Question>>(`/questions/${id}`, payload)).data.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/questions/${id}`)
  },

  async reorderOptions(questionId: string, optionIdsOrder: string[]): Promise<Question> {
    return (
      await api.put<ApiResponse<Question>>(`/questions/${questionId}/options/reorder`, {
        option_ids_order: optionIdsOrder,
      })
    ).data.data
  },

  async setCorrectOption(questionId: string, optionId: string): Promise<Question> {
    return (
      await api.put<ApiResponse<Question>>(`/questions/${questionId}/options/${optionId}`, {
        is_correct: true,
      })
    ).data.data
  },

  async preview(id: string) {
    return (
      await api.get<
        ApiResponse<{
          type: string
          content_html: string
          score_weight: number
          options: { option_key: string; text: string; media?: unknown }[]
          correct_keys: string[]
          answer_keys: string[]
          explanation: string | null
        }>
      >(`/questions/${id}/preview`)
    ).data.data
  },
}

export type { BankStatus, Question, QuestionBank, QuestionType, QuestionPayload, OptionPayload }
