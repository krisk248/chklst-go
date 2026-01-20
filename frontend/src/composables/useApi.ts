import axios, { AxiosInstance, AxiosError } from 'axios'
import { ref } from 'vue'

interface ApiResponse<T = any> {
  data: T
  status: number
  statusText: string
}

const apiClient: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
})

export function useApi() {
  const error = ref<AxiosError | null>(null)
  const isLoading = ref(false)

  const request = async <T = any>(
    method: string,
    url: string,
    data?: any,
    config?: any
  ): Promise<ApiResponse<T>> => {
    isLoading.value = true
    error.value = null

    try {
      const response = await apiClient({
        method,
        url,
        data,
        ...config,
      })
      return {
        data: response.data,
        status: response.status,
        statusText: response.statusText,
      }
    } catch (err) {
      error.value = err as AxiosError
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const get = <T = any>(url: string, config?: any): Promise<ApiResponse<T>> => {
    return request<T>('GET', url, undefined, config)
  }

  const post = <T = any>(
    url: string,
    data?: any,
    config?: any
  ): Promise<ApiResponse<T>> => {
    return request<T>('POST', url, data, config)
  }

  const put = <T = any>(
    url: string,
    data?: any,
    config?: any
  ): Promise<ApiResponse<T>> => {
    return request<T>('PUT', url, data, config)
  }

  const patch = <T = any>(
    url: string,
    data?: any,
    config?: any
  ): Promise<ApiResponse<T>> => {
    return request<T>('PATCH', url, data, config)
  }

  const del = <T = any>(url: string, config?: any): Promise<ApiResponse<T>> => {
    return request<T>('DELETE', url, undefined, config)
  }

  return {
    error,
    isLoading,
    request,
    get,
    post,
    put,
    patch,
    delete: del,
  }
}

