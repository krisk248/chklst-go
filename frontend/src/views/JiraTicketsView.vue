<template>
  <div class="space-y-6">
    <div class="page-header">
      <div>
        <h1 class="page-title">Jira Tickets</h1>
        <p class="text-gray-400 mt-2">View and deploy from your assigned Jira tickets</p>
      </div>
      <div class="flex items-center gap-3">
        <Button variant="secondary" @click="fetchTickets" :disabled="loading">
          <RefreshCw :class="['w-4 h-4', { 'animate-spin': loading }]" />
          Refresh
        </Button>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex items-center gap-6">
      <label class="flex items-center gap-2 cursor-pointer">
        <input
          v-model="todayOnly"
          type="checkbox"
          class="w-4 h-4 rounded"
          @change="fetchTickets"
        />
        <span class="text-sm text-gray-300">Today Only</span>
      </label>
      <div class="flex items-center gap-2">
        <label class="text-sm font-medium text-gray-300">Status:</label>
        <select
          v-model="statusFilter"
          @change="fetchTickets"
          class="bg-surface-light border border-surface-border rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:ring-2 focus:ring-accent"
        >
          <option value="">All</option>
          <option value="Deployed for Test">Deployed for Test</option>
          <option value="Approved for Deployment">Approved for Deployment</option>
          <option value="Done">Done</option>
        </select>
      </div>
    </div>

    <!-- Not Configured Warning -->
    <Card v-if="!jiraConfigured" title="Jira Not Configured">
      <div class="text-center py-8">
        <AlertCircle class="w-12 h-12 text-yellow-500 mx-auto mb-4" />
        <p class="text-gray-400 mb-4">
          Jira integration is not configured. Please configure it in Settings.
        </p>
        <router-link to="/settings">
          <Button variant="primary">
            <Settings class="w-4 h-4" />
            Go to Settings
          </Button>
        </router-link>
      </div>
    </Card>

    <!-- Loading State -->
    <div v-else-if="loading" class="text-center py-12">
      <RefreshCw class="w-8 h-8 animate-spin text-accent mx-auto mb-4" />
      <p class="text-gray-400">Loading tickets...</p>
    </div>

    <!-- Error State -->
    <Card v-else-if="errorMsg" title="Error">
      <div class="text-center py-8">
        <AlertCircle class="w-12 h-12 text-red-500 mx-auto mb-4" />
        <p class="text-red-400 mb-4">{{ errorMsg }}</p>
        <Button variant="secondary" @click="fetchTickets">
          <RefreshCw class="w-4 h-4" />
          Try Again
        </Button>
      </div>
    </Card>

    <!-- Empty State -->
    <Card v-else-if="tickets.length === 0" title="No Tickets">
      <div class="text-center py-8">
        <Inbox class="w-12 h-12 text-gray-500 mx-auto mb-4" />
        <p class="text-gray-400">No tickets assigned to you in project {{ jiraProject }}</p>
      </div>
    </Card>

    <!-- Tickets Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="ticket in tickets"
        :key="ticket.key"
        class="bg-surface-deeper border border-surface-border rounded-lg p-4 hover:border-accent transition-colors cursor-pointer"
        @click="selectTicket(ticket)"
      >
        <div class="flex items-start justify-between mb-3">
          <span class="text-accent font-mono font-bold">{{ ticket.key }}</span>
          <span
            :class="[
              'text-xs px-2 py-1 rounded-full',
              getStatusColor(ticket.status),
            ]"
          >
            {{ ticket.status }}
          </span>
        </div>
        <h3 class="text-white font-medium mb-2 line-clamp-2">{{ ticket.summary }}</h3>
        <div class="text-sm text-gray-400 space-y-1">
          <p v-if="ticket.environment">
            <span class="text-gray-500">Environment:</span> {{ ticket.environment }}
          </p>
          <p v-if="ticket.developers">
            <span class="text-gray-500">Developer:</span> {{ ticket.developers }}
          </p>
          <p v-if="ticket.priority">
            <span class="text-gray-500">Priority:</span> {{ ticket.priority }}
          </p>
        </div>
        <div class="mt-3 pt-3 border-t border-surface-light flex justify-between items-center">
          <span class="text-xs text-gray-500">{{ formatDate(ticket.updated) }}</span>
          <Button size="sm" variant="primary" @click.stop="deployFromTicket(ticket)">
            <Rocket class="w-3 h-3" />
            Deploy
          </Button>
        </div>
      </div>
    </div>

    <!-- Selected Ticket Modal -->
    <div
      v-if="selectedTicket"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      @click.self="selectedTicket = null"
    >
      <div class="bg-surface-deeper border border-surface-border rounded-lg max-w-2xl w-full mx-4 max-h-[80vh] overflow-y-auto">
        <div class="p-6">
          <div class="flex items-start justify-between mb-4">
            <div>
              <span class="text-accent font-mono font-bold text-lg">{{ selectedTicket.key }}</span>
              <h2 class="text-white text-xl font-bold mt-1">{{ selectedTicket.summary }}</h2>
            </div>
            <button @click="selectedTicket = null" class="text-gray-400 hover:text-white">
              <X class="w-6 h-6" />
            </button>
          </div>

          <div class="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label class="text-sm text-gray-500">Status</label>
              <p class="text-white">{{ selectedTicket.status }}</p>
            </div>
            <div>
              <label class="text-sm text-gray-500">Priority</label>
              <p class="text-white">{{ selectedTicket.priority }}</p>
            </div>
            <div>
              <label class="text-sm text-gray-500">Environment</label>
              <p class="text-white">{{ selectedTicket.environment || '-' }}</p>
            </div>
            <div>
              <label class="text-sm text-gray-500">Developer(s)</label>
              <p class="text-white">{{ selectedTicket.developers || '-' }}</p>
            </div>
            <div>
              <label class="text-sm text-gray-500">Build Server</label>
              <p class="text-white">{{ selectedTicket.build_server || '-' }}</p>
            </div>
            <div>
              <label class="text-sm text-gray-500">Deploy Server</label>
              <p class="text-white">{{ selectedTicket.deploy_server || '-' }}</p>
            </div>
            <div>
              <label class="text-sm text-gray-500">Database</label>
              <p class="text-white">{{ selectedTicket.database || '-' }}</p>
            </div>
            <div>
              <label class="text-sm text-gray-500">DB Backup</label>
              <p class="text-white">{{ selectedTicket.db_backup || '-' }}</p>
            </div>
          </div>

          <div v-if="selectedTicket.description" class="mb-4">
            <label class="text-sm text-gray-500">Description</label>
            <p class="text-white whitespace-pre-wrap mt-1">{{ selectedTicket.description }}</p>
          </div>

          <div v-if="selectedTicket.notes" class="mb-4">
            <label class="text-sm text-gray-500">Notes</label>
            <p class="text-white whitespace-pre-wrap mt-1">{{ selectedTicket.notes }}</p>
          </div>

          <div class="flex justify-end gap-3 pt-4 border-t border-surface-light">
            <Button variant="secondary" @click="selectedTicket = null">
              Close
            </Button>
            <Button variant="primary" @click="deployFromTicket(selectedTicket)">
              <Rocket class="w-4 h-4" />
              Start Deployment
            </Button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Card from '../components/ui/Card.vue'
import Button from '../components/ui/Button.vue'
import { useSettingsStore } from '../stores/settings'
import { useApi } from '../composables/useApi'
import { useToast } from '../composables/useToast'
import {
  RefreshCw,
  AlertCircle,
  Settings,
  Inbox,
  Rocket,
  X,
} from 'lucide-vue-next'

interface JiraTicket {
  key: string
  summary: string
  status: string
  priority: string
  assignee: string
  reporter: string
  created: string
  updated: string
  description: string
  build_server: string
  deploy_server: string
  database: string
  db_backup: string
  environment: string
  notes: string
  developers: string
}

const router = useRouter()
const settingsStore = useSettingsStore()
const { get } = useApi()
const { error } = useToast()

const loading = ref(false)
const errorMsg = ref('')
const tickets = ref<JiraTicket[]>([])
const selectedTicket = ref<JiraTicket | null>(null)
const statusFilter = ref('')
const todayOnly = ref(true) // Default to today only

const jiraConfigured = computed(() => {
  const s = settingsStore.settings
  return s.jira_url && s.jira_email && s.jira_token
})

const jiraProject = computed(() => settingsStore.settings.jira_project || 'PAT')

const fetchTickets = async () => {
  if (!jiraConfigured.value) return

  loading.value = true
  errorMsg.value = ''

  try {
    const params = new URLSearchParams()
    if (statusFilter.value) params.set('status', statusFilter.value)
    if (todayOnly.value) params.set('today', 'true')
    const queryString = params.toString() ? `?${params.toString()}` : ''

    const response = await get<{ tickets: JiraTicket[]; total: number }>(`/jira/tickets${queryString}`)
    tickets.value = response.data?.tickets || []
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : 'Failed to fetch tickets'
    errorMsg.value = message
    error(message)
  } finally {
    loading.value = false
  }
}

const selectTicket = (ticket: JiraTicket) => {
  selectedTicket.value = ticket
}

const deployFromTicket = (ticket: JiraTicket) => {
  // Navigate to Quick Deploy with pre-filled data
  router.push({
    name: 'quick-deploy',
    query: {
      jira_id: ticket.key,
      environment: ticket.environment || '',
      developer: ticket.developers || '',
      build_server: ticket.build_server || '',
      deploy_server: ticket.deploy_server || '',
      database: ticket.database || '',
      db_backup: ticket.db_backup || '',
      notes: ticket.notes || '',
    },
  })
}

const getStatusColor = (status: string) => {
  switch (status?.toLowerCase()) {
    case 'done':
      return 'bg-green-500/20 text-green-400'
    case 'in progress':
      return 'bg-blue-500/20 text-blue-400'
    case 'ready for deploy':
      return 'bg-yellow-500/20 text-yellow-400'
    case 'to do':
      return 'bg-gray-500/20 text-gray-400'
    default:
      return 'bg-gray-500/20 text-gray-400'
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

onMounted(async () => {
  await settingsStore.fetchSettings()
  if (jiraConfigured.value) {
    await fetchTickets()
  }
})
</script>
