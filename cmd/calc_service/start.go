package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	application "github.com/superlogarifm/goCalc-v3/application"
)

func main() {
	app := application.NewApp()
	app.StartServer()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting.")
}
