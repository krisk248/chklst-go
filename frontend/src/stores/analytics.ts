import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useApi } from '../composables/useApi'

export interface ProjectInsight {
  project_id: number
  name: string
  deployments: number
  success_rate: number
  jira_compliance: number
  dirty_patches: number
  high_churn_days: number
  over_freq_weeks: number
  failed: number
  busiest_day: string
  busiest_day_count: number
  health_score: number
  classification: string // good | attention | dirty
  flags: string[]
}

export interface SummaryCounts {
  deployments: number
  projects: number
  dirty_patches: number
  red_days: number
}

export interface WeekCount {
  label: string
  count: number
}

export interface DeveloperStat {
  name: string
  deployments: number
  success_rate: number
  jira_compliance: number
  projects: number
  failed: number
}

export interface EnvStat {
  name: string
  deployments: number
  success_rate: number
  failed: number
  no_jira: number
}

export interface MonthTrend {
  label: string
  deployments: number
  success_rate: number
  jira_compliance: number
  red_days: number
}

export interface Insights {
  period: string
  total_deployments: number
  total_projects: number
  success_rate: number
  jira_compliance: number
  dirty_patches: number
  projects: ProjectInsight[]
  developers: DeveloperStat[]
  environments: EnvStat[]
  this_week: SummaryCounts
  this_month: SummaryCounts
  weekly_frequency: WeekCount[]
  monthly_trends: MonthTrend[]
}

// normalizeInsights guarantees array fields are never null (defends the UI against
// a backend that serializes an empty slice as JSON null — see the project-health bug).
function normalizeInsights(d: Insights): Insights {
  d.projects = d.projects || []
  d.developers = d.developers || []
  d.environments = d.environments || []
  d.weekly_frequency = d.weekly_frequency || []
  d.monthly_trends = d.monthly_trends || []
  d.projects.forEach(p => { p.flags = p.flags || [] })
  return d
}

export const useAnalyticsStore = defineStore('analytics', () => {
  const insights = ref<Insights | null>(null)
  const narrative = ref<string>('')
  const narrativeSeconds = ref<number | null>(null)
  const isLoading = ref(false)
  const isAnalyzing = ref(false)
  const error = ref<string | null>(null)
  const { get } = useApi()

  const qs = (month: number, year: number) => {
    const p = new URLSearchParams()
    if (month) p.set('month', String(month))
    if (year) p.set('year', String(year))
    const s = p.toString()
    return s ? `?${s}` : ''
  }

  const fetchInsights = async (month: number, year: number) => {
    isLoading.value = true
    error.value = null
    try {
      const response = await get<Insights>(`/analytics/insights${qs(month, year)}`)
      insights.value = normalizeInsights(response.data)
    } catch (err) {
      error.value = 'Failed to load insights'
      console.error(err)
    } finally {
      isLoading.value = false
    }
  }

  const fetchNarrative = async (month: number, year: number): Promise<{ ok: boolean; error?: string }> => {
    isAnalyzing.value = true
    error.value = null
    try {
      const response = await get<{ narrative: string; generation_seconds: number; insights: Insights }>(
        `/analytics/narrative${qs(month, year)}`
      )
      narrative.value = response.data.narrative
      narrativeSeconds.value = response.data.generation_seconds
      insights.value = normalizeInsights(response.data.insights)
      return { ok: true }
    } catch (err: any) {
      const msg = err?.response?.data?.error || 'AI analysis failed'
      error.value = msg
      return { ok: false, error: msg }
    } finally {
      isAnalyzing.value = false
    }
  }

  return { insights, narrative, narrativeSeconds, isLoading, isAnalyzing, error, fetchInsights, fetchNarrative }
})
