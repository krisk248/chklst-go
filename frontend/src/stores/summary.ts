import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useApi } from '../composables/useApi'
import type { Deployment } from './deployments'

export interface DailySummary {
  id: number
  date: string
  status: string // draft | generated | approved | sent | failed
  notes: string
  planned: string
  roadblocks: string
  generated_subject: string
  generated_body: string
  recipients: string
  sent_at?: string | null
  send_error?: string
  generated_at?: string | null
  generation_seconds?: number
  generated_by_model?: string
  reviewed?: boolean
  locked_at?: string | null
  auto_generated?: boolean
}

export interface Holiday {
  id: number
  date: string
  name: string
}

export interface SummaryStats {
  total: number
  by_environment: Record<string, number>
  dirty_patches: string[]
  high_churn: string[]
  failed_count: number
}

export interface TodayResponse {
  summary: DailySummary
  deployments: Deployment[]
  stats: SummaryStats
}

export const useSummaryStore = defineStore('summary', () => {
  const summary = ref<DailySummary | null>(null)
  const deployments = ref<Deployment[]>([])
  const stats = ref<SummaryStats | null>(null)
  const isLoading = ref(false)
  const isGenerating = ref(false)
  const lastGenSeconds = ref<number | null>(null)
  const lastGenModel = ref<string | null>(null)
  const error = ref<string | null>(null)
  const { get, post, put, delete: deleteApi } = useApi()

  const fetchToday = async (date?: string) => {
    isLoading.value = true
    error.value = null
    try {
      const url = date ? `/summary/today?date=${date}` : '/summary/today'
      const response = await get<TodayResponse>(url)
      summary.value = response.data.summary
      deployments.value = response.data.deployments || []
      // Guard array/map fields against a JSON-null from an empty backend slice.
      const st = response.data.stats
      if (st) {
        st.dirty_patches = st.dirty_patches || []
        st.high_churn = st.high_churn || []
        st.by_environment = st.by_environment || {}
      }
      stats.value = st
    } catch (err) {
      error.value = 'Failed to load daily summary'
      console.error(err)
    } finally {
      isLoading.value = false
    }
  }

  const saveNotes = async (fields: Partial<DailySummary>) => {
    if (!summary.value) return false
    try {
      const response = await put<DailySummary>(`/summary/${summary.value.id}`, {
        ...summary.value,
        ...fields,
      })
      summary.value = response.data
      return true
    } catch (err) {
      error.value = 'Failed to save'
      console.error(err)
      return false
    }
  }

  const generate = async (model?: string): Promise<{ ok: boolean; error?: string }> => {
    if (!summary.value) return { ok: false, error: 'No summary loaded' }
    isGenerating.value = true
    error.value = null
    try {
      const response = await post<{ summary: DailySummary; generation_seconds: number; model: string }>(
        `/summary/${summary.value.id}/generate`,
        model ? { model } : {}
      )
      summary.value = response.data.summary
      lastGenSeconds.value = response.data.generation_seconds
      lastGenModel.value = response.data.model
      return { ok: true }
    } catch (err: any) {
      const msg = err?.response?.data?.error || 'Generation failed'
      error.value = msg
      return { ok: false, error: msg }
    } finally {
      isGenerating.value = false
    }
  }

  // --- Parson activity tracker ---
  const recent = ref<DailySummary[]>([])
  const fetchRecent = async () => {
    try {
      const response = await get<DailySummary[]>('/summary/recent')
      recent.value = response.data || []
    } catch (err) {
      console.error(err)
    }
  }

  // --- Holidays ---
  const holidays = ref<Holiday[]>([])
  const fetchHolidays = async () => {
    try {
      const response = await get<Holiday[]>('/holidays')
      holidays.value = response.data || []
    } catch (err) {
      console.error(err)
    }
  }
  const addHoliday = async (date: string, name: string) => {
    try {
      await post('/holidays', { date, name })
      await fetchHolidays()
      return true
    } catch { return false }
  }
  const deleteHoliday = async (id: number) => {
    try {
      await deleteApi(`/holidays/${id}`)
      holidays.value = holidays.value.filter(h => h.id !== id)
    } catch (err) { console.error(err) }
  }

  const isSending = ref(false)
  const send = async (): Promise<{ ok: boolean; error?: string }> => {
    if (!summary.value) return { ok: false, error: 'No summary loaded' }
    isSending.value = true
    error.value = null
    try {
      const response = await post<{ ok: boolean; summary: DailySummary }>(`/summary/${summary.value.id}/send`, {})
      summary.value = response.data.summary
      return { ok: true }
    } catch (err: any) {
      const msg = err?.response?.data?.error || 'Send failed'
      error.value = msg
      return { ok: false, error: msg }
    } finally {
      isSending.value = false
    }
  }

  return {
    summary,
    deployments,
    stats,
    isLoading,
    isGenerating,
    isSending,
    send,
    lastGenSeconds,
    lastGenModel,
    error,
    fetchToday,
    saveNotes,
    generate,
    recent,
    fetchRecent,
    holidays,
    fetchHolidays,
    addHoliday,
    deleteHoliday,
  }
})
