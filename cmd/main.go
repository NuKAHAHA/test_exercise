package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"awesomeProject1/internal/config"
	"awesomeProject1/internal/handler"
	"awesomeProject1/internal/repository"
	"awesomeProject1/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	logger.Info("Starting application")

	logger.Info("Loading configuration")
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load config", slog.String("error", err.Error()))
		log.Fatal("Failed to load config:", err)
	}
	logger.Info("Configuration loaded successfully")

	logger.Info("Connecting to PostgreSQL database",
		slog.String("host", cfg.DBHost),
		slog.String("port", cfg.DBPort),
		slog.String("dbname", cfg.DBName),
		slog.String("user", cfg.DBUser))

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: NewGormLogger(logger),
	})
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL", slog.String("error", err.Error()))
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	logger.Info("Successfully connected to PostgreSQL")

	//logger.Info("Running database migrations")
	//if err := gormDB.AutoMigrate(&models.Subscription{}); err != nil {
	//	logger.Error("Failed to migrate database", slog.String("error", err.Error()))
	//	log.Fatal("Failed to migrate database:", err)
	//}
	//logger.Info("Database migrations completed successfully")

	//we used traditional migrations

	logger.Info("Initializing repository and service layers")
	repo := repository.NewSubscriptionRepository(gormDB, logger)
	service := service.NewSubscriptionService(repo, logger)

	logger.Info("Initializing HTTP server")
	router := gin.Default()

	router.Use(RequestLoggingMiddleware(logger))

	subHandler := handler.NewSubscriptionHandler(service, logger)

	api := router.Group("/subscriptions")
	{
		api.POST("", subHandler.Create)
		api.GET("/:id", subHandler.GetByID)
		api.PUT("/:id", subHandler.Update)
		api.DELETE("/:id", subHandler.Delete)
		api.GET("", subHandler.List)
		api.POST("/aggregate", subHandler.Aggregate)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	logger.Info("Starting HTTP server", slog.String("port", cfg.ServerPort))

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", slog.String("error", err.Error()))
			log.Fatalf("listen: %s\n", err)
		}
	}()

	logger.Info("HTTP server started successfully", slog.String("port", cfg.ServerPort))

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Received shutdown signal, shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", slog.String("error", err.Error()))
		log.Fatal("Server forced to shutdown:", err)
	}
	logger.Info("Server shutdown completed successfully")
}

func RequestLoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			slog.String("method", param.Method),
			slog.String("path", param.Path),
			slog.Int("status", param.StatusCode),
			slog.Duration("latency", param.Latency),
			slog.String("client_ip", param.ClientIP),
			slog.String("user_agent", param.Request.UserAgent()),
		)
		return ""
	})
}

func NewGormLogger(logger *slog.Logger) *GormLogger {
	return &GormLogger{logger: logger}
}

type GormLogger struct {
	logger *slog.Logger
}

func (l *GormLogger) LogMode(logger.LogLevel) logger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.InfoContext(ctx, msg, slog.Any("data", data))
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.WarnContext(ctx, msg, slog.Any("data", data))
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.ErrorContext(ctx, msg, slog.Any("data", data))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		l.logger.ErrorContext(ctx, "Database query error",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Duration("elapsed", elapsed),
			slog.String("error", err.Error()))
	} else {
		l.logger.DebugContext(ctx, "Database query",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Duration("elapsed", elapsed))
	}
}
