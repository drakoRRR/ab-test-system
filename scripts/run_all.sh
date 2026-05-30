#!/usr/bin/env bash
# Master validation script — runs all diploma scenarios (A–E) in sequence.
#
# Prerequisites:
#   1. Backend running:     cd backend && make dev
#   2. Frontend running:    cd frontend && pnpm dev   (optional, for screenshots)
#   3. Demo server running: make demo-run
#   4. Seed completed:      make demo-setup
#
# Required env vars:
#   FIREBASE_TOKEN   — from DevTools → Cookies after login at localhost:3000
#   PROJECT_ID       — printed by make demo-setup
#   EXPERIMENT_ID    — printed by make demo-setup (checkout-btn experiment)
#
# Optional overrides:
#   API_URL          (default http://localhost:8080/api/v1)
#   DEMO_URL         (default http://localhost:8081)
#   SDK_URL          (default http://localhost:8080)
#   VUS              (default 50)
#   B_RUNS           (default 20)  — set to 10 to halve total time
#
# Usage:
#   export FIREBASE_TOKEN=... PROJECT_ID=... EXPERIMENT_ID=...
#   ./scripts/run_all.sh

set -euo pipefail

# ── Config ────────────────────────────────────────────────────────────────────
API_URL="${API_URL:-http://localhost:8080/api/v1}"
DEMO_URL="${DEMO_URL:-http://localhost:8081}"
SDK_URL="${SDK_URL:-http://localhost:8080}"
VUS="${VUS:-50}"
B_RUNS="${B_RUNS:-20}"
PROJECT_ID="${PROJECT_ID:?Set PROJECT_ID (printed by make demo-setup)}"
EXPERIMENT_ID="${EXPERIMENT_ID:?Set EXPERIMENT_ID (printed by make demo-setup)}"
FIREBASE_TOKEN="${FIREBASE_TOKEN:?Set FIREBASE_TOKEN (DevTools → Cookies)}"
AUTH_HEADER="Authorization: Bearer ${FIREBASE_TOKEN}"
SDK_API_KEY=$(grep SDK_API_KEY demo/.env 2>/dev/null | cut -d= -f2 || true)

RUN_ID=$(date +%Y%m%d_%H%M%S)
RESULTS="results/${RUN_ID}"
mkdir -p "${RESULTS}"

ST_Ed="SKIP"; ST_Db="SKIP"; ST_Dd="SKIP"; ST_Dac="SKIP"
ST_A="SKIP"; ST_C="SKIP"; ST_Eab="SKIP"; ST_B="SKIP"

# ── Helpers ───────────────────────────────────────────────────────────────────
info()    { echo ""; echo "▶ $*"; }
ok()      { echo "  ✓ $*"; }
fail()    { echo "  ✗ $*" >&2; }
require() {
  for cmd in "$@"; do
    command -v "${cmd}" &>/dev/null || { fail "missing: ${cmd}"; exit 1; }
  done
}

health_check() {
  local url=$1 label=$2
  if curl -sf "${url}" > /dev/null 2>&1; then
    ok "${label} is up"
  else
    fail "${label} is not responding at ${url}"
    exit 1
  fi
}

# ── Preflight ─────────────────────────────────────────────────────────────────
info "Preflight checks"
require k6 jq go curl
health_check "${SDK_URL}/health"            "Backend (${SDK_URL})"
health_check "${DEMO_URL}/health"           "Demo server (${DEMO_URL})"
[[ -n "${SDK_API_KEY}" ]] && ok "SDK_API_KEY found in demo/.env" \
  || { fail "demo/.env not found — run: make demo-setup"; exit 1; }
ok "Results directory: ${RESULTS}/"

# ── Scenario E.d — Go benchmark (fast, no network) ───────────────────────────
info "Scenario E.d — SDK benchmark (BenchmarkGetVariant, BenchmarkAssignBucket)"
BENCH_OUT="${RESULTS}/bench.txt"
(cd sdk && go test -bench=. -benchmem -count=5 -run='^$') | tee "${BENCH_OUT}"
ST_Ed="PASS"
ok "saved → ${BENCH_OUT}"

# ── Scenario D.b + D.d — SDK unit tests ───────────────────────────────────────
info "Scenario D.b — Sticky determinism (1000 users × 5 calls, algorithm level)"
info "Scenario D.d — Chi-square bucket uniformity (100k users → 10k buckets)"
SDK_TEST_OUT="${RESULTS}/sdk_tests.txt"
(cd sdk && go test \
  -run "TestEvaluateExperiment_StickyAcrossRepeatedCalls|TestAssignBucket_ChiSquareUniformity" \
  -v -count=1) | tee "${SDK_TEST_OUT}"
ST_Db="PASS"; ST_Dd="PASS"
ok "saved → ${SDK_TEST_OUT}"

# ── Scenario D.a + D.c — HTTP sticky test (demo server) ──────────────────────
info "Scenario D.a+D.c — Sticky bucketing via HTTP + distribution uniformity"
STICKY_OUT="${RESULTS}/sticky.txt"
(cd demo/backend && \
  DEMO_URL="${DEMO_URL}" EXPERIMENT_KEY=checkout-btn \
  go run ./cmd/sticky/) | tee "${STICKY_OUT}"
if grep -q "FAIL" "${STICKY_OUT}"; then
  ST_Dac="FAIL"
  fail "sticky test failed — check ${STICKY_OUT}"
else
  ST_Dac="PASS"
  ok "saved → ${STICKY_OUT}"
fi

# ── Scenario A — Known effect (etalon run) ────────────────────────────────────
info "Scenario A — Known effect (CR_control=10%, CR_treatment=12%, VUS=${VUS}, 10m)"
A_K6="${RESULTS}/scenario_a_k6.txt"
VUS="${VUS}" DURATION=10m \
  CR_CONTROL=0.10 CR_TREATMENT=0.12 \
  EXPERIMENT_KEY=checkout-btn DEMO_URL="${DEMO_URL}" \
  k6 run k6/scenarios/ab_test.js | tee "${A_K6}"

echo "  waiting 10s for consumer flush..."
sleep 10

A_JSON="${RESULTS}/scenario_a_analytics.json"
curl -sf "${API_URL}/projects/${PROJECT_ID}/experiments/${EXPERIMENT_ID}/analytics" \
  -H "${AUTH_HEADER}" | jq . > "${A_JSON}"
ST_A="PASS"
ok "k6 report  → ${A_K6}"
ok "analytics  → ${A_JSON}"

# ── Scenario C — Large effect + time-to-significance ─────────────────────────
info "Scenario C — Large effect (CR_control=10%, CR_treatment=15%), creating experiment..."
C_EXP_JSON=$(curl -sf -X POST "${API_URL}/projects/${PROJECT_ID}/experiments" \
  -H "${AUTH_HEADER}" -H "Content-Type: application/json" \
  -d "{
    \"key\": \"large-effect-${RUN_ID}\",
    \"name\": \"Large Effect ${RUN_ID}\",
    \"trafficPercent\": 1.0,
    \"variants\": [
      {\"key\": \"control\",   \"name\": \"Control\",   \"weight\": 50},
      {\"key\": \"treatment\", \"name\": \"Treatment\", \"weight\": 50}
    ]
  }")
C_EXP_ID=$(echo "${C_EXP_JSON}" | jq -r '.id')
curl -sf -X POST "${API_URL}/projects/${PROJECT_ID}/experiments/${C_EXP_ID}/start" \
  -H "${AUTH_HEADER}" > /dev/null

echo "  waiting 35s for SDK config refresh..."
sleep 35

C_CSV="${RESULTS}/large_effect.csv"
echo "elapsed_sec,total_exposures,cr_control,cr_treatment,p_value,significant" > "${C_CSV}"
START_TS=$(date +%s)

snapshot() {
  local elapsed=$(( $(date +%s) - START_TS ))
  local analytics
  analytics=$(curl -sf "${API_URL}/projects/${PROJECT_ID}/experiments/${C_EXP_ID}/analytics" \
    -H "${AUTH_HEADER}" 2>/dev/null || echo '{}')
  local total cr_ctl cr_trt p_val sig
  total=$(echo "${analytics}"  | jq '.total_exposures // 0')
  cr_ctl=$(echo "${analytics}" | jq '[.variants[]? | select(.is_control==true)  | .conversion_rate][0] // 0')
  cr_trt=$(echo "${analytics}" | jq '[.variants[]? | select(.is_control==false) | .conversion_rate][0] // 0')
  p_val=$(echo  "${analytics}" | jq '[.variants[]? | select(.is_control==false) | .p_value][0] // null')
  sig=$(echo    "${analytics}" | jq '[.variants[]? | select(.is_control==false) | .significant][0] // false')
  echo "${elapsed},${total},${cr_ctl},${cr_trt},${p_val},${sig}" >> "${C_CSV}"
  echo "  t=${elapsed}s  exposures=${total}  p_value=${p_val}  significant=${sig}"
}

( while true; do sleep 60; snapshot; done ) &
SNAP_PID=$!

VUS="${VUS}" DURATION=10m \
  CR_CONTROL=0.10 CR_TREATMENT=0.15 \
  EXPERIMENT_KEY="large-effect-${RUN_ID}" DEMO_URL="${DEMO_URL}" \
  k6 run k6/scenarios/large_effect.js --quiet

kill "${SNAP_PID}" 2>/dev/null || true
sleep 10; snapshot

ST_C="PASS"
ok "saved → ${C_CSV}"

# ── Scenario E.a+E.b — Platform performance (k6) ─────────────────────────────
info "Scenario E.a+E.b — Performance: /sdk/config latency + /sdk/events throughput (VUS=200)"
PERF_OUT="${RESULTS}/perf_k6.txt"
SDK_URL="${SDK_URL}" SDK_API_KEY="${SDK_API_KEY}" VUS=200 \
  k6 run k6/scenarios/perf.js | tee "${PERF_OUT}"
ST_Eab="PASS"
ok "saved → ${PERF_OUT}"

# ── Scenario B — Null effect, 20 independent runs ─────────────────────────────
info "Scenario B — Null effect (${B_RUNS} runs × 3m, ~$(( B_RUNS * 3 + B_RUNS / 2 )) min total)"
echo "  Note: if this takes >55 min the Firebase token may expire — re-export if needed."
B_CSV="${RESULTS}/null_effect.csv"
echo "run,experiment_key,exposures_control,exposures_treatment,conversions_control,conversions_treatment,cr_control,cr_treatment,p_value,significant" > "${B_CSV}"

for i in $(seq 1 "${B_RUNS}"); do
  KEY="null-effect-${i}-$(date +%s)"
  echo "[${i}/${B_RUNS}] ${KEY}"

  EXP_JSON=$(curl -sf -X POST "${API_URL}/projects/${PROJECT_ID}/experiments" \
    -H "${AUTH_HEADER}" -H "Content-Type: application/json" \
    -d "{
      \"key\": \"${KEY}\",
      \"name\": \"Null Effect ${i}\",
      \"trafficPercent\": 1.0,
      \"variants\": [
        {\"key\": \"control\",   \"name\": \"Control\",   \"weight\": 50},
        {\"key\": \"treatment\", \"name\": \"Treatment\", \"weight\": 50}
      ]
    }")
  EXP_ID=$(echo "${EXP_JSON}" | jq -r '.id')
  curl -sf -X POST "${API_URL}/projects/${PROJECT_ID}/experiments/${EXP_ID}/start" \
    -H "${AUTH_HEADER}" > /dev/null

  echo "  waiting 35s for SDK config refresh..."
  sleep 35

  VUS="${VUS}" DURATION=3m \
    CR_CONTROL=0.10 CR_TREATMENT=0.10 \
    EXPERIMENT_KEY="${KEY}" DEMO_URL="${DEMO_URL}" \
    k6 run k6/scenarios/null_effect.js --quiet

  sleep 10

  ANALYTICS=$(curl -sf \
    "${API_URL}/projects/${PROJECT_ID}/experiments/${EXP_ID}/analytics" \
    -H "${AUTH_HEADER}")

  CTL=$(echo "${ANALYTICS}" | jq '.variants[] | select(.is_control==true)')
  TRT=$(echo "${ANALYTICS}" | jq '.variants[] | select(.is_control==false)')

  ROW="${i},${KEY}"
  ROW+=",$(echo "${CTL}" | jq '.exposures')"
  ROW+=",$(echo "${TRT}" | jq '.exposures')"
  ROW+=",$(echo "${CTL}" | jq '.conversions')"
  ROW+=",$(echo "${TRT}" | jq '.conversions')"
  ROW+=",$(echo "${CTL}" | jq '.conversion_rate')"
  ROW+=",$(echo "${TRT}" | jq '.conversion_rate')"
  ROW+=",$(echo "${TRT}" | jq '.p_value // "null"')"
  ROW+=",$(echo "${TRT}" | jq '.significant // false')"
  echo "${ROW}" >> "${B_CSV}"

  P=$(echo "${TRT}" | jq '.p_value // "null"')
  SIG=$(echo "${TRT}" | jq '.significant // false')
  echo "  p_value=${P}  significant=${SIG}"
done

FP=$(tail -n +2 "${B_CSV}" | awk -F',' '$9!="null" && $9+0<0.05 {c++} END{print c+0}')
ST_B="PASS (FP=${FP}/${B_RUNS})"
ok "saved → ${B_CSV}"
echo "  False positives: ${FP}/${B_RUNS} (expected ≤1 at α=0.05)"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "════════════════════════════════════════════"
echo " Results: ${RESULTS}/"
echo "════════════════════════════════════════════"
printf "  %-12s %s\n" "Scenario" "Status"
printf "  %-12s %s\n" "────────" "──────"
printf "  %-12s %s\n" "E.d"    "${ST_Ed}"
printf "  %-12s %s\n" "D.b"    "${ST_Db}"
printf "  %-12s %s\n" "D.d"    "${ST_Dd}"
printf "  %-12s %s\n" "D.a+c"  "${ST_Dac}"
printf "  %-12s %s\n" "A"      "${ST_A}"
printf "  %-12s %s\n" "C"      "${ST_C}"
printf "  %-12s %s\n" "E.a+b"  "${ST_Eab}"
printf "  %-12s %s\n" "B"      "${ST_B}"
echo "════════════════════════════════════════════"
echo ""
echo "Next steps:"
echo "  1. Open admin UI → checkout-btn → Analytics → screenshot"
echo "  2. Import CSVs into thesis charts (R / Python / Excel)"
echo "  3. Copy bench.txt numbers into Section 4 table"
