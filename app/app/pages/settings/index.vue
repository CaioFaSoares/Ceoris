<script setup lang="ts">
// Metadados da página (SEO e Título da Aba do Navegador)
useSeoMeta({
  title: 'Setup e Provisionamento - Ceoris'
})
</script>

<template>
  <UDashboardPanel grow>
    <template #header>
      <UDashboardNavbar title="Configurações e Provisionamento" />
    </template>

    <template #body>
      <div class="p-4 flex flex-col gap-6 max-w-4xl mx-auto w-full">
        <SettingsSystemHealthCard />
        <SettingsDiscordWizardCard />
      </div>
    </template>
  </UDashboardPanel>
</template>

