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
      <Button variant="success" @click="showAddProjectDialog = true" class="shadow-lg shadow-green-500/20">
        <Plus class="w-4 h-4" />
        Add Project
      </Button>
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
                  : 'bg-[#404040] border-[#555555] text-gray-300 hover:border-blue-400 hover:bg-[#4a4a4a]',
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
                    : 'bg-[#555555] text-gray-400'
                ]">
                  {{ project.components?.length || 0 }} comp
                </span>
              </div>
            </button>
          </div>
          <!-- Enhanced Empty State -->
          <div v-else class="text-center py-12">
            <div class="p-4 bg-[#404040] rounded-full w-16 h-16 mx-auto mb-4 flex items-center justify-center">
              <FolderPlus class="w-8 h-8 text-gray-500" />
            </div>
            <p class="text-gray-400 mb-2">No projects yet</p>
            <p class="text-gray-500 text-sm mb-4">Create your first project to get started</p>
            <Button variant="success" size="sm" @click="showAddProjectDialog = true">
              <Plus class="w-4 h-4" />
              Add Project
            </Button>
          </div>
          <div v-if="projectsStore.projects.length > 0" class="flex gap-2 mt-4 pt-4 border-t border-[#555555]">
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
            <div class="border-t border-[#555555] pt-6 mt-6">
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
                  class="flex items-center justify-between bg-gradient-to-r from-[#404040] to-[#3a3a3a] p-4 rounded-lg border border-[#555555] hover:border-purple-400/50 transition-all group"
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
                        <span class="text-xs px-1.5 py-0.5 bg-[#555555] text-gray-400 rounded">{{ component.vcs_type }}</span>
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
              <div v-else class="text-center py-10 bg-[#404040]/50 rounded-lg border border-dashed border-[#555555]">
                <div class="p-3 bg-[#555555] rounded-full w-12 h-12 mx-auto mb-3 flex items-center justify-center">
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

            <div class="flex gap-3 justify-end pt-4 border-t border-[#555555]">
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
            <div class="p-4 bg-[#404040] rounded-full w-20 h-20 mx-auto mb-4 flex items-center justify-center">
              <MousePointerClick class="w-10 h-10 text-gray-500" />
            </div>
            <p class="text-gray-300 text-lg mb-2">No project selected</p>
            <p class="text-gray-500 text-sm">Select a project from the list to view and edit its details</p>
          </div>
        </Card>
      </div>
    </div>

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
import { Plus, Trash2, Copy, Save, Folder, FolderKanban, FolderPlus, Layers, Code2, Pencil, MousePointerClick } from 'lucide-vue-next'

const projectsStore = useProjectsStore()
const libraryStore = useLibraryStore()
const { copyToClipboard } = useClipboard()
const { success, error } = useToast()

const selectedProjectId = ref<number | null>(null)
const showAddProjectDialog = ref(false)
const showAddComponentDialog = ref(false)
const editingComponent = ref<Component | null>(null)

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

const handleDeleteProject = async () => {
  if (!selectedProjectId.value) return
  if (confirm('Are you sure you want to delete this project?')) {
    await projectsStore.deleteProject(selectedProjectId.value)
    selectedProjectId.value = null
    success('Project deleted successfully!')
  }
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

const handleDeleteComponent = async (componentId: number) => {
  if (!selectedProject.value || !selectedProject.value.id) return
  if (confirm('Are you sure you want to delete this component?')) {
    await projectsStore.deleteComponent(selectedProject.value.id, componentId)
    success('Component deleted successfully!')
  }
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

onMounted(async () => {
  if (projectsStore.projects.length === 0) {
    await projectsStore.fetchProjects()
  }
  if (libraryStore.presets.developers.length === 0) {
    await libraryStore.fetchPresets()
  }
})
</script>
