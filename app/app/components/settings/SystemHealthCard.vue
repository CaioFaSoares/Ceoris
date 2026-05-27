<script setup lang="ts">
import { useIntervalFn } from '@vueuse/core'

interface SystemHealth {
  services: {
    go_daemon: string
    discord_ws: string
    pocketbase: string
  }
  status: string
}

// 1. Fetch Inicial via useAsyncData
const { data: health, status, refresh } = await useAsyncData<SystemHealth>(
  'system-health',
  () => useApi('/api/system/health'),
  {
    // Fallback gracioso caso a API esteja fora do ar logo no primeiro acesso
    default: () => ({
      status: 'offline',
      services: { go_daemon: 'offline', discord_ws: 'offline', pocketbase: 'offline' }
    })
  }
)

// 2. Polling Assíncrono (Atualiza a cada 15 segundos em background)
const { pause, resume } = useIntervalFn(async () => {
  try {
    await refresh()
  } catch (error) {
    // Se o catch disparar, o servidor Go caiu totalmente ou a rede falhou.
    if (health.value) {
      health.value = {
        status: 'offline',
        services: { go_daemon: 'offline', discord_ws: 'offline', pocketbase: 'offline' }
      }
    }
  }
}, 15000, { immediateCallback: true })

// Desliga o polling automaticamente se o componente for destruído (mudança de página)
onUnmounted(() => pause())
</script>

<template>
  <UCard :ui="{ body: { padding: 'p-4 sm:p-6' } }">
    <template #header>
      <div class="flex items-center justify-between">
        <h3 class="text-base font-semibold leading-6 text-gray-900 dark:text-white">
          Monitor de Integridade do Sistema
        </h3>
        <div class="flex items-center gap-3">
          <UButton 
            color="neutral" 
            variant="ghost" 
            icon="i-heroicons-arrow-path" 
            size="xs"
            :loading="status === 'pending'"
            @click="refresh()"
          >
            Testar
          </UButton>
          <UIcon name="i-heroicons-cpu-chip" class="w-5 h-5 text-gray-500" />
        </div>
      </div>
    </template>

    <template #default>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        
        <div class="flex items-center justify-between p-4 rounded-lg bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700">
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">Ceoris Core Engine</p>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">Backend e PocketBase (Porta 12000)</p>
          </div>
          
          <div class="flex items-center gap-3">
            <span class="relative flex h-3 w-3" v-if="health?.services?.go_daemon === 'healthy'">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-3 w-3 bg-emerald-500"></span>
            </span>
            <UBadge 
              :color="health?.services?.go_daemon === 'healthy' ? 'green' : 'red'" 
              variant="subtle"
            >
              {{ (health?.services?.go_daemon || 'offline').toUpperCase() }}
            </UBadge>
          </div>
        </div>

        <div class="flex items-center justify-between p-4 rounded-lg bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700">
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">Discord Gateway Connection</p>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              Sessão WebSocket do Bot 
            </p>
          </div>
          
          <div class="flex items-center gap-3">
            <span class="relative flex h-3 w-3" v-if="health?.services?.discord_ws === 'connected'">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-3 w-3 bg-emerald-500"></span>
            </span>
            <UBadge 
              :color="health?.services?.discord_ws === 'connected' ? 'green' : 'red'" 
              variant="subtle"
            >
              {{ (health?.services?.discord_ws || 'offline').toUpperCase() }}
            </UBadge>
          </div>
        </div>

      </div>
    </template>
  </UCard>
</template>
