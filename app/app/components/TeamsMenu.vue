<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'

// Importando o estado global (PRD 1.2)
const { selectedGuildId, setGuildId } = useAppContext()

defineProps<{
  collapsed?: boolean
}>()

interface Guild {
  id: string
  name: string
  icon_url?: string
}

// 1. Busca dos Dados via API (PRD 1.1)
const { data: guilds, status } = await useAsyncData<Guild[]>(
  'discord-guilds',
  () => useApi('/api/discord/guilds')
)

// 2. Mágica de Inicialização Automática
// Se não houver guilda selecionada e a API retornar dados, seleciona a primeira por padrão
watchEffect(() => {
  if (!selectedGuildId.value && guilds.value && guilds.value.length > 0) {
    setGuildId(guilds.value[0].id)
  }
})

// 3. Mapeamento para o formato do DropdownMenu
const items = computed<DropdownMenuItem[][]>(() => {
  const guildItems = guilds.value ? guilds.value.map(guild => ({
    label: guild.name,
    // Se houver ícone no Discord usa avatar, senão usa ícone genérico
    avatar: guild.icon_url ? { src: guild.icon_url, alt: guild.name } : undefined,
    icon: !guild.icon_url ? 'i-heroicons-server' : undefined, 
    onSelect() {
      setGuildId(guild.id) // Atualiza o estado global ao clicar
    }
  })) : []

  return [
    guildItems,
    [
      {
        label: 'Adicionar Servidor',
        icon: 'i-heroicons-plus-circle'
      },
      {
        label: 'Sincronizar Dados',
        icon: 'i-heroicons-arrow-path'
      }
    ]
  ]
})

// 4. Computed auxiliar para exibir o servidor ativamente selecionado no botão principal
const activeGuild = computed(() => {
  return guilds.value?.find(g => g.id === selectedGuildId.value) || null
})
</script>

<template>
  <UDropdownMenu
    :items="items"
    :content="{ align: 'center', collisionPadding: 12 }"
    :ui="{ content: collapsed ? 'w-40' : 'w-(--reka-dropdown-menu-trigger-width)' }"
  >
    <UButton
      v-if="status === 'pending'"
      loading
      color="neutral"
      variant="ghost"
      block
      :square="collapsed"
    />
    
    <UButton
      v-else-if="activeGuild"
      v-bind="{
        label: collapsed ? undefined : activeGuild.name,
        avatar: activeGuild.icon_url ? { src: activeGuild.icon_url, alt: activeGuild.name } : undefined,
        icon: !activeGuild.icon_url ? 'i-heroicons-server' : undefined,
        trailingIcon: collapsed ? undefined : 'i-lucide-chevrons-up-down'
      }"
      color="neutral"
      variant="ghost"
      block
      :square="collapsed"
      class="data-[state=open]:bg-elevated"
      :class="[!collapsed && 'py-2']"
      :ui="{
        trailingIcon: 'text-dimmed'
      }"
    />

    <UButton
      v-else
      label="Nenhum Servidor"
      icon="i-heroicons-exclamation-triangle"
      color="error"
      variant="ghost"
      block
      :square="collapsed"
    />
  </UDropdownMenu>
</template>
