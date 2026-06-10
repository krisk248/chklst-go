<template>
  <div class="space-y-6">
    <!-- Enhanced Header -->
    <div class="page-header">
      <div>
        <div class="flex items-center gap-3">
          <div class="p-2 bg-blue-500/20 rounded-lg">
            <FolderKanban class="w-6 h-6 text-blue-400" />
          </div>
          <div>
            <h1 class="page-title">Projects</h1>
            <p class="text-gray-400 text-sm">{{ projectsStore.projects.length }} projects configured</p>
          </div>
        </div>
      </div>
      <div class="flex items-center gap-3">
        <Button variant="secondary" @click="exportProjects">
          <Download class="w-4 h-4" />
          Export TOML
        </Button>
        <Button variant="primary" @click="showImportModal = true">
          <Upload class="w-4 h-4" />
          Import TOML
        </Button>
        <Button variant="success" @click="showAddProjectDialog = true" class="shadow-lg shadow-green-500/20">
          <Plus class="w-4 h-4" />
          Add Project
        </Button>
      </div>
    </div>

    <div class="grid grid-cols-3 gap-6">
      <!-- Projects List -->
      <div class="col-span-1">
        <Card title="Projects">
          <template #header-right>
            <span class="text-xs px-2 py-1 bg-blue-500/20 text-blue-400 rounded-full">
              {{ projectsStore.projects.length }}
            </span>
          </template>
          <div class="space-y-2" v-if="projectsStore.projects.length > 0">
            <button
              v-for="project in projectsStore.projects"
              :key="project.id"
              @click="selectedProjectId = project.id"
              :class="[
                'w-full text-left px-4 py-3 rounded-lg border transition-all group',
                selectedProjectId === project.id
                  ? 'bg-gradient-to-r from-blue-600 to-blue-500 border-blue-400 text-white shadow-lg shadow-blue-500/30'
                  : 'bg-surface-light border-surface-border text-gray-300 hover:border-blue-400 hover:bg-surface-lighter',
              ]"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <Folder :class="[
                    'w-4 h-4 transition-colors',
                    selectedProjectId === project.id ? 'text-white' : 'text-gray-400 group-hover:text-blue-400'
                  ]" />
                  <span class="font-medium">{{ project.name }}</span>
                </div>
                <span :class="[
                  'text-xs px-2 py-0.5 rounded-full',
                  selectedProjectId === project.id
                    ? 'bg-white/20 text-white'
                    : 'bg-surface-border text-gray-400'
                ]">
                  {{ project.components?.length || 0 }} comp
                </span>
              </div>
            </button>
          </div>
          <!-- Enhanced Empty State -->
          <div v-else class="text-center py-12">
            <div class="p-4 bg-surface-light rounded-full w-16 h-16 mx-auto mb-4 flex items-center justify-center">
              <FolderPlus class="w-8 h-8 text-gray-500" />
            </div>
            <p class="text-gray-400 mb-2">No projects yet</p>
            <p class="text-gray-500 text-sm mb-4">Create your first project to get started</p>
            <Button variant="success" size="sm" @click="showAddProjectDialog = true">
              <Plus class="w-4 h-4" />
              Add Project
            </Button>
          </div>
          <div v-if="projectsStore.projects.length > 0" class="flex gap-2 mt-4 pt-4 border-t border-surface-border">
            <Button
              variant="danger"
              size="sm"
              @click="handleDeleteProject"
              :disabled="!selectedProjectId"
              class="flex-1"
            >
              <Trash2 class="w-4 h-4" />
              Delete
            </Button>
            <Button
              variant="secondary"
              size="sm"
              @click="handleCopyProject"
              :disabled="!selectedProjectId"
              class="flex-1"
            >
              <Copy class="w-4 h-4" />
              Copy
            </Button>
            <Button
              variant="primary"
              size="sm"
              @click="handleDuplicateProject"
              :disabled="!selectedProjectId"
              class="flex-1"
            >
              <CopyPlus class="w-4 h-4" />
              Duplicate
            </Button>
          </div>
        </Card>
      </div>

      <!-- Project Details -->
      <div class="col-span-2">
        <Card v-if="editingProject" title="Project Details">
          <template #header-right>
            <span class="text-xs px-2 py-1 bg-green-500/20 text-green-400 rounded-full">
              Editing
            </span>
          </template>
          <form @submit.prevent="handleSaveProject" class="space-y-4">
            <!-- Project Info Grid -->
            <div class="grid grid-cols-2 gap-4">
              <div class="col-span-2">
                <Input
                  v-model="editingProject.name"
                  label="Project Name"
                  required
                />
              </div>
              <Select
                v-model="editingProject.build_server"
                :options="libraryStore.presets.build_servers"
                label="Build Server"
                required
              />
              <Select
                v-model="editingProject.deploy_server"
                :options="libraryStore.presets.deploy_servers"
                label="Deploy Server"
                required
              />
              <Input
                v-model="editingProject.database_name"
                label="Database"
                required
              />
              <Select
                v-model="editingProject.environment"
                :options="libraryStore.presets.environments"
                label="Environment"
                required
              />
              <div class="col-span-2">
                <Input
                  v-model="editingProject.backup_location"
                  label="Backup Location"
                  placeholder="/backups/project"
                />
              </div>
            </div>

            <!-- Components Section -->
            <div class="border-t border-surface-border pt-6 mt-6">
              <div class="flex items-center justify-between mb-4">
                <div class="flex items-center gap-2">
                  <Layers class="w-5 h-5 text-purple-400" />
                  <h3 class="text-lg font-bold text-white">Components</h3>
                  <span class="text-xs px-2 py-0.5 bg-purple-500/20 text-purple-400 rounded-full">
                    {{ selectedProject?.components?.length || 0 }}
                  </span>
                </div>
                <Button
                  type="button"
                  variant="success"
                  size="sm"
                  @click="showAddComponentDialog = true"
                  class="shadow-md shadow-green-500/20"
                >
                  <Plus class="w-4 h-4" />
                  Add Component
                </Button>
              </div>

              <div class="space-y-3" v-if="selectedProject?.components?.length">
                <div
                  v-for="component in selectedProject?.components || []"
                  :key="component.id"
                  class="flex items-center justify-between bg-gradient-to-r from-surface-light to-[#3a3a3a] p-4 rounded-lg border border-surface-border hover:border-purple-400/50 transition-all group"
                >
                  <div class="flex items-center gap-3">
                    <div class="p-2 bg-purple-500/20 rounded-lg">
                      <Code2 class="w-4 h-4 text-purple-400" />
                    </div>
                    <div>
                      <p class="font-medium text-white group-hover:text-purple-300 transition-colors">{{ component.name }}</p>
                      <div class="flex items-center gap-2 mt-1">
                        <span class="text-xs text-gray-400">{{ component.developer }}</span>
                        <span class="text-gray-600">•</span>
                        <span class="text-xs px-1.5 py-0.5 bg-surface-border text-gray-400 rounded">{{ component.vcs_type }}</span>
                      </div>
                    </div>
                  </div>
                  <div class="flex gap-2 opacity-60 group-hover:opacity-100 transition-opacity">
                    <Button
                      type="button"
                      variant="secondary"
                      size="sm"
                      @click="editComponent(component)"
                    >
                      <Pencil class="w-3 h-3" />
                      Edit
                    </Button>
                    <Button
                      type="button"
                      variant="danger"
                      size="sm"
                      @click="handleDeleteComponent(component.id!)"
                    >
                      <Trash2 class="w-3 h-3" />
                    </Button>
                  </div>
                </div>
              </div>
              <!-- Empty Components State -->
              <div v-else class="text-center py-10 bg-surface-light/50 rounded-lg border border-dashed border-surface-border">
                <div class="p-3 bg-surface-border rounded-full w-12 h-12 mx-auto mb-3 flex items-center justify-center">
                  <Layers class="w-6 h-6 text-gray-400" />
                </div>
                <p class="text-gray-400 mb-1">No components yet</p>
                <p class="text-gray-500 text-sm mb-3">Add components to track deployments</p>
                <Button
                  type="button"
                  variant="success"
                  size="sm"
                  @click="showAddComponentDialog = true"
                >
                  <Plus class="w-4 h-4" />
                  Add First Component
                </Button>
              </div>
            </div>

            <div class="flex gap-3 justify-end pt-4 border-t border-surface-border">
              <Button variant="secondary" type="button" @click="selectedProjectId = null">
                Cancel
              </Button>
              <Button variant="success" type="submit" class="shadow-lg shadow-green-500/20">
                <Save class="w-4 h-4" />
                Save Project
              </Button>
            </div>
          </form>
        </Card>
        <!-- No Project Selected State -->
        <Card v-else title="Project Details">
          <div class="text-center py-16">
            <div class="p-4 bg-surface-light rounded-full w-20 h-20 mx-auto mb-4 flex items-center justify-center">
              <MousePointerClick class="w-10 h-10 text-gray-500" />
            </div>
            <p class="text-gray-300 text-lg mb-2">No project selected</p>
            <p class="text-gray-500 text-sm">Select a project from the list to view and edit its details</p>
          </div>
        </Card>
      </div>
    </div>

    <!-- Duplicate Project Dialog -->
    <Dialog :is-open="showDuplicateDialog" title="Duplicate Project" @close="showDuplicateDialog = false">
      <div class="space-y-3">
        <p class="text-sm text-gray-400">
          Creates an independent copy of <strong class="text-white">{{ selectedProject?.name }}</strong>
          with all {{ selectedProject?.components?.length || 0 }} component(s). Deployment history is not copied.
        </p>
        <Input v-model="duplicateName" label="New project name" @keyup.enter="confirmDuplicateProject" />
      </div>
      <template #footer>
        <Button variant="secondary" @click="showDuplicateDialog = false">Cancel</Button>
        <Button variant="primary" @click="confirmDuplicateProject">
          <CopyPlus class="w-4 h-4" /> Duplicate
        </Button>
      </template>
    </Dialog>

    <!-- Delete Project Confirmation -->
    <Dialog :is-open="showDeleteProjectDialog" title="Delete Project?" @close="showDeleteProjectDialog = false">
      <p class="text-sm text-gray-300">
        <strong class="text-white">{{ selectedProject?.name }}</strong> and its
        {{ selectedProject?.components?.length || 0 }} component(s) — including all of its
        deployment history — will be permanently deleted. <span class="text-red-400">This cannot be undone.</span>
      </p>
      <template #footer>
        <Button variant="secondary" @click="showDeleteProjectDialog = false">Cancel</Button>
        <Button variant="danger" @click="confirmDeleteProject">
          <Trash2 class="w-4 h-4" /> Delete Project
        </Button>
      </template>
    </Dialog>

    <!-- Delete Component Confirmation -->
    <Dialog :is-open="deleteComponentId !== null" title="Delete Component?" @close="deleteComponentId = null">
      <p class="text-sm text-gray-300">
        This component will be removed from <strong class="text-white">{{ selectedProject?.name }}</strong>.
        <span class="text-red-400">This cannot be undone.</span>
      </p>
      <template #footer>
        <Button variant="secondary" @click="deleteComponentId = null">Cancel</Button>
        <Button variant="danger" @click="confirmDeleteComponent">
          <Trash2 class="w-4 h-4" /> Delete Component
        </Button>
      </template>
    </Dialog>

    <!-- Add/Edit Project Dialog -->
    <Dialog
      :is-open="showAddProjectDialog"
      title="Add New Project"
      @close="showAddProjectDialog = false"
    >
      <div class="space-y-4">
        <Input
          v-model="newProject.name"
          label="Project Name"
          placeholder="e.g., BRHUB"
          required
        />
        <Select
          v-model="newProject.build_server"
          :options="libraryStore.presets.build_servers"
          label="Build Server"
          required
        />
        <Select
          v-model="newProject.deploy_server"
          :options="libraryStore.presets.deploy_servers"
          label="Deploy Server"
          required
        />
        <Input
          v-model="newProject.database_name"
          label="Database"
          required
        />
        <Select
          v-model="newProject.environment"
          :options="libraryStore.presets.environments"
          label="Environment"
          required
        />
        <Input
          v-model="newProject.backup_location"
          label="Backup Location"
          placeholder="/backups/project"
        />
      </div>

      <template #footer>
        <Button variant="secondary" @click="showAddProjectDialog = false">
          Cancel
        </Button>
        <Button variant="success" @click="handleAddProject">
          Create Project
        </Button>
      </template>
    </Dialog>

    <!-- Add/Edit Component Dialog -->
    <Dialog
      :is-open="showAddComponentDialog"
      :title="editingComponent ? 'Edit Component' : 'Add Component'"
      @close="showAddComponentDialog = false"
    >
      <div class="space-y-4">
        <Input
          v-model="newComponent.name"
          label="Component Name"
          required
        />
        <Select
          v-model="newComponent.developer"
          :options="libraryStore.presets.developers"
          label="Developer"
          required
        />
        <Select
          v-model="newComponent.vcs_type"
          :options="['git', 'svn']"
          label="VCS Type"
          required
        />
        <Input
          v-model="newComponent.vcs_url"
          label="VCS URL"
          placeholder="https://..."
          required
        />
        <Input
          v-model="newComponent.build_command"
          label="Build Command"
          placeholder="npm run build"
          required
        />
        <Input
          v-model="newComponent.component_url"
          label="Component URL"
          placeholder="https://..."
          required
        />
      </div>

      <template #footer>
        <Button variant="secondary" @click="showAddComponentDialog = false">
          Cancel
        </Button>
        <Button variant="success" @click="handleAddComponent">
          {{ editingComponent ? 'Update' : 'Add' }} Component
        </Button>
      </template>
    </Dialog>

    <!-- Import Modal -->
    <div
      v-if="showImportModal"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      @click.self="closeImportModal"
    >
      <div class="bg-surface-deeper border border-surface-border rounded-lg w-full max-w-2xl mx-4">
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-xl font-bold text-white">Import Projects TOML</h2>
            <button @click="closeImportModal" class="text-gray-400 hover:text-white">
              <X class="w-6 h-6" />
            </button>
          </div>

          <!-- Step 1: Upload/Paste -->
          <div v-if="importStep === 'upload'" class="space-y-4">
            <div>
              <label class="text-sm font-medium text-gray-300 block mb-2">Upload TOML file or paste content</label>
              <input
                type="file"
                accept=".toml"
                @change="handleFileUpload"
                class="w-full text-sm text-gray-400 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:bg-accent file:text-white file:cursor-pointer"
              />
            </div>

            <div>
              <label class="text-sm font-medium text-gray-300 block mb-2">Or paste TOML content</label>
              <textarea
                v-model="importContent"
                rows="10"
                class="w-full px-3 py-2 bg-surface-light border border-surface-border rounded-lg text-white text-sm font-mono focus:outline-none focus:border-accent"
                placeholder="[[projects]]
name = &quot;MyProject&quot;
build_server = &quot;192.168.1.100&quot;
deploy_server = &quot;192.168.1.200&quot;
..."
              ></textarea>
            </div>

            <div class="flex justify-end gap-3 pt-4 border-t border-surface-border">
              <Button variant="secondary" @click="closeImportModal">
                Cancel
              </Button>
              <Button
                variant="primary"
                @click="previewImport"
                :disabled="!importContent || previewing"
              >
                <Eye v-if="!previewing" class="w-4 h-4" />
                <RefreshCw v-else class="w-4 h-4 animate-spin" />
                {{ previewing ? 'Analyzing...' : 'Preview Import' }}
              </Button>
            </div>
          </div>

          <!-- Step 2: Preview & Confirm -->
          <div v-else-if="importStep === 'preview'" class="space-y-4">
            <!-- Summary -->
            <div class="grid grid-cols-3 gap-4">
              <div class="bg-surface-light p-4 rounded-lg border border-surface-border">
                <p class="text-2xl font-bold text-blue-400">{{ importPreview.total }}</p>
                <p class="text-sm text-gray-400">Total Projects</p>
              </div>
              <div class="bg-surface-light p-4 rounded-lg border border-green-500/30">
                <p class="text-2xl font-bold text-green-400">{{ importPreview.new_projects?.length || 0 }}</p>
                <p class="text-sm text-gray-400">New Projects</p>
              </div>
              <div class="bg-surface-light p-4 rounded-lg border border-yellow-500/30">
                <p class="text-2xl font-bold text-yellow-400">{{ importPreview.conflicts?.length || 0 }}</p>
                <p class="text-sm text-gray-400">Conflicts</p>
              </div>
            </div>

            <!-- New Projects List -->
            <div v-if="importPreview.new_projects?.length" class="bg-surface-light p-4 rounded-lg border border-green-500/30">
              <h4 class="text-sm font-medium text-green-400 mb-2">New Projects (will be created)</h4>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="name in importPreview.new_projects"
                  :key="name"
                  class="text-xs px-2 py-1 bg-green-500/20 text-green-300 rounded"
                >
                  {{ name }}
                </span>
              </div>
            </div>

            <!-- Conflicts List -->
            <div v-if="importPreview.conflicts?.length" class="bg-surface-light p-4 rounded-lg border border-yellow-500/30">
              <h4 class="text-sm font-medium text-yellow-400 mb-2">Conflicts (already exist)</h4>
              <div class="flex flex-wrap gap-2 mb-4">
                <span
                  v-for="name in importPreview.conflicts"
                  :key="name"
                  class="text-xs px-2 py-1 bg-yellow-500/20 text-yellow-300 rounded"
                >
                  {{ name }}
                </span>
              </div>

              <label class="text-sm font-medium text-gray-300 block mb-2">How to handle conflicts?</label>
              <select
                v-model="importMode"
                class="w-full px-3 py-2 bg-[#333333] border border-surface-border rounded-lg text-white focus:outline-none focus:border-accent"
              >
                <option value="skip">Skip - Keep existing, only add new projects</option>
                <option value="merge">Merge - Update existing project fields, merge components</option>
                <option value="overwrite">Overwrite - Replace existing projects completely</option>
              </select>
            </div>

            <div v-if="!importPreview.conflicts?.length" class="bg-surface-light p-4 rounded-lg border border-surface-border">
              <p class="text-gray-300">No conflicts detected. All projects are new and will be created.</p>
            </div>

            <div class="flex justify-between gap-3 pt-4 border-t border-surface-border">
              <Button variant="secondary" @click="importStep = 'upload'">
                <ArrowLeft class="w-4 h-4" />
                Back
              </Button>
              <div class="flex gap-3">
                <Button variant="secondary" @click="closeImportModal">
                  Cancel
                </Button>
                <Button
                  variant="success"
                  @click="executeImport"
                  :disabled="importing"
                >
                  <Upload v-if="!importing" class="w-4 h-4" />
                  <RefreshCw v-else class="w-4 h-4 animate-spin" />
                  {{ importing ? 'Importing...' : 'Import Projects' }}
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import Card from '../components/ui/Card.vue'
import Input from '../components/ui/Input.vue'
import Select from '../components/ui/Select.vue'
import Button from '../components/ui/Button.vue'
import Dialog from '../components/ui/Dialog.vue'
import { useProjectsStore, type Project, type Component } from '../stores/projects'
import { useLibraryStore } from '../stores/library'
import { useClipboard } from '../composables/useClipboard'
import { useToast } from '../composables/useToast'
import { useApi } from '../composables/useApi'
import { Plus, Trash2, Copy, CopyPlus, Save, Folder, FolderKanban, FolderPlus, Layers, Code2, Pencil, MousePointerClick, Download, Upload, X, RefreshCw, Eye, ArrowLeft } from 'lucide-vue-next'

const projectsStore = useProjectsStore()
const libraryStore = useLibraryStore()
const { copyToClipboard } = useClipboard()
const { success, error } = useToast()
const { get, post } = useApi()

const selectedProjectId = ref<number | null>(null)
const showAddProjectDialog = ref(false)
const showAddComponentDialog = ref(false)
const editingComponent = ref<Component | null>(null)

// Import/Export state
const showImportModal = ref(false)
const importStep = ref<'upload' | 'preview'>('upload')
const importContent = ref('')
const importMode = ref('skip')
const previewing = ref(false)
const importing = ref(false)
const importPreview = ref<{
  total: number
  new_projects: string[]
  conflicts: string[]
  has_conflicts: boolean
}>({
  total: 0,
  new_projects: [],
  conflicts: [],
  has_conflicts: false
})

const newProject = ref<Project>({
  name: '',
  build_server: '',
  deploy_server: '',
  database_name: '',
  environment: '',
  backup_location: '',
  components: [],
})

const newComponent = ref<Component>({
  name: '',
  developer: '',
  vcs_type: 'git',
  vcs_url: '',
  build_command: '',
  component_url: '',
})

// Local editing copy - prevents direct store mutation
const editingProject = ref<Project | null>(null)

// Watch for project selection changes and create a local copy
watch(selectedProjectId, (newId) => {
  if (newId) {
    const project = projectsStore.projects.find(p => p.id === newId)
    // Deep clone to prevent store mutation via v-model
    editingProject.value = project ? JSON.parse(JSON.stringify(project)) : null
  } else {
    editingProject.value = null
  }
}, { immediate: true })

// Computed for read-only access to store project (for components list updates)
const selectedProject = computed(() => {
  return projectsStore.projects.find(p => p.id === selectedProjectId.value) || null
})

const handleAddProject = async () => {
  if (!newProject.value.name) {
    error('Project name is required')
    return
  }

  const result = await projectsStore.createProject({
    ...newProject.value,
    components: [],
  })

  if (result) {
    success('Project created successfully!')
    showAddProjectDialog.value = false
    newProject.value = {
      name: '',
      build_server: '',
      deploy_server: '',
      database_name: '',
      environment: '',
      backup_location: '',
      components: [],
    }
    selectedProjectId.value = result.id || null
  } else {
    error('Failed to create project')
  }
}

const handleSaveProject = async () => {
  if (!editingProject.value || !editingProject.value.id) return

  const result = await projectsStore.updateProject(editingProject.value.id, {
    name: editingProject.value.name,
    build_server: editingProject.value.build_server,
    deploy_server: editingProject.value.deploy_server,
    database_name: editingProject.value.database_name,
    environment: editingProject.value.environment,
    backup_location: editingProject.value.backup_location,
  })

  if (result) {
    success('Project saved successfully!')
  } else {
    error('Failed to save project')
  }
}

const showDeleteProjectDialog = ref(false)

const handleDeleteProject = () => {
  if (!selectedProjectId.value) return
  showDeleteProjectDialog.value = true
}

const confirmDeleteProject = async () => {
  if (!selectedProjectId.value) return
  await projectsStore.deleteProject(selectedProjectId.value)
  selectedProjectId.value = null
  showDeleteProjectDialog.value = false
  success('Project deleted successfully!')
}

const handleCopyProject = async () => {
  if (!selectedProject.value) return
  const text = `Project: ${selectedProject.value.name}
Build Server: ${selectedProject.value.build_server}
Deploy Server: ${selectedProject.value.deploy_server}
Database: ${selectedProject.value.database_name}
Environment: ${selectedProject.value.environment}`

  if (await copyToClipboard(text)) {
    success('Project details copied!')
  } else {
    error('Failed to copy')
  }
}

const showDuplicateDialog = ref(false)
const duplicateName = ref('')

const handleDuplicateProject = () => {
  if (!selectedProject.value || !selectedProject.value.id) return
  duplicateName.value = `${selectedProject.value.name} (copy)`
  showDuplicateDialog.value = true
}

const confirmDuplicateProject = async () => {
  if (!selectedProject.value || !selectedProject.value.id) return
  showDuplicateDialog.value = false
  const clone = await projectsStore.duplicateProject(
    selectedProject.value.id,
    duplicateName.value.trim() || undefined
  )
  if (clone && clone.id) {
    selectedProjectId.value = clone.id
    success(`Duplicated as "${clone.name}"`)
  } else {
    error('Failed to duplicate project')
  }
}

const handleAddComponent = async () => {
  if (!selectedProject.value || !selectedProject.value.id) return

  if (editingComponent.value && editingComponent.value.id) {
    const result = await projectsStore.updateComponent(
      selectedProject.value.id,
      editingComponent.value.id,
      newComponent.value
    )
    if (result) {
      success('Component updated successfully!')
    } else {
      error('Failed to update component')
    }
  } else {
    const result = await projectsStore.addComponent(
      selectedProject.value.id,
      newComponent.value
    )
    if (result) {
      success('Component added successfully!')
    } else {
      error('Failed to add component')
    }
  }

  showAddComponentDialog.value = false
  resetComponentForm()
}

const deleteComponentId = ref<number | null>(null)

const handleDeleteComponent = (componentId: number) => {
  if (!selectedProject.value || !selectedProject.value.id) return
  deleteComponentId.value = componentId
}

const confirmDeleteComponent = async () => {
  if (!selectedProject.value?.id || deleteComponentId.value === null) return
  await projectsStore.deleteComponent(selectedProject.value.id, deleteComponentId.value)
  deleteComponentId.value = null
  success('Component deleted successfully!')
}

const editComponent = (component: Component) => {
  editingComponent.value = component
  newComponent.value = { ...component }
  showAddComponentDialog.value = true
}

const resetComponentForm = () => {
  editingComponent.value = null
  newComponent.value = {
    name: '',
    developer: '',
    vcs_type: 'git',
    vcs_url: '',
    build_command: '',
    component_url: '',
  }
}

// Import/Export functions
const exportProjects = async () => {
  try {
    const response = await get<string>('/projects/export', { responseType: 'text' })
    const blob = new Blob([response.data || ''], { type: 'application/toml' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'projects.toml'
    a.click()
    URL.revokeObjectURL(url)
    success('Projects exported!')
  } catch (err) {
    error('Failed to export projects')
  }
}

const handleFileUpload = (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (file) {
    const reader = new FileReader()
    reader.onload = (e) => {
      importContent.value = e.target?.result as string
    }
    reader.readAsText(file)
  }
}

const previewImport = async () => {
  if (!importContent.value) return

  previewing.value = true
  try {
    const response = await post<{
      total: number
      new_projects: string[]
      conflicts: string[]
      has_conflicts: boolean
    }>('/projects/import/preview', {
      content: importContent.value,
    })

    if (response.data) {
      importPreview.value = response.data
      importStep.value = 'preview'
    }
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : 'Failed to parse TOML'
    error(message)
  } finally {
    previewing.value = false
  }
}

const executeImport = async () => {
  if (!importContent.value) return

  importing.value = true
  try {
    const response = await post<{
      success: boolean
      message: string
      imported: number
      updated: number
      skipped: number
    }>('/projects/import', {
      content: importContent.value,
      mode: importMode.value,
    })

    if (response.data?.success) {
      const { imported, updated, skipped } = response.data
      success(`Import complete: ${imported} created, ${updated} updated, ${skipped} skipped`)
      await projectsStore.fetchProjects()
      closeImportModal()
    } else {
      error('Import failed')
    }
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : 'Import failed'
    error(message)
  } finally {
    importing.value = false
  }
}

const closeImportModal = () => {
  showImportModal.value = false
  importStep.value = 'upload'
  importContent.value = ''
  importMode.value = 'skip'
  importPreview.value = {
    total: 0,
    new_projects: [],
    conflicts: [],
    has_conflicts: false
  }
}

onMounted(async () => {
  if (projectsStore.projects.length === 0) {
    await projectsStore.fetchProjects()
  }
  if (libraryStore.presets.developers.length === 0) {
    await libraryStore.fetchPresets()
  }
})
</script>
