import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const email = ref<string | null>(localStorage.getItem('email'))

  const isAuthenticated = computed(() => !!token.value)

  function setAuth(newToken: string, newEmail: string) {
    token.value = newToken
    email.value = newEmail
    localStorage.setItem('token', newToken)
    localStorage.setItem('email', newEmail)
  }

  function logout() {
    token.value = null
    email.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('email')
  }

  return {
    token,
    email,
    isAuthenticated,
    setAuth,
    logout
  }
})
