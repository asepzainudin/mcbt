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
  attempt: { status: string; expires_at: string; attempt_no: number; submitted_at: string | null }
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

  async heartbeat(attemptId: string) {
    return (
      await api.post<
        ApiResponse<{
          server_time: string
          remaining_seconds: number
          is_expired: boolean
          attempt_status: string
          submitted_at: string | null
        }>
      >(`/candidate/attempts/${attemptId}/heartbeat`)
    ).data.data
  },

  async autosave(
    attemptId: string,
    answers: { question_id: string; value: string }[],
  ): Promise<{ saved_count: number }> {
    return (
      await api.post<ApiResponse<{ saved_count: number }>>(
        `/candidate/attempts/${attemptId}/autosave`,
        { answers },
      )
    ).data.data
  },

  async submit(attemptId: string, confirm = true): Promise<StartAttemptResult & { status: string; submitted_at: string | null }> {
    return (
      await api.post<ApiResponse<StartAttemptResult & { status: string; submitted_at: string | null }>>(
        `/candidate/attempts/${attemptId}/submit`,
        { confirm_submit: confirm },
      )
    ).data.data
  },

  async getDiscussion(attemptId: string) {
    return (
      await api.get<
        ApiResponse<{
          question_id: string
          section_name: string
          type: string
          text: string
          score_weight: number
          media?: { url: string } | null
          media_position: string
          options: { option_key: string; text: string; media_url?: string }[]
          correct_keys: string[]
          explanation: string | null
          answer_value: string
          is_correct: boolean | null
          score: number | null
          feedback: string | null
          is_flagged: boolean
        }[]>
      >(`/candidate/attempts/${attemptId}/discussion`)
    ).data.data ?? []
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
