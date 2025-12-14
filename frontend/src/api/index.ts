import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const api = axios.create({
  baseURL: '/api',
  timeout: 60000 // Increased for video uploads
})

api.interceptors.request.use((config) => {
  const authStore = useAuthStore()
  if (authStore.token) {
    config.headers.Authorization = `Bearer ${authStore.token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Auth APIs
export const authApi = {
  login: (username: string, password: string) => api.post('/auth/login', { username, password }),
  sendCode: (email: string) => api.post('/auth/send-code', { email }),
  verify: (email: string, code: string) => api.post('/auth/verify', { email, code })
}

// Legacy FaceSwap APIs (v1 - mock)
export const faceswapApi = {
  upload: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/faceswap/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  swap: (mediaId: string, faceIds: string[], model: string) =>
    api.post('/faceswap/swap', { media_id: mediaId, face_ids: faceIds, model }),
  getTask: (taskId: string) => api.get(`/faceswap/tasks/${taskId}`)
}

// Types for v2 API
export interface DetectedFace {
  index: number
  face_id: number           // VModel face ID
  bbox?: number[]           // [x, y, width, height] (optional)
  landmarks_str?: string    // Legacy field (optional)
  thumbnail?: string        // Face thumbnail URL
}

export interface DetectFacesResponse {
  code: number
  data?: {
    faces: DetectedFace[]
    detect_id?: string      // VModel detect_id for swap
    frame_image: string
  }
  msg?: string
}

export interface FaceSwapPair {
  source_image_url: string  // New face
  face_id: number           // VModel: target face ID
  landmarks_str?: string    // Legacy field (optional)
}

export interface CreateFaceSwapRequest {
  target_video_url: string
  detect_id?: string        // VModel: detection ID
  frame_image_url?: string  // Legacy field (optional)
  face_swaps: FaceSwapPair[]
  face_enhance?: boolean
}

export interface TaskStatusResponse {
  code: number
  data?: {
    task_id: string
    status: 'queuing' | 'processing' | 'completed' | 'failed'
    result_url?: string
    error?: string
  }
  msg?: string
}

// V2 APIs (VModel integration)
export const faceswapApiV2 = {
  // Upload video/image
  uploadMedia: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post<{ url: string; key: string }>('/v2/media/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 300000 // 5 min for large videos
    })
  },

  // Upload face image (for replacement)
  uploadFace: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post<{ url: string; key: string }>('/v2/media/upload/face', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  // Upload video frame for detection
  uploadFrame: (file: File | Blob) => {
    const formData = new FormData()
    formData.append('file', file, 'frame.jpg')
    return api.post<{ url: string; key: string }>('/v2/media/upload/frame', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  // Detect faces from image URL
  detectFaces: (imageUrl: string) =>
    api.post<DetectFacesResponse>('/v2/face/detect', { image_url: imageUrl }),

  // Detect faces from uploaded file
  detectFacesFromUpload: (file: File | Blob) => {
    const formData = new FormData()
    formData.append('file', file, 'frame.jpg')
    return api.post<DetectFacesResponse>('/v2/face/detect/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  // Create face swap task
  createSwapTask: (req: CreateFaceSwapRequest) =>
    api.post<{ code: number; data?: { task_id: string; status: string }; msg?: string }>(
      '/v2/faceswap/create',
      req
    ),

  // Get task status
  getTaskStatus: (taskId: string) =>
    api.get<TaskStatusResponse>(`/v2/faceswap/task/${taskId}`)
}

export default api
