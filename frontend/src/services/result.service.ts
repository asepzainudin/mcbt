import { api } from '../lib/axios'
import type { ApiResponse, Exam } from '../types/api'

export interface ExamResultRow {
  rank: number
  attempt_id: string
  student_id: string
  student_name: string
  nis: string
  class_name: string | null
  score: number | null
  passing_grade: number
  passed: boolean
}

export interface StudentResultRow {
  exam_id: string
  exam_title: string
  subject_name: string
  status: string
  score: number | null
  passing_grade: number
  results_published: boolean
  show_result_immediately: boolean
  has_essay: boolean
  essay_ungraded: boolean
  submitted_at: string | null
}

export const resultService = {
  async examResults(examId: string, classId?: string): Promise<ExamResultRow[]> {
    const params: Record<string, string> = {}
    if (classId) params.class_id = classId
    return (
      await api.get<ApiResponse<ExamResultRow[]>>(`/exams/${examId}/results`, { params })
    ).data.data ?? []
  },

  async myResults(): Promise<StudentResultRow[]> {
    return (
      await api.get<ApiResponse<StudentResultRow[]>>('/candidate/results')
    ).data.data ?? []
  },

  async publishResults(examId: string, published: boolean): Promise<void> {
    await api.patch(`/exams/${examId}/publish-results`, { published })
  },
}

export type { Exam }
