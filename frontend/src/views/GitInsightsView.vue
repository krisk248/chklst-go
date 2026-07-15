<template>
  <div class="space-y-4">
    <div class="page-header">
      <div>
        <h1 class="page-title text-xl">Git Insights</h1>
        <p class="text-gray-400 text-sm">Commit activity & CI/CD cadence from your GitHub repos</p>
      </div>
      <div class="flex items-center gap-2">
        <select v-model.number="days" @change="reload"
          class="px-3 py-2 text-sm bg-surface-light border border-surface-border rounded-lg text-white">
          <option :value="7">Last 7 days</option>
          <option :value="30">Last 30 days</option>
          <option :value="90">Last 90 days</option>
        </select>
      </div>
    </div>

    <p v-if="store.error" class="text-red-400 text-sm">
      {{ store.error }} — configure GitHub in <RouterLink to="/settings" class="text-accent underline">Settings</RouterLink>.
    </p>

    <Skeleton v-if="store.isLoading" :rows="5" />

    <template v-else-if="store.insights">
      <!-- headline numbers -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div class="p-4 bg-surface-deeper border border-surface-border rounded-lg">
          <p class="text-gray-400 text-xs">Commits</p>
          <p class="text-2xl font-bold text-white">{{ store.insights.total_commits }}</p>
          <p class="text-[11px] text-gray-500">last {{ store.insights.period_days }} days</p>
        </div>
        <div class="p-4 bg-surface-deeper border border-surface-border rounded-lg">
          <p class="text-gray-400 text-xs">Merges (CI/CD)</p>
          <p class="text-2xl font-bold text-green-400">{{ store.insights.total_merges }}</p>
        </div>
        <div class="p-4 bg-surface-deeper border border-surface-border rounded-lg">
          <p class="text-gray-400 text-xs">Active repos</p>
          <p class="text-2xl font-bold text-purple-400">{{ store.insights.by_repo.length }}</p>
        </div>
        <div class="p-4 bg-surface-deeper border border-surface-border rounded-lg">
          <p class="text-gray-400 text-xs">Contributors</p>
          <p class="text-2xl font-bold text-blue-400">{{ store.insights.authors.length }}</p>
        </div>
      </div>

      <!-- daily commit graph -->
      <Card title="Commits per day">
        <div class="flex items-end gap-0.5 h-28">
          <div v-for="d in store.insights.by_day" :key="d.date" class="flex-1 flex flex-col items-center justify-end h-full">
            <div class="w-full rounded-t bg-accent" :style="{ height: `${dayBar(d.count)}%` }" :title="`${d.label}: ${d.count}`"></div>
          </div>
        </div>
        <p class="text-[10px] text-gray-500 mt-1">{{ store.insights.by_day[0]?.label }} → {{ store.insights.by_day.at(-1)?.label }}</p>
      </Card>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <!-- contributors -->
        <Card title="Contributors">
          <div class="space-y-2">
            <div v-for="a in store.insights.authors" :key="a.name" class="flex items-center justify-between p-2 bg-surface rounded text-sm">
              <span class="text-white">{{ a.name }} <span class="text-gray-500 text-xs">· {{ a.repos }} repo(s)</span></span>
              <span class="text-xs">
                <span class="text-gray-300">{{ a.commits }} commits</span>
                <span v-if="a.merges" class="text-green-400"> · {{ a.merges }} merges</span>
                <span class="text-gray-500"> · last {{ a.last_active }}</span>
              </span>
            </div>
            <p v-if="store.insights.authors.length === 0" class="text-gray-400 text-sm text-center py-2">No commits in this period.</p>
          </div>
        </Card>

        <!-- by repo + cadence -->
        <Card title="By repository">
          <div class="space-y-2 mb-4">
            <div v-for="r in store.insights.by_repo.slice(0, 8)" :key="r.name" class="flex items-center justify-between p-2 bg-surface rounded text-sm">
              <span class="text-white">{{ r.name }}</span>
              <span class="text-xs text-gray-300">{{ r.commits }} commits<span v-if="r.merges" class="text-green-400"> · {{ r.merges }} merges</span></span>
            </div>
          </div>
          <p class="text-xs text-gray-400 mb-1">Commit cadence (by hour)</p>
          <div class="flex items-end gap-0.5 h-12">
            <div v-for="(n, h) in store.insights.by_hour" :key="h" class="flex-1 flex flex-col items-center justify-end h-full">
              <div class="w-full rounded-t" :class="h >= 18 || h < 6 ? 'bg-purple-500' : 'bg-accent'" :style="{ height: `${hourBar(n)}%` }" :title="`${h}:00 — ${n}`"></div>
            </div>
          </div>
          <p class="text-[10px] text-gray-500 mt-1">Purple = nights (6 PM–6 AM)</p>
        </Card>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import Card from '../components/ui/Card.vue'
import Skeleton from '../components/ui/Skeleton.vue'
import { useGitInsightsStore } from '../stores/gitInsights'

const store = useGitInsightsStore()
const days = ref(30)

const dayBar = (count: number) => {
  const max = Math.max(1, ...(store.insights?.by_day.map(d => d.count) || [1]))
  return Math.max(count > 0 ? 6 : 0, (count / max) * 100)
}
const hourBar = (count: number) => {
  const max = Math.max(1, ...(store.insights?.by_hour || [1]))
  return Math.max(count > 0 ? 8 : 0, (count / max) * 100)
}

const reload = () => store.fetchInsights(days.value)
onMounted(reload)
</script>
