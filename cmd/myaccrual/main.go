package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/zklevsha/go-musthave-diploma/internal/config"
	"github.com/zklevsha/go-musthave-diploma/internal/handler"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	log.Println("INFO main starting accural server")
	config := config.GetAccuralConfig()

	log.Printf("INFO main server config: RunAddr: %s", config.RunAddr)

	// Starting web server
	handler := handler.AccGetHandler(config, ctx)
	fmt.Printf("INFO main starting web server at %s\n", config.RunAddr)

	srv := &http.Server{
		Addr:    config.RunAddr,
		Handler: handler,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("CRITICAL failed to start web server: %s\n", err)
		}
	}()
	log.Print("INFO server Started\n")

	// Handling shutdown
	sig := <-done
	log.Printf("INFO main got a signal '%v', start shutting down...\n", sig)
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("ERROR server shutdown failed:%+v", err)
	}
	cancel()
	log.Print("INFO server exited properly")
}
