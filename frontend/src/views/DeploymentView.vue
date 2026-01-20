<template>
  <div class="h-full flex flex-col">
    <!-- Favorites Section - Takes top portion -->
    <div class="flex-shrink-0 mb-4">
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-sm font-medium text-gray-400">Quick Select</h2>
        <button
          @click="showAllProjects = !showAllProjects"
          class="text-xs text-[#4a9eff] hover:underline"
        >
          {{ showAllProjects ? 'Show less' : 'Browse all' }}
        </button>
      </div>

      <!-- Favorites Grid -->
      <div class="grid grid-cols-4 gap-2" v-if="!showAllProjects">
        <button
          v-for="fav in displayedFavorites"
          :key="fav.value"
          @click="selectProjectComponent(fav.value)"
          :class="[
            'p-3 rounded-lg border-2 transition-all text-left',
            selectedValue === fav.value
              ? 'bg-[#4a9eff]/20 border-[#4a9eff] ring-2 ring-[#4a9eff]/30'
              : 'bg-[#353535] border-[#444444] hover:border-[#4a9eff]/50'
          ]"
        >
          <div class="text-xs text-gray-400 truncate">{{ fav.projectName }}</div>
          <div class="text-sm font-medium text-white truncate">{{ fav.componentName }}</div>
          <div class="text-xs text-gray-500 mt-1">{{ fav.count || 0 }} deploys</div>
        </button>

        <!-- Add/Search button -->
        <button
          @click="showAllProjects = true"
          class="p-3 rounded-lg border-2 border-dashed border-[#555555] hover:border-[#4a9eff] transition-all flex flex-col items-center justify-center text-gray-400 hover:text-[#4a9eff]"
        >
          <Search class="w-5 h-5 mb-1" />
          <span class="text-xs">Search All</span>
        </button>
      </div>

      <!-- Searchable Project List (expanded view) -->
      <div v-else class="bg-[#353535] rounded-lg border border-[#444444] p-3">
        <div class="relative mb-3">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search projects or components..."
            class="w-full pl-9 pr-3 py-2 text-sm bg-[#2a2a2a] border border-[#555555] rounded-lg focus:outline-none focus:border-[#4a9eff] text-white placeholder-gray-500"
            autofocus
          />
        </div>
        <div class="grid grid-cols-4 gap-2 max-h-48 overflow-y-auto">
          <button
            v-for="option in filteredOptions"
            :key="option.value"
            @click="selectProjectComponent(option.value); showAllProjects = false"
            :class="[
              'p-2 rounded-lg border transition-all text-left',
              selectedValue === option.value
                ? 'bg-[#4a9eff]/20 border-[#4a9eff]'
                : 'bg-[#404040] border-[#555555] hover:border-[#4a9eff]/50'
            ]"
          >
            <div class="text-xs text-gray-400 truncate">{{ option.project.name }}</div>
            <div class="text-sm font-medium text-white truncate">{{ option.component.name }}</div>
          </button>
        </div>
      </div>
    </div>

    <!-- Main Deploy Form - Compact when favorite selected -->
    <div v-if="selectedOption" class="flex-1 flex flex-col">
      <!-- Selected Info Bar -->
      <div class="flex items-center gap-4 p-3 bg-[#4a9eff]/10 border border-[#4a9eff]/30 rounded-lg mb-4">
        <div class="flex-1">
          <span class="text-[#4a9eff] font-medium">{{ selectedOption.project.name }}</span>
          <span class="text-gray-400 mx-2">›</span>
          <span class="text-white font-medium">{{ selectedOption.component.name }}</span>
        </div>
        <div class="text-xs text-gray-400 flex gap-4">
          <span>Env: <span class="text-gray-300">{{ selectedOption.project.environment }}</span></span>
          <span>Build: <span class="text-gray-300">{{ selectedOption.project.build_server }}</span></span>
          <span>Deploy: <span class="text-gray-300">{{ selectedOption.project.deploy_server }}</span></span>
        </div>
        <button @click="clearSelection" class="text-gray-400 hover:text-white">
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Quick Input Row -->
      <div class="bg-[#353535] rounded-lg border border-[#444444] p-4">
        <div class="flex items-end gap-4">
          <!-- Patch ID with PAT- prefix -->
          <div class="flex-1">
            <label class="text-xs font-medium text-gray-400 block mb-1.5">Patch ID</label>
            <div class="flex items-center">
              <span class="px-3 py-2 bg-[#2a2a2a] border border-r-0 border-[#555555] rounded-l-lg text-gray-400 text-sm">
                PAT-
              </span>
              <input
                v-model="patchNumber"
                type="text"
                placeholder="123"
                class="flex-1 px-3 py-2 bg-[#404040] border border-[#555555] text-white text-sm focus:outline-none focus:border-[#4a9eff]"
                :class="noPatchId ? 'opacity-50' : ''"
                :disabled="noPatchId"
                @keydown.enter="handleSubmit"
              />
              <button
                @click="toggleNoPatchId"
                :class="[
                  'px-3 py-2 border border-l-0 border-[#555555] rounded-r-lg text-sm transition-all',
                  noPatchId
                    ? 'bg-orange-600 text-white'
                    : 'bg-[#2a2a2a] text-gray-400 hover:text-orange-400'
                ]"
                :title="noPatchId ? 'Patch ID skipped' : 'Click to skip Patch ID'"
              >
                <X class="w-4 h-4" />
              </button>
            </div>
            <div v-if="noPatchId" class="text-xs text-orange-400 mt-1">No Patch ID</div>
          </div>

          <!-- Status Toggles (compact) -->
          <div class="flex items-center gap-4">
            <div class="flex items-center gap-2">
              <span class="text-xs text-gray-400">Build:</span>
              <button
                @click="form.build_status = 'success'"
                :class="[
                  'px-2 py-1 text-xs rounded-l',
                  form.build_status === 'success' ? 'bg-green-600 text-white' : 'bg-[#404040] text-gray-400'
                ]"
              >✓</button>
              <button
                @click="form.build_status = 'failed'"
                :class="[
                  'px-2 py-1 text-xs rounded-r',
                  form.build_status === 'failed' ? 'bg-red-600 text-white' : 'bg-[#404040] text-gray-400'
                ]"
              >✗</button>
            </div>

            <div class="flex items-center gap-2">
              <span class="text-xs text-gray-400">Deploy:</span>
              <button
                @click="form.deploy_status = 'success'"
                :class="[
                  'px-2 py-1 text-xs rounded-l',
                  form.deploy_status === 'success' ? 'bg-green-600 text-white' : 'bg-[#404040] text-gray-400'
                ]"
              >✓</button>
              <button
                @click="form.deploy_status = 'failed'"
                :class="[
                  'px-2 py-1 text-xs rounded-r',
                  form.deploy_status === 'failed' ? 'bg-red-600 text-white' : 'bg-[#404040] text-gray-400'
                ]"
              >✗</button>
            </div>

            <div class="flex items-center gap-2">
              <span class="text-xs text-gray-400">DB:</span>
              <button
                @click="form.has_db_script = !form.has_db_script"
                :class="[
                  'px-2 py-1 text-xs rounded',
                  form.has_db_script ? 'bg-blue-600 text-white' : 'bg-[#404040] text-gray-400'
                ]"
              >{{ form.has_db_script ? 'Yes' : 'No' }}</button>
            </div>
          </div>

          <!-- Save Button -->
          <Button
            variant="success"
            @click="handleSubmit"
            :disabled="isSubmitting"
            class="px-6"
          >
            <Save class="w-4 h-4" />
            {{ isSubmitting ? 'Saving...' : 'Save & Copy' }}
          </Button>
        </div>

        <!-- DB Script Name (conditional) -->
        <div v-if="form.has_db_script" class="mt-3">
          <input
            v-model="form.database_script_name"
            type="text"
            placeholder="Script name, e.g., migration_001.sql"
            class="w-full px-3 py-2 text-sm bg-[#404040] border border-[#555555] rounded-lg text-white focus:outline-none focus:border-[#4a9eff]"
          />
        </div>

        <!-- Notes (expandable) -->
        <div class="mt-3">
          <button
            @click="showNotes = !showNotes"
            class="text-xs text-gray-400 hover:text-[#4a9eff]"
          >
            {{ showNotes ? '− Hide notes' : '+ Add notes' }}
          </button>
          <textarea
            v-if="showNotes"
            v-model="form.notes"
            placeholder="Additional notes..."
            rows="2"
            class="w-full mt-2 px-3 py-2 text-sm bg-[#404040] border border-[#555555] rounded-lg text-white focus:outline-none focus:border-[#4a9eff] resize-none"
          ></textarea>
        </div>
      </div>

      <!-- Success State -->
      <div v-if="saveSuccess" class="mt-4 p-4 bg-green-900/30 border border-green-600 rounded-lg flex items-center gap-4">
        <CheckCircle class="w-6 h-6 text-green-400" />
        <div class="flex-1">
          <div class="text-green-400 font-medium">Saved & Copied!</div>
          <div class="text-gray-400 text-sm">Ready to paste in Teams</div>
        </div>
        <Button variant="secondary" @click="clearForNext">
          <Plus class="w-4 h-4" />
          Deploy Another
        </Button>
      </div>
    </div>

    <!-- Empty State - No selection -->
    <div v-else class="flex-1 flex items-center justify-center">
      <div class="text-center text-gray-400">
        <Rocket class="w-12 h-12 mx-auto mb-3 opacity-30" />
        <p class="text-lg">Select a project above to deploy</p>
        <p class="text-sm mt-1">Or press <kbd class="px-1.5 py-0.5 bg-[#404040] rounded text-xs">Ctrl+R</kbd> to repeat last</p>
      </div>
    </div>

    <!-- Keyboard shortcuts hint -->
    <div class="flex-shrink-0 flex items-center justify-center gap-6 py-3 text-xs text-gray-500 border-t border-[#404040]">
      <span><kbd class="px-1 py-0.5 bg-[#353535] rounded">Ctrl+S</kbd> Save</span>
      <span><kbd class="px-1 py-0.5 bg-[#353535] rounded">Ctrl+R</kbd> Repeat Last</span>
      <span><kbd class="px-1 py-0.5 bg-[#353535] rounded">Ctrl+L</kbd> Clear</span>
      <span><kbd class="px-1 py-0.5 bg-[#353535] rounded">Enter</kbd> Quick Save</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import Button from '../components/ui/Button.vue'
import { useProjectsStore } from '../stores/projects'
import { useDeploymentsStore } from '../stores/deployments'
import { useSettingsStore } from '../stores/settings'
import { useClipboard, type DeploymentData } from '../composables/useClipboard'
import { useToast } from '../composables/useToast'
import { Save, CheckCircle, Plus, Search, X, Rocket } from 'lucide-vue-next'

const projectsStore = useProjectsStore()
const deploymentsStore = useDeploymentsStore()
const settingsStore = useSettingsStore()
const { copyToClipboard, formatForTeams } = useClipboard()
const { success, error } = useToast()

// State
const selectedValue = ref('')
const patchNumber = ref('')
const noPatchId = ref(false)
const showAllProjects = ref(false)
const searchQuery = ref('')
const showNotes = ref(false)
const isSubmitting = ref(false)
const saveSuccess = ref(false)

const form = ref({
  build_status: 'success' as 'success' | 'failed',
  deploy_status: 'success' as 'success' | 'failed',
  has_db_script: false,
  database_script_name: '',
  notes: '',
  deployed_by: 'Kannan',
})

// Track deployment counts per project/component for sorting favorites
const deploymentCounts = computed(() => {
  const counts: Record<string, number> = {}
  for (const d of deploymentsStore.deployments) {
    const key = `${d.project_id}-${d.component_id}`
    counts[key] = (counts[key] || 0) + 1
  }
  return counts
})

// Build favorites list (sorted by usage)
const displayedFavorites = computed(() => {
  return projectsStore.projectComponentOptions
    .map(opt => ({
      ...opt,
      projectName: opt.project.name,
      componentName: opt.component.name,
      count: deploymentCounts.value[opt.value] || 0,
    }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 7) // Top 7 + search button = 8 grid items
})

// Filtered options for search
const filteredOptions = computed(() => {
  if (!searchQuery.value.trim()) {
    return projectsStore.projectComponentOptions
  }
  const query = searchQuery.value.toLowerCase()
  return projectsStore.projectComponentOptions.filter(opt =>
    opt.project.name?.toLowerCase().includes(query) ||
    opt.component.name?.toLowerCase().includes(query)
  )
})

// Get selected option details
const selectedOption = computed(() => {
  if (!selectedValue.value) return null
  return projectsStore.projectComponentOptions.find(o => o.value === selectedValue.value)
})

// Computed patch ID
const fullPatchId = computed(() => {
  if (noPatchId.value) return null
  if (!patchNumber.value.trim()) return null
  return `PAT-${patchNumber.value.trim()}`
})

// Handlers
const selectProjectComponent = (value: string) => {
  selectedValue.value = value
  saveSuccess.value = false
  // Reset form for new selection
  patchNumber.value = ''
  noPatchId.value = false
  form.value.build_status = 'success'
  form.value.deploy_status = 'success'
  form.value.has_db_script = false
  form.value.database_script_name = ''
  form.value.notes = ''
  showNotes.value = false
}

const clearSelection = () => {
  selectedValue.value = ''
  saveSuccess.value = false
}

const toggleNoPatchId = () => {
  noPatchId.value = !noPatchId.value
  if (noPatchId.value) {
    patchNumber.value = ''
  }
}

const clearForNext = () => {
  patchNumber.value = ''
  noPatchId.value = false
  form.value.build_status = 'success'
  form.value.deploy_status = 'success'
  form.value.has_db_script = false
  form.value.database_script_name = ''
  form.value.notes = ''
  showNotes.value = false
  saveSuccess.value = false
}

const buildDeploymentData = (): DeploymentData => {
  const opt = selectedOption.value
  if (!opt) return {} as DeploymentData

  return {
    patchId: fullPatchId.value || undefined,
    project: opt.project.name,
    component: opt.component.name,
    environment: opt.project.environment,
    componentUrl: opt.component.component_url,
    buildServer: opt.project.build_server,
    buildStatus: form.value.build_status,
    vcsUrl: opt.component.vcs_url,
    deployServer: opt.project.deploy_server,
    buildBackup: opt.project.backup_location,
    databaseName: opt.project.database_name,
    deployStatus: form.value.deploy_status,
    databaseScript: form.value.has_db_script ? form.value.database_script_name : undefined,
    developer: opt.component.developer,
    deployedBy: form.value.deployed_by,
    timestamp: new Date().toISOString(),
    notes: form.value.notes || undefined,
  }
}

const handleSubmit = async () => {
  if (isSubmitting.value || !selectedOption.value) return
  if (saveSuccess.value) return // Already saved

  isSubmitting.value = true

  const opt = selectedOption.value
  const now = new Date()

  const deployment = {
    jira_id: fullPatchId.value,
    project_id: opt.projectId,
    component_id: opt.componentId,
    timestamp: now.toISOString(),
    environment: opt.project.environment,
    vcs_url: opt.component.vcs_url,
    developer_name: opt.component.developer,
    build_server: opt.project.build_server,
    deploy_server: opt.project.deploy_server,
    database_name: opt.project.database_name,
    database_script: form.value.has_db_script ? form.value.database_script_name : undefined,
    build_status: form.value.build_status,
    deploy_status: form.value.deploy_status,
    notes: form.value.notes,
    deployed_by: form.value.deployed_by,
  }

  try {
    const result = await deploymentsStore.createDeployment(deployment)
    if (result) {
      // Store for "Repeat Last"
      deploymentsStore.setLastDeployment(opt.projectId, opt.componentId)

      // Auto-copy to Teams format
      const data = buildDeploymentData()
      const text = formatForTeams(data)
      await copyToClipboard(text)

      saveSuccess.value = true
      success('Saved & copied to clipboard!')
    } else {
      error('Failed to save deployment')
    }
  } catch (err) {
    error('Failed to save deployment')
  } finally {
    isSubmitting.value = false
  }
}

const repeatLast = () => {
  const last = deploymentsStore.lastDeployment
  if (last) {
    const option = projectsStore.projectComponentOptions.find(
      o => o.projectId === last.project_id && o.componentId === last.component_id
    )
    if (option) {
      selectProjectComponent(option.value)
      success('Loaded last deployment')
    }
  }
}

// Keyboard shortcuts
const handleKeydown = (e: KeyboardEvent) => {
  if (e.ctrlKey || e.metaKey) {
    switch (e.key.toLowerCase()) {
      case 's':
        e.preventDefault()
        if (selectedOption.value && !saveSuccess.value) {
          handleSubmit()
        }
        break
      case 'r':
        e.preventDefault()
        repeatLast()
        break
      case 'l':
        e.preventDefault()
        clearSelection()
        break
    }
  }
}

onMounted(async () => {
  window.addEventListener('keydown', handleKeydown)

  if (projectsStore.projects.length === 0) {
    await projectsStore.fetchProjects()
  }
  if (deploymentsStore.deployments.length === 0) {
    await deploymentsStore.fetchDeployments()
  }
  await settingsStore.fetchSettings()
  form.value.deployed_by = settingsStore.settings?.default_deployed_by || 'Kannan'
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>
