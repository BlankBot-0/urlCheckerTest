package url_checker

import (
	"URLChecker/internal/config"
	"URLChecker/internal/logger"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"
)

type URLChecker struct {
	RateLimit          time.Duration
	BackoffCoefficient int
	MaxDelay           time.Duration
	l                  *logger.Loggers
}

func NewURLChecker(cfg config.Checker, l *logger.Loggers) *URLChecker {
	return &URLChecker{
		RateLimit:          cfg.RateLimit,
		BackoffCoefficient: cfg.BackoffCoefficient,
		MaxDelay:           cfg.MaxDelay,
		l:                  l,
	}
}

func (c *URLChecker) StartCheck(resChan chan CheckResult, urlString string) {
	rateLimit := c.RateLimit
	for {
		result, err := Check(urlString)
		if err != nil && rateLimit < c.MaxDelay {
			rateLimit = c.EvalRateLimitJitter(rateLimit)
			c.l.Warn(fmt.Sprintf("Couldn't reach %s, retrying after %ds", urlString, rateLimit))
		} else if err == nil {
			rateLimit = c.RateLimit
		}
		if rateLimit > c.MaxDelay {
			rateLimit = c.MaxDelay
		}

		resChan <- *result
		time.Sleep(rateLimit)
	}
}

func (c *URLChecker) StartChecks(resChan chan CheckResult, urls []string) {
	for _, urlString := range urls {
		go c.StartCheck(resChan, urlString)
	}
}

func (c *URLChecker) EvalRateLimitJitter(rateLimit time.Duration) time.Duration {
	seconds := rateLimit.Seconds()
	jitter := rand.NormFloat64() * 0.1 * seconds
	return time.Duration(seconds*float64(c.BackoffCoefficient) + jitter)
}

func Check(urlString string) (*CheckResult, error) {
	startTime := time.Now()
	res, err := http.Get(urlString)
	responseTime := time.Since(startTime)

	// Documentation: "Any returned error will be of type url.Error"
	if err != nil {
		return &CheckResult{
			StatusCode:   0,
			URL:          urlString,
			ResponseTime: 0,
			ErrMessage:   err.Error(),
		}, err
	}

	return &CheckResult{
		StatusCode:   res.StatusCode,
		URL:          urlString,
		ResponseTime: responseTime,
	}, err
}

type CheckResult struct {
	StatusCode   int
	URL          string
	ResponseTime time.Duration
	ErrMessage   string
}

func (r *CheckResult) String() string {
	if r.ErrMessage == "" {
		return fmt.Sprintf("URL: %s, Respose Time: %dms, StatusCode: %d", r.URL, r.ResponseTime.Milliseconds(), r.StatusCode)
	}
	return fmt.Sprintf("URL: %s, Error: %s",
		r.URL, r.ErrMessage)
}
