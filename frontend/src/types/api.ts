export interface ApiResponse<T = unknown> {
  success: boolean
  message: string
  data: T
  meta?: PaginationMeta | null
}

export interface ApiError {
  success: false
  message: string
  errors?: Record<string, string>
}

export interface PaginationMeta {
  page: number
  limit: number
  total_items: number
  total_pages: number
}

export interface LoginResponseData {
  user_id: string
  username: string
  roles: string[]
}

export interface MeResponseData {
  id: string
  name: string
  username: string
  roles: string[]
}

export interface RoleItem {
  id: string
  code: string
}

export type Semester = 'ODD' | 'EVEN'

export interface AcademicYear {
  id: string
  year: string
  semester: Semester
  start_date: string | null
  end_date: string | null
  is_active: boolean
}

export interface SchoolClass {
  id: string
  name: string
  grade_level: number | null
  academic_year_id: string
  academic_year?: AcademicYear
}

export interface Subject {
  id: string
  code: string
  name: string
  description: string | null
}

export interface User {
  id: string
  username: string
  name: string
  roles: string[]
}
