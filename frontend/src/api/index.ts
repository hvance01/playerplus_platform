import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000
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
  sendCode: (email: string) => api.post('/auth/send-code', { email }),
  verify: (email: string, code: string) => api.post('/auth/verify', { email, code })
}

// FaceSwap APIs
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

export default api
