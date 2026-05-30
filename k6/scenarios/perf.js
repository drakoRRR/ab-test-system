import http from 'k6/http'
import { check } from 'k6'
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js'

// Scenario E: Platform performance under load.
//
// Two sequential phases (startTime separates them):
//   config_latency   — GET /sdk/config (hot path: cached in Redis)     p95 target: <50ms
//   events_ingest    — POST /sdk/events (202 Accepted, async pipeline) p95 target: <100ms
//
// Required env vars:
//   SDK_URL      backend base URL, default http://localhost:8080
//   SDK_API_KEY  SDK key issued by the platform

export const options = {
  scenarios: {
    config_latency: {
      executor:  'constant-vus',
      vus:       parseInt(__ENV.VUS || '200'),
      duration:  '2m',
      exec:      'configLatency',
      startTime: '0s',
    },
    events_ingest: {
      executor:  'constant-vus',
      vus:       parseInt(__ENV.VUS || '200'),
      duration:  '2m',
      exec:      'eventsIngest',
      startTime: '2m30s',
    },
  },
  thresholds: {
    'http_req_duration{scenario:config_latency}': ['p(95)<50'],
    'http_req_duration{scenario:events_ingest}':  ['p(95)<100'],
    'http_req_failed':                            ['rate<0.01'],
  },
}

const SDK_URL = __ENV.SDK_URL     || 'http://localhost:8080'
const API_KEY = __ENV.SDK_API_KEY || ''

const CONFIG_HEADERS = { 'X-API-Key': API_KEY }
const EVENT_HEADERS  = { 'Content-Type': 'application/json', 'X-API-Key': API_KEY }

export function configLatency() {
  const res = http.get(`${SDK_URL}/sdk/config`, { headers: CONFIG_HEADERS })
  check(res, { 'config 200': (r) => r.status === 200 })
}

export function eventsIngest() {
  const events = Array.from({ length: 10 }, () => ({
    id:            uuidv4(),
    type:          'exposure',
    user_id:       uuidv4(),
    experiment_id: 'perf-exp',
    variant_id:    'perf-variant',
    name:          'exposure',
    value:         0,
    timestamp:     new Date().toISOString(),
  }))

  const res = http.post(
    `${SDK_URL}/sdk/events`,
    JSON.stringify({ events }),
    { headers: EVENT_HEADERS }
  )
  check(res, { 'events 202': (r) => r.status === 202 })
}
