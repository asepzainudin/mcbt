import { api } from '../lib/axios'
import type {
  ApiResponse,
  AssignResult,
  ExamSchedule,
  SchedulePayload,
} from '../types/api'

export const scheduleService = {
  async getByExam(examId: string): Promise<ExamSchedule | null> {
    const res = await api.get<ApiResponse<ExamSchedule | null>>(`/exams/${examId}/schedules`)
    return res.data.data
  },

  async create(examId: string, payload: SchedulePayload): Promise<ExamSchedule> {
    return (
      await api.post<ApiResponse<ExamSchedule>>(`/exams/${examId}/schedules`, payload)
    ).data.data
  },

  async update(id: string, payload: SchedulePayload): Promise<ExamSchedule> {
    return (await api.put<ApiResponse<ExamSchedule>>(`/schedules/${id}`, payload)).data.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/schedules/${id}`)
  },

  async generateToken(id: string): Promise<string> {
    return (
      await api.post<ApiResponse<{ token: string }>>(`/schedules/${id}/generate-token`)
    ).data.data.token
  },
}

export interface ParticipantRow {
  id: string
  student_id: string
  nis: string
  name: string
  class_name: string | null
  assigned_via: 'class' | 'individual'
}

export const participantService = {
  async list(examId: string): Promise<ParticipantRow[]> {
    return (await api.get<ApiResponse<ParticipantRow[]>>(`/exams/${examId}/participants`)).data.data
  },

  async assignClass(examId: string, classIds: string[]): Promise<AssignResult> {
    return (
      await api.post<ApiResponse<AssignResult>>(`/exams/${examId}/participants/assign-class`, {
        class_ids: classIds,
      })
    ).data.data
  },

  async assignIndividual(examId: string, studentIds: string[]): Promise<AssignResult> {
    return (
      await api.post<ApiResponse<AssignResult>>(`/exams/${examId}/participants/assign-individual`, {
        student_ids: studentIds,
      })
    ).data.data
  },

  async remove(examId: string, participantId: string): Promise<void> {
    await api.delete(`/exams/${examId}/participants/${participantId}`)
  },
}
