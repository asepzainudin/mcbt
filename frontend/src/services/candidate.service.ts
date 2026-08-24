import { api } from '../lib/axios'
import type { ApiResponse, CandidateExam, StartAttemptResult } from '../types/api'

export const candidateService = {
  async listExams(): Promise<CandidateExam[]> {
    return (await api.get<ApiResponse<CandidateExam[]>>('/candidate/exams')).data.data ?? []
  },

  async validateToken(examId: string, token: string): Promise<void> {
    await api.post(`/candidate/exams/${examId}/validate-token`, { token })
  },

  async start(examId: string, token?: string): Promise<StartAttemptResult> {
    return (
      await api.post<ApiResponse<StartAttemptResult>>(
        `/candidate/exams/${examId}/start`,
        token ? { token } : {},
      )
    ).data.data
  },
}

export interface AttemptQuestion {
  question_id: string
  section_name: string
  sequence: number
  type: string
  text: string
  score_weight: number
  media?: { url: string; file_name: string } | null
  media_position: 'before' | 'after'
  options: { option_key: string; text: string; media?: { url: string } | null }[]
  answer_value: string
  is_flagged: boolean
  answered_at: string | null
}

export interface AttemptSheet {
  attempt: { status: string; expires_at: string; attempt_no: number }
  questions: AttemptQuestion[]
}

export const attemptService = {
  async getQuestions(attemptId: string): Promise<AttemptSheet> {
    return (
      await api.get<ApiResponse<AttemptSheet>>(`/candidate/attempts/${attemptId}/questions`)
    ).data.data
  },

  async saveAnswer(attemptId: string, questionId: string, answerValue: string) {
    const client_timestamp = Math.floor(Date.now() / 1000)
    return (
      await api.post<ApiResponse<{ answered_at: string; question_id: string }>>(
        `/candidate/attempts/${attemptId}/answers`,
        { question_id: questionId, answer_value: answerValue, client_timestamp },
      )
    ).data.data
  },

  async setFlag(
    attemptId: string,
    questionId: string,
    flagged: boolean,
  ): Promise<{ is_flagged: boolean }> {
    if (flagged) {
      return (
        await api.post(`/candidate/attempts/${attemptId}/questions/${questionId}/flag`)
      ).data.data
    }
    return (
      await api.delete(`/candidate/attempts/${attemptId}/questions/${questionId}/flag`)
    ).data.data
  },
}
