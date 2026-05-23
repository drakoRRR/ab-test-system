// Domain types matching the backend OpenAPI spec

export type ExperimentStatus = 'draft' | 'running' | 'paused' | 'completed'

export interface Project {
  id: string
  organizationId?: string
  name: string
  description?: string
  createdAt: string
  updatedAt?: string
}

export interface CreateProjectRequest {
  name: string
  description?: string
}

export interface UpdateProjectRequest {
  name?: string
  description?: string
}

export interface FlagRule {
  type: 'percentage'
  value: number
}

export interface Flag {
  id: string
  projectId: string
  key: string
  name: string
  enabled: boolean
  rules?: FlagRule[]
  createdAt: string
  updatedAt: string
}

export interface CreateFlagRequest {
  key: string
  name: string
  rules?: FlagRule[]
}

export interface UpdateFlagRequest {
  name?: string
  enabled?: boolean
  rules?: FlagRule[]
}

export interface Variant {
  id: string
  key: string
  name: string
  weight: number
}

export interface CreateVariantRequest {
  key: string
  name: string
  weight: number
}

export interface Experiment {
  id: string
  projectId: string
  flagId?: string | null
  key: string
  name: string
  description?: string
  status: ExperimentStatus
  trafficPercent: number
  variants: Variant[]
  createdAt: string
  updatedAt: string
  startedAt?: string | null
  endedAt?: string | null
}

export interface CreateExperimentRequest {
  key: string
  name: string
  description?: string
  flagId?: string
  trafficPercent: number
  variants: CreateVariantRequest[]
}

export interface UpdateExperimentRequest {
  name?: string
  description?: string
  trafficPercent?: number
}

export interface VariantAnalytics {
  variantId: string
  variantKey: string
  variantName: string
  exposures: number
  conversions: number
  conversionRate: number
  isControl: boolean
  uplift?: number | null
  pValue?: number | null
  ciLow?: number | null
  ciHigh?: number | null
  significant?: boolean | null
}

export interface ExperimentAnalytics {
  experimentId: string
  totalExposures: number
  totalConversions: number
  variants: VariantAnalytics[]
}

export interface ApiKey {
  id: string
  projectId: string
  name: string
  prefix: string
  createdAt: string
  revokedAt?: string | null
}

export interface ApiKeyCreated extends ApiKey {
  key: string
}

export interface CreateApiKeyRequest {
  name: string
}

export interface Pagination {
  page: number
  limit: number
  total: number
  totalPages: number
}
