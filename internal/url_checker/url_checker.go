package url_checker

import (
	"URLChecker/internal/config"
	"URLChecker/internal/logger"
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
)

type URLChecker struct {
	Client      *http.Client
	RateLimiter *rate.Limiter
	loggers     *logger.Loggers
}

func NewURLChecker(cfg config.Checker, l *logger.Loggers) *URLChecker {
	r := rate.NewLimiter(rate.Every(cfg.RateLimit), 1)
	client := &http.Client{
		Timeout: cfg.Timeout,
	}
	return &URLChecker{
		Client:      client,
		RateLimiter: r,
		loggers:     l,
	}
}

func (c *URLChecker) StartCheck(ctx context.Context, wg *sync.WaitGroup, resChan chan fmt.Stringer, urlString string) {
	for {
		// Wait(ctx context.Context) returns non-nil error when Context is canceled,
		// or the expected wait time exceeds the Context's Deadline.
		err := c.RateLimiter.Wait(ctx)
		if err != nil {
			c.loggers.Info(fmt.Sprintf("stopping monitoring %s", urlString))
			wg.Done()
			return
		}
		resChan <- Check(c.Client, urlString)
	}
}

func (c *URLChecker) StartChecks(ctx context.Context, wg *sync.WaitGroup, resChan chan fmt.Stringer, urls []string) {
	for _, urlString := range urls {
		go c.StartCheck(ctx, wg, resChan, urlString)
	}
}

func Check(c *http.Client, urlString string) fmt.Stringer {
	startTime := time.Now()
	res, err := c.Get(urlString)
	responseTime := time.Since(startTime)

	// Documentation: "Any returned error will be of type url.Error"
	if err != nil {
		return &pingError{
			URL: urlString,
			Err: err,
		}
	}

	return &CheckResult{
		StatusCode:   res.StatusCode,
		URL:          urlString,
		ResponseTime: responseTime,
	}
}

type CheckResult struct {
	StatusCode   int
	URL          string
	ResponseTime time.Duration
}

func (r *CheckResult) String() string {
	return fmt.Sprintf("URL: %s, Respose Time: %dms, StatusCode: %d",
		r.URL, r.ResponseTime.Milliseconds(), r.StatusCode)
}

type pingError struct {
	URL string
	Err error
}

func (e pingError) String() string {
	return fmt.Sprintf("URL: %s, Error: %v", e.URL, e.Err.Error())
}
