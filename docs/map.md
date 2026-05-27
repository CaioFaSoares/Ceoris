Excelente ideia. Transformar aquelas dicas em um **Épico 0** formaliza a etapa de "limpeza e fundação" e evita que o projeto carregue lixo do template original. Isso cria um ambiente de desenvolvimento impecável antes mesmo de começarmos a plugar os dados reais.

Aqui está o **Roadmap Atualizado do Ceoris Dashboard**, agora começando com o Épico 0 estrutural:

---

# 🗺️ Roadmap de Produto: Ceoris Nuxt Dashboard (Atualizado)

## 📍 Épico 0: Preparação do Ambiente e Arquitetura Nuxt

**Objetivo:** Limpar o template base do Nuxt UI Dashboard, extrair os componentes de alto valor e criar a estrutura de pastas e reatividade definitiva do projeto.

* **PRD 0.1: Extração e Limpeza do Template**
* **Requisitos:** O template padrão vem com páginas de demonstração (`pages/inbox`, `pages/customers`, `pages/settings`). A missão aqui é "salvar" os componentes úteis (como o `<InboxList>`, modais de exclusão e tabelas complexas) movendo-os para a pasta global `/components`, e depois deletar essas páginas de demonstração.


* **PRD 0.2: Estrutura de Roteamento Base**
* **Requisitos:** Criar a árvore limpa de arquivos baseada nas nossas 4 áreas principais, deixando os arquivos em branco, apenas com o `<UPage>` e `<UPageHeader>`.
* **Arquivos a criar:** `pages/index.vue` (Home), `pages/setup.vue` (Provisionamento), `pages/inteligencia.vue` (Presença), e `pages/engajamento.vue` (Megafone).


* **PRD 0.3: Padrões de Reatividade (Substituindo o Streamlit)**
* **Requisitos:** Criar os exemplos base de composables para carregamento de dados. Ao invés do `st.cache_data` e `st.rerun()` do Python, vamos definir como usaremos o `useAsyncData` (para buscar dados no carregamento da página com cache nativo) e o `useIntervalFn` (do VueUse) caso precisemos de *auto-refresh* de dados na tela de Home.



---

## 📍 Épico 1: Integração API e Estado Global

**Objetivo:** Configurar a comunicação com o backend em Go (`localhost:12000`) e gerenciar a seleção de contexto (Servidor Discord).

* **PRD 1.1: Cliente de API (`useApi.ts`)**
* **Requisitos:** Criar um composable utilizando o `$fetch` nativo do Nuxt para padronizar as chamadas ao backend, injetando automaticamente headers e tratando erros globais.


* **PRD 1.2: Seletor Global de Servidor (Contexto)**
* **Requisitos:** Adaptar o componente `<TeamsMenu.vue>` do template. Ele fará um GET no Go para listar os servidores disponíveis e salvará o ID selecionado no estado global (`useState` ou Pinia), substituindo a lógica do `st.session_state` do Streamlit.


* **PRD 1.3: Layout e Sidebar Dinâmica**
* **Requisitos:** Modificar o `layouts/default.vue` e a `<UDashboardSidebar>`. Os links de navegação devem apontar para as páginas criadas no PRD 0.2.



---

## 📍 Épico 2: Dashboard e Onboarding (Migração `app.py`)

**Objetivo:** Recriar a tela inicial com métricas vitais e o assistente de instalação do bot.

* **PRD 2.1: Painel de Métricas (Index)**
* **Requisitos:** Utilizar os componentes `<HomeStats>` para renderizar os totais de Alunos, Canais Criados e Presenças, consumindo os endpoints do Go.


* **PRD 2.2: Wizard de Instalação e Status**
* **Requisitos:** Recriar os "step-cards" do Streamlit com `<UCard>` do Nuxt UI. Incluir o botão primário de autorização OAuth2 (`client_id`).
* **Saúde do Sistema:** Adicionar badges no cabeçalho superior (`<UHeader>`) que indicam se a API em Go e a conexão com o Discord estão "Online" ou "Offline".



---

## 📍 Épico 3: Setup e Provisionamento (Migração `2_server_setup.py`)

**Objetivo:** Interface de gerenciamento de turmas, cargos e criação em lote de canais 1-on-1.

* **PRD 3.1: Configuração de Turmas e Turnos**
* **Requisitos:** Formulários limpos usando `<UForm>`, `<USelectMenu>` (com busca de cargos do Discord) e `<UInput type="time">` para definir horários de check-in.


* **PRD 3.2: Tabela de Sincronização de Alunos**
* **Requisitos:** Reaproveitar a `<UTable>` rica extraída no Épico 0 para listar os alunos. Mostrar visualmente (tags/badges) quem já tem o "Canal 1-on-1" criado e quem está pendente.


* **PRD 3.3: Motor de Provisionamento em Lote**
* **Requisitos:** Botão de ação que dispara a rota de provisionamento. Para feedback visual, abrir um `<USlideover>` ou `<UProgress>` mostrando o andamento da criação dos canais no Discord em tempo real.



---

## 📍 Épico 4: Centro de Inteligência (Migração `3_intelligence_center.py`)

**Objetivo:** Visualização de frequência, gráficos de presença e exportação de CSV.

* **PRD 4.1: Visão Diária e Filtros Temporais**
* **Requisitos:** Implementar o seletor de datas (`<HomeDateRangePicker>`) e tabs (`<UTabs>`) para alternar entre "Roster Diário" e "Relatório Histórico".


* **PRD 4.2: Gráficos de Frequência**
* **Requisitos:** Adaptar o componente `<HomeChart>` do template (Chart.js) para mostrar as estatísticas de Completo, Atrasado e Falta, substituindo os gráficos pesados do Plotly.


* **PRD 4.3: Exportação de CSV Transparente**
* **Requisitos:** Tabela completa de registros de ponto e um `<UButton>` que faz download direto do CSV gerado pelo Go, sem recarregar a tela ou abrir novas abas.



---

## 📍 Épico 5: Hub de Engajamento (Migração `4_engagement_hub.py`)

**Objetivo:** Envio de comunicados (Broadcasts) direcionados e megafone.

* **PRD 5.1: Histórico e Layout de Inbox**
* **Requisitos:** Utilizar o layout extraído de `pages/inbox` no Épico 0. À esquerda, a lista de mensagens enviadas/agendadas; à direita, a pré-visualização da mensagem selecionada.


* **PRD 5.2: Compositor de Megafone (Editor)**
* **Requisitos:** Formulário de criação de mensagem. Usar `<UTextarea>` (ou um editor Markdown simples), e um seletor múltiplo (`<USelectMenu multiple>`) para definir as turmas alvo da mensagem.


* **PRD 5.3: Ações Rápidas (Clonar)**
* **Requisitos:** Botão no histórico de mensagens que preenche o formulário de composição instantaneamente com os dados de uma mensagem anterior.



---

Como desenvolvedor, começar pelo **Épico 0** é incrivelmente satisfatório porque você toma controle do repositório. Quer começar a executar o **PRD 0.1** (descobrindo quais componentes da demo atual do Nuxt Dashboard você deve puxar para a pasta `/components` e quais páginas deletar)?