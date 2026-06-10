import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useApi } from '../composables/useApi'

export const useAuthStore = defineStore('auth', () => {
  const authEnabled = ref(false)
  const authenticated = ref(true) // optimistic until status says otherwise
  const checked = ref(false)
  const { get, post } = useApi()

  const checkStatus = async () => {
    try {
      const r = await get<{ auth_enabled: boolean; authenticated: boolean }>('/auth/status')
      authEnabled.value = r.data.auth_enabled
      authenticated.value = r.data.authenticated
    } catch {
      /* network error — keep optimistic defaults */
    } finally {
      checked.value = true
    }
  }

  const login = async (password: string): Promise<{ ok: boolean; error?: string }> => {
    try {
      await post('/auth/login', { password })
      authenticated.value = true
      return { ok: true }
    } catch (err: any) {
      return { ok: false, error: err?.response?.data?.error || 'Login failed' }
    }
  }

  const logout = async () => {
    try { await post('/auth/logout', {}) } catch { /* ignore */ }
    authenticated.value = false
  }

  // Called by the axios 401 interceptor when a session expires mid-use.
  const onUnauthorized = () => { authenticated.value = false }

  return { authEnabled, authenticated, checked, checkStatus, login, logout, onUnauthorized }
})
