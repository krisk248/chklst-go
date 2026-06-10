<template>
  <div class="space-y-6">
    <div class="page-header">
      <div>
        <h1 class="page-title">Library</h1>
        <p class="text-gray-400 mt-2">Manage library presets and defaults</p>
      </div>
      <div class="flex items-center gap-3">
        <Button variant="secondary" @click="exportLibrary">
          <Download class="w-4 h-4" />
          Export TOML
        </Button>
        <Button variant="primary" @click="showImportModal = true">
          <Upload class="w-4 h-4" />
          Import TOML
        </Button>
      </div>
    </div>

    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <Card title="Developers">
        <div class="text-center text-3xl font-bold text-accent">
          {{ libraryStore.presets.developers.length }}
        </div>
      </Card>
      <Card title="Build Servers">
        <div class="text-center text-3xl font-bold text-blue-400">
          {{ libraryStore.presets.build_servers.length }}
        </div>
      </Card>
      <Card title="Deploy Servers">
        <div class="text-center text-3xl font-bold text-green-400">
          {{ libraryStore.presets.deploy_servers.length }}
        </div>
      </Card>
      <Card title="Environments">
        <div class="text-center text-3xl font-bold text-yellow-400">
          {{ libraryStore.presets.environments.length }}
        </div>
      </Card>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Developers -->
      <Card title="Developers">
        <div class="space-y-2">
          <div
            v-for="(dev, idx) in libraryStore.presets.developers"
            :key="`dev-${idx}`"
            class="flex items-center justify-between bg-surface-light p-3 rounded-lg border border-surface-border"
          >
            <span class="text-white">{{ dev }}</span>
            <Button
              variant="danger"
              size="sm"
              @click="libraryStore.removeDeveloper(dev)"
            >
              <Trash2 class="w-4 h-4" />
            </Button>
          </div>
          <div v-if="libraryStore.presets.developers.length === 0" class="text-center py-8 text-gray-400">
            <p>No developers in library</p>
          </div>
        </div>
        <div class="mt-4 pt-4 border-t border-surface-border">
          <form @submit.prevent="addDeveloper" class="flex gap-2">
            <Input
              v-model="newDeveloper"
              placeholder="Add developer"
              size="sm"
            />
            <Button type="submit" variant="success">
              <Plus class="w-4 h-4" />
            </Button>
          </form>
        </div>
      </Card>

      <!-- Build Servers -->
      <Card title="Build Servers">
        <div class="space-y-2">
          <div
            v-for="(server, idx) in libraryStore.presets.build_servers"
            :key="`build-${idx}`"
            class="flex items-center justify-between bg-surface-light p-3 rounded-lg border border-surface-border"
          >
            <span class="text-white">{{ server }}</span>
            <Button
              variant="danger"
              size="sm"
              @click="libraryStore.removeBuildServer(server)"
            >
              <Trash2 class="w-4 h-4" />
            </Button>
          </div>
          <div v-if="libraryStore.presets.build_servers.length === 0" class="text-center py-8 text-gray-400">
            <p>No build servers in library</p>
          </div>
        </div>
        <div class="mt-4 pt-4 border-t border-surface-border">
          <form @submit.prevent="addBuildServer" class="flex gap-2">
            <Input
              v-model="newBuildServer"
              placeholder="Add build server"
              size="sm"
            />
            <Button type="submit" variant="success">
              <Plus class="w-4 h-4" />
            </Button>
          </form>
        </div>
      </Card>

      <!-- Deploy Servers -->
      <Card title="Deploy Servers">
        <div class="space-y-2">
          <div
            v-for="(server, idx) in libraryStore.presets.deploy_servers"
            :key="`deploy-${idx}`"
            class="flex items-center justify-between bg-surface-light p-3 rounded-lg border border-surface-border"
          >
            <span class="text-white">{{ server }}</span>
            <Button
              variant="danger"
              size="sm"
              @click="libraryStore.removeDeployServer(server)"
            >
              <Trash2 class="w-4 h-4" />
            </Button>
          </div>
          <div v-if="libraryStore.presets.deploy_servers.length === 0" class="text-center py-8 text-gray-400">
            <p>No deploy servers in library</p>
          </div>
        </div>
        <div class="mt-4 pt-4 border-t border-surface-border">
          <form @submit.prevent="addDeployServer" class="flex gap-2">
            <Input
              v-model="newDeployServer"
              placeholder="Add deploy server"
              size="sm"
            />
            <Button type="submit" variant="success">
              <Plus class="w-4 h-4" />
            </Button>
          </form>
        </div>
      </Card>

      <!-- Environments -->
      <Card title="Environments">
        <div class="space-y-2">
          <div
            v-for="(env, idx) in libraryStore.presets.environments"
            :key="`env-${idx}`"
            class="flex items-center justify-between bg-surface-light p-3 rounded-lg border border-surface-border"
          >
            <span class="text-white">{{ env }}</span>
            <Button
              variant="danger"
              size="sm"
              @click="libraryStore.removeEnvironment(env)"
            >
              <Trash2 class="w-4 h-4" />
            </Button>
          </div>
          <div v-if="libraryStore.presets.environments.length === 0" class="text-center py-8 text-gray-400">
            <p>No environments in library</p>
          </div>
        </div>
        <div class="mt-4 pt-4 border-t border-surface-border">
          <form @submit.prevent="addEnvironment" class="flex gap-2">
            <Input
              v-model="newEnvironment"
              placeholder="Add environment"
              size="sm"
            />
            <Button type="submit" variant="success">
              <Plus class="w-4 h-4" />
            </Button>
          </form>
        </div>
      </Card>
    </div>

    <!-- Import Modal -->
    <div
      v-if="showImportModal"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      @click.self="showImportModal = false"
    >
      <div class="bg-surface-deeper border border-surface-border rounded-lg w-full max-w-lg mx-4">
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-xl font-bold text-white">Import Library TOML</h2>
            <button @click="showImportModal = false" class="text-gray-400 hover:text-white">
              <X class="w-6 h-6" />
            </button>
          </div>

          <div class="space-y-4">
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
                rows="8"
                class="w-full px-3 py-2 bg-surface-light border border-surface-border rounded-lg text-white text-sm font-mono focus:outline-none focus:border-accent"
                placeholder="[library]
developers = [&quot;Dev1&quot;, &quot;Dev2&quot;]
build_servers = [&quot;192.168.1.149&quot;]
..."
              ></textarea>
            </div>

            <div>
              <label class="text-sm font-medium text-gray-300 block mb-2">Import Mode</label>
              <select
                v-model="importMode"
                class="w-full px-3 py-2 bg-surface-light border border-surface-border rounded-lg text-white focus:outline-none focus:border-accent"
              >
                <option value="merge">Merge (add new, keep existing)</option>
                <option value="overwrite">Overwrite (replace all)</option>
              </select>
            </div>
          </div>

          <div class="flex justify-end gap-3 mt-6 pt-4 border-t border-surface-border">
            <Button variant="secondary" @click="showImportModal = false">
              Cancel
            </Button>
            <Button
              variant="success"
              @click="importLibrary"
              :disabled="!importContent || importing"
            >
              <Upload v-if="!importing" class="w-4 h-4" />
              <RefreshCw v-else class="w-4 h-4 animate-spin" />
              {{ importing ? 'Importing...' : 'Import' }}
            </Button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Card from '../components/ui/Card.vue'
import Input from '../components/ui/Input.vue'
import Button from '../components/ui/Button.vue'
import { useLibraryStore } from '../stores/library'
import { useApi } from '../composables/useApi'
import { useToast } from '../composables/useToast'
import { Trash2, Plus, Download, Upload, X, RefreshCw } from 'lucide-vue-next'

const libraryStore = useLibraryStore()
const { get, post } = useApi()
const { success, error } = useToast()

const newDeveloper = ref('')
const newBuildServer = ref('')
const newDeployServer = ref('')
const newEnvironment = ref('')

// Import state
const showImportModal = ref(false)
const importContent = ref('')
const importMode = ref('merge')
const importing = ref(false)

const addDeveloper = async () => {
  if (newDeveloper.value.trim()) {
    await libraryStore.addDeveloper(newDeveloper.value.trim())
    success('Developer added!')
    newDeveloper.value = ''
  }
}

const addBuildServer = async () => {
  if (newBuildServer.value.trim()) {
    await libraryStore.addBuildServer(newBuildServer.value.trim())
    success('Build server added!')
    newBuildServer.value = ''
  }
}

const addDeployServer = async () => {
  if (newDeployServer.value.trim()) {
    await libraryStore.addDeployServer(newDeployServer.value.trim())
    success('Deploy server added!')
    newDeployServer.value = ''
  }
}

const addEnvironment = async () => {
  if (newEnvironment.value.trim()) {
    await libraryStore.addEnvironment(newEnvironment.value.trim())
    success('Environment added!')
    newEnvironment.value = ''
  }
}

const exportLibrary = async () => {
  try {
    const response = await get<string>('/library/export', { responseType: 'text' })
    const blob = new Blob([response.data || ''], { type: 'application/toml' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'library.toml'
    a.click()
    URL.revokeObjectURL(url)
    success('Library exported!')
  } catch (err) {
    error('Failed to export library')
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

const importLibrary = async () => {
  if (!importContent.value) return

  importing.value = true
  try {
    const response = await post<{ success: boolean; message: string }>('/library/import', {
      content: importContent.value,
      mode: importMode.value,
    })

    if (response.data?.success) {
      success(response.data.message || 'Library imported!')
      await libraryStore.fetchPresets()
      showImportModal.value = false
      importContent.value = ''
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

onMounted(async () => {
  if (libraryStore.presets.developers.length === 0) {
    await libraryStore.fetchPresets()
  }
})
</script>
