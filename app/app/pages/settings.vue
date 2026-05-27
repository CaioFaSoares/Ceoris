<script setup lang="ts">
// Metadados da página (SEO e Título da Aba do Navegador)
useSeoMeta({
  title: 'Setup e Provisionamento - Ceoris'
})
</script>

<template>
  <UDashboardPage>
    <UDashboardPanel grow>
      <UDashboardNavbar title="Configurações e Provisionamento" />

      <UDashboardPanelContent>
        <div class="p-4 flex flex-col gap-6 max-w-4xl mx-auto w-full">
          <SettingsSystemHealthCard />
          <SettingsDiscordWizardCard />
        </div>
      </UDashboardPanelContent>
    </UDashboardPanel>
  </UDashboardPage>
</template>

