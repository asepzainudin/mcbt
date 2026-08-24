import { api } from '../lib/axios'
import type { ApiResponse, MediaRef } from '../types/api'

export type MediaUploadType = 'QUESTION_IMAGE' | 'OPTION_IMAGE'

export async function uploadMedia(file: File, type: MediaUploadType): Promise<MediaRef> {
  const form = new FormData()
  form.append('file', file)
  form.append('type', type)
  const res = await api.post<ApiResponse<MediaRef>>('/media/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return res.data.data
}

export const mediaService = { uploadMedia }
