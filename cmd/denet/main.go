package main

import (
	"context"
	"denet/internal/config"
	"denet/internal/http-server/handlers/info"
	"denet/internal/http-server/handlers/leaderboard"
	"denet/internal/http-server/handlers/login"
	"denet/internal/http-server/handlers/referrer"
	"denet/internal/http-server/handlers/task"
	"denet/internal/http-server/handlers/users/save"
	middlewares "denet/internal/http-server/middleware"
	"denet/internal/lib/logger/sl"
	"denet/internal/storage/postgres"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	storage, err := postgres.New(cfg.StoragePath)
	if err != nil {
		log.Error("Failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/users/login", login.NewLogin(log, storage))

	router.Route("/users/", func(r chi.Router) {
		r.Use(middlewares.ValidateJWT)
		r.Post("/create", save.New(log, storage))
		r.Get("/{id}/status", info.NewUserInfo(log, storage))
		r.Get("/leaderboard", leaderboard.NewLeaderboard(log, storage))
		r.Post("/{id}/task/complete", task.NewTask(log, storage))
		r.Post("/{id}/task/referrer", referrer.NewReferalTask(log, storage))
	})

	// router.Post("/users", save.New(log, storage))
	// router.Get("/users/{id}/status", info.NewUserInfo(log, storage))
	// router.Get("/users/leaderboard", leaderboard.NewLeaderboard(log, storage))
	// router.Post("/users/{id}/task/complete", task.NewTask(log, storage))
	// router.Post("/users/{id}/task/referrer", referrer.NewReferalTask(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}
	log.Info("server stopped")
	// if err := srv.ListenAndServe(); err != nil {
	// 	log.Error("failed to start server")
	// }
	// log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	fmt.Println(env)
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
