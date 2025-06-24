package generator

import (
	"log"
	"math/rand"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/viragtripathi/http-loadgen/internal/config"
	"github.com/viragtripathi/http-loadgen/internal/metrics"
	"github.com/viragtripathi/http-loadgen/internal/request"
)

type tuple struct {
	Subject string
	Object  string
}

func RunGenerator(dryRun bool) {
	cfg := config.AppConfig.Workload
	duration := time.Duration(cfg.DurationSec) * time.Second
	endTime := time.Now().Add(duration)

	// Initialize HTTP client with pooling settings
	request.InitClientFromConfig(request.ClientOptions{
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConns,
		MaxOpenConns:        cfg.MaxOpenConns,
		RequestTimeoutSec:   cfg.RequestTimeoutSec,
	})

	writeWorkers := 1
	readWorkers := cfg.ReadRatio
	totalWorkers := writeWorkers + readWorkers

	var wg sync.WaitGroup
	tupleChan := make(chan tuple, 10000)

	var allowedCount, deniedCount, failedWrites, readCount, writeCount int64

	log.Printf("🚧 Load generation for %v with %d total workers (%d writers, %d readers)...",
		duration, totalWorkers, writeWorkers, readWorkers)

	funcMap := templateFunctions()

	var progressLock sync.Mutex
	lastPrint := time.Now()

	// Phase 1: Start write workers
	for i := 0; i < writeWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for time.Now().Before(endTime) {
				objectID := uuid.New().String()
				subjectID := uuid.New().String()
				subjectFull := "user:" + subjectID

				if !dryRun {
					data := map[string]any{
						"object":     objectID,
						"subject":    subjectID,
						"subject_id": subjectFull,
						"uuid":       funcMap["uuid"],
						"timestamp":  funcMap["timestamp"],
						"randInt":    funcMap["randInt"],
						"WriteAPI":   config.AppConfig.API.WriteAPI,
					}

					resp, err := request.ExecuteWithRetry(
						config.AppConfig.Requests.WriteTemplate,
						data,
						funcMap,
						cfg.MaxRetries,
						cfg.RetryDelayMillis,
					)

					if err != nil || resp.StatusCode >= 300 {
						log.Printf("❌ Write failed: %v", err)
						failedWrites++
					} else {
						for j := 0; j < cfg.ReadRatio; j++ {
							tupleChan <- tuple{Subject: subjectID, Object: objectID}
						}
						writeCount++
					}
				}

				progressLock.Lock()
				if time.Since(lastPrint) > 10*time.Second {
					log.Printf("📊 Progress: writes=%d reads=%d allowed=%d denied=%d failed=%d",
						writeCount, readCount, allowedCount, deniedCount, failedWrites)
					lastPrint = time.Now()
				}
				progressLock.Unlock()
			}
		}(i)
	}

	// Phase 2: Start read workers
	for i := 0; i < readWorkers; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for time.Now().Before(endTime) {
				select {
				case t := <-tupleChan:
					if !dryRun {
						data := map[string]any{
							"object":     t.Object,
							"subject":    t.Subject,
							"subject_id": "user:" + t.Subject,
							"uuid":       funcMap["uuid"],
							"timestamp":  funcMap["timestamp"],
							"randInt":    funcMap["randInt"],
							"ReadAPI":    config.AppConfig.API.ReadAPI,
						}

						resp, err := request.ExecuteWithRetry(
							config.AppConfig.Requests.ReadTemplate,
							data,
							funcMap,
							cfg.MaxRetries,
							cfg.RetryDelayMillis,
						)

						if err != nil {
							log.Printf("❌ Read failed: %v", err)
							deniedCount++
						} else {
							if resp.StatusCode == 200 {
								metrics.PermissionCheckCounter.WithLabelValues("allowed").Inc()
								allowedCount++
							} else {
								metrics.PermissionCheckCounter.WithLabelValues("denied").Inc()
								deniedCount++
							}
							resp.Body.Close()
						}
						readCount++
					}

					progressLock.Lock()
					if time.Since(lastPrint) > 10*time.Second {
						log.Printf("📊 Progress: writes=%d reads=%d allowed=%d denied=%d failed=%d",
							writeCount, readCount, allowedCount, deniedCount, failedWrites)
						lastPrint = time.Now()
					}
					progressLock.Unlock()
				default:
					time.Sleep(5 * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()

	log.Println("✅ Load generation and permission checks complete")
	log.Printf("⏱️ Duration: %v", duration)
	log.Printf("⚙️  Concurrency: %d", totalWorkers)
	log.Printf("🚦 Checks/sec:  %d", cfg.ChecksPerSecond)
	log.Printf("🧪 Mode:        %s", map[bool]string{true: "DRY RUN", false: "LIVE"}[dryRun])
	log.Printf("📈 Allowed:     %d", allowedCount)
	log.Printf("📉 Denied:      %d", deniedCount)
	log.Printf("📤 Writes:      %d", writeCount)
	log.Printf("👁️  Reads:      %d", readCount)
	if writeCount > 0 {
		log.Printf("📊 Read/Write ratio: %.1f:1", float64(readCount)/float64(writeCount))
	}
	log.Printf("🚨 Failed writes: %d", failedWrites)

	if dryRun {
		log.Println("⚠️  Dry-run mode: No requests were sent.")
	}
}

func templateFunctions() template.FuncMap {
	return template.FuncMap{
		"uuid": func() string {
			return uuid.New().String()
		},
		"timestamp": func() string {
			return time.Now().Format(time.RFC3339)
		},
		"randInt": func(min, max int) string {
			return strconv.Itoa(rand.Intn(max-min+1) + min)
		},
	}
}
