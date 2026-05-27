<script setup lang="ts">
import { z } from 'zod'

useSeoMeta({
  title: 'Mapeamento de Cargos - Ceoris'
})

const { selectedGuildId } = useAppContext()
const toast = useToast()
const router = useRouter()

// 1. Definição do Schema de Validação (Zod)
const schema = z.object({
  squad_roles: z.array(z.string()).default([]),
  mentor_roles: z.array(z.string()).default([]),
  skill_roles: z.array(z.string()).default([])
}).refine(data => {
  // Validação Cruzada: Impede que um mesmo cargo seja Turma e Mentor
  const intersection = data.squad_roles.filter(role => data.mentor_roles.includes(role))
  return intersection.length === 0
}, {
  message: "Um cargo não pode ser Turma e Mentor simultaneamente.",
  path: ["mentor_roles"] // Aponta o erro no campo de mentores
})

type Schema = z.output<typeof schema>
const state = reactive<Schema>({ squad_roles: [], mentor_roles: [], skill_roles: [] })
const isSubmitting = ref(false)

// 2. Fetch dos Cargos do Discord (Opções do Select)
const { data: discordRoles, status: loadingRoles } = useLazyAsyncData(
  'discord-all-roles',
  () => selectedGuildId.value 
    ? useApi(`/api/discord/guilds/${selectedGuildId.value}/roles`)
    : Promise.resolve([]),
  { watch: [selectedGuildId], default: () => [] }
)

// 3. Fetch do Mapeamento Salvo (Para preencher o state inicial)
const { status: loadingMapping } = useLazyAsyncData(
  'guild-role-mapping',
  async () => {
    if (!selectedGuildId.value) return null
    const mapping = await useApi<Schema>(`/api/config/guilds/${selectedGuildId.value}/mapping`)
    // Atualiza o state com os dados vindos do banco
    if (mapping) {
      state.squad_roles = mapping.squad_roles || []
      state.mentor_roles = mapping.mentor_roles || []
      state.skill_roles = mapping.skill_roles || []
    }
    return mapping
  },
  { watch: [selectedGuildId] }
)

// 4. Submissão do Formulário
async function onSubmit(event: { data: Schema }) {
  isSubmitting.value = true
  try {
    await useApi(`/api/config/guilds/${selectedGuildId.value}/mapping`, {
      method: 'PATCH',
      body: event.data
    })
    
    toast.add({ 
      title: 'Mapeamento Salvo!', 
      description: 'A taxonomia foi atualizada. Agora você pode configurar os turnos.',
      color: 'green' 
    })
    
    // Redireciona de forma fluida para a próxima etapa (Base de Alunos)
    router.push('/settings/members')
  } catch (error) {
    // Erros capturados pelo useApi global
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <UDashboardPanel grow>
    <template #header>
      <UDashboardNavbar title="Mapeamento de Cargos" />
    </template>

    <template #body>
      <div class="p-4 flex flex-col gap-6 max-w-4xl mx-auto w-full">
        <UAlert 
          v-if="!selectedGuildId" 
          icon="i-heroicons-exclamation-triangle" 
          title="Selecione um Servidor" 
          description="Selecione uma Guilda no menu lateral para mapear os cargos."
          color="yellow" 
          class="mb-6"
        />

        <div v-else-if="loadingMapping === 'pending'" class="flex justify-center py-12">
          <UIcon name="i-heroicons-arrow-path" class="w-8 h-8 animate-spin text-gray-400" />
        </div>

        <div v-else class="w-full">
          <div class="mb-8">
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">Taxonomia do Servidor</h2>
            <p class="text-gray-500 dark:text-gray-400">Classifique os cargos do seu Discord para que o Ceoris saiba exatamente quem monitorar e quem possui privilégios de acesso aos Cadernos.</p>
          </div>

          <UForm :schema="schema" :state="state" @submit="onSubmit" class="space-y-6">
            
            <UCard>
              <div class="mb-4">
                <div class="flex items-center gap-2">
                  <UIcon name="i-heroicons-user-group" class="text-blue-500 w-5 h-5" />
                  <h3 class="font-semibold">Cargos de Turmas (Squads)</h3>
                </div>
                <p class="text-sm text-gray-500">Selecione os cargos que representam salas de aula. Eles receberão controle de presença e canais 1-on-1.</p>
              </div>
              <UFormField name="squad_roles">
                <USelectMenu
                  v-model="state.squad_roles"
                  :items="discordRoles"
                  multiple
                  searchable
                  value-key="id"
                  label-key="name"
                  placeholder="Ex: T15, Turma Backend..."
                  :loading="loadingRoles === 'pending'"
                  class="w-full"
                />
              </UFormField>
            </UCard>

            <UCard>
              <div class="mb-4">
                <div class="flex items-center gap-2">
                  <UIcon name="i-heroicons-academic-cap" class="text-purple-500 w-5 h-5" />
                  <h3 class="font-semibold">Cargos de Mentores (Staff)</h3>
                </div>
                <p class="text-sm text-gray-500">Selecione a equipe pedagógica. Eles terão acesso automático de leitura aos canais e Cadernos dos alunos.</p>
              </div>
              <UFormField name="mentor_roles">
                <USelectMenu
                  v-model="state.mentor_roles"
                  :items="discordRoles"
                  multiple
                  searchable
                  value-key="id"
                  label-key="name"
                  placeholder="Ex: Mentor, Professor, Admin..."
                  :loading="loadingRoles === 'pending'"
                  class="w-full"
                />
              </UFormField>
            </UCard>

            <UCard>
              <div class="mb-4">
                <div class="flex items-center gap-2">
                  <UIcon name="i-heroicons-tag" class="text-emerald-500 w-5 h-5" />
                  <h3 class="font-semibold">Cargos Secundários (Skills / Trilhas)</h3>
                </div>
                <p class="text-sm text-gray-500">Cargos que não ativam recursos, mas que você deseja rastrear no banco de dados para relatórios ou filtros futuros.</p>
              </div>
              <UFormField name="skill_roles">
                <USelectMenu
                  v-model="state.skill_roles"
                  :items="discordRoles"
                  multiple
                  searchable
                  value-key="id"
                  label-key="name"
                  placeholder="Ex: Aluno Backend, React Native..."
                  :loading="loadingRoles === 'pending'"
                  class="w-full"
                />
              </UFormField>
            </UCard>

            <div class="flex justify-end gap-4 mt-8 pb-12">
              <UButton type="submit" color="primary" size="lg" :loading="isSubmitting">
                Salvar Mapeamento e Avançar
                <template #trailing>
                  <UIcon name="i-heroicons-arrow-right" class="w-5 h-5" />
                </template>
              </UButton>
            </div>

          </UForm>
        </div>
      </div>
    </template>
  </UDashboardPanel>
</template>
