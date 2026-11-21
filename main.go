package main

import (
	dbconnection "PlataTest/internal/db_connection"
	"PlataTest/internal/repository"
	"PlataTest/internal/service"
	"PlataTest/internal/transport"
	"PlataTest/internal/worker"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const clientTimeout = 10 * time.Second
const shutdownTimeout = 5 * time.Second

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Printf("db dsn is not set")
		return
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Printf("api key is not set")
		return

	}
	pool, err := dbconnection.NewPool(ctx, dsn)
	if err != nil {
		log.Printf("connect postgres: %v", err)
		return
	}
	defer pool.Close()
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		log.Printf("create migration: %v", err)
		return
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Printf("migrate db: %v", err)
		return
	}
	m.Close()

	rep := repository.NewRepository(pool)
	ser := service.NewService(rep)
	handlers := transport.NewHandler(ser)
	r := transport.SetupRouter(handlers)
	client := &http.Client{
		Timeout: clientTimeout,
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := worker.RunPriceUpdater(
			ctx,
			rep,
			client,
			apiKey,
		)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("update worker: %v", err)
		}
	}()
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	go func() {
		log.Print("server is ready to accept conns")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen and serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("shutdown signal recieved")

	stCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(stCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
	wg.Wait()
	log.Printf("server stopped")
}
