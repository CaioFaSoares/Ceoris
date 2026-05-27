<script setup lang="ts">
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
  () => selectedGuildId.value ? useApi(`/api/config/guilds/${selectedGuildId.value}/students`) : Promise.resolve([]), 
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

// 3. Mutação: Disparar Sincronização Completa
async function handleSyncDiscord() {
  if (!selectedGuildId.value) return

  isSyncing.value = true
  try {
    // Aciona a Sincronização Avançada no Go Daemon (Processa alunos e gerentes)
    await useApi(`/api/sync/guilds/${selectedGuildId.value}/advanced`, {
      method: 'POST',
      body: {
        students: { active: true },
        managers: []
      }
    })
    
    toast.add({
      title: 'Sincronização Concluída',
      description: 'A base de alunos e turmas foi mapeada com sucesso.',
      color: 'green'
    })
    
    // Recarrega a tabela automaticamente para mostrar os novos alunos
    await refresh()
  } catch (error) {
    // Erros já são tratados globalmente
  } finally {
    isSyncing.value = false
  }
}
</script>

<template>
  <UDashboardPanel grow>
    <template #header>
      <UDashboardNavbar title="Base de Alunos (Membros)">
        <template #right>
          <UButton
            label="Sincronizar Discord"
            icon="i-heroicons-arrow-path"
            color="primary"
            :loading="isSyncing"
            :disabled="!selectedGuildId || isSyncing"
            @click="handleSyncDiscord"
          />
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
                      :class="row.original.channel_id ? 'bg-emerald-400 animate-ping' : 'bg-amber-400'"
                    ></span>
                    <span 
                      class="relative inline-flex rounded-full h-2.5 w-2.5"
                      :class="row.original.channel_id ? 'bg-emerald-500' : 'bg-amber-500'"
                    ></span>
                  </span>
                  <span class="text-sm font-medium" :class="row.original.channel_id ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400'">
                    {{ row.original.channel_id ? 'Pronto' : 'Pendente' }}
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
