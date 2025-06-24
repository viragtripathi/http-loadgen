package request

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"text/template"
	"time"
	"log"

	"github.com/viragtripathi/http-loadgen/internal/metrics"
)

type TemplatedRequest struct {
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Body    string            `yaml:"body"`
	Headers map[string]string `yaml:"headers"`
}

type ClientOptions struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxOpenConns        int
	RequestTimeoutSec   int
}

var (
	sharedClient *http.Client
	clientOnce   sync.Once
)

func InitClientFromConfig(opts ClientOptions) {
	clientOnce.Do(func() {
		tr := &http.Transport{
			MaxIdleConns:        opts.MaxIdleConns,
			MaxIdleConnsPerHost: opts.MaxIdleConnsPerHost,
			MaxConnsPerHost:     opts.MaxOpenConns,
			IdleConnTimeout:     300 * time.Second,
		}
		sharedClient = &http.Client{
			Transport: tr,
			Timeout:   time.Duration(opts.RequestTimeoutSec) * time.Second,
		}
	})
}

func ExecuteWithTemplate(req TemplatedRequest, data map[string]any, funcMap template.FuncMap) (*http.Response, error) {
	if sharedClient == nil {
		return nil, fmt.Errorf("HTTP client not initialized — call InitClientFromConfig() first")
	}

	urlRendered, err := renderTemplate("url", req.URL, data, funcMap)
	if err != nil {
		return nil, fmt.Errorf("failed to render URL template: %w", err)
	}
	bodyRendered, err := renderTemplate("body", req.Body, data, funcMap)
	if err != nil {
		return nil, fmt.Errorf("failed to render body template: %w", err)
	}

	httpReq, err := http.NewRequest(req.Method, urlRendered, bytes.NewBufferString(bodyRendered))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	return sharedClient.Do(httpReq)
}

func ExecuteWithRetry(req TemplatedRequest, data map[string]any, funcMap template.FuncMap, maxRetries int, retryDelayMillis int) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		start := time.Now()

		resp, err = ExecuteWithTemplate(req, data, funcMap)
		metrics.RetryAttempts.Inc()

        // Consider retryable:
        // - Network error (err != nil)
        // - Status >= 500
        // - Status == 429 (Too Many Requests)
		status := getStatus(resp)
		shouldRetry := err != nil || status >= 500 || status == 429

		if !shouldRetry {
			log.Printf("✅ Request succeeded after %d attempt(s) (status=%v)", attempt, status)
			metrics.RetrySuccess.Inc()
			metrics.RetryDuration.Observe(time.Since(start).Seconds())
			return resp, nil
		}

        // If retry is needed and more attempts remain
		if attempt < maxRetries {
			log.Printf("🔁 Retry %d: Request failed (status=%v, error=%v)", attempt, status, err)
			time.Sleep(time.Duration(retryDelayMillis) * time.Millisecond)
		}
	}

	// All retries failed
	log.Printf("❌ Final failure after %d retries: %v", maxRetries, err)
	return resp, err
}

func getStatus(resp *http.Response) int {
	if resp != nil {
		return resp.StatusCode
	}
	return 0
}

func renderTemplate(name, tmpl string, data map[string]any, funcMap template.FuncMap) (string, error) {
	t, err := template.New(name).Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
