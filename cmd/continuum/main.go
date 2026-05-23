package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/anarudhan/continuum/internal/api/handlers"
	"github.com/anarudhan/continuum/internal/api/middleware"
	"github.com/anarudhan/continuum/internal/models"
	ws "github.com/anarudhan/continuum/internal/websocket"
)

func main() {
	// Load configuration
	databaseURL := getEnv("CONTINUUM_DATABASE_URL", "postgres://continuum:continuum@localhost:5432/continuum?sslmode=disable")
	redisURL := getEnv("CONTINUUM_REDIS_URL", "redis://localhost:6379")
	port := getEnv("CONTINUUM_PORT", "8080")

	// Initialize database
	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	store := &models.Store{DB: db}
	if err := store.RunMigrations(context.Background()); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Redis
	redisOpt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpt)
	defer redisClient.Close()

	// Initialize stores
	agentStore := models.NewAgentStore(store)
	memoryStore := models.NewMemoryStore(store)
	sessionStore := models.NewSessionStore(store)

	// Create default agent if none exists
	ensureDefaultAgent(context.Background(), agentStore)

	// Initialize handlers
	memoryHandler := handlers.NewMemoryHandler(memoryStore)
	sessionHandler := handlers.NewSessionHandler(sessionStore)
	agentHandler := handlers.NewAgentHandler(agentStore)
	healthHandler := handlers.NewHealthHandler(db, redisClient)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(agentStore)
	rateLimiter := middleware.NewRateLimiter(redisClient)

	// Setup router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.SecurityHeaders())

	// WebSocket endpoint
	wsServer := ws.NewServer()
	go wsServer.Run()
	router.GET("/ws", wsServer.HandleConnection)

	// Health endpoints (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)
	router.GET("/live", healthHandler.Live)

	// API v1 routes
	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware.RequireAPIKey())
	v1.Use(rateLimiter.Limit())
	{
		// Memories
		v1.POST("/memories", memoryHandler.Create)
		v1.GET("/memories", memoryHandler.List)
		v1.GET("/memories/search", memoryHandler.Search)
		v1.GET("/memories/:id", memoryHandler.Get)
		v1.DELETE("/memories/:id", memoryHandler.Delete)

		// Sessions
		v1.POST("/sessions", sessionHandler.Create)
		v1.GET("/sessions", sessionHandler.List)
		v1.GET("/sessions/:id", sessionHandler.Get)
		v1.POST("/sessions/:id/end", sessionHandler.End)

		// Agents (admin-only for full listing)
		v1.GET("/agents", authMiddleware.RequireAdmin(), agentHandler.List)
	}

	// Admin routes (admin-only)
	admin := router.Group("/admin")
	admin.Use(authMiddleware.RequireAPIKey())
	admin.Use(authMiddleware.RequireAdmin())
	admin.POST("/agents", agentHandler.Create)

	// Start server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Continuum server started on port %s", port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func ensureDefaultAgent(ctx context.Context, agentStore *models.AgentStore) {
	agents, err := agentStore.List(ctx)
	if err != nil {
		log.Printf("Warning: failed to list agents: %v", err)
		return
	}

	if len(agents) == 0 {
		agent, apiKey, err := agentStore.Create(ctx, "default")
		if err != nil {
			log.Printf("Warning: failed to create default agent: %v", err)
			return
		}
		log.Printf("Created default agent: %s", agent.ID)
		log.Printf("API Key hint: ctm_...%s (save this — shown once)", apiKey[len(apiKey)-4:])
	}
}
