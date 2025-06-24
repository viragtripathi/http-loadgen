package generator

import (
	"testing"

	"github.com/viragtripathi/http-loadgen/internal/config"
)

func TestRunGenerator_DryRun(t *testing.T) {
	config.AppConfig.Workload.DurationSec = 1
	config.AppConfig.Workload.Concurrency = 2
	config.AppConfig.Workload.ChecksPerSecond = 100
	config.AppConfig.Workload.ReadRatio = 2
	config.AppConfig.Workload.MaxRetries = 1
	config.AppConfig.Workload.RetryDelayMillis = 10
	config.AppConfig.Workload.RequestTimeoutSec = 1
	config.AppConfig.Workload.MaxOpenConns = 10
	config.AppConfig.Workload.MaxIdleConns = 10

	RunGenerator(true)
}
