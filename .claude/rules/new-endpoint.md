# Adding a New API Endpoint

This project is **OpenAPI-first**: the spec is the source of truth. Never write handler code before the spec exists.

---

## Step-by-Step

### 1. Define the endpoint in `api/<domain>.yml`

Add the path, operation, request/response schemas. Every operation **must** have an `operationId`.

```yaml
# api/experiment.yml
paths:
  /projects/{projectId}/experiments:
    post:
      operationId: CreateExperiment
      summary: Create a new experiment
      tags: [Experiments]
      security:
        - BearerAuth: []
      parameters:
        - $ref: './base.yml#/components/parameters/ProjectIdParam'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ExperimentCreateRequest'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Experiment'
        '400':
          $ref: './base.yml#/components/responses/BadRequest'
        '401':
          $ref: './base.yml#/components/responses/Unauthorized'
```

**If this is a new domain file**, add it to `api/redocly.yaml` — otherwise `make generate-api` silently skips it:

```yaml
# api/redocly.yaml
apis:
  base:
    root: ./base.yml
  user:
    root: ./user.yml
  experiment:               # add this
    root: ./experiment.yml
```

### 2. Run codegen

```bash
make generate-api
```

This bundles all specs into `api/_build/openapi.yml` and regenerates `internal/handler/http/api/codegen/api.gen.go`.

The generated `StrictServerInterface` will now include a new method:

```go
// auto-generated — DO NOT EDIT
type StrictServerInterface interface {
    CreateUser(ctx context.Context, request CreateUserRequestObject) (CreateUserResponseObject, error)
    // ... existing methods ...
    CreateExperiment(ctx context.Context, request CreateExperimentRequestObject) (CreateExperimentResponseObject, error)
}
```

The project will not compile until every method in `StrictServerInterface` is implemented — use this as your checklist.

### 3. Implement the handler method

**If the domain package already exists** (`internal/handler/http/api/experiment/`), add the method to its implementation file:

```go
// internal/handler/http/api/experiment/experiment.go
func (h *Handler) CreateExperiment(
    ctx context.Context,
    request gen.CreateExperimentRequestObject,
) (gen.CreateExperimentResponseObject, error) {
    if request.Body == nil {
        return gen.CreateExperiment400JSONResponse{Message: ptr("request body is required")}, nil
    }

    exp, err := h.service.CreateExperiment(ctx, toDomainCreateRequest(request))
    if err != nil {
        return nil, err
    }

    return gen.CreateExperiment201JSONResponse(toAPIExperiment(exp)), nil
}
```

**If this is a new domain**, create two files:

```go
// internal/handler/http/api/experiment/handler.go
package experiment

import (
    "context"

    "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
)

// Service interface defined here, at the point of use — not in services/ or infra/
type Service interface {
    CreateExperiment(ctx context.Context, req experiment.CreateRequest) (experiment.Experiment, error)
    GetExperiment(ctx context.Context, id uuid.UUID) (experiment.Experiment, error)
}

type Handler struct {
    service Service
}

func NewHandler(service Service) *Handler {
    return &Handler{service: service}
}
```

```go
// internal/handler/http/api/experiment/experiment.go
package experiment

import (
    "context"

    gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func (h *Handler) CreateExperiment(
    ctx context.Context,
    request gen.CreateExperimentRequestObject,
) (gen.CreateExperimentResponseObject, error) {
    // ...
}
```

### 4. Add converters in the same domain package

Converters between domain types and API (generated) types live in `internal/handler/http/api/<domain>/convert.go`.
They are pure functions with no error handling — if the handler has done its job, conversion cannot fail.

```go
// internal/handler/http/api/experiment/convert.go
package experiment

import (
    "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
    gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func toDomainCreateRequest(r gen.CreateExperimentRequestObject) experiment.CreateRequest {
    return experiment.CreateRequest{
        Name:            r.Body.Name,
        Hypothesis:      derefString(r.Body.Hypothesis),
        TrafficPercent:  r.Body.TrafficPercent,
    }
}

func toAPIExperiment(e experiment.Experiment) gen.Experiment {
    return gen.Experiment{
        Id:             e.ID.String(),
        Name:           e.Name,
        Status:         gen.ExperimentStatus(e.Status),
        TrafficPercent: e.TrafficPercent,
        CreatedAt:      &e.CreatedAt,
    }
}
```

### 5. Register the new domain handler in the aggregate server

`internal/handler/http/api/server.go` holds the struct that satisfies the full `StrictServerInterface` by embedding all domain handlers:

```go
// internal/handler/http/api/server.go
package api

import (
    "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/experiment"
    "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/user"
)

// Server satisfies gen.StrictServerInterface via embedding.
// Adding a new domain = add a new embedded field.
type Server struct {
    *user.Handler
    *experiment.Handler
}

func NewServer(
    user *user.Handler,
    exp  *experiment.Handler,
) *Server {
    return &Server{
        Handler:  user,
        Handler:  exp,   // embed by field name to avoid ambiguity if needed
    }
}
```

### 6. Wire in `cmd/server/main.go`

```go
// cmd/server/main.go
experimentRepo := postgres.NewExperimentRepo(db)
experimentSvc  := experimentservice.NewService(experimentRepo)
experimentH    := experimenthandler.NewHandler(experimentSvc)

server := api.NewServer(userH, experimentH)
strict := gen.NewStrictHandler(server, middlewares)
router := gen.HandlerWithOptions(strict, gen.GorillaServerOptions{
    Router: mux.NewRouter(),
})
```

---

## Handler Conventions

**Nil body check**: always check `request.Body == nil` before accessing fields — return 400, not a panic.

```go
if request.Body == nil {
    return gen.CreateExperiment400JSONResponse{Message: ptr("request body is required")}, nil
}
```

**Return typed response objects**, not raw `http.ResponseWriter` writes. The strict handler pattern gives you compile-time checked response types (`Create201JSONResponse`, `Create400JSONResponse`, etc.).

**Never return a domain error directly** from a handler method. Map it first:

```go
exp, err := h.service.CreateExperiment(ctx, req)
if err != nil {
    switch {
    case errors.Is(err, domain.ErrConflict):
        return gen.CreateExperiment409JSONResponse{Message: ptr(err.Error())}, nil
    default:
        return nil, err   // 500 — let the strict handler's ResponseErrorHandlerFunc deal with it
    }
}
```

**Converters assume preconditions are met**: `toDomainXxx()` functions do not validate — validation is OpenAPI middleware's job (required fields, enums, formats). The handler guards nil body; converters just map fields.

**Validation layering** (do not duplicate across layers):

```
OpenAPI middleware  →  request schema validation (required, format, enum)
Handler             →  nil body guard, auth context extraction
Service             →  business rules (state machine, ownership, limits)
Infra               →  persistence constraints (unique, FK)
```

---

## Gotchas

- Adding a new `api/<domain>.yml` without registering it in `api/redocly.yaml` → `make generate-api` succeeds but your endpoints are missing from the bundle. No error, no warning.
- `StrictServerInterface` grows with every new operationId. The project will not compile until all methods are implemented — intended behaviour, use it as a checklist.
- `operationId` drives the generated type names (`CreateExperiment` → `CreateExperimentRequestObject`, `CreateExperiment201JSONResponse`). Keep operationIds consistent: `VerbNoun` (`CreateExperiment`, `ListFlags`, `StartExperiment`).
- Path parameters in the URL (`{projectId}`) map to `request.ProjectId` in the request object — check the generated type to confirm the exact field name.
