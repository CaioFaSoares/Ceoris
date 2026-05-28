/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // 1. Criar o Administrador (Superuser no v0.23+)
  const adminEmail = $os.getenv("PB_ADMIN_EMAIL");
  const adminPassword = $os.getenv("PB_ADMIN_PASSWORD");

  if (adminEmail && adminPassword) {
    try {
      app.findAuthRecordByEmail("_superusers", adminEmail);
    } catch (_) {
      const superusers = app.findCollectionByNameOrId("_superusers");
      const admin = new Record(superusers);
      admin.set("email", adminEmail);
      admin.setPassword(adminPassword);
      app.save(admin);
      console.log("🚀 Admin criado via automação!");
    }
  }

  // 2. Definir as Coleções Base (Sem relacionamentos)
  const baseCollections = [
    {
        "name": "users",
        "type": "auth",
        "fields": [
            {
                "name": "name",
                "type": "text",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "avatar",
                "type": "file",
                "required": false,
                "presentable": false,
                "unique": false,
                "mimeTypes": [
                    "image/jpeg",
                    "image/png",
                    "image/svg+xml",
                    "image/gif",
                    "image/webp"
                ],
                "thumbs": null,
                "maxSelect": 1,
                "maxSize": 5242880,
                "protected": false
            }
        ],
        "indexes": [],
        "listRule": "id = @request.auth.id",
        "viewRule": "id = @request.auth.id",
        "createRule": "",
        "updateRule": "id = @request.auth.id",
        "deleteRule": "id = @request.auth.id",
        "options": {
            "allowEmailAuth": true,
            "allowOAuth2Auth": true,
            "allowUsernameAuth": true,
            "exceptEmailDomains": null,
            "manageRule": null,
            "minPasswordLength": 8,
            "onlyEmailDomains": null,
            "onlyVerified": false,
            "requireEmail": false
        }
    },
    {
        "name": "guilds",
        "type": "base",
        "fields": [
            {
                "name": "discord_id",
                "type": "text",
                "required": true,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "name",
                "type": "text",
                "required": true,
                "presentable": true,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "status",
                "type": "select",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSelect": 1,
                "values": [
                    "active",
                    "archived"
                ]
            },
            {
                "name": "announcement_channel_id",
                "type": "text",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "squad_roles",
                "type": "json",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSize": 2000000
            },
            {
                "name": "mentor_roles",
                "type": "json",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSize": 2000000
            },
            {
                "name": "skill_roles",
                "type": "json",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSize": 2000000
            }
        ],
        "indexes": [
            "CREATE UNIQUE INDEX `idx_guilds_discord_id` ON `guilds` (`discord_id`)"
        ],
        "listRule": "",
        "viewRule": "",
        "createRule": "",
        "updateRule": "",
        "deleteRule": "",
        "options": {}
    },
    {
        "name": "broadcasts",
        "type": "base",
        "fields": [
            {
                "name": "content",
                "type": "text",
                "required": true,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "target_type",
                "type": "text",
                "required": true,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "target_roles",
                "type": "json",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSize": 2000000
            },
            {
                "name": "status",
                "type": "select",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSelect": 1,
                "values": [
                    "scheduled",
                    "processing",
                    "completed",
                    "failed"
                ]
            },
            {
                "name": "schedule_time",
                "type": "date",
                "required": true,
                "presentable": false,
                "unique": false,
                "min": "",
                "max": ""
            },
            {
                "name": "guild_id",
                "type": "text",
                "required": true,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "metrics_sent",
                "type": "number",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": 0,
                "max": null,
                "noDecimal": true
            },
            {
                "name": "metrics_errors",
                "type": "number",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": 0,
                "max": null,
                "noDecimal": true
            }
        ],
        "indexes": [],
        "listRule": "",
        "viewRule": "",
        "createRule": "",
        "updateRule": "",
        "deleteRule": "",
        "options": {}
    }
];

  // Salva as coleções base primeiro (para gerar os IDs delas no banco)
  baseCollections.forEach((colDef) => {
    try {
      const collection = new Collection(colDef);
      app.save(collection);
    } catch (err) {
      console.log("Ignorando criação da base col", colDef.name, err.message);
    }
  });

  // Helper para buscar o ID real da coleção gerado pelo PocketBase
  const getColId = (name) => {
    try {
      return app.findCollectionByNameOrId(name).id;
    } catch (e) {
      return name; // Fallback
    }
  };

  // 3. Criar as Coleções Relacionais
  const relationalCollections = [
    {
        "name": "roles",
        "type": "base",
        "fields": [
            {
                "name": "discord_id",
                "type": "text",
                "required": true,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "name",
                "type": "text",
                "required": true,
                "presentable": true,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "guild_id",
                "type": "relation",
                "required": false,
                "presentable": false,
                "unique": false,
                "collectionId": getColId("guilds"),
                "cascadeDelete": false,
                "minSelect": null,
                "maxSelect": 1,
                "displayFields": null
            },
            {
                "name": "shift",
                "type": "select",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSelect": 1,
                "values": [
                    "morning",
                    "afternoon",
                    "night"
                ]
            },
            {
                "name": "check_in_time",
                "type": "text",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "checkout_cooldown",
                "type": "number",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "noDecimal": true
            },
            {
                "name": "is_monitored",
                "type": "bool",
                "required": false,
                "presentable": false,
                "unique": false
            },
            {
                "name": "is_active",
                "type": "bool",
                "required": false,
                "presentable": false,
                "unique": false
            },
            {
                "name": "is_staff",
                "type": "bool",
                "required": false,
                "presentable": false,
                "unique": false
            },
            {
                "name": "squad_channel_id",
                "type": "text",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            }
        ],
        "indexes": [
            "CREATE UNIQUE INDEX `idx_roles_discord_id` ON `roles` (`discord_id`)"
        ],
        "listRule": "",
        "viewRule": "",
        "createRule": "",
        "updateRule": "",
        "deleteRule": "",
        "options": {}
    },
    {
        "name": "activities",
        "type": "base",
        "fields": [
            {
                "name": "guild_id",
                "type": "relation",
                "required": false,
                "presentable": false,
                "unique": false,
                "collectionId": getColId("guilds"),
                "cascadeDelete": false,
                "minSelect": null,
                "maxSelect": 1,
                "displayFields": null
            },
            {
                "name": "title",
                "type": "text",
                "required": true,
                "presentable": true,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "description",
                "type": "editor",
                "required": false,
                "presentable": false,
                "unique": false,
                "convertUrls": false
            },
            {
                "name": "type",
                "type": "select",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSelect": 1,
                "values": [
                    "announcement",
                    "task",
                    "feedback_request"
                ]
            },
            {
                "name": "due_date",
                "type": "date",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": "",
                "max": ""
            },
            {
                "name": "status",
                "type": "select",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSelect": 1,
                "values": [
                    "draft",
                    "published",
                    "archived"
                ]
            }
        ],
        "indexes": [],
        "listRule": "",
        "viewRule": "",
        "createRule": "",
        "updateRule": "",
        "deleteRule": "",
        "options": {}
    },
    {
        "name": "managers",
        "type": "base",
        "fields": [
            {
                "name": "discord_id",
                "type": "text",
                "required": true,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "name",
                "type": "text",
                "required": true,
                "presentable": true,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "role",
                "type": "select",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSelect": 1,
                "values": [
                    "admin",
                    "mentor",
                    "pedagogy"
                ]
            },
            {
                "name": "guilds",
                "type": "relation",
                "required": false,
                "presentable": false,
                "unique": false,
                "collectionId": getColId("guilds"),
                "cascadeDelete": false,
                "minSelect": null,
                "maxSelect": null,
                "displayFields": null
            },
            {
                "name": "user_id",
                "type": "relation",
                "required": false,
                "presentable": false,
                "unique": false,
                "collectionId": getColId("users"),
                "cascadeDelete": true,
                "minSelect": null,
                "maxSelect": 1,
                "displayFields": null
            }
        ],
        "indexes": [
            "CREATE UNIQUE INDEX `idx_managers_discord_id` ON `managers` (`discord_id`)"
        ],
        "listRule": "",
        "viewRule": "",
        "createRule": "",
        "updateRule": "",
        "deleteRule": "",
        "options": {}
    },
    {
        "name": "students",
        "type": "base",
        "fields": [
            {
                "name": "discord_id",
                "type": "text",
                "required": true,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "username",
                "type": "text",
                "required": true,
                "presentable": true,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "nickname",
                "type": "text",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "role_id",
                "type": "relation",
                "required": false,
                "presentable": false,
                "unique": false,
                "collectionId": getColId("roles"),
                "cascadeDelete": false,
                "minSelect": null,
                "maxSelect": 1,
                "displayFields": null
            },
            {
                "name": "guild_id",
                "type": "relation",
                "required": false,
                "presentable": false,
                "unique": false,
                "collectionId": getColId("guilds"),
                "cascadeDelete": false,
                "minSelect": null,
                "maxSelect": 1,
                "displayFields": null
            },
            {
                "name": "channel_id",
                "type": "text",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "status",
                "type": "select",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSelect": 1,
                "values": [
                    "active",
                    "inactive",
                    "dropped"
                ]
            },
            {
                "name": "secondary_roles",
                "type": "relation",
                "required": false,
                "presentable": false,
                "unique": false,
                "collectionId": getColId("roles"),
                "cascadeDelete": false,
                "minSelect": null,
                "maxSelect": null,
                "displayFields": null
            },
            {
                "name": "shift",
                "type": "select",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSelect": 1,
                "values": [
                    "morning",
                    "afternoon",
                    "night"
                ]
            },
            {
                "name": "user_id",
                "type": "relation",
                "required": false,
                "presentable": false,
                "unique": false,
                "collectionId": getColId("users"),
                "cascadeDelete": true,
                "minSelect": null,
                "maxSelect": 1,
                "displayFields": null
            }
        ],
        "indexes": [
            "CREATE UNIQUE INDEX `idx_students_discord_guild` ON `students` (`discord_id`, `guild_id`)"
        ],
        "listRule": "",
        "viewRule": "",
        "createRule": "",
        "updateRule": "",
        "deleteRule": "",
        "options": {}
    },
    {
        "name": "attendances",
        "type": "base",
        "fields": [
            {
                "name": "student_id",
                "type": "relation",
                "required": false,
                "presentable": false,
                "unique": false,
                "collectionId": getColId("students"),
                "cascadeDelete": false,
                "minSelect": null,
                "maxSelect": 1,
                "displayFields": null
            },
            {
                "name": "date",
                "type": "date",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": "",
                "max": ""
            },
            {
                "name": "clock_in",
                "type": "date",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": "",
                "max": ""
            },
            {
                "name": "clock_out",
                "type": "date",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": "",
                "max": ""
            },
            {
                "name": "status",
                "type": "select",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSelect": 1,
                "values": [
                    "pending_checkout",
                    "completed",
                    "absent",
                    "justified",
                    "late"
                ]
            },
            {
                "name": "source",
                "type": "select",
                "required": false,
                "presentable": false,
                "unique": false,
                "maxSelect": 1,
                "values": [
                    "discord_bot",
                    "manual_override"
                ]
            },
            {
                "name": "notes",
                "type": "text",
                "required": false,
                "presentable": false,
                "unique": false,
                "min": null,
                "max": null,
                "pattern": ""
            },
            {
                "name": "checkout_prompt_sent",
                "type": "bool",
                "required": false,
                "presentable": false,
                "unique": false
            }
        ],
        "indexes": [],
        "listRule": "",
        "viewRule": "",
        "createRule": "",
        "updateRule": "",
        "deleteRule": "",
        "options": {}
    }
];

  relationalCollections.forEach((colDef) => {
    try {
      // Re-evaluate relation collectionIds just before saving
      colDef.fields.forEach(f => {
        if (f.type === "relation" && typeof f.collectionId === "string") {
            try {
                const actualCol = app.findCollectionByNameOrId(f.collectionId);
                f.collectionId = actualCol.id;
            } catch(e) {}
        }
      });
      const collection = new Collection(colDef);
      app.save(collection);
    } catch (err) {
      console.log("Erro ao criar coleção relacional", colDef.name, err.message);
    }
  });

  console.log("🚀 Todas as coleções criadas com sucesso!");

}, (app) => {
  const collections = ["attendances", "students", "managers", "activities", "roles", "users", "guilds", "broadcasts"];
  collections.forEach((name) => {
    try {
      const col = app.findCollectionByNameOrId(name);
      app.delete(col);
    } catch (_) {}
  });
});
