/**
 * 文件相关 API 接口
 */
import { http, type ApiResponse } from '../../utils/http'
import type { FileUploadResponse } from '../../types/file'

/**
 * 上传文件
 */
export const uploadFile = async (file: File): Promise<FileUploadResponse> => {
  const formData = new FormData()
  formData.append('file', file)

  const response = await http.post<ApiResponse<FileUploadResponse>>('/file/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })

  if (response.data.code === 0 && response.data.data) {
    return response.data.data
  }
  throw new Error(response.data.message || '文件上传失败')
}
