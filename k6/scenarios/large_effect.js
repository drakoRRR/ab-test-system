import http from 'k6/http'
import { check, sleep } from 'k6'
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js'

// Scenario C: Large effect — CR_CONTROL=0.10, CR_TREATMENT=0.15 (+50% relative uplift).
// Used to measure time-to-significance: p-value should cross 0.05 within ~2-3 minutes.
// Run via scripts/snapshot_analytics.sh which polls /analytics every 60s in parallel.

export const options = {
  vus:      parseInt(__ENV.VUS      || '50'),
  duration: __ENV.DURATION          || '10m',
  thresholds: {
    http_req_failed:   ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
}

const DEMO_URL       = __ENV.DEMO_URL       || 'http://localhost:8081'
const EXPERIMENT_KEY = __ENV.EXPERIMENT_KEY || 'large-effect'

const CONVERSION_RATES = {
  control:   parseFloat(__ENV.CR_CONTROL   || '0.10'),
  treatment: parseFloat(__ENV.CR_TREATMENT || '0.15'),
}

const HEADERS = { 'Content-Type': 'application/json' }

export default function () {
  const userId   = uuidv4()
  const visitRes = http.get(
    `${DEMO_URL}/visit?user_id=${userId}&experiment_key=${EXPERIMENT_KEY}`
  )
  check(visitRes, { 'visit 200': (r) => r.status === 200 })

  const body    = visitRes.json()
  const variant = body ? body.variant : null
  if (!variant) return

  sleep(Math.random() * 1.5 + 0.5)

  const rate = CONVERSION_RATES[variant] ?? 0.0
  if (Math.random() < rate) {
    const convRes = http.post(
      `${DEMO_URL}/convert`,
      JSON.stringify({ user_id: userId, event: 'purchase', value: 1.0 }),
      { headers: HEADERS }
    )
    check(convRes, { 'convert 200': (r) => r.status === 200 })
  }
}
