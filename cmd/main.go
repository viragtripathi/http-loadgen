package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/viragtripathi/http-loadgen/cmd/generator"
	"github.com/viragtripathi/http-loadgen/internal/config"
	"github.com/viragtripathi/http-loadgen/internal/metrics"
)

func safeStatus(resp *http.Response) int {
	if resp != nil {
		return resp.StatusCode
	}
	return -1
}

func main() {
	writeAPI := flag.String("write-api", "", "Base URL for Write API (overrides config)")
    readAPI := flag.String("read-api", "", "Base URL for Read API (overrides config)")
	concurrency := flag.Int("concurrency", 0, "Override number of concurrent workers")
	checksPerSecond := flag.Int("checks-per-second", 0, "Override checks per second")
	duration := flag.Int("duration-sec", 0, "Override duration in seconds")
	readRatio := flag.Int("read-ratio", 0, "Override read/write ratio (e.g. 100 = 100:1)")
	dryRun := flag.Bool("dry-run", false, "Simulate workload without API calls")
	workloadConfig := flag.String("workload-config", "config/config.yaml", "Path to workload config")

	requestTimeout := flag.Int("request-timeout", 5, "Per-request timeout in seconds")
	maxRetries := flag.Int("max-retries", 3, "Override max retries for API calls")
	retryDelay := flag.Int("retry-delay", 200, "Override delay (ms) between retries")
	maxOpenConns := flag.Int("max-open-conns", 100, "Max open HTTP connections (default: 100)")
	maxIdleConns := flag.Int("max-idle-conns", 100, "Max idle HTTP connections (default: 100)")
	logFile := flag.String("log-file", "", "Path to log output file")
	serveMetrics := flag.Bool("serve-metrics", false, "Keep Prometheus metrics endpoint alive after run")
	verbose := flag.Bool("verbose", true, "Enable verbose logging")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `
📦 http-loadgen: Generic HTTP workload simulator

Usage:
  ./http-loadgen [flags]

Options:
  -write-api           Base URL for Write API (overrides config)
  -read-api            Base URL for Read API (overrides config)
  -workload-config     Path to workload config file (default: config/config.yaml)
  -concurrency         Number of concurrent workers (overrides config file)
  -checks-per-second   Max permission checks per second (overrides config file)
  -duration-sec        Run for this many seconds (default from config file)
  -read-ratio          Read-to-write ratio (e.g. 100 = 100 reads per write)
  -request-timeout     Per-request timeout in seconds
  -max-retries         Max retry attempts for API calls
  -retry-delay         Retry delay in milliseconds
  -max-open-conns      Max open HTTP connections
  -max-idle-conns      Max idle HTTP connections
  -log-file            Path to write logs to (default: stdout only)
  -serve-metrics       Keep Prometheus metrics endpoint alive after run
  -dry-run             Skip actual writes and permission checks
  -verbose             Enable verbose logging for reads/writes
`)
	}

	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	// Logging setup
	if *logFile != "" {
		f, err := os.Create(*logFile)
		if err != nil {
			log.Fatalf("❌ Failed to create log file: %v", err)
		}
		defer f.Close()

		if *verbose {
			log.SetOutput(io.MultiWriter(os.Stdout, f))
		} else {
			log.SetOutput(f)
		}
	} else if !*verbose {
		log.SetOutput(io.Discard)
	}

	// Load config file
	if err := config.LoadConfig(*workloadConfig); err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Set base URLs
    if *writeAPI != "" {
        config.AppConfig.API.WriteAPI = *writeAPI
    }
    if *readAPI != "" {
        config.AppConfig.API.ReadAPI = *readAPI
    }

    if config.AppConfig.API.WriteAPI == "" || config.AppConfig.API.ReadAPI == "" {
        log.Fatalf("❌ Both write-api and read-api must be specified via config or CLI")
    }

	// CLI overrides
	if *concurrency > 0 {
		config.AppConfig.Workload.Concurrency = *concurrency
	}
	if *checksPerSecond > 0 {
		config.AppConfig.Workload.ChecksPerSecond = *checksPerSecond
	}
	if *duration > 0 {
		config.AppConfig.Workload.DurationSec = *duration
	}
	if *readRatio > 0 {
		config.AppConfig.Workload.ReadRatio = *readRatio
	}
	if *requestTimeout > 0 {
		config.AppConfig.Workload.RequestTimeoutSec = *requestTimeout
	}
	if *maxRetries > 0 {
		config.AppConfig.Workload.MaxRetries = *maxRetries
	}
	if *retryDelay > 0 {
		config.AppConfig.Workload.RetryDelayMillis = *retryDelay
	}
	if *maxOpenConns > 0 {
		config.AppConfig.Workload.MaxOpenConns = *maxOpenConns
	}
	if *maxIdleConns > 0 {
		config.AppConfig.Workload.MaxIdleConns = *maxIdleConns
	}

	config.AppConfig.Workload.Verbose = *verbose

	// Readiness check
	if !*dryRun {
		healthURL := config.AppConfig.API.ReadAPI + "/health/alive"
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(healthURL)
		if err != nil || resp == nil || resp.StatusCode != 200 {
			log.Fatalf(`❌ Unable to reach Read API at %s

Details:
- Error: %v
- HTTP Status: %v
`, config.AppConfig.API.ReadAPI, err, safeStatus(resp))
		}
	}

	metrics.Init()
	generator.RunGenerator(*dryRun)

	if *serveMetrics {
		fmt.Println("📊 Prometheus metrics available at http://localhost:2112/metrics")
		fmt.Println("🔁 Waiting indefinitely for Prometheus to scrape. Ctrl+C to exit.")
		select {}
	}
}
