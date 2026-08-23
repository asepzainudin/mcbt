import { api } from '../lib/axios'
import type {
  AcademicYear,
  ApiResponse,
  PaginationMeta,
  SchoolClass,
  Semester,
  Subject,
} from '../types/api'

export interface ListParams {
  page?: number
  limit?: number
  search?: string
  academic_year_id?: string
}

function clean(params: ListParams): Record<string, string | number> {
  const out: Record<string, string | number> = {}
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== '') out[k] = v as string | number
  }
  return out
}

export interface AcademicYearPayload {
  year: string
  semester: Semester
}

export interface ClassPayload {
  name: string
  academic_year_id: string
}

export interface SubjectPayload {
  code: string
  name: string
  description?: string
}

type ListResult<T> = { data: T[]; meta: PaginationMeta | null }

async function list<T>(resource: string, params: ListParams): Promise<ListResult<T>> {
  const res = await api.get<ApiResponse<T[]>>(`/${resource}`, {
    params: clean(params),
  })
  return { data: res.data.data, meta: res.data.meta ?? null }
}

export const masterDataService = {
  academicYears: {
    list: (p: ListParams) => list<AcademicYear>('academic-years', p),
    create: async (payload: AcademicYearPayload) =>
      (await api.post<ApiResponse<AcademicYear>>('/academic-years', payload)).data.data,
    update: async (id: string, payload: AcademicYearPayload) =>
      (await api.put<ApiResponse<AcademicYear>>(`/academic-years/${id}`, payload)).data.data,
    activate: async (id: string) =>
      (await api.patch<ApiResponse<AcademicYear>>(`/academic-years/${id}/activate`)).data.data,
    remove: async (id: string) => {
      await api.delete(`/academic-years/${id}`)
    },
  },

  classes: {
    list: (p: ListParams) => list<SchoolClass>('classes', p),
    create: async (payload: ClassPayload) =>
      (await api.post<ApiResponse<SchoolClass>>('/classes', payload)).data.data,
    update: async (id: string, payload: ClassPayload) =>
      (await api.put<ApiResponse<SchoolClass>>(`/classes/${id}`, payload)).data.data,
    remove: async (id: string) => {
      await api.delete(`/classes/${id}`)
    },
  },

  subjects: {
    list: (p: ListParams) => list<Subject>('subjects', p),
    create: async (payload: SubjectPayload) =>
      (await api.post<ApiResponse<Subject>>('/subjects', payload)).data.data,
    update: async (id: string, payload: SubjectPayload) =>
      (await api.put<ApiResponse<Subject>>(`/subjects/${id}`, payload)).data.data,
    remove: async (id: string) => {
      await api.delete(`/subjects/${id}`)
    },
  },
}
