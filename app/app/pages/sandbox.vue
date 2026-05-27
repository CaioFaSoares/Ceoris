<script setup lang="ts">
import { useIntervalFn } from '@vueuse/core'

useSeoMeta({
  title: 'Sandbox Reactividade - Ceoris'
})

// O useAsyncData fará o Isomorphic Fetch na primeira carga e será o nosso cache.
const { data: serverHealth, status, error, refresh } = await useAsyncData(
  'sandbox-health',
  () => useApi('/api/system/health')
)

// Inicia um loop que chama o refresh() a cada 5 segundos
const { pause, resume, isActive } = useIntervalFn(() => {
  refresh()
}, 5000)

async function triggerError() {
  try {
    await useApi('/api/rota-quebrada')
  } catch (err) {
    console.log('Erro capturado localmente, mas o Toast já deve ter aparecido globalmente.')
  }
}

// Testando o Contexto Global (CEORIS-012)
const { selectedGuildId, setGuildId, clearGuildId } = useAppContext()

// Podemos expor o objeto de data de forma visual para a interface.
</script>

<template>
  <UDashboardPage>
    <UDashboardPanel grow>
      <UDashboardNavbar title="Sandbox: Polling e Fetching">
        <template #right>
          <div class="flex items-center gap-2">
            <UButton color="error" variant="outline" icon="i-lucide-alert-triangle" @click="triggerError">
              Forçar Erro 404
            </UButton>
            <UButton :color="isActive ? 'error' : 'primary'" @click="isActive ? pause() : resume()">
              {{ isActive ? 'Pausar Polling' : 'Retomar Polling' }}
            </UButton>
          </div>
        </template>
      </UDashboardNavbar>

      <UDashboardPanelContent>
        <div class="flex flex-col gap-4 max-w-2xl mx-auto py-8">
          <UCard>
            <template #header>
              <h2 class="font-semibold text-lg">Testando Padrões do CEORIS-003</h2>
              <p class="text-sm text-gray-500">
                Abaixo está o resultado reativo de um endpoint do Go Server.
              </p>
            </template>
            
            <div v-if="status === 'pending'" class="text-yellow-500 flex items-center gap-2">
              <UIcon name="i-lucide-loader-2" class="animate-spin" /> Carregando...
            </div>
            
            <div v-else-if="status === 'error'" class="text-red-500">
              Erro ao acessar o Go Daemon: {{ error?.message }}
            </div>
            
            <div v-else>
              <div class="mb-2">
                <strong>Status do Polling:</strong> 
                <UBadge :color="isActive ? 'green' : 'orange'">
                  {{ isActive ? 'Ativo (a cada 5s)' : 'Pausado' }}
                </UBadge>
              </div>
              <pre class="bg-gray-100 dark:bg-gray-800 p-4 rounded-lg overflow-auto text-sm">{{ serverHealth }}</pre>
            </div>
            
            <template #footer>
              <div class="text-xs text-gray-400">
                Abra a aba Network (F12) e veja as requisições acontecendo de forma silenciosa.
              </div>
            </template>
          </UCard>

          <UCard>
            <template #header>
              <h2 class="font-semibold text-lg">Testando Contexto Global (CEORIS-012)</h2>
              <p class="text-sm text-gray-500">
                Estado compartilhado da Guilda selecionada usando `useCookie`.
              </p>
            </template>
            
            <div class="mb-4">
              <strong>Guilda Atual:</strong> 
              <span class="ml-2 font-mono" :class="selectedGuildId ? 'text-primary' : 'text-gray-500'">
                {{ selectedGuildId || 'Nenhuma selecionada' }}
              </span>
            </div>

            <div class="flex items-center gap-2">
              <UButton color="primary" @click="setGuildId('123456789')">
                Setar Guilda (123456789)
              </UButton>
              <UButton color="white" variant="solid" @click="clearGuildId">
                Limpar Guilda
              </UButton>
            </div>
            
            <template #footer>
              <div class="text-xs text-gray-400">
                Navegue para outra página e volte, ou dê F5 (Reload) para testar a persistência do Cookie!
              </div>
            </template>
          </UCard>
        </div>
      </UDashboardPanelContent>
    </UDashboardPanel>
  </UDashboardPage>
</template>
