<template>
  <div class="space-y-4">
    <div class="page-header">
      <div>
        <h1 class="page-title text-xl">Daily Summary</h1>
        <p class="text-gray-400 text-sm">Auto-pulled deployments + your notes → written by <span class="text-blue-400">Parson</span></p>
      </div>
      <div class="flex items-center gap-2">
        <input
          v-model="selectedDate"
          type="date"
          class="px-3 py-2 text-sm bg-surface-light border border-surface-border rounded-lg text-white focus:outline-none focus:border-accent"
          @change="reload"
        />
        <span
          v-if="store.summary"
          :class="['text-xs px-2 py-1 rounded-full', statusClass(store.summary.status)]"
        >
          {{ store.summary.status }}
        </span>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <!-- Left: today's deployments + notes -->
      <div class="space-y-4">
        <Card :title="`Deployments (${store.deployments.length})`">
          <div class="space-y-2 max-h-64 overflow-y-auto">
            <div
              v-for="d in store.deployments"
              :key="d.id"
              class="flex items-center justify-between p-2 bg-surface rounded text-sm"
            >
              <div class="flex items-center gap-2 min-w-0">
                <span class="font-mono text-blue-400 shrink-0">{{ d.jira_id || '(no JIRA)' }}</span>
                <span class="text-white truncate" :title="`${projectName(d.project_id)}${componentName(d.component_id) ? ' (' + componentName(d.component_id) + ')' : ''}`">
                  {{ projectName(d.project_id) }}<span v-if="componentName(d.component_id)" class="text-gray-400"> ({{ componentName(d.component_id) }})</span>
                </span>
              </div>
              <div class="flex items-center gap-1 shrink-0">
                <span
                  class="text-[10px] px-1.5 py-0.5 rounded"
                  :class="d.deploy_status === 'success' ? 'bg-green-500/20 text-green-300' : d.deploy_status === 'failed' ? 'bg-red-500/20 text-red-300' : 'bg-gray-500/20 text-gray-300'"
                >{{ d.deploy_status || 'pending' }}</span>
                <span class="text-xs px-2 py-0.5 rounded bg-surface-lighter text-gray-300">{{ d.environment || '—' }}</span>
              </div>
            </div>
            <p v-if="store.deployments.length === 0" class="text-gray-400 text-center py-3 text-sm">
              No deployments recorded for this day.
            </p>
          </div>
          <!-- Stats / flags -->
          <div v-if="store.stats" class="mt-3 pt-3 border-t border-surface-border space-y-1 text-xs">
            <div class="flex flex-wrap gap-2">
              <span class="px-2 py-1 rounded bg-blue-500/20 text-blue-300">Total: {{ store.stats.total }}</span>
              <span
                v-for="(n, env) in store.stats.by_environment"
                :key="env"
                class="px-2 py-1 rounded bg-purple-500/20 text-purple-300"
              >{{ env }}: {{ n }}</span>
              <span v-if="store.stats.failed_count" class="px-2 py-1 rounded bg-red-500/20 text-red-300">
                Failed: {{ store.stats.failed_count }}
              </span>
            </div>
            <p v-if="store.stats.dirty_patches.length" class="text-yellow-400">
              ⚠ No-JIRA patches: {{ store.stats.dirty_patches.join(', ') }}
            </p>
            <p v-if="store.stats.high_churn.length" class="text-orange-400">
              ⚠ High churn: {{ store.stats.high_churn.join(', ') }}
            </p>
          </div>
        </Card>

        <Card title="Your Notes">
          <div class="space-y-3">
            <div>
              <label class="block text-xs text-gray-400 mb-1">Extra activities (not captured as deployments)</label>
              <textarea v-model="form.notes" rows="3" class="note-area" placeholder="e.g. Migrated Tomcat 11 with OpenJDK 21..."></textarea>
            </div>
            <div>
              <label class="block text-xs text-gray-400 mb-1">
                Planned activities — this week <span class="text-gray-500">(persists across days)</span>
              </label>
              <textarea v-model="weeklyPlanned" rows="2" class="note-area" placeholder="Standing plan for the week (set once, shows every day)..."></textarea>
            </div>
            <div>
              <label class="block text-xs text-gray-400 mb-1">Planned activities — today <span class="text-gray-500">(specific to this day)</span></label>
              <textarea v-model="form.planned" rows="2" class="note-area" placeholder="Anything planned specifically for today..."></textarea>
            </div>
            <div>
              <label class="block text-xs text-gray-400 mb-1">Roadblocks / Suggestions</label>
              <textarea v-model="form.roadblocks" rows="2" class="note-area" placeholder="Any blockers... (blank = 'No roadblocks')"></textarea>
            </div>
            <Button variant="secondary" size="sm" @click="saveNotes">
              <Save class="w-4 h-4" /> Save Notes
            </Button>
          </div>
        </Card>
      </div>

      <!-- Right: generate + preview -->
      <Card title="Activity Report">
        <div class="space-y-3">
          <div class="flex items-end gap-2">
            <div class="flex-1">
              <label class="block text-xs text-gray-400 mb-1">Model</label>
              <select
                v-model="selectedModel"
                class="w-full px-3 py-2 text-sm bg-surface-light border border-surface-border rounded-lg text-white focus:outline-none focus:border-accent"
              >
                <option v-for="m in models" :key="m.name" :value="m.name">{{ m.name }}</option>
                <option v-if="models.length === 0" :value="selectedModel">{{ selectedModel }}</option>
              </select>
            </div>
            <Button variant="primary" @click="generate" :disabled="store.isGenerating">
              <RefreshCw v-if="store.isGenerating" class="w-4 h-4 animate-spin" />
              <Sparkles v-else class="w-4 h-4" />
              {{ store.isGenerating ? 'Parson is writing...' : 'Generate with Parson' }}
            </Button>
          </div>
          <p class="text-xs text-gray-500">Switch models and regenerate to compare output (e.g. e4b vs 12b).</p>
          <p v-if="store.lastGenSeconds !== null && !store.isGenerating" class="text-xs text-green-400">
            ✓ Generated in {{ store.lastGenSeconds.toFixed(1) }}s with {{ store.lastGenModel }}
          </p>
          <p v-if="store.error" class="text-red-400 text-sm">{{ store.error }}</p>

          <template v-if="store.summary && store.summary.generated_body">
            <div>
              <label class="block text-xs text-gray-400 mb-1">Subject</label>
              <input v-model="form.generated_subject" class="note-area" />
            </div>
            <div>
              <label class="block text-xs text-gray-400 mb-1">Body (editable)</label>
              <textarea v-model="form.generated_body" rows="16" class="note-area font-mono text-sm"></textarea>
            </div>
            <div>
              <label class="block text-xs text-gray-400 mb-1">Recipients (blank = use Settings → Email defaults)</label>
              <input v-model="recipients" class="note-area" placeholder="override To, comma-separated" />
            </div>
            <div class="flex gap-2 flex-wrap">
              <Button variant="secondary" size="sm" @click="saveNotes">
                <Save class="w-4 h-4" /> Save Edits
              </Button>
              <Button variant="secondary" size="sm" @click="copyReport" :class="justCopied && '!bg-green-700 !text-white'">
                <Check v-if="justCopied" class="w-4 h-4" />
                <Copy v-else class="w-4 h-4" />
                {{ justCopied ? 'Copied!' : 'Copy' }}
              </Button>
              <Button variant="success" size="sm" @click="sendReport" :disabled="store.isSending">
                <RefreshCw v-if="store.isSending" class="w-4 h-4 animate-spin" />
                <Send v-else class="w-4 h-4" />
                {{ store.isSending ? 'Sending...' : 'Send now' }}
              </Button>
            </div>
            <p v-if="store.summary?.status === 'sent'" class="text-xs text-green-400">
              ✓ Sent{{ store.summary.sent_at ? ' at ' + new Date(store.summary.sent_at).toLocaleString() : '' }}
            </p>
            <p class="text-xs text-gray-500">
              Configure SMTP in Settings → Email. The 6 PM/8 PM auto-scheduler comes next.
            </p>
          </template>

          <div v-else class="text-center py-10 text-gray-500 text-sm">
            <Sparkles class="w-8 h-8 mx-auto mb-2 opacity-40" />
            Add your notes, then click <strong>Generate with AI</strong>.
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import Card from '../components/ui/Card.vue'
import Button from '../components/ui/Button.vue'
import { useSummaryStore } from '../stores/summary'
import { useProjectsStore } from '../stores/projects'
import { useSettingsStore } from '../stores/settings'
import { useToast } from '../composables/useToast'
import { useClipboard } from '../composables/useClipboard'
import { Save, RefreshCw, Sparkles, Copy, Check, Send } from 'lucide-vue-next'

const store = useSummaryStore()
const projectsStore = useProjectsStore()
const settingsStore = useSettingsStore()
const { success, error } = useToast()
const { copyToClipboard } = useClipboard()

const models = ref<{ name: string }[]>([])
const selectedModel = ref('gemma4:e4b')
const weeklyPlanned = ref('')
const recipients = ref('')

const todayStr = () => {
  const d = new Date()
  const m = `${d.getMonth() + 1}`.padStart(2, '0')
  const day = `${d.getDate()}`.padStart(2, '0')
  return `${d.getFullYear()}-${m}-${day}`
}

const selectedDate = ref(todayStr())
const form = ref({ notes: '', planned: '', roadblocks: '', generated_subject: '', generated_body: '' })

const projectName = (id: number) => projectsStore.getProjectName(id)
const componentName = (id?: number) => {
  if (!id) return ''
  const n = projectsStore.getComponentName(id)
  return n && !n.startsWith('Component ') ? n : ''
}

const syncForm = () => {
  const s = store.summary
  if (!s) return
  form.value = {
    notes: s.notes || '',
    planned: s.planned || '',
    roadblocks: s.roadblocks || '',
    generated_subject: s.generated_subject || '',
    generated_body: s.generated_body || '',
  }
  recipients.value = s.recipients || ''
}

const sendReport = async () => {
  // Persist the latest edits + recipient override, then send.
  await store.saveNotes({ ...form.value, recipients: recipients.value })
  const result = await store.send()
  if (result.ok) { success('Report sent'); syncForm() }
  else error(result.error || 'Send failed')
}

const reload = async () => {
  await store.fetchToday(selectedDate.value)
  syncForm()
}

const saveNotes = async () => {
  // Persist the day's notes (summary) and the standing weekly plan (settings).
  const ok = await store.saveNotes(form.value)
  await settingsStore.saveSettings({ planned_weekly: weeklyPlanned.value })
  if (ok) { success('Saved'); syncForm() }
  else error('Failed to save')
}

const generate = async () => {
  // Persist notes first so the model sees the latest input.
  await store.saveNotes(form.value)
  const result = await store.generate(selectedModel.value)
  if (result.ok) { success(`Report generated (${selectedModel.value})`); syncForm() }
  else error(result.error || 'Generation failed')
}

const justCopied = ref(false)

const copyReport = async () => {
  const text = `Subject: ${form.value.generated_subject}\n\n${form.value.generated_body}`
  if (await copyToClipboard(text)) {
    justCopied.value = true
    setTimeout(() => (justCopied.value = false), 2000)
    success('Report copied')
  }
  else error('Failed to copy')
}

const statusClass = (status: string) => {
  switch (status) {
    case 'generated': return 'bg-blue-500/20 text-blue-300'
    case 'approved': return 'bg-green-500/20 text-green-300'
    case 'sent': return 'bg-green-600/30 text-green-300'
    case 'failed': return 'bg-red-500/20 text-red-300'
    default: return 'bg-gray-500/20 text-gray-300'
  }
}

watch(() => store.summary?.id, syncForm)

onMounted(async () => {
  if (projectsStore.projects.length === 0) await projectsStore.fetchProjects()
  await reload()
  // Populate the model selector from installed Ollama models + current setting.
  await settingsStore.fetchSettings()
  selectedModel.value = settingsStore.settings.ai_model || 'gemma4:e4b'
  weeklyPlanned.value = settingsStore.settings.planned_weekly || ''
  const res = await settingsStore.fetchAIModels()
  models.value = res.installed || []
})
</script>

<style scoped>
.note-area {
  width: 100%;
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  background: #404040;
  border: 1px solid #555555;
  border-radius: 0.5rem;
  color: white;
  resize: vertical;
}
.note-area:focus {
  outline: none;
  border-color: #4a9eff;
}
</style>
