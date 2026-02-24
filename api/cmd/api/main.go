package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"clothes-shop/api/internal/auth"
	"clothes-shop/api/internal/config"
	"clothes-shop/api/internal/db"
	"clothes-shop/api/internal/httpapi"
	"clothes-shop/api/internal/migrate"
	"clothes-shop/api/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer pool.Close()

	if cfg.AutoMigrate {
		exists, err := db.TableExists(ctx, pool, "roles")
		if err != nil {
			log.Fatalf("migration precheck error: %v", err)
		}
		if exists {
			log.Printf("auto-migrate enabled; schema appears present, skipping")
		} else {
			if err := migrate.Up(cfg.DatabaseURL, "/migrations"); err != nil {
				log.Fatalf("migration error: %v", err)
			}
		}
	}

	st := store.New(pool)
	authSvc := auth.NewService(cfg.JWTSecret, cfg.JWTAccessTTL)
	srv := httpapi.New(cfg, st, authSvc)

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      srv.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("api listening on %s", cfg.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

