package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"erm-dokter/internal/config"
	"erm-dokter/internal/di"
	"erm-dokter/internal/pkg/database"
	"erm-dokter/internal/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New()

	if err := cfg.Validate(); err != nil {
		log.Error("Konfigurasi tidak valid: %v", err)
		os.Exit(1)
	}

	db, err := database.InitMySQL(
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, log,
	)
	if err != nil {
		log.Error("Database MySQL connection failed: %v", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("Berhasil terhubung ke database MySQL")

	routeConfig := di.ProvideRouteConfig(db, cfg, log)
	routeConfig.Setup()
	handler := routeConfig.BuildHandler(cfg.CORSOrigin)

	serverAddr := fmt.Sprintf(":%s", cfg.AppPort)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("Starting server '%s' (%s mode) on %s...", cfg.AppName, cfg.AppEnv, serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	log.Info("Server stopped gracefully")
}
