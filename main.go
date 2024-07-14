package main

import (
	"URLChecker/internal/config"
	"URLChecker/internal/logger"
	"URLChecker/internal/url_checker"
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	cfg := config.MustLoad()
	rateLimiter := rate.NewLimiter(rate.Every(cfg.RateLimit), 1)
	loggers := logger.New(cfg)
	defer loggers.CloseTarget()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	checker := url_checker.NewURLChecker(cfg.Checker, loggers, rateLimiter)
	messages := make(chan fmt.Stringer)

	var wg sync.WaitGroup
	wg.Add(len(cfg.URLs))
	checker.StartChecks(ctx, &wg, messages, cfg.URLs)

	go func() {
		for msg := range messages {
			switch v := msg.(type) {
			case *url_checker.CheckResult:
				loggers.Info(v)
			case *url_checker.PingError:
				loggers.Warn(v)
			}
		}
	}()

	<-ctx.Done()
	loggers.Info("shutting down...")

	wg.Wait()
	loggers.Info("shut down successfully")
}
