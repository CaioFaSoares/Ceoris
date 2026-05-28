<script setup lang="ts">
import { z } from 'zod'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits(['update:modelValue', 'saved'])

const { selectedGuildId } = useAppContext()
const toast = useToast()

// 1. CORREÇÃO DO V-MODEL DO SLIDEOVER
// O Nuxt UI novo usa v-model:open internamente.
const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const isSubmitting = ref(false)

const schema = z.object({
  role_id: z.string().min(1, 'Selecione um cargo do Discord'),
  shift: z.enum(['morning', 'afternoon', 'night'], { required_error: 'Selecione um turno' }),
  check_in_time: z.string().regex(/^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$/, 'Use o formato HH:MM')
})
type Schema = z.output<typeof schema>

const state = reactive<Schema>({ 
  role_id: '', 
  shift: 'morning', 
  check_in_time: '09:00' 
})

// 2. CORREÇÃO DO SUSPENSE (useLazyAsyncData) e Curto-circuito
const { data: dbRoles, status: loadingRoles } = useLazyAsyncData<any[]>(
  'config-roles-list',
  () => selectedGuildId.value ? useApi(`/api/guilds/${selectedGuildId.value}/squads`) : Promise.resolve([]),
  { watch: [selectedGuildId], default: () => [] }
)

const availableRoles = computed(() => {
  if (!dbRoles.value) return []
  return dbRoles.value.sort((a, b) => a.name.localeCompare(b.name))
})

// Adaptação dos turnos para o novo formato do Nuxt UI
const shiftItems = [
  { label: 'Manhã', value: 'morning' },
  { label: 'Tarde', value: 'afternoon' },
  { label: 'Noite', value: 'night' }
]

async function onSubmit(event: { data: Schema }) {
  isSubmitting.value = true
  try {
    // Usamos o PATCH exato que o backend em Go já possui
    await useApi(`/api/guilds/${selectedGuildId.value}/squads/${event.data.role_id}`, {
      method: 'PATCH',
      body: {
        shift: event.data.shift,
        check_in_time: event.data.check_in_time,
        is_monitored: true,
        is_active: true
      }
    })
    
    toast.add({ title: 'Turma configurada com sucesso!', color: 'green' })
    emit('saved') // Avisa a página pai para recarregar a tabela
    isOpen.value = false // Fecha o modal
    state.role_id = '' // Reseta o form
  } catch (err) {
    // Erros já tratados globalmente
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <USlideover v-model:open="isOpen">
    <template #content>
      <UCard class="flex flex-col flex-1" :ui="{ body: { base: 'flex-1' }, ring: '', divide: 'divide-y divide-gray-100 dark:divide-gray-800' }">
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-base font-semibold leading-6 text-gray-900 dark:text-white">
              Configurar Nova Turma
            </h3>
            <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark-20-solid" class="-my-1" @click="isOpen = false" />
          </div>
        </template>

        <UForm :schema="schema" :state="state" class="space-y-6" @submit="onSubmit">
          
          <UFormField label="Cargo no Discord (Squad)" name="role_id">
            <USelectMenu
              v-model="state.role_id"
              :items="availableRoles"
              value-key="id"
              label-key="name"
              :loading="loadingRoles === 'pending'"
              searchable
              placeholder="Selecione o cargo correspondente"
              class="w-full"
            >
              <template #empty>
                Nenhum cargo encontrado. (Faça o Sync primeiro)
              </template>
            </USelectMenu>
          </UFormField>

          <UFormField label="Turno de Estudo" name="shift">
            <USelect
              v-model="state.shift"
              :items="shiftItems"
              value-key="value"
              label-key="label"
              class="w-full"
            />
          </UFormField>

          <UFormField label="Horário do Check-in (HH:MM)" name="check_in_time">
            <UInput v-model="state.check_in_time" type="time" icon="i-heroicons-clock" class="w-full" />
          </UFormField>

          <div class="flex justify-end gap-3 pt-4">
            <UButton color="white" variant="solid" @click="isOpen = false">Cancelar</UButton>
            <UButton type="submit" color="primary" :loading="isSubmitting">Salvar Regra</UButton>
          </div>
        </UForm>
      </UCard>
    </template>
  </USlideover>
</template>
