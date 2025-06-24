#!/bin/bash
set -e

if [[ "$1" == "--help" ]]; then
  echo ""
  echo "📦 run.sh: Start environment and run load test or benchmark matrix"
  echo ""
  echo "Usage:"
  echo "  ./scripts/run.sh [--benchmark] [--mode local|cloud]"
  echo ""
  echo "Options:"
  echo "  --benchmark     Run predefined benchmark matrix"
  echo "  --mode          Select mode: 'local' (default) or 'cloud'"
  echo ""
  exit 0
fi

MODE="local"
BENCHMARK_MODE=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --mode)
      MODE="$2"
      shift 2
      ;;
    --benchmark)
      BENCHMARK_MODE=true
      shift
      ;;
    *)
      shift
      ;;
  esac
done

TIMEOUT_CMD="timeout"
if ! command -v timeout >/dev/null 2>&1; then
  if command -v gtimeout >/dev/null 2>&1; then
    TIMEOUT_CMD="gtimeout"
  else
    echo "❌ Neither 'timeout' nor 'gtimeout' found. Please install one (e.g. 'brew install coreutils')"
    exit 1
  fi
fi

# Optional: used by docker-compose
if [[ "$MODE" == "cloud" ]]; then
  export API_CONFIG_PATH=./api/config.cloud.yaml
else
  export API_CONFIG_PATH=./api/config.local.yaml
fi

APP_BINARY="./http-loadgen"
OUTPUT_CSV="./benchmark_results.csv"

echo "🧼 Cleaning up old containers..."
docker-compose down -v --remove-orphans

echo "🚀 Starting containers (mode: $MODE)..."
docker-compose up -d

echo "⏳ Waiting for environment to stabilize..."
sleep 5

if [[ "$BENCHMARK_MODE" == true ]]; then
  echo "📈 Running benchmark matrix..."
  echo "timestamp,duration_sec,concurrency,checks_per_sec,read_ratio,allowed,denied,writes,reads,failed" > "$OUTPUT_CSV"

matrix=(
  "30 5 500 100"
  "45 10 1000 100"
  "60 10 1000 10"
)

  for row in "${matrix[@]}"; do
    read -r DURATION CONC CHECKS RATIO <<< "$row"
    echo "🔄 Benchmark: ${DURATION}s, ${CONC} workers, ${CHECKS} checks/sec, ${RATIO}:1"

    LOG="bench_${DURATION}s_${CONC}_${RATIO}.log"
    START=$(date +%s)

    timeout $((DURATION + 30)) bash -c "
      $APP_BINARY \
        --duration-sec=$DURATION \
        --concurrency=$CONC \
        --checks-per-second=$CHECKS \
        --read-ratio=$RATIO \
        --max-retries=3 \
        --retry-delay=200 \
        --max-open-conns=200 \
        --max-idle-conns=200 \
        --request-timeout=10 \
        --workload-config=./config/config.yaml \
        --log-file=$LOG \
        --verbose=false
    "
    EXIT=$?
    if [[ "$EXIT" -ne 0 ]]; then
      echo "⚠️  http-loadgen exited with code $EXIT"
    fi

    END=$(date +%s)
    ELAPSED=$((END - START))

    ALLOWED=$(grep "📈 Allowed" "$LOG" | awk '{print $NF}' || echo "0")
    DENIED=$(grep "📉 Denied" "$LOG" | awk '{print $NF}' || echo "0")
    WRITES=$(grep "📤 Writes" "$LOG" | awk '{print $NF}' || echo "0")
    READS=$(grep "👁️  Reads" "$LOG" | awk '{print $NF}' || echo "0")
    FAILED=$(grep "🚨 Failed writes" "$LOG" | awk '{print $NF}' || echo "0")

    echo "$(date +%Y-%m-%dT%H:%M:%S),${DURATION},${CONC},${CHECKS},${RATIO},${ALLOWED},${DENIED},${WRITES},${READS},${FAILED}" >> "$OUTPUT_CSV"
    echo "✅ Done in ${ELAPSED}s → allowed=${ALLOWED} denied=${DENIED} writes=${WRITES} reads=${READS}"

    echo "♻️ Cooldown between runs..."
    sleep 5
  done

  echo "📊 Benchmark complete. Results written to $OUTPUT_CSV"
else
  echo "🔥 Running single workload..."
  $APP_BINARY \
    --duration-sec=30 \
    --concurrency=10 \
    --checks-per-second=1000 \
    --read-ratio=100 \
    --max-retries=3 \
    --retry-delay=200 \
    --max-open-conns=100 \
    --max-idle-conns=100 \
    --request-timeout=10 \
    --workload-config=./config/config.yaml \
    --log-file=run.log \
    --verbose=false

  echo "✅ Workload completed. See run.log for details."
fi
