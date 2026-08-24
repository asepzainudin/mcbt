import { api } from '../lib/axios'
import type {
  ApiResponse,
  ExamSection,
  MapQuestionsPayload,
  SectionPayload,
  SectionQuestion,
} from '../types/api'

export const sectionService = {
  async listByExam(examId: string): Promise<ExamSection[]> {
    return (await api.get<ApiResponse<ExamSection[]>>(`/exams/${examId}/sections`)).data.data
  },

  async create(examId: string, payload: SectionPayload): Promise<ExamSection> {
    return (
      await api.post<ApiResponse<ExamSection>>(`/exams/${examId}/sections`, payload)
    ).data.data
  },

  async update(id: string, payload: SectionPayload): Promise<ExamSection> {
    return (await api.put<ApiResponse<ExamSection>>(`/sections/${id}`, payload)).data.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/sections/${id}`)
  },

  async mapQuestions(
    sectionId: string,
    payload: MapQuestionsPayload,
  ): Promise<{ mapped_count: number; skipped: number }> {
    return (
      await api.post<ApiResponse<{ mapped_count: number; skipped: number }>>(
        `/sections/${sectionId}/questions`,
        payload,
      )
    ).data.data
  },

  async listQuestions(sectionId: string): Promise<SectionQuestion[]> {
    return (await api.get<ApiResponse<SectionQuestion[]>>(`/sections/${sectionId}/questions`)).data
      .data
  },

  async removeQuestion(sectionId: string, questionId: string): Promise<void> {
    await api.delete(`/sections/${sectionId}/questions/${questionId}`)
  },
}
