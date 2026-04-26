import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/api/index.js'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isAdmin = computed(() => user.value?.role === 'admin')
  const isDev = computed(() => user.value?.role === 'developer' || user.value?.role === 'admin')

  function setAuth(tokenStr, userData) {
    token.value = tokenStr
    user.value = userData
    localStorage.setItem('token', tokenStr)
    localStorage.setItem('user', JSON.stringify(userData))
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  async function fetchProfile() {
    try {
      const res = await api.get('/auth/profile')
      user.value = res.data.data
      localStorage.setItem('user', JSON.stringify(res.data.data))
    } catch (e) {
      logout()
    }
  }

  return { token, user, isAdmin, isDev, setAuth, logout, fetchProfile }
})
