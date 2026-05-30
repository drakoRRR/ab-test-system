#!/usr/bin/env bash
# Quick validation — all scenarios in ~20 minutes.
#
# Prerequisites (must be running):
#   cd backend && make dev
#   make demo-run
#   make demo-setup   (writes demo/.env, prints PROJECT_ID + EXPERIMENT_ID)
#
# Usage:
#   export FIREBASE_TOKEN=<token from DevTools → Cookies>
#   export PROJECT_ID=<uuid>
#   export EXPERIMENT_ID=<uuid>   # checkout-btn experiment
#   ./scripts/run_quick.sh

set -euo pipefail

API_URL="${API_URL:-http://localhost:8080/api/v1}"
DEMO_URL="${DEMO_URL:-http://localhost:8081}"
SDK_URL="${SDK_URL:-http://localhost:8080}"
VUS="${VUS:-50}"
PROJECT_ID="${PROJECT_ID:?Set PROJECT_ID}"
EXPERIMENT_ID="${EXPERIMENT_ID:?Set EXPERIMENT_ID}"
FIREBASE_TOKEN="${FIREBASE_TOKEN:?Set FIREBASE_TOKEN}"
AUTH="Authorization: Bearer ${FIREBASE_TOKEN}"
SDK_API_KEY=$(grep SDK_API_KEY demo/.env 2>/dev/null | cut -d= -f2 || true)

RUN_ID=$(date +%Y%m%d_%H%M%S)
OUT="results/${RUN_ID}"
mkdir -p "${OUT}"

info() { echo ""; echo "▶ $*"; }
ok()   { echo "  ✓ $*"; }

is_up() {
  local code
  code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 "$1" 2>/dev/null || echo "000")
  [[ "${code}" != "000" ]]
}

# ── Preflight ─────────────────────────────────────────────────────────────────
info "Preflight"
for cmd in k6 jq go curl; do
  command -v "${cmd}" &>/dev/null || { echo "missing: ${cmd}"; exit 1; }
done
is_up "${SDK_URL}"  || { echo "Backend not responding at ${SDK_URL}";  exit 1; }
is_up "${DEMO_URL}/health" || { echo "Demo not responding at ${DEMO_URL}"; exit 1; }
[[ -n "${SDK_API_KEY}" ]] || { echo "demo/.env missing — run: make demo-setup"; exit 1; }
ok "all services up  |  results → ${OUT}/"

# ── E.d benchmark ─────────────────────────────────────────────────────────────
info "E.d — SDK benchmark"
(cd sdk && go test -bench=. -benchmem -count=3 -run='^$') | tee "${OUT}/bench.txt"

# ── D.b + D.d — SDK tests ─────────────────────────────────────────────────────
info "D.b + D.d — Sticky determinism + chi-square uniformity"
(cd sdk && go test \
  -run "TestEvaluateExperiment_StickyAcrossRepeatedCalls|TestAssignBucket_ChiSquareUniformity" \
  -v -count=1) | tee "${OUT}/sdk_tests.txt"

# ── D.a + D.c — HTTP sticky ───────────────────────────────────────────────────
info "D.a + D.c — HTTP consistency + distribution (1000 users)"
(cd demo/backend && DEMO_URL="${DEMO_URL}" EXPERIMENT_KEY=checkout-btn \
  go run ./cmd/sticky/) | tee "${OUT}/sticky.txt"

# ── A — Known effect ──────────────────────────────────────────────────────────
info "A — Known effect  CR_ctrl=10%  CR_trt=12%  3 min"
VUS="${VUS}" DURATION=3m CR_CONTROL=0.10 CR_TREATMENT=0.12 \
  EXPERIMENT_KEY=checkout-btn DEMO_URL="${DEMO_URL}" \
  k6 run k6/scenarios/ab_test.js | tee "${OUT}/scenario_a_k6.txt"
sleep 10
curl -sf "${API_URL}/projects/${PROJECT_ID}/experiments/${EXPERIMENT_ID}/analytics" \
  -H "${AUTH}" | jq . > "${OUT}/scenario_a_analytics.json"
ok "analytics → ${OUT}/scenario_a_analytics.json"

# ── C — Large effect + p-value curve ─────────────────────────────────────────
info "C — Large effect  CR_ctrl=10%  CR_trt=15%  3 min  (snapshot every 30s)"
C_KEY="large-effect-${RUN_ID}"
C_ID=$(curl -sf -X POST "${API_URL}/projects/${PROJECT_ID}/experiments" \
  -H "${AUTH}" -H "Content-Type: application/json" \
  -d "{\"key\":\"${C_KEY}\",\"name\":\"Large Effect\",\"trafficPercent\":1.0,
       \"variants\":[{\"key\":\"control\",\"name\":\"Control\",\"weight\":50},
                     {\"key\":\"treatment\",\"name\":\"Treatment\",\"weight\":50}]}" \
  | jq -r '.id')
curl -sf -X POST "${API_URL}/projects/${PROJECT_ID}/experiments/${C_ID}/start" \
  -H "${AUTH}" > /dev/null

echo "  waiting 15s for SDK config refresh..."
sleep 15

C_CSV="${OUT}/large_effect.csv"
echo "elapsed_sec,total_exposures,cr_control,cr_treatment,p_value,significant" > "${C_CSV}"
START_TS=$(date +%s)

snap() {
  local el=$(( $(date +%s) - START_TS ))
  local a
  a=$(curl -sf "${API_URL}/projects/${PROJECT_ID}/experiments/${C_ID}/analytics" \
    -H "${AUTH}" 2>/dev/null || echo '{}')
  local tot cr_c cr_t pv sig
  tot=$(echo "${a}"  | jq '.total_exposures // 0')
  cr_c=$(echo "${a}" | jq '[.variants[]? | select(.is_control==true)  | .conversion_rate][0] // 0')
  cr_t=$(echo "${a}" | jq '[.variants[]? | select(.is_control==false) | .conversion_rate][0] // 0')
  pv=$(echo   "${a}" | jq '[.variants[]? | select(.is_control==false) | .p_value][0] // null')
  sig=$(echo  "${a}" | jq '[.variants[]? | select(.is_control==false) | .significant][0] // false')
  echo "${el},${tot},${cr_c},${cr_t},${pv},${sig}" >> "${C_CSV}"
  echo "  t=${el}s  exposures=${tot}  p_value=${pv}  significant=${sig}"
}

( while true; do sleep 30; snap; done ) &
SNAP_PID=$!

VUS="${VUS}" DURATION=3m CR_CONTROL=0.10 CR_TREATMENT=0.15 \
  EXPERIMENT_KEY="${C_KEY}" DEMO_URL="${DEMO_URL}" \
  k6 run k6/scenarios/large_effect.js --quiet

kill "${SNAP_PID}" 2>/dev/null || true
sleep 10; snap
ok "p-value curve → ${C_CSV}"

# ── B — Null effect, 5 runs ───────────────────────────────────────────────────
info "B — Null effect  5 runs × 1.5 min  (~12 min total)"
B_CSV="${OUT}/null_effect.csv"
echo "run,experiment_key,exposures_c,exposures_t,cr_control,cr_treatment,p_value,significant" \
  > "${B_CSV}"

for i in $(seq 1 5); do
  KEY="null-effect-${i}-$(date +%s)"
  echo "[${i}/5] ${KEY}"

  EID=$(curl -sf -X POST "${API_URL}/projects/${PROJECT_ID}/experiments" \
    -H "${AUTH}" -H "Content-Type: application/json" \
    -d "{\"key\":\"${KEY}\",\"name\":\"Null ${i}\",\"trafficPercent\":1.0,
         \"variants\":[{\"key\":\"control\",\"name\":\"Control\",\"weight\":50},
                       {\"key\":\"treatment\",\"name\":\"Treatment\",\"weight\":50}]}" \
    | jq -r '.id')
  curl -sf -X POST "${API_URL}/projects/${PROJECT_ID}/experiments/${EID}/start" \
    -H "${AUTH}" > /dev/null

  echo "  waiting 15s for SDK config refresh..."
  sleep 15

  VUS="${VUS}" DURATION=1m30s CR_CONTROL=0.10 CR_TREATMENT=0.10 \
    EXPERIMENT_KEY="${KEY}" DEMO_URL="${DEMO_URL}" \
    k6 run k6/scenarios/null_effect.js --quiet

  sleep 10

  AN=$(curl -sf "${API_URL}/projects/${PROJECT_ID}/experiments/${EID}/analytics" \
    -H "${AUTH}")
  CTL=$(echo "${AN}" | jq '.variants[] | select(.is_control==true)')
  TRT=$(echo "${AN}" | jq '.variants[] | select(.is_control==false)')
  PV=$(echo "${TRT}" | jq '.p_value // "null"')
  SIG=$(echo "${TRT}" | jq '.significant // false')

  echo "${i},${KEY},$(echo "${CTL}"|jq '.exposures'),$(echo "${TRT}"|jq '.exposures'),$(echo "${CTL}"|jq '.conversion_rate'),$(echo "${TRT}"|jq '.conversion_rate'),${PV},${SIG}" \
    >> "${B_CSV}"
  echo "  p_value=${PV}  significant=${SIG}"
done

FP=$(tail -n +2 "${B_CSV}" | awk -F',' '$7!="null" && $7+0<0.05 {c++} END{print c+0}')
ok "null effect CSV → ${B_CSV}  (false positives: ${FP}/5)"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "════════════════════════════════════════"
echo " Done — results: ${OUT}/"
echo "════════════════════════════════════════"
echo "  bench.txt              SDK benchmark"
echo "  sdk_tests.txt          D.b sticky + D.d chi-square"
echo "  sticky.txt             D.a HTTP consistency + D.c distribution"
echo "  scenario_a_k6.txt      k6 report (known effect)"
echo "  scenario_a_analytics.json  uplift / p-value / CI"
echo "  large_effect.csv       p-value curve over time"
echo "  null_effect.csv        false positive rate (${FP}/5)"
echo "════════════════════════════════════════"
echo ""
echo "Screenshot checklist:"
echo "  □ Admin UI → checkout-btn → Analytics dashboard"
