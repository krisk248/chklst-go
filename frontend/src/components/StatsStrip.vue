<template>
  <div v-if="loaded" class="flex flex-wrap items-center gap-2 mb-3">
    <span class="strip-chip">
      <Activity class="w-3.5 h-3.5 text-accent" />
      <strong class="text-white">{{ week?.deployments ?? 0 }}</strong>&nbsp;this week
    </span>
    <span class="strip-chip">
      <CalendarDays class="w-3.5 h-3.5 text-purple-400" />
      <strong class="text-white">{{ month?.deployments ?? 0 }}</strong>&nbsp;this month
    </span>
    <span v-if="redFlags > 0" class="strip-chip !border-red-500/40 text-red-300">
      <AlertTriangle class="w-3.5 h-3.5" />
      <strong>{{ redFlags }}</strong>&nbsp;red flag(s) this week
    </span>
    <span v-else class="strip-chip text-green-300">
      <CheckCircle class="w-3.5 h-3.5" />
      no red flags this week
    </span>
    <RouterLink to="/daily-summary" class="strip-chip hover:border-accent transition-colors">
      <Sparkles class="w-3.5 h-3.5 text-blue-300" />
      Parson: {{ parsonLabel }}
    </RouterLink>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useApi } from '../composables/useApi'
import { Activity, CalendarDays, AlertTriangle, CheckCircle, Sparkles } from 'lucide-vue-next'

interface Counts { deployments: number; projects: number; dirty_patches: number; red_days: number }

const { get } = useApi()
const loaded = ref(false)
const week = ref<Counts | null>(null)
const month = ref<Counts | null>(null)
const parsonStatus = ref('')

const redFlags = computed(() => (week.value?.red_days ?? 0) + (week.value?.dirty_patches ?? 0))

const parsonLabel = computed(() => {
  switch (parsonStatus.value) {
    case 'sent': return 'today’s report sent ✓'
    case 'generated': return 'report ready — sends tonight'
    case 'failed': return 'send failed — check Daily Summary'
    default: return 'report not generated yet'
  }
})

onMounted(async () => {
  try {
    const [insights, today] = await Promise.all([
      get<{ this_week: Counts; this_month: Counts }>('/analytics/insights'),
      get<{ summary: { status: string } }>('/summary/today'),
    ])
    week.value = insights.data.this_week
    month.value = insights.data.this_month
    parsonStatus.value = today.data.summary?.status || ''
    loaded.value = true
  } catch {
    // Strip is purely informational — stay hidden on failure.
  }
})
</script>

<style scoped>
.strip-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.3rem 0.65rem;
  font-size: 0.75rem;
  color: #d1d5db;
  background: #2f2f2f;
  border: 1px solid #4a4a4a;
  border-radius: 9999px;
}
</style>
