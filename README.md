# http-loadgen

![Latest Release](https://img.shields.io/github/v/release/viragtripathi/http-loadgen)
![Build](https://github.com/viragtripathi/http-loadgen/actions/workflows/release.yml/badge.svg)
![Go Version](https://img.shields.io/badge/go-1.24-blue)
![License](https://img.shields.io/github/license/viragtripathi/http-loadgen)
![Docker Pulls](https://img.shields.io/docker/pulls/virag/http-loadgen)

![http-loadgen banner](http-loadgen.png)

---

`http-loadgen` is a high-throughput, configurable **HTTP benchmarking and workload simulation tool**. It can test any read/write HTTP API using configurable templates, retry logic, and Prometheus metrics.

---

## 🚀 Quick Start

```bash
./samples/ory/keto/scripts/run.sh
```

This will launch a sample environment (Keto + HAProxy) and run a default test.

---

## 📊 Benchmark Matrix

```bash
./samples/ory/keto/scripts/run.sh --benchmark
```

Runs a predefined matrix of configurations. Results are saved to:

```
samples/ory/keto/scripts/benchmark_results.csv
```

---

## 🧪 Examples

### 🔐 Ory Keto (permission system)

Sample config: [`samples/ory/keto/config/config.yaml`](samples/ory/keto/config/config.yaml)

Run:

```bash
./http-loadgen \
  --workload-config=samples/ory/keto/config/config.yaml \
  --log-file=run.log
```

---

### 🧪 Fake API (no dependencies)

Includes a built-in `read`/`write` test server. Start both with:

```bash
make run-fake
```

Or dry-run without sending real requests:

```bash
./http-loadgen --workload-config=samples/fakeapi/config/dry-run.yaml --dry-run
```

---

## 📁 Folder Layout

```
cmd/              # CLI entrypoint (main.go)
config/           # Example default loadgen config
samples/          # Test suites: ory/keto, fakeapi, hydra, etc.
internal/         # Core engine: request execution, metrics, config
scripts/          # Optional root-level scripts
```

---

## 🔧 CLI Flags

| Flag                  | Description                              |
|-----------------------|------------------------------------------|
| `--duration-sec`      | Duration to run the test                 |
| `--concurrency`       | Number of concurrent workers             |
| `--checks-per-second` | Max read requests per second             |
| `--read-ratio`        | Read to write ratio (e.g., 100 = 100:1)  |
| `--workload-config`   | Path to YAML config file                 |
| `--log-file`          | Where to write logs                      |
| `--verbose`           | Enable detailed logging                  |
| `--max-retries`       | Retry attempts per request               |
| `--retry-delay`       | Delay between retries (ms)               |
| `--request-timeout`   | Timeout per HTTP request (sec)           |
| `--max-open-conns`    | Max HTTP connections                     |
| `--max-idle-conns`    | Max idle connections                     |
| `--serve-metrics`     | Keep Prometheus metrics server alive     |
| `--dry-run`           | Run logic without making real HTTP calls |

---

## 📦 Build

```bash
make build
```

---

## 🍎 macOS Gatekeeper (Quarantine) Fix

After downloading a macOS binary, you may need:

```bash
xattr -d com.apple.quarantine ./http-loadgen_darwin_arm64
chmod +x ./http-loadgen_darwin_arm64
./http-loadgen_darwin_arm64 --help
```

Or allow via:
>  System Settings → Privacy & Security → Allow Anyway

---

## 🐳 Docker Usage

### 🔧 Build locally

```bash
docker build -t http-loadgen:latest .
```

### 🧪 Run with built-in config

```bash
docker run --rm virag/http-loadgen:latest \
  --workload-config=/app/config/config.yaml \
  --log-file=/app/run.log
```

### ⚙️ Mount local config

```bash
docker run --rm \
  -v $(pwd)/config:/app/config \
  virag/http-loadgen:latest \
  --workload-config=/app/config/custom.yaml \
  --log-file=/app/run.log
```

### 🔍 Dry-run mode

```bash
docker run --rm virag/http-loadgen:latest \
  --workload-config=/app/config/config.yaml \
  --dry-run
```

---

## 📈 Prometheus Metrics

Run with:

```bash
--serve-metrics
```

Then scrape:

```
http://localhost:2112/metrics
```

Exposed metrics include:

- `retry_attempts_total`
- `retry_success_total`
- `permission_check_total`
- `retry_duration_seconds`

---

## 📖 Why Load Test APIs

`http-loadgen` tests actual API behavior, not just databases — including:

- Latency
- Rate limiting
- Retry behavior
- API performance under concurrency

It mirrors how real clients behave.

---

## 📚 References

- [Ory Keto](https://www.ory.sh/docs/keto)
- [CockroachDB](https://www.cockroachlabs.com/docs/)
- [Prometheus](https://prometheus.io/)
