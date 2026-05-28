package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chantry/server/internal/config"
	"chantry/server/internal/cron"
	"chantry/server/internal/discord"
	"chantry/server/internal/handlers"
	pbclient "chantry/server/internal/pocketbase"
	"chantry/server/internal/usecases"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	runFiberApp()
}

// runFiberApp executes the original Discord Integration daemon (Fiber API)
func runFiberApp() {
	log.Println("🚀 Starting Discord Go Daemon...")

	// 1. Load Configurations from environment (fails early if invalid)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ FATAL: Erro ao carregar as configurações: %v", err)
	}

	// 2. Initialize Discord Client Service
	discordService, err := discord.NewDiscordService(cfg.DiscordBotToken)
	if err != nil {
		log.Fatalf("❌ FATAL: Erro ao inicializar o cliente Discord: %v", err)
	}
	log.Println("✅ Cliente Discord inicializado com sucesso")

	// 3. Initialize PocketBase Client & Authenticate
	pbClient := pbclient.NewClient(cfg.PocketBaseURL)
	if err := pbClient.Authenticate(cfg.PBAdminEmail, cfg.PBAdminPassword); err != nil {
		log.Fatalf("❌ FATAL: Erro ao autenticar no PocketBase: %v", err)
	}
	log.Println("✅ PocketBase Client autenticado com sucesso")

	// 4. Initialize HTTP handlers/controllers
	pbRepo := pbclient.NewRepository(pbClient)

	provisionUsecase := usecases.NewProvisionUsecase(discordService, pbRepo)
	provisionHandler := handlers.NewProvisionHandler(provisionUsecase)

	syncUsecase := usecases.NewSyncUsecase(discordService, pbRepo, provisionUsecase)
	syncHandler := handlers.NewSyncHandler(syncUsecase)

	discordHandler := handlers.NewDiscordHandler(discordService)

	configHandler := handlers.NewConfigHandler(pbRepo, discordService)
	reportUsecase := usecases.NewReportUsecase(pbRepo)
	reportHandler := handlers.NewReportHandler(pbRepo, reportUsecase)

	testUsecase := usecases.NewTestUsecase(pbRepo, discordService)
	testHandler := handlers.NewTestHandler(testUsecase, pbRepo)

	broadcastUsecase := usecases.NewBroadcastUsecase(discordService, pbRepo)
	broadcastHandler := handlers.NewBroadcastHandler(broadcastUsecase)

	uiUsecase := usecases.NewUIUsecase(discordService, pbRepo)
	uiHandler := handlers.NewUIHandler(uiUsecase)

	systemHandler := handlers.NewSystemHandler(pbRepo, discordService, cfg.DiscordAppID)

	// Initialize AttendanceUsecase and register Interaction handlers for Gateway
	attendanceUsecase := usecases.NewAttendanceUsecase(pbRepo)
	discord.RegisterInteractionHandlers(discordService.Session, attendanceUsecase)
	log.Println("✅ Handlers de Interação do Discord registrados")

	// Start background async Broadcast worker
	cron.StartBroadcastWorker(pbRepo, discordService)

	// Set global timezone
	loc, err := time.LoadLocation(cfg.Timezone)
	if err == nil {
		time.Local = loc
		log.Printf("🌍 Timezone set to %s", cfg.Timezone)
	} else {
		log.Printf("⚠️ Failed to load timezone %s: %v", cfg.Timezone, err)
	}

	// Initialize Fiber App with dynamic configuration
	app := fiber.New(fiber.Config{
		AppName: "Chantry Go Daemon v0.1.0",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Logger Middleware for basic connection tracking
	app.Use(logger.New())

	// CORS Middleware to allow Streamlit frontend query the health route securely
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Healthcheck Route
	app.Get("/api/health", func(c *fiber.Ctx) error {
		now := time.Now()
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":    "ok",
			"service":   "chantry-go-daemon",
			"timestamp": now.Format(time.RFC3339),
			"timezone":  cfg.Timezone,
		})
	})

	// Nível 0: Global
	api := app.Group("/api")
	api.Get("/system/health", systemHandler.HandleGetHealth)

	// Nível 1: Descoberta de Domínio
	api.Get("/guilds", discordHandler.HandleGetGuilds)

	// Nível 2: Recursos da Guilda (Guild-Centric)
	guilds := api.Group("/guilds/:id")

	// Taxonomia & Setup
	guilds.Get("/discord/roles", discordHandler.HandleGetGuildRoles)
	guilds.Get("/discord/members", discordHandler.HandleGetGuildMembers)
	guilds.Get("/discord/categories", discordHandler.HandleGetCategories)
	guilds.Post("/discord/categories", discordHandler.HandleCreateCategory)
	guilds.Get("/mapping", configHandler.HandleGetGuildMapping)
	guilds.Patch("/mapping", configHandler.HandleUpdateGuildMapping)

	// Regras de Negócio (Squads)
	guilds.Get("/squads", configHandler.HandleGetGuildRolesConfig)
	guilds.Patch("/squads/:roleId", configHandler.HandleUpdateRoleConfig)
	guilds.Patch("/squads/:roleId/channel", configHandler.HandleUpdateSquadChannel)
	guilds.Patch("/config", configHandler.HandleUpdateGuildConfig)

	// Vida Acadêmica (Estudantes & Sincronização)
	guilds.Get("/students", configHandler.HandleGetGuildStudents)
	guilds.Post("/sync", syncHandler.HandleAdvancedSync)
	guilds.Post("/sync/members", syncHandler.HandleSyncMembers)

	// Operações em Lote (Assíncronas)
	guilds.Post("/provision", provisionHandler.HandleProvisionChannels)
	guilds.Post("/heal", provisionHandler.HandleHealChannels)
	guilds.Get("/ui/provision-page", provisionHandler.HandleGetProvisionPageData)

	// Targeted Broadcast Route
	guilds.Post("/broadcast/send", broadcastHandler.HandleSendBroadcast)
	guilds.Get("/ui/broadcast-page", broadcastHandler.HandleGetBroadcastPageData)
	api.Post("/broadcasts", broadcastHandler.HandleCreateBroadcast)
	api.Delete("/broadcasts/:id", broadcastHandler.HandleCancelBroadcast)

	// BFF UI Routes
	api.Get("/ui/squads/:roleId", uiHandler.HandleSquadDashboard)

	// Analytical Report Routes (Daily Attendance Dashboard)
	guilds.Get("/reports/attendances", reportHandler.HandleGetAttendances)
	guilds.Get("/reports/export", reportHandler.HandleExportReport)

	// Sandbox Dry Run Test Routes
	api.Post("/test/attendance/trigger", testHandler.HandleTestAttendanceTrigger)
	guilds.Get("/test/managers", testHandler.HandleGetManagers)

	// Start dynamic cron background scheduler
	cron.StartDynamicCron(pbRepo, discordService, cfg.Timezone)
	// Start background clock-out check ticker
	cron.StartClockOutTicker(pbRepo, discordService, cfg.Timezone)

	// Get port from environment or fallback to 12000
	port := os.Getenv("PORT")
	if port == "" {
		port = "12000"
	}

	// Start Fiber in a non-blocking goroutine
	go func() {
		log.Printf("🚀 Fiber app listening on port %s...", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("⚠️ Fiber server listening stopped: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("⚡ Gracefully shutting down daemon...")

	// Close Discord Gateway connection
	if err := discordService.Close(); err != nil {
		log.Printf("⚠️ Warning: erro ao fechar sessão do Discord: %v", err)
	} else {
		log.Println("✅ Conexão WebSocket do Discord fechada com sucesso")
	}

	// Shutdown Fiber app
	if err := app.Shutdown(); err != nil {
		log.Printf("⚠️ Warning: erro ao desligar Fiber: %v", err)
	} else {
		log.Println("✅ Servidor Fiber finalizado com sucesso")
	}
}
