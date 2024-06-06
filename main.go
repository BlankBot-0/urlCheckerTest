package main

import (
	"URLChecker/internal/config"
	"URLChecker/internal/logger"
	"URLChecker/internal/url_checker"
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	cfg := config.MustLoad()
	rate.NewLimiter(rate.Every(cfg.RateLimit), 1)
	logInfo := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	logError := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	logWarn := log.New(os.Stderr, "WARN\t", log.Ldate|log.Ltime|log.Lshortfile)

	loggers := logger.New(logInfo, logWarn, logError)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	checker := url_checker.NewURLChecker(cfg.Checker, loggers)
	messages := make(chan fmt.Stringer)

	var wg sync.WaitGroup
	wg.Add(len(cfg.URLs))
	checker.StartChecks(ctx, &wg, messages, cfg.URLs)

	go func() {
		for msg := range messages {
			loggers.Info(msg.String())
		}
	}()

	<-ctx.Done()
	loggers.Info("shutting down...")

	wg.Wait()
	loggers.Info("shut down successfully")
}
