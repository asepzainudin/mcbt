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
  email: string
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

export interface Teacher {
  id: string
  user_id: string
  nip: string | null
  phone: string | null
  address: string | null
  user?: Pick<User, 'id' | 'username' | 'name' | 'email'>
}

export interface Student {
  id: string
  user_id: string
  nis: string
  class_id: string | null
  phone: string | null
  address: string | null
  user?: Pick<User, 'id' | 'username' | 'name' | 'email'>
  class?: SchoolClass | null
}

export interface ImportRowError {
  row: number
  field: string
  reason: string
}

export interface ImportResult {
  imported_count: number
  skipped: ImportRowError[]
}

export interface User {
  id: string
  username: string
  name: string
  email: string
  roles: string[]
}

export type QuestionType =
  | 'MULTIPLE_CHOICE'
  | 'TRUE_FALSE'
  | 'MULTIPLE_ANSWER'
  | 'ESSAY'
  | 'SHORT_ANSWER'

export interface MediaRef {
  id: string
  file_path: string
  file_name: string
  mime_type: string
}

export type BankStatus = 'draft' | 'published' | 'archived'

export interface QuestionBank {
  id: string
  code: string
  title: string
  subject_id: string
  academic_year_id: string | null
  status: BankStatus
  description: string | null
  subject?: Subject
  academic_year?: AcademicYear | null
}

export interface QuestionOption {
  id: string
  option_key: string
  label: string
  text: string
  content: string
  is_correct: boolean
  position: number
  media_id: string | null
  media?: MediaRef | null
}

export interface Question {
  id: string
  question_bank_id: string
  type: QuestionType
  question_type: string
  text: string
  content: string
  score_weight: number
  points: number
  explanation: string | null
  answer_keys: string[]
  media_id: string | null
  options?: QuestionOption[]
}

export interface BankPayload {
  code: string
  title: string
  subject_id: string
  academic_year_id?: string | null
  description?: string | null
}

export interface OptionPayload {
  option_key?: string
  text: string
  is_correct: boolean
}

export interface QuestionPayload {
  type: QuestionType
  text: string
  score_weight: number
  explanation?: string | null
  options?: OptionPayload[]
  answer_keys?: string[]
}
