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
  url: string
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
  media_position: 'before' | 'after'
  media?: MediaRef | null
  options?: QuestionOption[]
  is_used: boolean
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
  media_id?: string | null
}

export interface QuestionPayload {
  type: QuestionType
  text: string
  score_weight: number
  explanation?: string | null
  media_id?: string | null
  media_position?: 'before' | 'after'
  options?: OptionPayload[]
  answer_keys?: string[]
}

export type ExamStatus = 'draft' | 'published' | 'closed'

export interface Exam {
  id: string
  title: string
  description: string | null
  subject_id: string
  academic_year_id: string | null
  question_bank_id: string | null
  status: ExamStatus
  duration_minutes: number
  max_attempts: number
  passing_grade: number
  randomize_questions: boolean
  randomize_options: boolean
  allow_backtrack: boolean
  auto_submit: boolean
  show_result_immediately: boolean
  allow_discussion: boolean
  negative_marking: boolean
  negative_value: number
  token_enabled: boolean
  attempts_count: number
  exam_token?: string | null
  subject?: Subject
  academic_year?: AcademicYear | null
  question_bank?: QuestionBank | null
}

export interface ExamPayload {
  title: string
  description?: string | null
  subject_id: string
  academic_year_id?: string | null
  question_bank_id?: string | null
}

export interface ExamSettingsPayload {
  duration_minutes: number
  max_attempts: number
  passing_grade: number
  randomize_questions: boolean
  randomize_options: boolean
  allow_backtrack: boolean
  auto_submit: boolean
  show_result_immediately: boolean
  allow_discussion: boolean
  negative_marking: boolean
  negative_value: number
  token_enabled: boolean
}

export interface ExamSection {
  id: string
  exam_id: string
  name: string
  sequence: number
  question_count?: number
}

export interface SectionQuestion {
  id: string
  type: string
  text: string
  score_weight: number
  answer_keys: string[]
  question_bank_id: string
}

export interface SectionPayload {
  name: string
  sequence: number
}

export interface MapQuestionsPayload {
  question_bank_ids: string[]
  total_random_questions: number
}

export interface ExamSchedule {
  id: string
  exam_id: string
  start_time: string
  end_time: string
  token: string
}

export interface SchedulePayload {
  start_time: string
  end_time: string
  token?: string
}

export interface ExamParticipant {
  id: string
  exam_id: string
  student_id: string
  assigned_via: 'class' | 'individual'
  nis: string
  name: string
  class_name: string | null
}

export interface AssignResult {
  assigned: number
  skipped: number
}

export interface CandidateExam {
  exam_id: string
  title: string
  subject_code: string
  subject_name: string
  duration_minutes: number
  start_time: string
  end_time: string
  token_enabled: boolean
  show_result_immediately: boolean
  max_attempts: number
  passing_grade: number
  attempts_used: number
  active_attempt_id: string | null
  active_expires_at: string | null
  last_status: string | null
  score: number | null
  submitted_at: string | null
  has_essay: boolean
  essay_ungraded: boolean
}

export interface StartAttemptResult {
  attempt_id: string
  started_at: string
  expires_at: string
  attempt_no: number
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

export type ReportStatus = 'pending' | 'reviewing' | 'resolved' | 'rejected'

export interface QuestionReport {
  id: string
  attempt_id: string
  question_id: string
  student_id: string
  reason: string
  status: ReportStatus
  resolution?: string | null
  resolved_by?: string | null
  resolved_at?: string | null
  created_at: string
}
