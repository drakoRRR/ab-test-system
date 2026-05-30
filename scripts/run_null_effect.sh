#!/usr/bin/env bash
# Scenario B: 20 independent null-effect runs.
# Each run creates a fresh experiment, runs k6 with CR_CONTROL=CR_TREATMENT=0.10,
# then extracts analytics and appends a row to a CSV.
#
# Usage:
#   export FIREBASE_TOKEN=<token>
#   export PROJECT_ID=<uuid>
#   ./scripts/run_null_effect.sh
#
# Optional overrides:
#   API_URL   (default http://localhost:8080/api/v1)
#   DEMO_URL  (default http://localhost:8081)
#   VUS       (default 50)
#   DURATION  (default 3m)
#   RUNS      (default 20)

set -euo pipefail

API_URL="${API_URL:-http://localhost:8080/api/v1}"
DEMO_URL="${DEMO_URL:-http://localhost:8081}"
VUS="${VUS:-50}"
DURATION="${DURATION:-3m}"
RUNS="${RUNS:-20}"
PROJECT_ID="${PROJECT_ID:?Set PROJECT_ID to your platform project UUID}"
FIREBASE_TOKEN="${FIREBASE_TOKEN:?Set FIREBASE_TOKEN from DevTools → Cookies}"

AUTH_HEADER="Authorization: Bearer ${FIREBASE_TOKEN}"

mkdir -p results
OUTPUT="results/null_effect_$(date +%Y%m%d_%H%M%S).csv"

echo "run,experiment_key,exposures_control,exposures_treatment,conversions_control,conversions_treatment,cr_control,cr_treatment,p_value,significant" \
  > "${OUTPUT}"

echo "Starting ${RUNS} null-effect runs → ${OUTPUT}"
echo ""

for i in $(seq 1 "${RUNS}"); do
  KEY="null-effect-${i}-$(date +%s)"
  echo "[${i}/${RUNS}] experiment key: ${KEY}"

  # ── Create experiment ──────────────────────────────────────────────────────
  EXP_JSON=$(curl -sf -X POST "${API_URL}/projects/${PROJECT_ID}/experiments" \
    -H "${AUTH_HEADER}" \
    -H "Content-Type: application/json" \
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

  # ── Start experiment ───────────────────────────────────────────────────────
  curl -sf -X POST "${API_URL}/projects/${PROJECT_ID}/experiments/${EXP_ID}/start" \
    -H "${AUTH_HEADER}" > /dev/null

  # ── Wait for demo SDK to poll the new experiment config (PollInterval=30s) ─
  echo "  waiting 35s for SDK config refresh..."
  sleep 35

  # ── Run k6 ────────────────────────────────────────────────────────────────
  VUS="${VUS}" DURATION="${DURATION}" \
    CR_CONTROL=0.10 CR_TREATMENT=0.10 \
    EXPERIMENT_KEY="${KEY}" DEMO_URL="${DEMO_URL}" \
    k6 run k6/scenarios/null_effect.js --quiet

  # ── Wait for Kafka consumer to flush events to PostgreSQL ─────────────────
  echo "  waiting for consumer flush..."
  sleep 10

  # ── Fetch analytics ────────────────────────────────────────────────────────
  ANALYTICS=$(curl -sf \
    "${API_URL}/projects/${PROJECT_ID}/experiments/${EXP_ID}/analytics" \
    -H "${AUTH_HEADER}")

  CTL=$(echo "${ANALYTICS}"  | jq '.variants[] | select(.is_control == true)')
  TRT=$(echo "${ANALYTICS}"  | jq '.variants[] | select(.is_control == false)')

  EXP_CTL=$(echo "${CTL}" | jq '.exposures')
  EXP_TRT=$(echo "${TRT}" | jq '.exposures')
  CNV_CTL=$(echo "${CTL}" | jq '.conversions')
  CNV_TRT=$(echo "${TRT}" | jq '.conversions')
  CR_CTL=$(echo  "${CTL}" | jq '.conversion_rate')
  CR_TRT=$(echo  "${TRT}" | jq '.conversion_rate')
  P_VALUE=$(echo "${TRT}" | jq '.p_value // "null"')
  SIG=$(echo     "${TRT}" | jq '.significant // false')

  echo "${i},${KEY},${EXP_CTL},${EXP_TRT},${CNV_CTL},${CNV_TRT},${CR_CTL},${CR_TRT},${P_VALUE},${SIG}" \
    >> "${OUTPUT}"

  echo "  p_value=${P_VALUE}  significant=${SIG}"
  echo ""
done

echo "Done. Results saved to ${OUTPUT}"
echo ""

# ── Summary ────────────────────────────────────────────────────────────────
TOTAL=$(tail -n +2 "${OUTPUT}" | wc -l | tr -d ' ')
FP=$(tail -n +2 "${OUTPUT}" | awk -F',' '$9 != "null" && $9+0 < 0.05 {count++} END {print count+0}')
echo "False positives (p < 0.05): ${FP} / ${TOTAL}"
echo "Expected at α=0.05: ≤1 out of 20"
