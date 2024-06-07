package url_checker

import (
	"URLChecker/internal/config"
	"URLChecker/internal/logger"
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestCheck(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	client := &http.Client{Timeout: time.Second}
	msg := Check(client, ts.URL).String()
	expected := &CheckResult{
		StatusCode:   200,
		URL:          ts.URL,
		ResponseTime: 0,
	}
	if msg != expected.String() {
		t.Errorf("got %s, expected %s", msg, expected.String())
	}

	redirectServers := make([]*httptest.Server, 11)
	redirectServers[0] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for i := range 10 {
		redirectServers[i+1] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, redirectServers[i].URL, http.StatusTemporaryRedirect)
		}))
	}
	defer func() {
		for i := range redirectServers {
			redirectServers[i].Close()
		}
	}()

	msg = Check(client, redirectServers[10].URL).String()
	expectedErr := &pingError{
		URL: redirectServers[10].URL,
		Err: fmt.Errorf("Get \"%s\": stopped after 10 redirects", redirectServers[0].URL),
	}
	if msg != expectedErr.String() {
		t.Errorf("got %s, expected %s", msg, expectedErr.String())
	}
}

func TestCheckTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer ts.Close()

	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	msg := Check(client, ts.URL).String()
	expectedError := fmt.Errorf("Get \"%s\": context deadline exceeded (Client.Timeout exceeded while awaiting headers)",
		ts.URL)
	expected := pingError{
		URL: ts.URL,
		Err: expectedError,
	}
	if msg != expected.String() {
		t.Errorf("got %s, expected %s", msg, expected.String())
	}
}

func TestURLChecker(t *testing.T) {
	httpStatuses := []int{
		http.StatusOK,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusTemporaryRedirect,
	}
	servers := make([]*httptest.Server, 11)
	for i := range servers {
		status := httpStatuses[i%len(httpStatuses)]

		servers[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if i%len(httpStatuses) == 3 {
				http.Redirect(w, r, servers[0].URL, http.StatusTemporaryRedirect)
			} else {
				w.WriteHeader(status)
			}
		}))
	}

	urls := make([]string, len(servers))
	for i := range urls {
		urls[i] = servers[i].URL
	}
	urlResults := make(map[string]fmt.Stringer)
	for i, url := range urls {
		status := httpStatuses[i%len(httpStatuses)]

		if i%len(httpStatuses) == 3 {
			status = http.StatusOK
		}
		urlResults[url] = &CheckResult{
			StatusCode:   status,
			URL:          servers[i].URL,
			ResponseTime: 0,
		}
	}

	cfg := config.Config{
		Env: "local",
		Checker: config.Checker{
			URLs:      urls,
			RateLimit: time.Second,
			Timeout:   2 * time.Second,
		},
	}
	rate.NewLimiter(rate.Every(cfg.RateLimit), 1)
	loggers := logger.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	checker := NewURLChecker(cfg.Checker, loggers)
	messages := make(chan fmt.Stringer)

	var wg sync.WaitGroup
	wg.Add(len(cfg.URLs))
	checker.StartChecks(ctx, &wg, messages, cfg.URLs)

	go func() {
		for msg := range messages {
			tmpUrl := msg.String()[5:27]
			expectedReport := urlResults[tmpUrl]
			expected, ok1 := expectedReport.(*CheckResult)
			msgResult, ok2 := msg.(*CheckResult)
			if ok1 && ok2 {
				expected.ResponseTime = msgResult.ResponseTime
			}
			if urlResults[tmpUrl].String() != msg.String() {
				t.Errorf("got %s, expected %s", msg.String(), urlResults[tmpUrl].String())
			}
			loggers.Info(msg.String())
		}
	}()

	go func() {
		time.Sleep(6 * time.Second)
		cancel()
	}()

	<-ctx.Done()
	wg.Wait()
	loggers.Info("shut down successfully")
}
