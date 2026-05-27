<script setup lang="ts">
// Metadados da página (SEO e Título da Aba do Navegador)
useSeoMeta({
  title: 'Centro de Inteligência - Ceoris'
})
</script>

<template>
  <UDashboardPanel grow>
    <template #header>
      <UDashboardNavbar title="Centro de Inteligência">
        <template #right>
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div class="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">
        Módulo em construção...
      </div>
    </template>
  </UDashboardPanel>
</template>
