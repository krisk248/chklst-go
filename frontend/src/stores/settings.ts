import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useApi } from '../composables/useApi'

export interface AppSettings {
  default_deployed_by: string
  excel_export_path: string
  auto_clear_after_save: boolean

  // Jira Integration
  jira_url: string
  jira_email: string
  jira_token: string
  jira_project: string

  // Webhook Notifications
  teams_webhook_url: string
  webhooks_enabled: boolean

  // AI / Ollama (Parson)
  ai_enabled: boolean
  ollama_url: string
  ai_model: string
  planned_weekly: string
  parson_system_prompt: string
  parson_temperature: number

  // Email (SMTP)
  smtp_host: string
  smtp_port: number
  smtp_user: string
  smtp_password: string
  smtp_from: string
  smtp_to: string
  smtp_cc: string
  smtp_security: string
  smtp_test_to: string

  // Scheduler
  summary_schedule_enabled: boolean
  summary_generate_time: string
  summary_send_time: string
  summary_weekdays: string
  summary_auto_send: boolean
}

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<AppSettings>({
    default_deployed_by: 'Kannan',
    excel_export_path: '/reports',
    auto_clear_after_save: false,
    // Jira Integration
    jira_url: '',
    jira_email: '',
    jira_token: '',
    jira_project: 'PAT',
    // Webhooks
    teams_webhook_url: '',
    webhooks_enabled: false,
    // AI / Ollama (Parson)
    ai_enabled: false,
    ollama_url: 'http://localhost:11434',
    ai_model: 'gemma4:e4b',
    planned_weekly: '',
    parson_system_prompt: '',
    parson_temperature: 0.4,
    // Email (SMTP)
    smtp_host: '',
    smtp_port: 587,
    smtp_user: '',
    smtp_password: '',
    smtp_from: '',
    smtp_to: '',
    smtp_cc: '',
    smtp_security: 'tls',
    smtp_test_to: '',
    // Scheduler
    summary_schedule_enabled: false,
    summary_generate_time: '18:30',
    summary_send_time: '20:00',
    summary_weekdays: '0,1,2,3,4',
    summary_auto_send: true,
  })
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const { get, post } = useApi()

  // Result of the last "Test AI" call, surfaced in the Settings UI.
  interface AITestResult {
    ok: boolean
    version?: string
    model?: string
    sample?: string
    stage?: string
    error?: string
  }

  interface AIModel {
    name: string
    size_gb?: number
    size_hint?: string
    note?: string
  }
  interface AIModelsResult {
    installed: AIModel[]
    suggested: AIModel[]
    base_url?: string
    error?: string
  }

  const fetchAIModels = async (): Promise<AIModelsResult> => {
    try {
      const response = await get<AIModelsResult>('/ai/models')
      return response.data
    } catch (err: any) {
      const data = err?.response?.data
      if (data) return data as AIModelsResult
      return { installed: [], suggested: [], error: 'Could not reach the server' }
    }
  }

  const testEmail = async (): Promise<{ ok: boolean; error?: string; to?: string[]; cc?: string[] }> => {
    try {
      const response = await post<{ ok: boolean; to?: string[]; cc?: string[] }>('/email/test', {})
      return response.data
    } catch (err: any) {
      return { ok: false, error: err?.response?.data?.error || 'Email test failed' }
    }
  }

  const fetchParsonDefaults = async (): Promise<{ system_prompt: string; temperature: number }> => {
    const response = await get<{ system_prompt: string; temperature: number }>('/ai/defaults')
    return response.data
  }

  const testAI = async (): Promise<AITestResult> => {
    try {
      const response = await post<AITestResult>('/ai/test', {})
      return response.data
    } catch (err: any) {
      // The endpoint returns structured JSON even on failure (4xx/5xx).
      const data = err?.response?.data
      if (data) return data as AITestResult
      return { ok: false, stage: 'request', error: 'Could not reach the server' }
    }
  }

  const fetchSettings = async () => {
    isLoading.value = true
    error.value = null
    try {
      const response = await get<AppSettings>('/settings')
      settings.value = response.data || settings.value
    } catch (err) {
      error.value = 'Failed to fetch settings'
      console.error(err)
    } finally {
      isLoading.value = false
    }
  }

  const saveSettings = async (newSettings: Partial<AppSettings>) => {
    isLoading.value = true
    error.value = null
    try {
      const updated = { ...settings.value, ...newSettings }
      const response = await post<AppSettings>('/settings', updated)
      settings.value = response.data || updated
      return true
    } catch (err) {
      error.value = 'Failed to save settings'
      console.error(err)
      return false
    } finally {
      isLoading.value = false
    }
  }

  return {
    settings,
    isLoading,
    error,
    fetchSettings,
    saveSettings,
    testAI,
    fetchAIModels,
    fetchParsonDefaults,
    testEmail,
  }
})
