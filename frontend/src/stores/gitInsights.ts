import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useApi } from '../composables/useApi'

export interface AuthorStat { name: string; commits: number; repos: number; merges: number; last_active: string }
export interface DayCount { label: string; date: string; count: number }
export interface RepoStat { name: string; commits: number; merges: number }

export interface GitInsights {
  period_days: number
  total_commits: number
  total_merges: number
  repos: number
  authors: AuthorStat[]
  by_day: DayCount[]
  by_repo: RepoStat[]
  by_hour: number[]
}

export const useGitInsightsStore = defineStore('gitInsights', () => {
  const insights = ref<GitInsights | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const { get, post } = useApi()

  const fetchInsights = async (days = 30) => {
    isLoading.value = true
    error.value = null
    try {
      const r = await get<GitInsights>(`/git/insights?days=${days}`)
      insights.value = r.data
    } catch (err: any) {
      error.value = err?.response?.data?.error || 'Failed to load git insights'
      insights.value = null
    } finally {
      isLoading.value = false
    }
  }

  const test = async (): Promise<{ ok: boolean; error?: string; repos_found?: number; owner?: string }> => {
    try {
      const r = await post<{ ok: boolean; repos_found: number; owner: string }>('/git/test', {})
      return r.data
    } catch (err: any) {
      return { ok: false, error: err?.response?.data?.error || 'Test failed' }
    }
  }

  return { insights, isLoading, error, fetchInsights, test }
})
