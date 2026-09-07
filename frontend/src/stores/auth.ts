import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { authApi } from '@/api'
import type { User } from '@/api/types'
import { getToken, setToken } from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(getToken())
  const ready = ref(false)

  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'ADMIN')

  function setSession(newToken: string, newUser: User) {
    token.value = newToken
    user.value = newUser
    setToken(newToken)
  }

  function clear() {
    token.value = null
    user.value = null
    setToken(null)
  }

  async function login(loginId: string, password: string) {
    const res = await authApi.login(loginId, password)
    setSession(res.token, res.user)
  }

  async function register(payload: {
    username: string
    email: string
    password: string
    display_name?: string
  }) {
    const res = await authApi.register(payload)
    setSession(res.token, res.user)
  }

  /** Called once on app start to resolve the current user from a stored token. */
  async function bootstrap() {
    if (!token.value) {
      ready.value = true
      return
    }
    try {
      user.value = await authApi.me()
    } catch {
      clear()
    } finally {
      ready.value = true
    }
  }

  return {
    user,
    token,
    ready,
    isAuthenticated,
    isAdmin,
    login,
    register,
    bootstrap,
    clear,
  }
})
