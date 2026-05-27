<script setup lang="ts">
import { computed } from 'vue'

interface SystemHealthPayload {
  env: {
    discord_client_id: string
  }
}

// Fetch lazy com chave própria para não conflitar com SystemHealthCard
const { data: healthPayload, status } = useLazyAsyncData<any>(
  'discord-wizard-config',
  () => useApi('/api/system/health')
)

// 2. Validação reativa
const hasClientId = computed(() => !!healthPayload.value?.env?.discord_client_id)

// 3. Montagem dinâmica da URL do Discord baseada no ID retornado pelo Go
const oauthUrl = computed(() => {
  const clientId = healthPayload.value?.env?.discord_client_id
  if (!clientId) return '#'
  // Mesma URL de permissões e escopos utilizada no Streamlit original
  return `https://discord.com/oauth2/authorize?client_id=${clientId}&permissions=8&integration_type=0&scope=bot+applications.commands`
})
</script>

<template>
  <UCard :ui="{ body: { padding: 'p-4 sm:p-6' } }">
    <template #header>
      <div class="flex items-center gap-2">
        <UIcon name="i-simple-icons-discord" class="w-6 h-6 text-[#5865F2]" />
        <h3 class="text-base font-semibold leading-6 text-gray-900 dark:text-white">
          Integração com o Servidor Discord
        </h3>
      </div>
    </template>

    <template #default>
      <div class="space-y-8">
        <p class="text-sm text-gray-600 dark:text-gray-300">
          Siga o assistente rápido para habilitar as funcionalidades de gerenciamento do Ceoris no seu servidor.
        </p>

        <div class="flex gap-4">
          <div class="flex-shrink-0 flex items-center justify-center w-8 h-8 rounded-full bg-blue-100 text-blue-600 dark:bg-blue-900 dark:text-blue-300 font-bold">
            1
          </div>
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white">Convite de Instalação</h4>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400 mb-4">
              Autorize o Bot a entrar no seu servidor com permissões administrativas completas.
            </p>
            
            <UButton
              :to="hasClientId ? oauthUrl : undefined"
              target="_blank"
              :disabled="!hasClientId || status === 'pending'"
              :loading="status === 'pending'"
              icon="i-simple-icons-discord"
              label="Convidar Bot para o Discord"
              style="background-color: #5865F2; color: white;"
              size="lg"
            />
            
            <p v-if="!hasClientId && status !== 'pending'" class="mt-2 text-xs text-red-500 flex items-center gap-1">
              <UIcon name="i-heroicons-exclamation-triangle" />
              O 'discord_client_id' não foi encontrado no servidor Go. Verifique seu arquivo .env.
            </p>
          </div>
        </div>

        <div class="flex gap-4">
          <div class="flex-shrink-0 flex items-center justify-center w-8 h-8 rounded-full bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400 font-bold">
            2
          </div>
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white">Mapeamento de Cargos</h4>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400 mb-4">
              Classifique os cargos do seu servidor entre Turmas, Mentores e Trilhas para que o Ceoris saiba quem é quem.
            </p>
            <UButton
              to="/settings/roles"
              icon="i-heroicons-tag"
              label="Taxonomia"
              color="gray"
              variant="solid"
            />
          </div>
        </div>

        <div class="flex gap-4">
          <div class="flex-shrink-0 flex items-center justify-center w-8 h-8 rounded-full bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400 font-bold">
            3
          </div>
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white">Regras de Turno</h4>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400 mb-4">
              Configure os horários de check-in para cada uma das Turmas mapeadas.
            </p>
            <UButton
              to="/settings/squads"
              icon="i-heroicons-cog-8-tooth"
              label="Configurar Turmas"
              color="gray"
              variant="solid"
            />
          </div>
        </div>

        <div class="flex gap-4">
          <div class="flex-shrink-0 flex items-center justify-center w-8 h-8 rounded-full bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400 font-bold">
            4
          </div>
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white">Inicialização de Dados</h4>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400 mb-4">
              Finalize realizando a primeira sincronização para espelhar os membros do Discord no painel.
            </p>
            <UButton
              to="/settings/members"
              icon="i-heroicons-users"
              label="Sincronizar Alunos"
              color="gray"
              variant="solid"
            />
          </div>
        </div>

      </div>
    </template>
  </UCard>
</template>
