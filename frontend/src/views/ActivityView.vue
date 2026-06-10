<template>
  <div class="space-y-4">
    <div class="page-header">
      <div>
        <h1 class="page-title text-xl">Parson Activity</h1>
        <p class="text-gray-400 text-sm">What Parson did each day — generation time, review, lock, and send status</p>
      </div>
      <Button variant="secondary" size="sm" @click="reload">
        <RefreshCw class="w-4 h-4" /> Refresh
      </Button>
    </div>

    <Card>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="text-left text-gray-400 border-b border-surface-border">
              <th class="py-2 pr-4">Date</th>
              <th class="py-2 pr-4">Status</th>
              <th class="py-2 pr-4">Source</th>
              <th class="py-2 pr-4">Generated</th>
              <th class="py-2 pr-4">Time</th>
              <th class="py-2 pr-4">Model</th>
              <th class="py-2 pr-4">Reviewed</th>
              <th class="py-2 pr-4">Locked</th>
              <th class="py-2 pr-4">Sent</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in store.recent" :key="row.id" class="border-b border-[#3a3a3a]">
              <td class="py-2 pr-4 text-white font-medium">{{ row.date }}</td>
              <td class="py-2 pr-4">
                <span class="text-xs px-2 py-0.5 rounded-full" :class="statusClass(row.status)">{{ row.status }}</span>
              </td>
              <td class="py-2 pr-4 text-gray-300">{{ row.auto_generated ? 'Scheduled' : (row.generated_at ? 'Manual' : '—') }}</td>
              <td class="py-2 pr-4 text-gray-300">{{ fmt(row.generated_at) }}</td>
              <td class="py-2 pr-4 text-gray-300">{{ row.generation_seconds ? row.generation_seconds.toFixed(0) + 's' : '—' }}</td>
              <td class="py-2 pr-4 text-gray-400">{{ row.generated_by_model || '—' }}</td>
              <td class="py-2 pr-4">
                <span :class="row.reviewed ? 'text-green-400' : 'text-gray-500'">{{ row.reviewed ? 'Yes' : 'No' }}</span>
              </td>
              <td class="py-2 pr-4 text-gray-300">{{ fmt(row.locked_at) }}</td>
              <td class="py-2 pr-4 text-gray-300">
                <span v-if="row.status === 'failed'" class="text-red-400" :title="row.send_error">failed</span>
                <span v-else>{{ fmt(row.sent_at) }}</span>
              </td>
            </tr>
            <tr v-if="store.recent.length === 0">
              <td colspan="9" class="py-6 text-center text-gray-500">No activity yet.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import Card from '../components/ui/Card.vue'
import Button from '../components/ui/Button.vue'
import { useSummaryStore } from '../stores/summary'
import { RefreshCw } from 'lucide-vue-next'

const store = useSummaryStore()

const fmt = (iso?: string | null) => {
  if (!iso) return '—'
  const d = new Date(iso)
  return isNaN(d.getTime()) ? '—' : d.toLocaleString()
}

const statusClass = (status: string) => {
  switch (status) {
    case 'generated': return 'bg-blue-500/20 text-blue-300'
    case 'sent': return 'bg-green-600/30 text-green-300'
    case 'failed': return 'bg-red-500/20 text-red-300'
    default: return 'bg-gray-500/20 text-gray-300'
  }
}

const reload = () => store.fetchRecent()
onMounted(reload)
</script>
