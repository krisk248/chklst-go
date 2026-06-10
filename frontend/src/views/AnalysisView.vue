<template>
  <div class="space-y-4">
    <div class="page-header">
      <div>
        <h1 class="page-title text-xl">Deployment Analysis</h1>
        <p class="text-gray-400 text-sm">Insights and trends from your deployment data</p>
      </div>
      <div class="flex items-center gap-2">
        <Select
          v-model="selectedMonth"
          :options="monthOptions"
          label=""
          class="w-32"
        />
        <Select
          v-model="selectedYear"
          :options="yearOptions"
          label=""
          class="w-24"
        />
      </div>
    </div>

    <!-- AI Project Health (Parson) -->
    <Card title="Project Health — AI Analysis">
      <template #header-right>
        <div class="flex items-center gap-2">
          <span v-if="analyticsStore.insights" class="text-xs text-gray-400">{{ analyticsStore.insights.period }}</span>
          <Button variant="primary" size="sm" @click="runAnalysis" :disabled="analyticsStore.isAnalyzing">
            <RefreshCw v-if="analyticsStore.isAnalyzing" class="w-4 h-4 animate-spin" />
            <Sparkles v-else class="w-4 h-4" />
            {{ analyticsStore.isAnalyzing ? 'Parson is analyzing...' : 'Generate AI Insights' }}
          </Button>
        </div>
      </template>

      <!-- This week / this month summary + frequency graph -->
      <div v-if="analyticsStore.insights" class="mb-4 space-y-3">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div class="p-3 bg-surface-deep rounded-lg">
            <p class="text-xs text-gray-400 mb-1">This Week</p>
            <p class="text-white text-sm">
              <span class="text-lg font-bold">{{ analyticsStore.insights.this_week.deployments }}</span> deploys ·
              {{ analyticsStore.insights.this_week.projects }} projects
            </p>
            <p class="text-[11px] mt-1">
              <span :class="analyticsStore.insights.this_week.red_days ? 'text-red-400' : 'text-gray-500'">{{ analyticsStore.insights.this_week.red_days }} red day(s)</span> ·
              <span :class="analyticsStore.insights.this_week.dirty_patches ? 'text-red-400' : 'text-gray-500'">{{ analyticsStore.insights.this_week.dirty_patches }} no-JIRA</span>
            </p>
          </div>
          <div class="p-3 bg-surface-deep rounded-lg">
            <p class="text-xs text-gray-400 mb-1">This Month</p>
            <p class="text-white text-sm">
              <span class="text-lg font-bold">{{ analyticsStore.insights.this_month.deployments }}</span> deploys ·
              {{ analyticsStore.insights.this_month.projects }} projects
            </p>
            <p class="text-[11px] mt-1">
              <span :class="analyticsStore.insights.this_month.red_days ? 'text-red-400' : 'text-gray-500'">{{ analyticsStore.insights.this_month.red_days }} red day(s)</span> ·
              <span :class="analyticsStore.insights.this_month.dirty_patches ? 'text-red-400' : 'text-gray-500'">{{ analyticsStore.insights.this_month.dirty_patches }} no-JIRA</span>
            </p>
          </div>
        </div>

        <!-- Frequency graph: deployments per week (last 10 weeks) -->
        <div class="p-3 bg-surface-deep rounded-lg">
          <p class="text-xs text-gray-400 mb-2">Deployment frequency (per week)</p>
          <div class="flex items-end gap-1 h-24">
            <div v-for="w in analyticsStore.insights.weekly_frequency" :key="w.label" class="flex-1 flex flex-col items-center justify-end h-full">
              <span class="text-[10px] text-gray-300 mb-0.5">{{ w.count }}</span>
              <div class="w-full rounded-t" :class="w.count > 3 ? 'bg-orange-500' : 'bg-accent'"
                :style="{ height: `${freqBarHeight(w.count)}%` }" :title="`${w.label}: ${w.count}`"></div>
              <span class="text-[9px] text-gray-500 mt-1">{{ w.label }}</span>
            </div>
          </div>
          <p class="text-[10px] text-gray-500 mt-1">Orange = more than 3 patches that week.</p>
        </div>

        <!-- Month-over-month trend (last 6 months) -->
        <div class="p-3 bg-surface-deep rounded-lg">
          <p class="text-xs text-gray-400 mb-2">Month-over-month trend</p>
          <div class="grid grid-cols-6 gap-2">
            <div v-for="m in analyticsStore.insights.monthly_trends" :key="m.label" class="text-center">
              <p class="text-sm font-bold text-white">{{ m.deployments }}</p>
              <p class="text-[10px]" :class="m.jira_compliance >= 80 ? 'text-green-400' : 'text-yellow-400'">{{ m.jira_compliance }}% JIRA</p>
              <p class="text-[10px]" :class="m.red_days ? 'text-red-400' : 'text-gray-500'">{{ m.red_days }} red</p>
              <p class="text-[9px] text-gray-500 mt-0.5">{{ m.label }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Project health rows (deterministic, worst first) -->
      <div v-if="analyticsStore.insights" class="space-y-2">
        <div
          v-for="p in analyticsStore.insights.projects"
          :key="p.project_id"
          class="p-2 bg-surface rounded"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2 min-w-0">
              <span class="text-xs px-2 py-0.5 rounded-full font-medium" :class="classClass(p.classification)">
                {{ classLabel(p.classification) }}
              </span>
              <span class="text-white font-medium truncate">{{ p.name }}</span>
              <span class="text-xs text-gray-500">{{ p.deployments }} deploys</span>
            </div>
            <div class="flex items-center gap-3 shrink-0">
              <div class="w-28 bg-gray-700 rounded-full h-2">
                <div class="h-2 rounded-full" :class="scoreBar(p.health_score)" :style="{ width: `${p.health_score}%` }"></div>
              </div>
              <span class="text-sm text-gray-300 w-14 text-right">{{ p.health_score }}/100</span>
            </div>
          </div>
          <div v-if="p.flags && p.flags.length" class="mt-1 flex flex-wrap gap-1">
            <span v-for="f in p.flags" :key="f" class="text-[10px] px-1.5 py-0.5 rounded bg-yellow-500/15 text-yellow-300">{{ f }}</span>
          </div>
        </div>
        <p v-if="analyticsStore.insights.projects.length === 0" class="text-gray-400 text-center py-3 text-sm">No deployments in this period.</p>

        <!-- Developer & environment breakdowns (period-scoped) -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-3 mt-3">
          <div class="p-3 bg-surface-deep rounded-lg">
            <p class="text-xs text-gray-400 mb-2">By developer</p>
            <div class="space-y-1.5 max-h-44 overflow-y-auto">
              <div v-for="dev in analyticsStore.insights.developers" :key="dev.name" class="flex items-center justify-between text-sm">
                <span class="text-white truncate">{{ dev.name }} <span class="text-gray-500 text-xs">({{ dev.projects }} proj)</span></span>
                <span class="text-xs shrink-0">
                  <span class="text-gray-300">{{ dev.deployments }} dep</span> ·
                  <span :class="dev.jira_compliance >= 80 ? 'text-green-400' : 'text-yellow-400'">{{ dev.jira_compliance }}% JIRA</span>
                  <span v-if="dev.failed" class="text-red-400"> · {{ dev.failed }} failed</span>
                </span>
              </div>
            </div>
          </div>
          <div class="p-3 bg-surface-deep rounded-lg">
            <p class="text-xs text-gray-400 mb-2">By environment</p>
            <div class="space-y-1.5">
              <div v-for="env in analyticsStore.insights.environments" :key="env.name" class="flex items-center justify-between text-sm">
                <span class="text-white">{{ env.name }}</span>
                <span class="text-xs">
                  <span class="text-gray-300">{{ env.deployments }} dep</span> ·
                  <span :class="env.success_rate >= 95 ? 'text-green-400' : 'text-yellow-400'">{{ env.success_rate }}% ok</span>
                  <span v-if="env.no_jira" class="text-red-400"> · {{ env.no_jira }} no-JIRA</span>
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Parson narrative -->
      <div v-if="analyticsStore.narrative" class="mt-4 pt-4 border-t border-surface-border">
        <div class="flex items-center gap-2 mb-2">
          <span class="text-sm font-medium text-blue-300">Parson's Analysis</span>
          <span v-if="analyticsStore.narrativeSeconds" class="text-xs text-gray-500">({{ analyticsStore.narrativeSeconds.toFixed(0) }}s)</span>
        </div>
        <pre class="whitespace-pre-wrap text-sm text-gray-200 font-sans bg-surface-deep p-3 rounded-lg">{{ analyticsStore.narrative }}</pre>
      </div>
      <p v-if="analyticsStore.error" class="text-red-400 text-sm mt-2">{{ analyticsStore.error }}</p>
    </Card>

    <!-- Overview Stats -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
      <Card class="!p-3">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-blue-500/20 rounded">
            <Activity class="w-5 h-5 text-blue-400" />
          </div>
          <div>
            <p class="text-gray-400 text-xs">Total Deployments</p>
            <p class="text-xl font-bold text-white">{{ totalDeployments }}</p>
          </div>
        </div>
      </Card>
      <Card class="!p-3">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-green-500/20 rounded">
            <CheckCircle class="w-5 h-5 text-green-400" />
          </div>
          <div>
            <p class="text-gray-400 text-xs">Success Rate</p>
            <p class="text-xl font-bold text-green-400">{{ successRate }}%</p>
          </div>
        </div>
      </Card>
      <Card class="!p-3">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-purple-500/20 rounded">
            <Users class="w-5 h-5 text-purple-400" />
          </div>
          <div>
            <p class="text-gray-400 text-xs">Active Developers</p>
            <p class="text-xl font-bold text-purple-400">{{ activeDevelopers }}</p>
          </div>
        </div>
      </Card>
      <Card class="!p-3">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-orange-500/20 rounded">
            <Folder class="w-5 h-5 text-orange-400" />
          </div>
          <div>
            <p class="text-gray-400 text-xs">Active Projects</p>
            <p class="text-xl font-bold text-orange-400">{{ activeProjects }}</p>
          </div>
        </div>
      </Card>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <!-- Top Deployers -->
      <Card title="Top Deployers">
        <div class="space-y-2">
          <div
            v-for="(dev, idx) in topDeployers"
            :key="dev.name"
            class="flex items-center justify-between p-2 bg-surface rounded"
          >
            <div class="flex items-center gap-2">
              <span class="w-6 h-6 rounded-full bg-accent flex items-center justify-center text-xs font-bold">
                {{ idx + 1 }}
              </span>
              <span class="text-white text-sm">{{ dev.name }}</span>
            </div>
            <div class="flex items-center gap-3">
              <div class="w-24 bg-gray-700 rounded-full h-2">
                <div
                  class="bg-accent h-2 rounded-full"
                  :style="{ width: `${(dev.count / maxDeployerCount) * 100}%` }"
                ></div>
              </div>
              <span class="text-gray-400 text-sm w-8 text-right">{{ dev.count }}</span>
            </div>
          </div>
          <p v-if="topDeployers.length === 0" class="text-gray-400 text-center py-2">No data</p>
        </div>
      </Card>

      <!-- Most Deployed Projects -->
      <Card title="Most Deployed Projects">
        <div class="space-y-2">
          <div
            v-for="(proj, idx) in topProjects"
            :key="proj.name"
            class="flex items-center justify-between p-2 bg-surface rounded"
          >
            <div class="flex items-center gap-2">
              <span class="w-6 h-6 rounded-full bg-green-500 flex items-center justify-center text-xs font-bold">
                {{ idx + 1 }}
              </span>
              <span class="text-white text-sm">{{ proj.name }}</span>
            </div>
            <div class="flex items-center gap-3">
              <div class="w-24 bg-gray-700 rounded-full h-2">
                <div
                  class="bg-green-500 h-2 rounded-full"
                  :style="{ width: `${(proj.count / maxProjectCount) * 100}%` }"
                ></div>
              </div>
              <span class="text-gray-400 text-sm w-8 text-right">{{ proj.count }}</span>
            </div>
          </div>
          <p v-if="topProjects.length === 0" class="text-gray-400 text-center py-2">No data</p>
        </div>
      </Card>

      <!-- Deployments by Environment -->
      <Card title="Deployments by Environment">
        <div class="space-y-2">
          <div
            v-for="env in environmentStats"
            :key="env.name"
            class="flex items-center justify-between p-2 bg-surface rounded"
          >
            <div class="flex items-center gap-2">
              <Server class="w-4 h-4 text-gray-400" />
              <span class="text-white text-sm">{{ env.name }}</span>
            </div>
            <div class="flex items-center gap-3">
              <div class="w-24 bg-gray-700 rounded-full h-2">
                <div
                  class="bg-purple-500 h-2 rounded-full"
                  :style="{ width: `${(env.count / maxEnvCount) * 100}%` }"
                ></div>
              </div>
              <span class="text-gray-400 text-sm w-8 text-right">{{ env.count }}</span>
            </div>
          </div>
          <p v-if="environmentStats.length === 0" class="text-gray-400 text-center py-2">No environment data</p>
        </div>
      </Card>

      <!-- JIRA Compliance -->
      <Card title="JIRA ID Compliance">
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-gray-400 text-sm">With JIRA ID</span>
            <span class="text-green-400 font-bold">{{ jiraCompliance.with }} ({{ jiraComplianceRate }}%)</span>
          </div>
          <div class="w-full bg-gray-700 rounded-full h-3">
            <div
              class="bg-green-500 h-3 rounded-full"
              :style="{ width: `${jiraComplianceRate}%` }"
            ></div>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-gray-400 text-sm">Without JIRA ID</span>
            <span class="text-red-400 font-bold">{{ jiraCompliance.without }}</span>
          </div>
          <div class="mt-3 p-2 bg-surface rounded text-xs text-gray-400">
            Deployments with valid JIRA IDs help maintain traceability and audit compliance.
          </div>
        </div>
      </Card>

      <!-- Peak Deployment Times -->
      <Card title="Peak Deployment Hours">
        <div class="space-y-2">
          <div
            v-for="hour in peakHours"
            :key="hour.hour"
            class="flex items-center justify-between p-2 bg-surface rounded"
          >
            <div class="flex items-center gap-2">
              <Clock class="w-4 h-4 text-gray-400" />
              <span class="text-white text-sm">{{ hour.label }}</span>
            </div>
            <div class="flex items-center gap-3">
              <div class="w-24 bg-gray-700 rounded-full h-2">
                <div
                  class="bg-orange-500 h-2 rounded-full"
                  :style="{ width: `${(hour.count / maxHourCount) * 100}%` }"
                ></div>
              </div>
              <span class="text-gray-400 text-sm w-8 text-right">{{ hour.count }}</span>
            </div>
          </div>
        </div>
      </Card>

      <!-- Daily Distribution -->
      <Card title="Deployments by Day of Week">
        <div class="space-y-2">
          <div
            v-for="day in dailyStats"
            :key="day.name"
            class="flex items-center justify-between p-2 bg-surface rounded"
          >
            <span class="text-white text-sm w-20">{{ day.name }}</span>
            <div class="flex items-center gap-3 flex-1 ml-2">
              <div class="flex-1 bg-gray-700 rounded-full h-2">
                <div
                  class="bg-cyan-500 h-2 rounded-full"
                  :style="{ width: `${(day.count / maxDayCount) * 100}%` }"
                ></div>
              </div>
              <span class="text-gray-400 text-sm w-8 text-right">{{ day.count }}</span>
            </div>
          </div>
        </div>
      </Card>
    </div>

    <!-- Failed Deployments Analysis -->
    <Card title="Failed Deployments Analysis">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="p-3 bg-surface rounded">
          <div class="flex items-center gap-2 mb-2">
            <XCircle class="w-4 h-4 text-red-400" />
            <span class="text-gray-400 text-sm">Build Failures</span>
          </div>
          <p class="text-2xl font-bold text-red-400">{{ failedStats.buildFailed }}</p>
          <p class="text-xs text-gray-500">{{ failedStats.buildFailedRate }}% of all deployments</p>
        </div>
        <div class="p-3 bg-surface rounded">
          <div class="flex items-center gap-2 mb-2">
            <AlertTriangle class="w-4 h-4 text-orange-400" />
            <span class="text-gray-400 text-sm">Deploy Failures</span>
          </div>
          <p class="text-2xl font-bold text-orange-400">{{ failedStats.deployFailed }}</p>
          <p class="text-xs text-gray-500">{{ failedStats.deployFailedRate }}% of all deployments</p>
        </div>
        <div class="p-3 bg-surface rounded">
          <div class="flex items-center gap-2 mb-2">
            <TrendingUp class="w-4 h-4 text-green-400" />
            <span class="text-gray-400 text-sm">Overall Success</span>
          </div>
          <p class="text-2xl font-bold text-green-400">{{ failedStats.successCount }}</p>
          <p class="text-xs text-gray-500">{{ successRate }}% success rate</p>
        </div>
      </div>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import Card from '../components/ui/Card.vue'
import Select from '../components/ui/Select.vue'
import Button from '../components/ui/Button.vue'
import { useDeploymentsStore } from '../stores/deployments'
import { useProjectsStore } from '../stores/projects'
import { useAnalyticsStore } from '../stores/analytics'
import { parseTimestamp } from '../lib/utils'
import { Sparkles, RefreshCw } from 'lucide-vue-next'
import {
  Activity,
  CheckCircle,
  Users,
  Folder,
  Server,
  Clock,
  XCircle,
  AlertTriangle,
  TrendingUp
} from 'lucide-vue-next'

const deploymentsStore = useDeploymentsStore()
const projectsStore = useProjectsStore()
const analyticsStore = useAnalyticsStore()

const selectedMonth = ref('All')
const selectedYear = ref(new Date().getFullYear().toString())

const monthOptions = ['All', 'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December']

// Period as numbers for the analytics API (month 0 = all).
const monthNum = computed(() => selectedMonth.value === 'All' ? 0 : monthOptions.indexOf(selectedMonth.value))
const yearNum = computed(() => parseInt(selectedYear.value) || 0)

const runAnalysis = async () => {
  await analyticsStore.fetchNarrative(monthNum.value, yearNum.value)
}

const classLabel = (c: string) => c === 'good' ? 'Healthy' : c === 'attention' ? 'Watch' : 'Dirty'
const classClass = (c: string) =>
  c === 'good' ? 'bg-green-500/20 text-green-300'
  : c === 'attention' ? 'bg-yellow-500/20 text-yellow-300'
  : 'bg-red-500/20 text-red-300'
const scoreBar = (s: number) => s >= 80 ? 'bg-green-500' : s >= 55 ? 'bg-yellow-500' : 'bg-red-500'

const freqBarHeight = (count: number) => {
  const max = Math.max(1, ...(analyticsStore.insights?.weekly_frequency?.map(w => w.count) || [1]))
  return Math.max(count > 0 ? 6 : 0, (count / max) * 100)
}

// Refresh the deterministic health whenever the period changes.
watch([monthNum, yearNum], () => analyticsStore.fetchInsights(monthNum.value, yearNum.value))

const yearOptions = computed(() => {
  const currentYear = new Date().getFullYear()
  return Array.from({ length: 5 }, (_, i) => (currentYear - i).toString())
})

const filteredDeployments = computed(() => {
  let data = [...deploymentsStore.deployments]

  if (selectedMonth.value !== 'All') {
    const monthIndex = monthOptions.indexOf(selectedMonth.value) - 1
    data = data.filter(d => {
      const date = parseTimestamp(d.timestamp)
      return date !== null && date.getMonth() === monthIndex
    })
  }

  data = data.filter(d => {
    const date = parseTimestamp(d.timestamp)
    return date !== null && date.getFullYear() === parseInt(selectedYear.value)
  })

  return data
})

const totalDeployments = computed(() => filteredDeployments.value.length)

const successRate = computed(() => {
  if (totalDeployments.value === 0) return 0
  const success = filteredDeployments.value.filter(d => d.deploy_status === 'success').length
  return Math.round((success / totalDeployments.value) * 100)
})

const activeDevelopers = computed(() => {
  const devs = new Set(filteredDeployments.value.map(d => d.developer_name || d.deployed_by).filter(Boolean))
  return devs.size
})

const activeProjects = computed(() => {
  const projs = new Set(filteredDeployments.value.map(d => d.project_id))
  return projs.size
})

// Use centralized helper from store
const getProjectName = projectsStore.getProjectName

// Top Deployers
const topDeployers = computed(() => {
  const counts: Record<string, number> = {}
  filteredDeployments.value.forEach(d => {
    const name = d.deployed_by || d.developer_name || 'Unknown'
    counts[name] = (counts[name] || 0) + 1
  })
  return Object.entries(counts)
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 5)
})

const maxDeployerCount = computed(() => Math.max(...topDeployers.value.map(d => d.count), 1))

// Top Projects
const topProjects = computed(() => {
  const counts: Record<number, number> = {}
  filteredDeployments.value.forEach(d => {
    counts[d.project_id] = (counts[d.project_id] || 0) + 1
  })
  return Object.entries(counts)
    .map(([id, count]) => ({ name: getProjectName(parseInt(id)), count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 5)
})

const maxProjectCount = computed(() => Math.max(...topProjects.value.map(p => p.count), 1))

// Environment Stats
const environmentStats = computed(() => {
  const counts: Record<string, number> = {}
  filteredDeployments.value.forEach(d => {
    const env = d.environment || 'Unknown'
    counts[env] = (counts[env] || 0) + 1
  })
  return Object.entries(counts)
    .filter(([name]) => name !== 'Unknown')
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count)
})

const maxEnvCount = computed(() => Math.max(...environmentStats.value.map(e => e.count), 1))

// JIRA Compliance
const jiraCompliance = computed(() => {
  const withJira = filteredDeployments.value.filter(d => d.jira_id && d.jira_id.trim() !== '').length
  return {
    with: withJira,
    without: totalDeployments.value - withJira
  }
})

const jiraComplianceRate = computed(() => {
  if (totalDeployments.value === 0) return 0
  return Math.round((jiraCompliance.value.with / totalDeployments.value) * 100)
})

// Peak Hours
const peakHours = computed(() => {
  const counts: Record<number, number> = {}
  filteredDeployments.value.forEach(d => {
    const date = parseTimestamp(d.timestamp)
    if (!date) return
    const hour = date.getHours()
    counts[hour] = (counts[hour] || 0) + 1
  })
  return Object.entries(counts)
    .map(([hour, count]) => ({
      hour: parseInt(hour),
      label: `${parseInt(hour).toString().padStart(2, '0')}:00 - ${(parseInt(hour) + 1).toString().padStart(2, '0')}:00`,
      count
    }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 5)
})

const maxHourCount = computed(() => Math.max(...peakHours.value.map(h => h.count), 1))

// Daily Stats
const dailyStats = computed(() => {
  const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
  const counts: Record<number, number> = { 0: 0, 1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0 }
  filteredDeployments.value.forEach(d => {
    const date = parseTimestamp(d.timestamp)
    if (!date) return
    const day = date.getDay()
    counts[day] = (counts[day] || 0) + 1
  })
  return days.map((name, idx) => ({ name, count: counts[idx] }))
})

const maxDayCount = computed(() => Math.max(...dailyStats.value.map(d => d.count), 1))

// Failed Stats
const failedStats = computed(() => {
  const buildFailed = filteredDeployments.value.filter(d => d.build_status === 'failed').length
  const deployFailed = filteredDeployments.value.filter(d => d.deploy_status === 'failed').length
  const successCount = filteredDeployments.value.filter(d => d.deploy_status === 'success').length

  return {
    buildFailed,
    deployFailed,
    successCount,
    buildFailedRate: totalDeployments.value ? Math.round((buildFailed / totalDeployments.value) * 100) : 0,
    deployFailedRate: totalDeployments.value ? Math.round((deployFailed / totalDeployments.value) * 100) : 0
  }
})

onMounted(async () => {
  if (projectsStore.projects.length === 0) {
    await projectsStore.fetchProjects()
  }
  if (deploymentsStore.deployments.length === 0) {
    await deploymentsStore.fetchDeployments()
  }
  analyticsStore.fetchInsights(monthNum.value, yearNum.value)
})
</script>
