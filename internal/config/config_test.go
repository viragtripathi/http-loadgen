package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmp := `api:
  write_api: "http://localhost:4467"
  read_api: "http://localhost:4466"
workload:
  concurrency: 5
  checks_per_second: 1000
  read_ratio: 100
  duration_sec: 10
  max_retries: 3
  retry_delay_ms: 200
  request_timeout_sec: 10
  max_open_conns: 100
  max_idle_conns: 100
`

	tmpFile := "test_config.yaml"
	if err := os.WriteFile(tmpFile, []byte(tmp), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if AppConfig.API.WriteAPI != "http://localhost:4467" {
		t.Errorf("unexpected write API: %s", AppConfig.API.WriteAPI)
	}
}
