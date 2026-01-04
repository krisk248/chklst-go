import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useApi } from '../composables/useApi'

export interface Deployment {
  id?: number
  jira_id?: string  // Optional - can be empty
  timestamp: string
  project_id: number
  component_id?: number  // Nullable for legacy data
  environment?: string
  vcs_url?: string
  developer_name?: string
  build_server?: string
  deploy_server?: string
  database_name?: string
  db_backup_location?: string
  database_script?: string
  previous_build_backup?: string
  build_status: string
  deploy_status: string
  notes?: string
  deployed_by?: string
  created_at?: string
  updated_at?: string
}

export const useDeploymentsStore = defineStore('deployments', () => {
  const deployments = ref<Deployment[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const { get, post, put, delete: deleteApi } = useApi()

  const fetchDeployments = async () => {
    isLoading.value = true
    error.value = null
    try {
      const response = await get<Deployment[]>('/deployments')
      deployments.value = response.data || []
    } catch (err) {
      error.value = 'Failed to fetch deployments'
      console.error(err)
    } finally {
      isLoading.value = false
    }
  }

  const createDeployment = async (deployment: Deployment): Promise<Deployment | null> => {
    isLoading.value = true
    error.value = null
    try {
      const response = await post<Deployment>('/deployments', deployment)
      const newDeployment = response.data
      if (newDeployment.id) {
        deployments.value.push(newDeployment)
      }
      return newDeployment
    } catch (err) {
      error.value = 'Failed to create deployment'
      console.error(err)
      return null
    } finally {
      isLoading.value = false
    }
  }

  const updateDeployment = async (id: number, deployment: Partial<Deployment>) => {
    isLoading.value = true
    error.value = null
    try {
      const response = await put<Deployment>(`/deployments/${id}`, deployment)
      const idx = deployments.value.findIndex(d => d.id === id)
      if (idx > -1) {
        deployments.value[idx] = response.data
      }
      return response.data
    } catch (err) {
      error.value = 'Failed to update deployment'
      console.error(err)
      return null
    } finally {
      isLoading.value = false
    }
  }

  const deleteDeployment = async (id: number) => {
    isLoading.value = true
    error.value = null
    try {
      await deleteApi(`/deployments/${id}`)
      deployments.value = deployments.value.filter(d => d.id !== id)
    } catch (err) {
      error.value = 'Failed to delete deployment'
      console.error(err)
    } finally {
      isLoading.value = false
    }
  }

  const recentDeployments = computed(() => {
    return deployments.value
      .sort((a, b) => new Date(b.created_at || '').getTime() - new Date(a.created_at || '').getTime())
      .slice(0, 10)
  })

  const getDeploymentsByProject = (projectId: number) => {
    return deployments.value.filter(d => d.project_id === projectId)
  }

  const getDeploymentsByMonth = (month: number, year: number) => {
    return deployments.value.filter(d => {
      const date = new Date(d.timestamp)
      return date.getMonth() === month - 1 && date.getFullYear() === year
    })
  }

  // Store last deployment for "Repeat Last" feature
  const lastDeployment = ref<{ project_id: number; component_id: number } | null>(null)

  const setLastDeployment = (projectId: number, componentId: number) => {
    lastDeployment.value = { project_id: projectId, component_id: componentId }
  }

  return {
    deployments,
    isLoading,
    error,
    lastDeployment,
    fetchDeployments,
    createDeployment,
    updateDeployment,
    deleteDeployment,
    recentDeployments,
    getDeploymentsByProject,
    getDeploymentsByMonth,
    setLastDeployment,
  }
})
