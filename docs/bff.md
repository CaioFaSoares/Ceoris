# Padrão BFF (Backend-For-Frontend) - Camada de DTOs

Este documento define os contratos de dados e as respostas das rotas (DTOs) que o Backend em Go expõe para o Nuxt Frontend.

## Filosofia

O Backend em Go atua como um BFF, consumindo os dados brutos e aninhados do PocketBase (`models.*Record`) e transformando-os em respostas rasas e pré-computadas para a UI (`dto.*Response`).
Isso garante que o Frontend (Nuxt) se preocupe estritamente com a renderização visual, deixando as verificações lógicas, extração de objetos `expand`, e validações a cargo do Go.

---

## Mapeamento de Rotas e Contratos (Fase 2)

### 1. Lista de Alunos de uma Guilda
- **Rota:** `GET /api/guilds/:id/students`
- **Uso no Nuxt:** `app/app/pages/settings/members.vue`
- **Descrição:** Retorna a lista de alunos formatada e achatada para exibição na tabela.
- **DTO de Resposta (`StudentListResponse`):**
  ```json
  [
    {
      "id": "abc123xyz",
      "discord_id": "1234567890",
      "username": "caio.soares",
      "nickname": "Caio Soares",
      "role_name": "T15",                 // Extraído de expand.role_id.name
      "guild_id": "guild123",
      "status": "active",
      "shift": "morning",
      "channel_id": "111222333444",       // [Legado] Mantido temporariamente para o front atual
      "has_1on1": true,                   // [Novo] Computado pelo Go
      "channel_status": "ready"           // [Novo] "ready" ou "pending"
    }
  ]
  ```

### 2. Configurações de Cargos (Roles) da Guilda
- **Rota:** `GET /api/guilds/:id/squads`
- **Uso no Nuxt:** Renderização de tabelas e opções de seleção em `squads.vue`.
- **Descrição:** Retorna as configurações de turnos e cooldowns associadas a cada Cargo mapeado do Discord.
- **DTO de Resposta (`RoleConfigResponse`):**
  ```json
  [
    {
      "id": "role123",
      "discord_id": "0987654321",
      "name": "Backend Staff",
      "shift": "afternoon",
      "check_in_time": "14:00",
      "checkout_cooldown": 240,
      "is_monitored": true,
      "is_active": true,
      "is_staff": false,
      "squad_channel_id": "888999777"
    }
  ]
  ```

### 3. Setup de Taxonomia da Guilda
- **Rota:** `GET /api/guilds/:id/mapping`
- **Uso no Nuxt:** `app/app/pages/settings/roles.vue`
- **Descrição:** Retorna exclusivamente as IDs dos cargos classificados (Turmas, Mentores, Skills) na raiz do JSON, omitindo dados de infraestrutura.
- **DTO de Resposta (`GuildMappingResponse`):**
  ```json
  {
    "squad_roles": ["roleID_1", "roleID_2"],
    "mentor_roles": ["roleID_3"],
    "skill_roles": ["roleID_4"]
  }
  ```

### 4. Relatório Diário de Presenças
- **Rota:** `GET /api/guilds/:id/reports/attendances?date=YYYY-MM-DD&role_id=xxx`
- **Uso no Nuxt:** (Futuras implementações de relatórios e dashboards).
- **Descrição:** Achatamento do relatório de ponto, onde as informações vitais do aluno são colocadas lado a lado com os horários.
- **DTO de Resposta (`AttendanceListResponse`):**
  ```json
  [
    {
      "attendance_id": "att123",
      "student_name": "caio.soares",
      "student_nickname": "Caio Soares",
      "date": "2026-05-27",
      "clock_in": "2026-05-27T10:00:00Z",
      "clock_out": "2026-05-27T18:00:00Z",
      "status": "completed",
      "source": "discord_bot"
    }
  ]
  ```

### 5. Rotas Assíncronas de Ação (Fire-and-Forget)
- **Rotas:** 
  - `POST /api/guilds/:id/sync`
  - `POST /api/guilds/:id/sync/members`
  - `POST /api/guilds/:id/provision`
  - `POST /api/guilds/:id/heal`
- **Descrição:** Essas rotas disparam Goroutines em background para processos demorados e evitam Timeout no Nuxt.
- **Contrato de Resposta (Status `202 Accepted`):**
  ```json
  {
    "data": {
      "status": "processing",
      "message": "A tarefa foi iniciada em background"
    }
  }
  ```

---

## Roteiro Incremental
À medida que novas necessidades da UI surgirem, as propriedades não deverão ser criadas como getters no Vue. Em vez disso, a lógica deve ser transposta para o diretório `server/internal/dto/*_dto.go` e mapeada diretamente na conversão dos dados do PocketBase.
