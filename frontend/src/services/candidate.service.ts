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
