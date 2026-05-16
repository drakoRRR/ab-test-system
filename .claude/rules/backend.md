# Go Backend Rules

Rules for writing backend Go code in this project.
Apply these to all code under `apps/ab-test-system/backend/`.

---

## Layer Architecture

```
cmd/server/          entry point: wire dependencies, start server, nothing else
internal/
  domain/            pure models, domain errors, value objects — zero external deps
  services/          business logic — orchestrates domain, calls repos via interfaces
  handler/http/      HTTP layer — translates HTTP ↔ service calls, no business logic
  infra/             interface implementations (PostgreSQL, Kafka, Firebase, etc.)
pkg/                 shared utilities importable by other modules
```

### What belongs where

**`domain/`**
- Structs that model the business (Experiment, Flag, Variant, Event)
- Domain errors (`ErrNotFound`, `ErrAlreadyRunning`, etc.)
- No imports outside stdlib

**`services/`**
- All business logic lives here
- Defines its own interfaces for repositories and external services
- Has no knowledge of HTTP, SQL, or Kafka — only domain types
- Calls infrastructure through interfaces defined in this package

**`handler/http/`**
- Reads HTTP request, calls service, writes HTTP response
- Maps service errors to HTTP status codes
- No `if`, `for`, or business conditions — only translation
- Depends on service interfaces, not concrete implementations

**`infra/`**
- Implements interfaces defined in `services/`
- Knows about SQL, Kafka, Firebase — the domain layer does not
- One sub-package per external system: `infra/postgres/`, `infra/kafka/`, `infra/gcp/`

---

## Interfaces

Define interfaces **at the point of use** (in the consumer package), not at the point of implementation.

```go
// CORRECT: interface lives in the services package, next to the code that uses it
package services

type ExperimentRepository interface {
    Create(ctx context.Context, exp domain.Experiment) error
    GetByID(ctx context.Context, id uuid.UUID) (domain.Experiment, error)
    ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Experiment, error)
}

type Service struct {
    repo ExperimentRepository
}
```

```go
// WRONG: interface lives in infra next to its own implementation
package infra

type ExperimentRepository interface { ... }  // do not do this
```

Keep interfaces small — one method is fine. Split large interfaces rather than merging unrelated methods.

---

## Error Handling

Domain errors are defined in `domain/` and sentinel-based:

```go
// domain/errors.go
var (
    ErrNotFound        = errors.New("not found")
    ErrAlreadyRunning  = errors.New("experiment is already running")
    ErrInvalidState    = errors.New("invalid state transition")
    ErrConflict        = errors.New("conflict")
)
```

Wrap errors at every layer boundary with operation context:

```go
// infra layer — wrap DB error
func (r *experimentRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Experiment, error) {
    var exp domain.Experiment
    err := r.db.QueryRowContext(ctx, query, id).Scan(...)
    if errors.Is(err, sql.ErrNoRows) {
        return domain.Experiment{}, fmt.Errorf("experimentRepo.GetByID: %w", domain.ErrNotFound)
    }
    if err != nil {
        return domain.Experiment{}, fmt.Errorf("experimentRepo.GetByID: %w", err)
    }
    return exp, nil
}
```

```go
// handler layer — map domain errors to HTTP status
func mapError(err error) int {
    switch {
    case errors.Is(err, domain.ErrNotFound):
        return http.StatusNotFound
    case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrAlreadyRunning):
        return http.StatusConflict
    default:
        return http.StatusInternalServerError
    }
}
```

Rules:
- Never swallow errors — always return or log with context
- Never use `panic` for control flow
- Always check errors from `defer` if the function can fail (e.g., `rows.Close()`)

---

## Constructors

Use `New` functions for dependency injection. Never use `init()` with side effects.

```go
type Service struct {
    repo    ExperimentRepository
    flagSvc FlagService
    log     zerolog.Logger
}

func New(repo ExperimentRepository, flagSvc FlagService, log zerolog.Logger) *Service {
    return &Service{repo: repo, flagSvc: flagSvc, log: log}
}
```

- All dependencies are explicit constructor parameters
- No package-level variables that hold state
- No global config access inside functions

---

## Business Logic

Business logic lives **only** in `services/` and `domain/`. The handler layer must not contain conditions that implement business rules.

```go
// CORRECT: business rule in service
func (s *Service) StartExperiment(ctx context.Context, id uuid.UUID) error {
    exp, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return fmt.Errorf("StartExperiment: %w", err)
    }
    if exp.Status != domain.StatusDraft {
        return fmt.Errorf("StartExperiment: %w", domain.ErrInvalidState)
    }
    exp.Status = domain.StatusRunning
    exp.StartedAt = time.Now()
    return s.repo.Update(ctx, exp)
}

// WRONG: business rule leaking into handler
func (h *Handler) StartExperiment(w http.ResponseWriter, r *http.Request) {
    exp := fetchExp(r)
    if exp.Status != "draft" {  // business rule in handler — do not do this
        http.Error(w, "not draft", 409)
        return
    }
    ...
}
```

---

## Unit Tests — Table-Driven

All unit tests use table-driven style with `t.Run`.

```go
func TestService_StartExperiment(t *testing.T) {
    tests := []struct {
        name       string
        experiment domain.Experiment
        wantErr    error
    }{
        {
            name:       "transitions draft to running",
            experiment: domain.Experiment{Status: domain.StatusDraft},
            wantErr:    nil,
        },
        {
            name:       "rejects already running experiment",
            experiment: domain.Experiment{Status: domain.StatusRunning},
            wantErr:    domain.ErrInvalidState,
        },
        {
            name:       "rejects completed experiment",
            experiment: domain.Experiment{Status: domain.StatusCompleted},
            wantErr:    domain.ErrInvalidState,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := &mockExperimentRepo{exp: tt.experiment}
            svc := New(repo)

            err := svc.StartExperiment(context.Background(), tt.experiment.ID)

            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
                return
            }
            require.NoError(t, err)
        })
    }
}
```

Rules:
- Test file lives next to the code it tests (`experiment_service_test.go`)
- Use `require` (fatal on failure) for setup and preconditions; `assert` for independent checks
- Test pure functions and domain logic without mocks
- Mock only external interfaces (repository, Kafka producer) — never mock domain logic
- Integration tests (DB, Kafka) live in `infra/` sub-packages and use testcontainers

---

## Context

`context.Context` is always the **first parameter** of any function that does I/O or can be cancelled.

```go
// CORRECT
func (s *Service) GetExperiment(ctx context.Context, id uuid.UUID) (domain.Experiment, error)

// WRONG
func (s *Service) GetExperiment(id uuid.UUID) (domain.Experiment, error)
```

Never store context in a struct. Pass it explicitly on every call.

---

## Naming

- Interfaces: describe behaviour, not implementation — `ExperimentRepository`, not `IExperimentRepository` or `ExperimentRepositoryInterface`
- Constructors: `New` when there is one per package, `NewXxx` when there are several
- Error variables: `ErrXxx` — exported, defined in `domain/`
- Avoid stuttering: `domain.Experiment`, not `domain.DomainExperiment`
- Acronyms: `userID`, `projectID`, `httpServer` — follow Go conventions

---

## Dependency Direction

```
handler  →  services
handler  →  domain
services →  domain
infra    →  domain
cmd      →  handler, services, infra   (wiring only)
```

**No layer imports another layer except `domain`.** This is enforced by Go's implicit interface satisfaction: `infra/` never needs to import `services/` to implement its interfaces — it just defines the same method signatures, and the compiler checks compatibility at the wiring site in `cmd/`.

```go
// services/experiment/service.go — defines the interface it needs
package experiment

type Repository interface {
    GetByID(ctx context.Context, id uuid.UUID) (domain.Experiment, error)
    Update(ctx context.Context, exp domain.Experiment) error
}

// infra/postgres/experiment_repo.go — implements matching methods, NO import of services/
package postgres

type ExperimentRepo struct{ db *sql.DB }

func (r *ExperimentRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Experiment, error) { ... }
func (r *ExperimentRepo) Update(ctx context.Context, exp domain.Experiment) error              { ... }

// cmd/server/main.go — only place that imports all layers and wires them together
package main

func main() {
    db   := postgres.Connect(cfg)
    repo := postgres.NewExperimentRepo(db)          // *postgres.ExperimentRepo
    svc  := experiment.NewService(repo)             // repo satisfies experiment.Repository implicitly
    h    := handler.NewExperimentHandler(svc)
}
```

Violations to catch:
- `services/` importing anything from `infra/` — always wrong
- `infra/` importing anything from `services/` or `handler/` — always wrong
- `handler/` importing anything from `infra/` — always wrong
- Any two non-`domain` packages importing each other — circular, breaks the boundary

---

## Miscellaneous

- Run `go vet ./...` and `golangci-lint run` before considering any code complete
- No `interface{}` / `any` in public APIs without a documented reason
- Prefer explicit over implicit — no magic, no reflection in business logic
- Config is loaded once at startup and passed as a typed struct; never read env vars inside business logic
- Logging uses `zerolog` — always log with context fields, never with `fmt.Printf`
