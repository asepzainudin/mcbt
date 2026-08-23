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

export interface User {
  id: string
  username: string
  name: string
  roles: string[]
}
