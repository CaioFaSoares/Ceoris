import type { FetchOptions } from 'ofetch'

/**
 * Cliente HTTP unificado do Ceoris.
 * Wrapper em volta do $fetch que trata erros globalmente e lida
 * com o roteamento interno (Nitro SSR) vs externo (Client).
 */
export const useApi = async <T>(
  request: string,
  opts?: FetchOptions<'json'>
): Promise<T> => {
  const config = useRuntimeConfig()
  const toast = useToast()

  // Nitro server should use config.apiBase (internal docker network)
  // Browser should use config.public.apiBase (external localhost network)
  const baseURL = import.meta.server ? config.apiBase : config.public.apiBase

  try {
    const response = await $fetch<{ data?: T, error?: string } | T>(request, {
      baseURL: baseURL as string,
      ...opts,
      
      onResponseError({ response }) {
        if (import.meta.client) {
          const errorMessage = response._data?.error || response._data?.message || 'Falha ao comunicar com o Go Daemon.'
          
          toast.add({
            id: `api_error_${response.status}`,
            title: 'Erro de Comunicação',
            description: errorMessage,
            color: 'error',
            icon: 'i-lucide-alert-circle'
          })
        }
      }
    })

    // O DESEMPACOTADOR (The Unwrapper)
    if (response && typeof response === 'object' && 'data' in response) {
      return response.data as T
    }

    return response as T
  } catch (error) {
    throw error
  }
}
