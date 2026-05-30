#!/usr/bin/env bash
# Scenario C: Run large_effect.js and snapshot /analytics every 60s.
# Produces a CSV with (elapsed_sec, total_exposures, p_value, significant)
# to plot the p-value curve over time.
#
# Usage:
#   export FIREBASE_TOKEN=<token>
#   export PROJECT_ID=<uuid>
#   export EXPERIMENT_ID=<uuid>   # created by seed or API beforehand
#   ./scripts/snapshot_analytics.sh
#
# Optional overrides:
#   API_URL          (default http://localhost:8080/api/v1)
#   DEMO_URL         (default http://localhost:8081)
#   VUS              (default 50)
#   DURATION         (default 10m)
#   SNAPSHOT_INTERVAL (default 60)

set -euo pipefail

API_URL="${API_URL:-http://localhost:8080/api/v1}"
DEMO_URL="${DEMO_URL:-http://localhost:8081}"
VUS="${VUS:-50}"
DURATION="${DURATION:-10m}"
SNAPSHOT_INTERVAL="${SNAPSHOT_INTERVAL:-60}"
PROJECT_ID="${PROJECT_ID:?Set PROJECT_ID}"
EXPERIMENT_ID="${EXPERIMENT_ID:?Set EXPERIMENT_ID (create via API or seed first)}"
FIREBASE_TOKEN="${FIREBASE_TOKEN:?Set FIREBASE_TOKEN}"

AUTH_HEADER="Authorization: Bearer ${FIREBASE_TOKEN}"

mkdir -p results
OUTPUT="results/large_effect_$(date +%Y%m%d_%H%M%S).csv"

echo "elapsed_sec,total_exposures,cr_control,cr_treatment,p_value,significant" \
  > "${OUTPUT}"

START_TS=$(date +%s)

# ── Snapshot loop (background) ─────────────────────────────────────────────
snapshot_loop() {
  while true; do
    sleep "${SNAPSHOT_INTERVAL}"

    ELAPSED=$(( $(date +%s) - START_TS ))

    ANALYTICS=$(curl -sf \
      "${API_URL}/projects/${PROJECT_ID}/experiments/${EXPERIMENT_ID}/analytics" \
      -H "${AUTH_HEADER}" 2>/dev/null || echo '{}')

    TOTAL=$(echo "${ANALYTICS}" | jq '.total_exposures // 0')
    CR_CTL=$(echo "${ANALYTICS}" | jq '[.variants[]? | select(.is_control == true)  | .conversion_rate][0] // 0')
    CR_TRT=$(echo "${ANALYTICS}" | jq '[.variants[]? | select(.is_control == false) | .conversion_rate][0] // 0')
    P_VAL=$(echo  "${ANALYTICS}" | jq '[.variants[]? | select(.is_control == false) | .p_value][0] // null')
    SIG=$(echo    "${ANALYTICS}" | jq '[.variants[]? | select(.is_control == false) | .significant][0] // false')

    echo "${ELAPSED},${TOTAL},${CR_CTL},${CR_TRT},${P_VAL},${SIG}" >> "${OUTPUT}"
    echo "  t=${ELAPSED}s  exposures=${TOTAL}  p_value=${P_VAL}  significant=${SIG}"
  done
}

snapshot_loop &
SNAPSHOT_PID=$!

# ── Run k6 ────────────────────────────────────────────────────────────────
echo "Running large_effect.js → ${OUTPUT}"
VUS="${VUS}" DURATION="${DURATION}" \
  DEMO_URL="${DEMO_URL}" \
  k6 run k6/scenarios/large_effect.js

# ── Stop snapshot loop ─────────────────────────────────────────────────────
kill "${SNAPSHOT_PID}" 2>/dev/null || true

# ── Final snapshot ─────────────────────────────────────────────────────────
sleep 10
ELAPSED=$(( $(date +%s) - START_TS ))
ANALYTICS=$(curl -sf \
  "${API_URL}/projects/${PROJECT_ID}/experiments/${EXPERIMENT_ID}/analytics" \
  -H "${AUTH_HEADER}")
TOTAL=$(echo "${ANALYTICS}"  | jq '.total_exposures')
CR_CTL=$(echo "${ANALYTICS}" | jq '[.variants[] | select(.is_control == true)  | .conversion_rate][0]')
CR_TRT=$(echo "${ANALYTICS}" | jq '[.variants[] | select(.is_control == false) | .conversion_rate][0]')
P_VAL=$(echo  "${ANALYTICS}" | jq '[.variants[] | select(.is_control == false) | .p_value][0]')
SIG=$(echo    "${ANALYTICS}" | jq '[.variants[] | select(.is_control == false) | .significant][0]')
echo "${ELAPSED},${TOTAL},${CR_CTL},${CR_TRT},${P_VAL},${SIG}" >> "${OUTPUT}"

echo ""
echo "Done. Results saved to ${OUTPUT}"
echo "Final: exposures=${TOTAL}  p_value=${P_VAL}  significant=${SIG}"
