<template>
  <div class="space-y-6">
    <div class="page-header">
      <div>
        <h1 class="page-title">Settings</h1>
        <p class="text-gray-400 mt-2">Manage application settings and preferences</p>
      </div>
    </div>

    <form @submit.prevent="handleSaveSettings">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Settings Navigation -->
        <div class="lg:col-span-1">
          <Card title="Settings">
            <nav class="space-y-2">
              <button
                v-for="item in settingsSections"
                :key="item.id"
                type="button"
                @click="activeSection = item.id"
                :class="[
                  'w-full text-left px-4 py-3 rounded-lg transition-all',
                  activeSection === item.id
                    ? 'bg-accent text-white'
                    : 'bg-surface-light text-gray-300 hover:bg-surface-border',
                ]"
              >
                {{ item.label }}
              </button>
            </nav>
          </Card>
        </div>

        <!-- Settings Content -->
        <div class="lg:col-span-2">
          <!-- User Defaults Section -->
          <Card v-if="activeSection === 'user'" title="User Defaults">
            <div class="space-y-4">
              <Select
                v-model="settingsForm.default_deployed_by"
                :options="developerOptions"
                label="Default Deployed By"
                required
              />
              <p class="text-sm text-gray-400 mt-4">
                This user will be pre-selected in the deployment form.
              </p>
            </div>
          </Card>

          <!-- Export Settings Section -->
          <Card v-if="activeSection === 'export'" title="Export Settings">
            <div class="space-y-4">
              <Input
                v-model="settingsForm.excel_export_path"
                label="Excel Export Path"
                placeholder="/reports/exports"
                required
              />
              <p class="text-sm text-gray-400 mt-4">
                Specify the default directory for exported Excel files.
              </p>
            </div>
          </Card>

          <!-- Jira Integration Section -->
          <Card v-if="activeSection === 'jira'" title="Jira Integration">
            <div class="space-y-4">
              <Input
                v-model="settingsForm.jira_url"
                label="Jira URL"
                placeholder="https://yourcompany.atlassian.net"
              />
              <Input
                v-model="settingsForm.jira_email"
                label="Jira Email"
                type="email"
                placeholder="user@company.com"
              />
              <Input
                v-model="settingsForm.jira_token"
                label="API Token"
                type="password"
                placeholder="Your Jira API token"
              />
              <Input
                v-model="settingsForm.jira_project"
                label="Project Key"
                placeholder="PAT"
              />
              <div class="pt-4">
                <Button
                  type="button"
                  variant="secondary"
                  @click="testJiraConnection"
                  :disabled="testingJira"
                >
                  <RefreshCw v-if="testingJira" class="w-4 h-4 animate-spin" />
                  <Zap v-else class="w-4 h-4" />
                  {{ testingJira ? 'Testing...' : 'Test Connection' }}
                </Button>
              </div>
            </div>
          </Card>

          <!-- Webhooks Section -->
          <Card v-if="activeSection === 'webhooks'" title="Webhook Notifications">
            <div class="space-y-4">
              <div class="space-y-2">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    v-model="settingsForm.webhooks_enabled"
                    type="checkbox"
                    class="w-4 h-4"
                  />
                  <span class="text-white">Enable Webhook Notifications</span>
                </label>
              </div>
              <Input
                v-model="settingsForm.teams_webhook_url"
                label="Microsoft Teams Webhook URL"
                placeholder="https://outlook.office.com/webhook/..."
                :disabled="!settingsForm.webhooks_enabled"
              />
              <p class="text-sm text-gray-400">
                Notifications will be sent when deployments are completed.
              </p>
              <div class="pt-4">
                <Button
                  type="button"
                  variant="secondary"
                  @click="testTeamsWebhook"
                  :disabled="!settingsForm.webhooks_enabled || !settingsForm.teams_webhook_url || testingWebhook"
                >
                  <RefreshCw v-if="testingWebhook" class="w-4 h-4 animate-spin" />
                  <Bell v-else class="w-4 h-4" />
                  {{ testingWebhook ? 'Sending...' : 'Test Teams Webhook' }}
                </Button>
              </div>
            </div>
          </Card>

          <!-- Parson (local AI) Section -->
          <Card v-if="activeSection === 'ai'" title="Parson — local AI (Ollama)">
            <div class="space-y-4">
              <div class="space-y-2">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    v-model="settingsForm.ai_enabled"
                    type="checkbox"
                    class="w-4 h-4"
                  />
                  <span class="text-white">Enable Parson</span>
                </label>
                <p class="text-sm text-gray-400">
                  Parson is your local AI assistant — it runs on a local Ollama server, so your
                  data never leaves this machine. Run <code class="text-blue-400">./AI.sh</code> once
                  to install a model.
                </p>
              </div>
              <Input
                v-model="settingsForm.ollama_url"
                label="Ollama URL"
                placeholder="http://localhost:11434"
                :disabled="!settingsForm.ai_enabled"
              />

              <!-- Model picker: installed models with guidance -->
              <div>
                <div class="flex items-center justify-between mb-1">
                  <label class="text-sm font-medium text-gray-300">Model</label>
                  <button type="button" class="text-xs text-blue-400 hover:underline" @click="loadModels">
                    Refresh models
                  </button>
                </div>
                <select
                  v-model="settingsForm.ai_model"
                  :disabled="!settingsForm.ai_enabled"
                  class="w-full px-3 py-2 text-sm bg-surface-light border border-surface-border rounded-lg text-white focus:outline-none focus:border-accent"
                >
                  <option v-for="m in installedModels" :key="m.name" :value="m.name">
                    {{ m.name }}{{ m.size_gb ? ` (${m.size_gb.toFixed(1)} GB)` : '' }}
                  </option>
                  <option v-if="installedModels.length === 0" :value="settingsForm.ai_model">
                    {{ settingsForm.ai_model }} (not detected)
                  </option>
                </select>
                <p v-if="currentModelNote" class="text-xs text-blue-300 mt-1">{{ currentModelNote }}</p>
              </div>

              <div class="flex items-center gap-2 pt-1">
                <Button
                  type="button"
                  variant="secondary"
                  @click="testAIConnection"
                  :disabled="!settingsForm.ai_enabled || testingAI"
                >
                  <RefreshCw v-if="testingAI" class="w-4 h-4 animate-spin" />
                  <Zap v-else class="w-4 h-4" />
                  {{ testingAI ? 'Testing (can take a minute)...' : 'Test Parson' }}
                </Button>
              </div>

              <!-- Agent Configuration -->
              <div class="mt-2 p-3 bg-surface-deep rounded-lg border border-surface-lighter space-y-3">
                <div class="flex items-center justify-between">
                  <p class="text-sm font-medium text-white">Agent Configuration</p>
                  <span class="text-xs text-gray-500">Parson runs a single-step report workflow</span>
                </div>

                <div>
                  <label class="block text-xs text-gray-400 mb-1">Temperature ({{ settingsForm.parson_temperature }})</label>
                  <input
                    v-model.number="settingsForm.parson_temperature"
                    type="range" min="0" max="1" step="0.1"
                    :disabled="!settingsForm.ai_enabled"
                    class="w-full"
                  />
                  <p class="text-[11px] text-gray-500">Lower = more consistent/structured. 0.4 is a good default for reports.</p>
                </div>

                <div>
                  <div class="flex items-center justify-between mb-1">
                    <label class="text-xs text-gray-400">System prompt (how Parson writes)</label>
                    <button type="button" class="text-xs text-blue-400 hover:underline" @click="loadDefaultPrompt">
                      Load default into editor
                    </button>
                  </div>
                  <textarea
                    v-model="settingsForm.parson_system_prompt"
                    :disabled="!settingsForm.ai_enabled"
                    rows="10"
                    placeholder="Leave blank to use the built-in default prompt, or paste/edit your own here."
                    class="w-full px-3 py-2 text-xs font-mono bg-surface-light border border-surface-border rounded-lg text-white focus:outline-none focus:border-accent resize-y"
                  ></textarea>
                  <p class="text-[11px] text-gray-500">Blank = built-in default. The report format ("Patch Deployments Completed…", planned/roadblock defaults) lives here — edit carefully.</p>
                </div>
              </div>

              <!-- Models you have / can use -->
              <div class="mt-2 p-3 bg-surface rounded-lg text-xs space-y-2">
                <div>
                  <p class="text-gray-300 font-medium mb-1">Installed models</p>
                  <ul v-if="installedModels.length" class="space-y-1">
                    <li v-for="m in installedModels" :key="m.name" class="flex justify-between gap-2">
                      <span class="text-white">{{ m.name }}</span>
                      <span class="text-gray-400 text-right">{{ m.note || '—' }}</span>
                    </li>
                  </ul>
                  <p v-else class="text-gray-500">None detected (is Ollama running? click Refresh).</p>
                </div>
                <div class="pt-2 border-t border-surface-border">
                  <p class="text-gray-300 font-medium mb-1">Other models you can use (pull with <code class="text-blue-400">ollama pull &lt;name&gt;</code>)</p>
                  <ul class="space-y-1">
                    <li v-for="m in suggestedModels" :key="m.name" class="flex justify-between gap-2">
                      <span class="text-white">{{ m.name }} <span class="text-gray-500">{{ m.size_hint }}</span></span>
                      <span class="text-gray-400 text-right">{{ m.note }}</span>
                    </li>
                  </ul>
                  <p class="text-gray-500 mt-1">Full guide: <code class="text-blue-400">docs/AI_MODEL_GUIDE.md</code></p>
                </div>
              </div>
              <!-- Test result -->
              <div
                v-if="aiTestResult"
                :class="[
                  'mt-2 p-3 rounded-lg text-sm border',
                  aiTestResult.ok
                    ? 'bg-green-900/30 border-green-600 text-green-300'
                    : 'bg-red-900/30 border-red-600 text-red-300',
                ]"
              >
                <template v-if="aiTestResult.ok">
                  ✓ Connected to Ollama {{ aiTestResult.version }} — model
                  <strong>{{ aiTestResult.model }}</strong> responded:
                  “{{ aiTestResult.sample }}”
                </template>
                <template v-else>
                  ✗ {{ aiTestResult.error }}
                  <span v-if="aiTestResult.stage" class="text-gray-400">(stage: {{ aiTestResult.stage }})</span>
                </template>
              </div>
            </div>
          </Card>

          <!-- Email (SMTP) Section -->
          <Card v-if="activeSection === 'email'" title="Email (SMTP)">
            <div class="space-y-4">
              <p class="text-sm text-gray-400">
                Used to send the daily report. Use your mail provider's SMTP server and an
                app-password. Nothing is sent unless you click Send / Test.
              </p>
              <div class="grid grid-cols-2 gap-4">
                <Input v-model="settingsForm.smtp_host" label="SMTP Host" placeholder="smtp.office365.com" />
                <Input v-model.number="settingsForm.smtp_port" label="Port" type="number" placeholder="587" />
              </div>
              <div class="space-y-1">
                <label class="text-sm font-medium text-gray-300">Security</label>
                <Select v-model="settingsForm.smtp_security" :options="['starttls', 'tls', 'none']" />
                <p class="text-[11px] text-gray-500">587 → starttls · 465 → tls · 25 → none</p>
              </div>
              <Input v-model="settingsForm.smtp_user" label="Username" placeholder="you@example.com" />
              <Input v-model="settingsForm.smtp_password" label="Password / App Password" type="password" placeholder="••••••••" />
              <Input v-model="settingsForm.smtp_from" label="From" placeholder="you@example.com" />
              <Input v-model="settingsForm.smtp_to" label="To (comma-separated)" placeholder="team@example.com" />
              <Input v-model="settingsForm.smtp_cc" label="Cc (comma-separated)" placeholder="lead@example.com, manager@example.com" />

              <div class="pt-2 border-t border-surface-lighter">
                <Input v-model="settingsForm.smtp_test_to" label="Test recipient (only for Send Test Email)" placeholder="security@example.com" />
                <p class="text-[11px] text-gray-500 mt-1">Test emails go here (no Cc). Real reports use To + Cc above.</p>
              </div>
              <div class="pt-2">
                <Button type="button" variant="secondary" @click="testEmailConnection" :disabled="testingEmail || !settingsForm.smtp_host">
                  <RefreshCw v-if="testingEmail" class="w-4 h-4 animate-spin" />
                  <Bell v-else class="w-4 h-4" />
                  {{ testingEmail ? 'Sending test...' : 'Send Test Email' }}
                </Button>
              </div>
            </div>
          </Card>

          <!-- Schedule Section -->
          <Card v-if="activeSection === 'schedule'" title="Daily Report Schedule">
            <div class="space-y-4">
              <label class="flex items-center gap-2 cursor-pointer">
                <input v-model="settingsForm.summary_schedule_enabled" type="checkbox" class="w-4 h-4" />
                <span class="text-white">Enable automatic daily report</span>
              </label>
              <p class="text-sm text-gray-400">
                Parson auto-generates the report at the generate time, then sends it at the send time
                on the selected days. Requires Parson (AI) + Email (SMTP) to be configured.
              </p>

              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium text-gray-300 mb-1">Generate at</label>
                  <input v-model="settingsForm.summary_generate_time" type="time"
                    :disabled="!settingsForm.summary_schedule_enabled"
                    class="w-full px-3 py-2 text-sm bg-surface-light border border-surface-border rounded-lg text-white" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-300 mb-1">Send at</label>
                  <input v-model="settingsForm.summary_send_time" type="time"
                    :disabled="!settingsForm.summary_schedule_enabled"
                    class="w-full px-3 py-2 text-sm bg-surface-light border border-surface-border rounded-lg text-white" />
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-300 mb-2">Active days</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="(d, idx) in weekdayLabels"
                    :key="idx"
                    type="button"
                    @click="toggleWeekday(idx)"
                    :disabled="!settingsForm.summary_schedule_enabled"
                    :class="[
                      'px-3 py-1.5 rounded-lg text-sm border transition-all',
                      isWeekdayOn(idx)
                        ? 'bg-accent border-accent text-white'
                        : 'bg-surface-light border-surface-border text-gray-300',
                    ]"
                  >{{ d }}</button>
                </div>
              </div>

              <label class="flex items-center gap-2 cursor-pointer">
                <input v-model="settingsForm.summary_auto_send" type="checkbox" class="w-4 h-4"
                  :disabled="!settingsForm.summary_schedule_enabled" />
                <span class="text-white">Auto-send at send time</span>
              </label>
              <p class="text-[11px] text-gray-500">
                If off, Parson still drafts the report at generate time, but you send it manually from the Daily Summary tab.
              </p>

              <!-- Holidays -->
              <div class="pt-3 border-t border-surface-lighter">
                <p class="text-sm font-medium text-white mb-1">Holidays</p>
                <p class="text-[11px] text-gray-500 mb-2">Parson skips generating/sending on these dates.</p>
                <div class="flex gap-2 mb-2">
                  <input v-model="newHolidayDate" type="date"
                    class="px-3 py-2 text-sm bg-surface-light border border-surface-border rounded-lg text-white" />
                  <input v-model="newHolidayName" type="text" placeholder="Name (e.g. Eid)"
                    class="flex-1 px-3 py-2 text-sm bg-surface-light border border-surface-border rounded-lg text-white" />
                  <Button type="button" variant="secondary" size="sm" @click="addHoliday" :disabled="!newHolidayDate">Add</Button>
                </div>
                <div class="space-y-1 max-h-40 overflow-y-auto">
                  <div v-for="h in summaryStore.holidays" :key="h.id"
                    class="flex items-center justify-between p-2 bg-surface rounded text-sm">
                    <span class="text-white">{{ h.date }} <span class="text-gray-400">{{ h.name }}</span></span>
                    <button type="button" class="text-red-400 hover:text-red-300 text-xs" @click="summaryStore.deleteHoliday(h.id)">Remove</button>
                  </div>
                  <p v-if="summaryStore.holidays.length === 0" class="text-gray-500 text-xs">No holidays added.</p>
                </div>
              </div>
            </div>
          </Card>

          <!-- Display Settings Section -->
          <Card v-if="activeSection === 'display'" title="Display Settings">
            <div class="space-y-4">
              <div class="space-y-2">
                <label class="text-sm font-medium text-gray-300">Theme</label>
                <Select
                  v-model="displaySettings.theme"
                  :options="['Dark', 'Light']"
                />
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium text-gray-300">Items Per Page</label>
                <Select
                  v-model="displaySettings.itemsPerPage"
                  :options="['10', '25', '50', '100']"
                />
              </div>
            </div>
          </Card>

          <!-- Advanced Settings Section -->
          <Card v-if="activeSection === 'advanced'" title="Advanced Settings">
            <div class="space-y-4">
              <div class="space-y-2">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    v-model="advancedSettings.enableNotifications"
                    type="checkbox"
                    class="w-4 h-4"
                  />
                  <span class="text-white">Enable Browser Notifications</span>
                </label>
              </div>
              <div class="space-y-2">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    v-model="advancedSettings.enableAutoRefresh"
                    type="checkbox"
                    class="w-4 h-4"
                  />
                  <span class="text-white">Enable Auto-refresh</span>
                </label>
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium text-gray-300">Auto-refresh Interval (seconds)</label>
                <Input
                  v-model.number="advancedSettings.refreshInterval"
                  type="number"
                  min="5"
                  max="300"
                />
              </div>
            </div>
          </Card>

          <!-- Save Button -->
          <div class="mt-6">
            <Button variant="success" type="submit">
              <Save class="w-4 h-4" />
              Save Settings
            </Button>
          </div>
        </div>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import Card from '../components/ui/Card.vue'
import Input from '../components/ui/Input.vue'
import Select from '../components/ui/Select.vue'
import Button from '../components/ui/Button.vue'
import { useSettingsStore } from '../stores/settings'
import { useSummaryStore } from '../stores/summary'
import { useLibraryStore } from '../stores/library'
import { useApi } from '../composables/useApi'
import { useToast } from '../composables/useToast'
import { Save, Zap, Bell, RefreshCw } from 'lucide-vue-next'

const settingsStore = useSettingsStore()
const summaryStore = useSummaryStore()
const libraryStore = useLibraryStore()

const newHolidayDate = ref('')
const newHolidayName = ref('')
const addHoliday = async () => {
  if (!newHolidayDate.value) return
  await summaryStore.addHoliday(newHolidayDate.value, newHolidayName.value)
  newHolidayDate.value = ''
  newHolidayName.value = ''
}
const { post } = useApi()
const { success, error } = useToast()

const activeSection = ref('user')
const testingJira = ref(false)
const testingWebhook = ref(false)
const testingAI = ref(false)
const testingEmail = ref(false)
const aiTestResult = ref<{ ok: boolean; version?: string; model?: string; sample?: string; stage?: string; error?: string } | null>(null)

interface AIModel { name: string; size_gb?: number; size_hint?: string; note?: string }
const installedModels = ref<AIModel[]>([])
const suggestedModels = ref<AIModel[]>([])

const currentModelNote = computed(() => {
  const m = installedModels.value.find(x => x.name === settingsForm.value.ai_model)
  return m?.note || ''
})

const loadModels = async () => {
  const res = await settingsStore.fetchAIModels()
  installedModels.value = res.installed || []
  suggestedModels.value = res.suggested || []
}

const weekdayLabels = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
const isWeekdayOn = (idx: number) =>
  settingsForm.value.summary_weekdays.split(',').map(s => s.trim()).includes(String(idx))
const toggleWeekday = (idx: number) => {
  const set = new Set(settingsForm.value.summary_weekdays.split(',').map(s => s.trim()).filter(Boolean))
  const key = String(idx)
  if (set.has(key)) set.delete(key); else set.add(key)
  settingsForm.value.summary_weekdays = Array.from(set).map(Number).sort((a, b) => a - b).join(',')
}

const testEmailConnection = async () => {
  // Persist SMTP config first so the backend tests what the user sees.
  await handleSaveSettings()
  testingEmail.value = true
  try {
    const res = await settingsStore.testEmail()
    if (res.ok) success(`Test email sent to ${(res.to || []).join(', ')}`)
    else error(res.error || 'Email test failed')
  } finally {
    testingEmail.value = false
  }
}

const loadDefaultPrompt = async () => {
  const def = await settingsStore.fetchParsonDefaults()
  settingsForm.value.parson_system_prompt = def.system_prompt
  if (!settingsForm.value.parson_temperature) settingsForm.value.parson_temperature = def.temperature
}

const settingsSections = [
  { id: 'user', label: 'User Defaults' },
  { id: 'export', label: 'Export Settings' },
  { id: 'jira', label: 'Jira Integration' },
  { id: 'webhooks', label: 'Webhooks' },
  { id: 'ai', label: 'Parson (AI)' },
  { id: 'email', label: 'Email (SMTP)' },
  { id: 'schedule', label: 'Schedule' },
  { id: 'display', label: 'Display' },
  { id: 'advanced', label: 'Advanced' },
]

const settingsForm = ref({
  default_deployed_by: '',
  excel_export_path: '',
  jira_url: '',
  jira_email: '',
  jira_token: '',
  jira_project: 'PAT',
  teams_webhook_url: '',
  webhooks_enabled: false,
  ai_enabled: false,
  ollama_url: 'http://localhost:11434',
  ai_model: 'gemma4:e4b',
  planned_weekly: '', // carried through so saving Settings doesn't wipe it (edited in Daily Summary)
  parson_system_prompt: '',
  parson_temperature: 0.4,
  smtp_host: '',
  smtp_port: 587,
  smtp_user: '',
  smtp_password: '',
  smtp_from: '',
  smtp_to: '',
  smtp_cc: '',
  smtp_security: 'tls',
  smtp_test_to: '',
  summary_schedule_enabled: false,
  summary_generate_time: '18:30',
  summary_send_time: '20:00',
  summary_weekdays: '0,1,2,3,4',
  summary_auto_send: true,
})

const displaySettings = ref({
  theme: 'Dark',
  itemsPerPage: '25',
})

const advancedSettings = ref({
  enableNotifications: true,
  enableAutoRefresh: false,
  refreshInterval: 30,
})

const developerOptions = computed(() => libraryStore.presets.developers)

const handleSaveSettings = async () => {
  const result = await settingsStore.saveSettings(settingsForm.value)

  if (result) {
    success('Settings saved successfully!')
  } else {
    error('Failed to save settings')
  }
}

const testJiraConnection = async () => {
  // First save settings so the backend has the latest config
  await handleSaveSettings()

  testingJira.value = true
  try {
    const response = await post<{ success: boolean; message?: string; error?: string }>('/jira/test', {})
    if (response.data?.success) {
      success('Jira connection successful!')
    } else {
      error(response.data?.error || 'Jira connection failed')
    }
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : 'Connection test failed'
    error(message)
  } finally {
    testingJira.value = false
  }
}

const testTeamsWebhook = async () => {
  // First save settings so the backend has the latest config
  await handleSaveSettings()

  testingWebhook.value = true
  try {
    const response = await post<{ success: boolean; message?: string; error?: string }>('/webhooks/test/teams', {})
    if (response.data?.success) {
      success('Test notification sent!')
    } else {
      error(response.data?.error || 'Webhook test failed')
    }
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : 'Webhook test failed'
    error(message)
  } finally {
    testingWebhook.value = false
  }
}

const testAIConnection = async () => {
  // Persist current AI config first so the backend tests what the user sees.
  await handleSaveSettings()

  testingAI.value = true
  aiTestResult.value = null
  try {
    const result = await settingsStore.testAI()
    aiTestResult.value = result
    if (result.ok) {
      success(`Parson ready — Ollama ${result.version}, ${result.model}`)
    } else {
      error(result.error || 'Parson test failed')
    }
  } finally {
    testingAI.value = false
  }
}

onMounted(async () => {
  if (libraryStore.presets.developers.length === 0) {
    await libraryStore.fetchPresets()
  }
  await settingsStore.fetchSettings()

  // Populate form with loaded settings
  const s = settingsStore.settings
  settingsForm.value = {
    default_deployed_by: s.default_deployed_by,
    excel_export_path: s.excel_export_path,
    jira_url: s.jira_url || '',
    jira_email: s.jira_email || '',
    jira_token: s.jira_token || '',
    jira_project: s.jira_project || 'PAT',
    teams_webhook_url: s.teams_webhook_url || '',
    webhooks_enabled: s.webhooks_enabled || false,
    ai_enabled: s.ai_enabled || false,
    ollama_url: s.ollama_url || 'http://localhost:11434',
    ai_model: s.ai_model || 'gemma4:e4b',
    planned_weekly: s.planned_weekly || '',
    parson_system_prompt: s.parson_system_prompt || '',
    parson_temperature: s.parson_temperature || 0.4,
    smtp_host: s.smtp_host || '',
    smtp_port: s.smtp_port || 587,
    smtp_user: s.smtp_user || '',
    smtp_password: s.smtp_password || '',
    smtp_from: s.smtp_from || '',
    smtp_to: s.smtp_to || '',
    smtp_cc: s.smtp_cc || '',
    smtp_security: s.smtp_security || 'tls',
    smtp_test_to: s.smtp_test_to || '',
    summary_schedule_enabled: s.summary_schedule_enabled || false,
    summary_generate_time: s.summary_generate_time || '18:30',
    summary_send_time: s.summary_send_time || '20:00',
    summary_weekdays: s.summary_weekdays || '0,1,2,3,4',
    summary_auto_send: s.summary_auto_send ?? true,
  }

  loadModels()
  summaryStore.fetchHolidays()

  // Always show Parson's actual system prompt: if the user hasn't set a custom one,
  // pre-fill the editor with the built-in default so it's visible (and editable).
  if (!settingsForm.value.parson_system_prompt) {
    try {
      const def = await settingsStore.fetchParsonDefaults()
      settingsForm.value.parson_system_prompt = def.system_prompt
    } catch { /* Ollama/server may be down; leave blank */ }
  }
})
</script>
