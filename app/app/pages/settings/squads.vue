<script setup lang="ts">
useSeoMeta({
  title: 'Gestão de Turmas e Turnos - Ceoris'
})

const { selectedGuildId } = useAppContext()
const toast = useToast()

// Definição das colunas da tabela
const columns = [
  { accessorKey: 'name', header: 'Cargo / Turma' },
  { accessorKey: 'shift', header: 'Turno' },
  { accessorKey: 'check_in_time', header: 'Horário Check-in' },
  { id: 'actions', header: '' }
]

// Busca das Regras já Salvas (Puxa do Go Daemon / PocketBase)
const { data: allRoles, status, refresh } = useLazyAsyncData<any[]>(
  'configured-roles',
  () => selectedGuildId.value ? useApi(`/api/guilds/${selectedGuildId.value}/squads`) : Promise.resolve([]),
  { watch: [selectedGuildId], default: () => [] }
)

// Filtra apenas os cargos que estão ativamente configurados/monitorados
const configuredRoles = computed(() => {
  if (!allRoles.value) return []
  return allRoles.value.filter(r => r.is_monitored === true || (r.shift && r.shift !== ''))
})

// Controle do Slideover
const isSlideoverOpen = ref(false)

// Função disparada quando o Slideover salva algo com sucesso
const handleRoleSaved = async () => {
  await refresh() // Atualiza a tabela na hora sem F5
}

// Controle de Edição
const editingRole = ref<any>(null)

const openAddRole = () => {
  editingRole.value = null
  isSlideoverOpen.value = true
}

const editRoleConfig = (role: any) => {
  editingRole.value = role
  isSlideoverOpen.value = true
}

// Função para remover monitoramento de uma turma
const isDeleting = ref<string | null>(null)
const removeRoleConfig = async (role: any) => {
  isDeleting.value = role.id
  try {
    await useApi(`/api/guilds/${selectedGuildId.value}/squads/${role.id}`, {
      method: 'PATCH',
      body: {
        is_monitored: false,
        shift: "", // Limpa o turno
        check_in_time: "" // Limpa o horário
      }
    })
    toast.add({ title: 'Configuração removida com sucesso!', color: 'gray' })
    await refresh()
  } catch (err) {
    // Tratado no wrapper
  } finally {
    isDeleting.value = null
  }
}
</script>

<template>
  <UDashboardPanel grow>
    <template #header>
      <UDashboardNavbar title="Configuração de Turmas e Turnos">
        <template #right>
            <UButton
              label="Adicionar Turma"
              icon="i-heroicons-plus"
              color="primary"
              @click="openAddRole"
              :disabled="!selectedGuildId"
            />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div class="max-w-5xl mx-auto w-full p-4 flex flex-col gap-6">
        <UAlert 
          v-if="!selectedGuildId" 
          icon="i-heroicons-exclamation-triangle" 
          title="Selecione um Servidor" 
          description="Você precisa selecionar um servidor do Discord no menu lateral esquerdo antes de configurar as turmas."
          color="yellow" 
        />

        <UCard v-else class="flex-1" :ui="{ body: { padding: '' } }">
          <template #header>
            <div class="flex items-center justify-between">
              <div>
                <h3 class="text-base font-semibold leading-6 text-gray-900 dark:text-white">Turmas Monitoradas</h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  Cargos do servidor que o Ceoris está acompanhando para controle de ponto diário.
                </p>
              </div>
            </div>
          </template>

          <UTable
            :data="configuredRoles"
            :columns="columns"
            :loading="status === 'pending'"
          >
            <template #shift-cell="{ row }">
              <UBadge color="gray" variant="soft" class="capitalize">
                {{ row.original.shift === 'morning' ? 'Manhã' : row.original.shift === 'afternoon' ? 'Tarde' : 'Noite' }}
              </UBadge>
            </template>
            
            <template #check_in_time-cell="{ row }">
              <span class="font-mono text-gray-600 dark:text-gray-300">{{ row.original.check_in_time }}</span>
            </template>
            
            <template #actions-cell="{ row }">
              <div class="flex items-center gap-2">
                <UButton 
                  color="gray" 
                  variant="ghost" 
                  icon="i-heroicons-pencil-square" 
                  size="xs" 
                  @click="editRoleConfig(row.original)"
                />
                <UButton 
                  color="red" 
                  variant="ghost" 
                  icon="i-heroicons-trash" 
                  size="xs" 
                  :loading="isDeleting === row.original.id"
                  @click="removeRoleConfig(row.original)"
                />
              </div>
            </template>
          </UTable>
        </UCard>
      </div>

      <SettingsRoleFormSlideover 
        v-model="isSlideoverOpen" 
        :role="editingRole"
        @saved="handleRoleSaved" 
      />
    </template>
  </UDashboardPanel>
</template>
