/**
 * Gerenciador de Contexto Global do Ceoris.
 * Substitui o st.session_state do Streamlit.
 */
export const useAppContext = () => {
  // Usamos useCookie para garantir que o F5 não destrua o contexto da guilda.
  // O cookie armazenará apenas a string do ID do Discord.
  const selectedGuildId = useCookie<string | null>('ceoris_selected_guild', {
    default: () => null,
    watch: true, // Garante reatividade caso o cookie mude
    maxAge: 60 * 60 * 24 * 7 // Persiste por 7 dias
  })

  // Funções auxiliares (Actions) para manipular o estado de forma limpa
  const setGuildId = (guildId: string) => {
    selectedGuildId.value = guildId
  }

  const clearGuildId = () => {
    selectedGuildId.value = null
  }

  // Retornamos tanto a variável reativa quanto os métodos de mutação
  return {
    selectedGuildId,
    setGuildId,
    clearGuildId
  }
}
