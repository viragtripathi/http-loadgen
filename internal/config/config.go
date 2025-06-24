package config

import (
	"fmt"
	"os"
	"log"

	"github.com/viragtripathi/http-loadgen/internal/request"
	"gopkg.in/yaml.v3"
)

type Config struct {
	API struct {
		WriteAPI string `yaml:"write_api"`
		ReadAPI  string `yaml:"read_api"`
	} `yaml:"api"`

	Workload struct {
		Concurrency       int  `yaml:"concurrency"`
		ChecksPerSecond   int  `yaml:"checks_per_second"`
		ReadRatio         int  `yaml:"read_ratio"`
		DurationSec       int  `yaml:"duration_sec"`
		MaxRetries        int  `yaml:"max_retries"`
		RetryDelayMillis  int  `yaml:"retry_delay_ms"`
		RequestTimeoutSec int  `yaml:"request_timeout_sec"`
		MaxOpenConns      int  `yaml:"max_open_conns"`
		MaxIdleConns      int  `yaml:"max_idle_conns"`
		Verbose           bool `yaml:"verbose"`
	} `yaml:"workload"`

	Requests struct {
		WriteTemplate request.TemplatedRequest `yaml:"write"`
		ReadTemplate  request.TemplatedRequest `yaml:"read"`
	} `yaml:"requests"`
}

var AppConfig Config

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("failed to read config file: %v", err)
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &AppConfig); err != nil {
		log.Printf("failed to unmarshal config: %v", err)
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
