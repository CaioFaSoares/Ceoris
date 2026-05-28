<script setup lang="ts">
import { useIntervalFn } from '@vueuse/core'

const { selectedGuildId } = useAppContext()
const toast = useToast()

// Configurações da Tabela
const columns = [
  { accessorKey: 'nickname', header: 'Nome / Apelido' },
  { accessorKey: 'role_name', header: 'Turma (Squad)' },
  { id: 'has_1on1', header: 'Status Canal 1-on-1' }
]

// Estado Local
const q = ref('') // Search query
const page = ref(1)
const pageCount = ref(10) // Quantidade por página
const isSyncing = ref(false)

// 1. Fetch dos Alunos (Lazy para não bloquear renderização)
const { data: students, status, refresh } = useLazyAsyncData<any[]>(
  'guild-students',
  () => selectedGuildId.value ? useApi(`/api/guilds/${selectedGuildId.value}/students`) : Promise.resolve([]), 
  { 
    watch: [selectedGuildId],
    default: () => [] 
  }
)

// 2. Lógica de Filtro (Search Client-side) e Paginação
const filteredRows = computed(() => {
  if (!students.value) return []
  
  let filtered = students.value
  
  // Filtro de Busca
  if (q.value) {
    filtered = filtered.filter((person: any) => {
      // Cria uma string combinada de valores fáceis de buscar
      const searchString = `${person.nickname || ''} ${person.role_name || ''}`.toLowerCase()
      return searchString.includes(q.value.toLowerCase())
    })
  }
  
  // Ordena pelo nome para manter organizado
  return filtered.sort((a: any, b: any) => (a.nickname || '').localeCompare(b.nickname || ''))
})

const paginatedRows = computed(() => {
  const start = (page.value - 1) * pageCount.value
  const end = start + pageCount.value
  return filteredRows.value.slice(start, end)
})

// 3. Funções de Background (Fire-and-Forget)
async function handleProvisionChannels() {
  if (!selectedGuildId.value) return
  try {
    await useApi(`/api/guilds/${selectedGuildId.value}/provision`, { method: 'POST' })
    toast.add({
      title: 'Ação Iniciada',
      description: 'Provisionamento em background iniciado!',
      color: 'blue'
    })
    triggerPolling()
  } catch (error) {}
}

async function handleHealChannels() {
  if (!selectedGuildId.value) return
  try {
    await useApi(`/api/guilds/${selectedGuildId.value}/heal`, { method: 'POST' })
    toast.add({
      title: 'Ação Iniciada',
      description: 'Processo de Auto-Healing iniciado em background.',
      color: 'blue'
    })
    triggerPolling()
  } catch (error) {}
}

async function handleSyncDiscord() {
  if (!selectedGuildId.value) return

  isSyncing.value = true
  try {
    // Aciona a Sincronização Avançada no Go Daemon (Processa alunos e gerentes)
    await useApi(`/api/guilds/${selectedGuildId.value}/sync`, {
      method: 'POST',
      body: {
        students: { active: true },
        managers: []
      }
    })
    
    toast.add({
      title: 'Sincronização em Background',
      description: 'A base de alunos e turmas começou a ser sincronizada. Isso ocorrerá silenciosamente.',
      color: 'blue'
    })
    triggerPolling()
  } catch (error) {
    // Erros já são tratados globalmente
  } finally {
    isSyncing.value = false
  }
}

// 4. Dropdown Actions e Motor de Polling
const batchActions = [
  [{
    label: 'Provisionar Canais Pendentes',
    icon: 'i-heroicons-server-stack',
    click: handleProvisionChannels
  }, {
    label: 'Corrigir Permissões (Heal)',
    icon: 'i-heroicons-wrench-screwdriver',
    click: handleHealChannels
  }]
]

const pollingCycles = ref(0)
const maxPollingCycles = 12

const { pause: stopPolling, resume: startPolling, isActive: isPollingActive } = useIntervalFn(async () => {
  pollingCycles.value++
  await refresh()
  
  // Regra 1: Parada Precoce por Sucesso
  if (students.value && students.value.every((s: any) => s.has_1on1 === true)) {
    stopPolling()
    pollingCycles.value = 0
    toast.add({
      title: 'Todos os canais prontos!',
      description: 'As ações em lote concluíram com sucesso.',
      color: 'green'
    })
    return
  }

  // Regra 2: A Guilhotina de Tempo
  if (pollingCycles.value >= maxPollingCycles) {
    stopPolling()
    pollingCycles.value = 0
    
    const hasPendings = students.value && students.value.some((s: any) => s.has_1on1 === false)
    if (hasPendings) {
      toast.add({
        title: 'Operação Finalizada',
        description: 'A operação finalizou, mas alguns alunos continuam pendentes. Verifique os logs se o erro persistir.',
        color: 'yellow'
      })
    } else {
      toast.add({
        title: 'Operação Finalizada',
        description: 'As tarefas de background parecem ter concluído.',
        color: 'green'
      })
    }
  }
}, 5000, { immediate: false })

function triggerPolling() {
  pollingCycles.value = 0
  startPolling()
}
</script>

<template>
  <UDashboardPanel grow>
    <template #header>
      <UDashboardNavbar title="Base de Alunos (Membros)">
        <template #right>
          <div class="flex gap-2 items-center">
            <span v-if="isPollingActive" class="flex items-center gap-2 text-sm text-blue-600 dark:text-blue-400 font-medium mr-2 bg-blue-50 dark:bg-blue-900/30 px-3 py-1.5 rounded-full border border-blue-200 dark:border-blue-800">
              <UIcon name="i-heroicons-arrow-path" class="animate-spin w-4 h-4" />
              Atualizando ao vivo...
            </span>
            <UDropdown :items="batchActions" :disabled="!selectedGuildId || isSyncing">
              <UButton color="white" trailing-icon="i-heroicons-chevron-down-20-solid" label="Ações em Lote" />
            </UDropdown>
            <UButton
              label="Sincronizar Discord"
              icon="i-heroicons-arrow-path"
              color="primary"
              :loading="isSyncing"
              :disabled="!selectedGuildId || isSyncing"
              @click="handleSyncDiscord"
            />
          </div>
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div class="max-w-6xl mx-auto w-full p-4 flex flex-col gap-6">
        <UAlert 
          v-if="!selectedGuildId" 
          icon="i-heroicons-exclamation-triangle" 
          title="Selecione um Servidor" 
          description="Selecione uma Guilda no menu lateral esquerdo para visualizar e sincronizar os alunos."
          color="yellow" 
        />

        <div v-else class="flex flex-col flex-1 w-full space-y-4">
          
          <div class="flex justify-between items-center bg-gray-50 dark:bg-gray-800/50 p-4 rounded-lg border border-gray-200 dark:border-gray-700">
            <UInput 
              v-model="q" 
              icon="i-heroicons-magnifying-glass-20-solid" 
              placeholder="Buscar alunos por nome ou turma..." 
              class="w-full max-w-md"
              size="md"
            />
            
            <div class="text-sm text-gray-600 dark:text-gray-300 flex items-center gap-2">
              <UIcon name="i-heroicons-users" class="w-5 h-5 text-gray-400" />
              Total Sincronizado: <span class="font-bold text-gray-900 dark:text-white">{{ filteredRows.length }}</span>
            </div>
          </div>

          <UCard :ui="{ body: { padding: '' } }" class="flex-1">
            <UTable
              :data="paginatedRows"
              :columns="columns"
              :loading="status === 'pending'"
            >
              <template #nickname-cell="{ row }">
                <div class="flex items-center gap-3">
                  <UAvatar :alt="row.original.nickname || row.original.username || 'A'" size="sm" />
                  <span class="font-medium text-gray-900 dark:text-white">{{ row.original.nickname || row.original.username || 'Sem Nome' }}</span>
                </div>
              </template>

              <template #role_name-cell="{ row }">
                <UBadge color="gray" variant="subtle" class="font-medium truncate max-w-[200px]">{{ row.original.role_name || 'Sem Turma' }}</UBadge>
              </template>

              <template #has_1on1-cell="{ row }">
                <div class="flex items-center gap-2">
                  <span class="relative flex h-2.5 w-2.5">
                    <span 
                      class="absolute inline-flex h-full w-full rounded-full opacity-75"
                      :class="row.original.has_1on1 ? 'bg-emerald-400 animate-ping' : 'bg-rose-400'"
                    ></span>
                    <span 
                      class="relative inline-flex rounded-full h-2.5 w-2.5"
                      :class="row.original.has_1on1 ? 'bg-emerald-500' : 'bg-rose-500'"
                    ></span>
                  </span>
                  <span class="text-sm font-medium" :class="row.original.has_1on1 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'">
                    {{ row.original.has_1on1 ? 'Pronto' : 'Pendente' }}
                  </span>
                </div>
              </template>
            </UTable>

            <div class="flex justify-end px-4 py-3 border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/20">
              <UPagination
                v-model="page"
                :page-count="pageCount"
                :total="filteredRows.length"
              />
            </div>
          </UCard>
        </div>
      </div>
    </template>
  </UDashboardPanel>
</template>
