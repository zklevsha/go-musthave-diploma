package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"

	"github.com/zklevsha/go-musthave-diploma/internal/config"
	"github.com/zklevsha/go-musthave-diploma/internal/db"
	"github.com/zklevsha/go-musthave-diploma/internal/handler"
	"github.com/zklevsha/go-musthave-diploma/internal/processor"
)

func main() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	log.Println("INFO main starting server")
	config := config.GetConfig()

	//hiding DSN password (for logs)
	re := regexp.MustCompile(":[a-zA-Z]+@")
	dsnLog := re.ReplaceAllString(config.DSN, ":******@")

	log.Printf("INFO main server config: RunAddr: %s, AccrualURL: %s, DSN: %s",
		config.RunAddr, config.AccrualURL, dsnLog)

	// initiazing storage
	s := &db.DBConnector{DSN: config.DSN, Ctx: ctx}
	err := s.Init()
	if err != nil {
		log.Panicf("CRITICAL failed to init connection to database: %s", err.Error())
	}
	defer s.Close()

	//Starting order`s proccessor
	p := processor.Processor{
		Delay:   config.AccrualDelay,
		Ctx:     ctx,
		Wg:      &wg,
		Storage: s,
		Accrual: config.AccrualURL,
	}
	wg.Add(1)
	go p.Start()

	// Starting web server
	handler := handler.GetHandler(config, ctx, s)
	fmt.Printf("INFO main starting web server at %s\n", config.RunAddr)

	srv := &http.Server{
		Addr:    config.RunAddr,
		Handler: handler,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("CRITICAL Failed to start web server: %s\n", err)
		}
	}()
	log.Print("INFO server Started\n")

	// Handling shutdown
	sig := <-done
	log.Printf("INFO main got a signal '%v', start shutting down...\n", sig)
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("ERROR server shutdown Failed:%+v", err)
	}
	cancel()
	wg.Wait()
	log.Print("INFO server exited properly")
}
