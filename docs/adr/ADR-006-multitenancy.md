# ADR-006: Multi-tenancy — Shared DB, Row-Level Isolation

**Status:** Accepted  
**Date:** 2026-05-16

## Context

The platform supports multiple organizations, each with its own projects, flags, and experiments. Data isolation between tenants must be guaranteed.

Options:
1. **Shared DB, shared schema, row-level isolation** — single PostgreSQL instance, `org_id` / `project_id` on every table
2. **Shared DB, schema-per-tenant** — a separate PostgreSQL schema (namespace) per organization
3. **DB-per-tenant** — a separate database / connection pool per organization

## Decision

**Shared DB, row-level isolation** enforced via `project_id` on all queries.

## Rationale

| Option | Isolation | Operational complexity | Suitable for |
|---|---|---|---|
| Row-level | Logical | Low | <1 000 tenants |
| Schema-per-tenant | Physical | Medium | 100–10 000 tenants |
| DB-per-tenant | Full | High | Enterprise / regulated |

For both the thesis and the real-world target scenario (a self-hosted platform inside a single company), row-level isolation is sufficient.

## Enforcement

Isolation is enforced at the **service layer**, not at the DB level:

```go
// CORRECT: always filter by project_id from the request context
func (r *flagRepo) List(ctx context.Context, projectID uuid.UUID) ([]Flag, error) {
    const q = `SELECT * FROM flags WHERE project_id = $1`
    return r.db.QueryContext(ctx, q, projectID)
}

// WRONG: never run SELECT * FROM flags without a filter
```

`project_id` is extracted from JWT claims (via middleware) or from the URL path parameter, and the backend verifies that the authenticated user has access to that project.

## Access Control

```
Organization
  └── Project (org_id FK)
        ├── Member has role: admin | member | viewer
        └── SDK Key (project-scoped)

Rules:
- admin:   full CRUD within their own organization
- member:  can manage flags/experiments within their projects
- viewer:  read-only
- SDK Key: access only to /sdk/* endpoints for its own project_id
```

## Consequences

- Every repo/service method takes `projectID` (or `orgID`) as an explicit parameter — never from global context
- When adding a new endpoint, always verify that the `project_id` filter is present in the query
- PostgreSQL Row-Level Security (RLS) is not used — application-layer enforcement is sufficient and easier to test
