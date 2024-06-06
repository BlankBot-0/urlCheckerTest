package main

import (
	"URLChecker/internal/config"
	"URLChecker/internal/logger"
	"URLChecker/internal/url_checker"
	"log"
	"os"
)

func main() {
	cfg := config.MustLoad()

	logInfo := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	logError := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	logWarn := log.New(os.Stderr, "WARN\t", log.Ldate|log.Ltime|log.Lshortfile)

	loggers := logger.New(logInfo, logWarn, logError)

	checker := url_checker.NewURLChecker(cfg.Checker, loggers)
	messages := make(chan url_checker.CheckResult)
	checker.StartChecks(messages, cfg.URLs)
	//go checker.StartCheck(messages, cfg.URLs[0])
	for {
		select {
		case result := <-messages:
			loggers.Info(result.String())
		}
	}
}
