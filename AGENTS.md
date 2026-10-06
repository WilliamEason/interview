# AGENTS.md

Guidance for humans and AI agents working in this repository. Keep it current: if you change a convention, change this file in the same commit.

## Purpose

A KYC (Know Your Customer) service for a FinTech platform. Its core responsibility:

> Given a user identifier (UUID) and an intended action, decide whether that user has passed the KYC checks required for that action.

**First use case — "Can User A BUY?"**

```
Trading service --"User A: BUY"--> KYC service --> "Can User A buy?"
                                                   --> "Does User A have a valid sanctions screening with no hits?"
```

A user may BUY only if they have a sanctions screening that is (1) **valid** (completed and not expired) and (2) has **no hits**. Anything else (no screening, expired, pending, hits, unknown user) is a denial, and the response must say which reason applied.

Fail closed: if the answer cannot be determined, the answer is "no".

## Status

Scaffolding only. Packages contain `doc.go` files describing intent; no domain logic exists yet. Do not add behaviour beyond the current task.

## Tech stack and tools

| Concern | Choice |
|---|---|
| Language | Go 1.26 (`go.mod` is the source of truth) |
| HTTP | stdlib `net/http` (`ServeMux` with method/path patterns) |
| Logging | stdlib `log/slog`, JSON handler |
| UUIDs | `github.com/google/uuid` |
| Testing | stdlib `testing` (table-driven); `testify` is not used unless agreed |
| Formatting | `go fmt` |
| Static checks | `go vet` |
| Build / run / test | the `go` toolchain only; no Makefile or external task runner |

Prefer the standard library. Adding a third-party dependency needs a reason stated in the PR.

Common commands:

```
go fmt ./...            # format
go vet ./...            # static checks
go build ./...          # compile everything
go test ./...           # run tests (add -race where cgo is available)
go run ./cmd/kycd       # run the service
go mod tidy             # after adding/removing dependencies
```

## Project layout

```
cmd/
  kycd/                  Composition root. Wires adapters to use cases, starts the server. No logic.
internal/
  domain/                Pure business logic. Imports stdlib only (plus uuid). No I/O, no JSON tags.
    user/                UserID value object (UUID).
    kyc/                 Sanctions screening, KYC action, eligibility decision, repository port.
  application/           Use cases. Orchestrates domain objects through ports.
    eligibility/         "Can this user perform this action?" use case.
  infrastructure/        Adapters implementing domain ports.
    memory/              In-memory repositories (tests, local dev).
  interfaces/            Inbound adapters (how the outside world calls us).
    httpapi/             HTTP handlers and JSON request/response structs.
```

### Dependency rule

Dependencies point inward only:

```
interfaces ──> application ──> domain <── infrastructure
                    ^                          
cmd/kycd wires everything
```

- `domain` imports nothing from this repo except other `domain` packages.
- `application` imports `domain` only. It never imports `infrastructure` or `interfaces`.
- `infrastructure` and `interfaces` import `domain` and/or `application`, never each other.
- Only `cmd/kycd` knows about concrete implementations.

Enforce with an architecture test (a `_test.go` that runs `go list -deps` / `go/build` over `internal/domain/...` and `internal/application/...` and fails on forbidden imports), so the rule is checked by plain `go test`.

## Domain-driven design conventions

- **Ubiquitous language.** Use the terms in the glossary below in code, tests, and docs. Do not invent synonyms.
- **Value objects** are immutable, validated in their constructor, and compared by value (e.g. `user.ID`).
- **Entities / aggregates** have identity and enforce their own invariants. Fields are unexported; mutate through methods that return errors.
- **Domain services / policies** hold rules that do not belong to a single entity (e.g. the rule deciding whether a screening permits an action).
- **Repositories** are interfaces defined in `domain` (consumer side), named for the aggregate (`kyc.ScreeningRepository`), implemented in `infrastructure`.
- **Time is injected.** Domain code takes `now time.Time` (or a `Clock` interface); never call `time.Now()` in `domain`.
- **Errors.** Use sentinel/typed errors for domain outcomes (e.g. `ErrScreeningNotFound`); wrap with `%w`; check with `errors.Is/As`.
- **Denials are not errors.** "User cannot BUY" is a normal decision value (with reasons), not a Go `error`. Errors mean we could not decide.

### Glossary

| Term | Meaning |
|---|---|
| User | A customer identified by a UUID. Owned by another system; we only reference the ID. |
| Action | Something a user wants to do that requires KYC clearance. Initially `BUY`. |
| Sanctions screening | A check of a user against sanctions lists, producing a status and any hits. |
| Hit | A potential or confirmed match against a sanctions list. Any hit blocks the action. |
| Valid screening | Completed, and not past its expiry at decision time. |
| Eligibility decision | The result of asking "can this user do this action?": allowed/denied plus reasons. |

## API contract (JSON)

Inbound/outbound payloads are plain Go structs with `json` tags living in `internal/interfaces/httpapi`. They are transport DTOs only:

- Never put `json` tags on domain types. Map DTO <-> domain explicitly.
- Use `snake_case` JSON field names. Timestamps are RFC 3339 UTC. IDs are UUID strings.
- Validate and parse at the edge (e.g. bad UUID -> 400) before calling the use case.
- Responses for a denial are `200` with `"allowed": false` and machine-readable reason codes; `4xx/5xx` are for bad requests and failures to decide.

Planned first endpoint (shape is provisional):

```
POST /v1/eligibility
{ "user_id": "<uuid>", "action": "BUY" }
->
{ "user_id": "<uuid>", "action": "BUY", "allowed": false, "reasons": ["SANCTIONS_SCREENING_MISSING"] }
```

## Testing

- Domain: table-driven unit tests, no mocks needed (pure code, inject `now`).
- Application: test against in-memory repositories from `infrastructure/memory` or small hand-written fakes.
- HTTP: `net/http/httptest` against the real handler.
- Cover the denial paths explicitly: missing, expired, pending, hits present, unknown user, repository failure (must fail closed).
- Use `go test -race ./...` where cgo is available (CI/Linux). Tests must be deterministic (no wall clock, no network).

## Code style and best practices

- Run `go fmt ./... && go vet ./... && go test ./...` before finishing any change.
- Accept interfaces, return structs. Define interfaces where they are consumed, keep them small.
- Pass `context.Context` as the first parameter to anything that does or may do I/O.
- No global state, no `init()` side effects, no package-level mutable variables.
- Constructors validate; avoid exported struct fields on domain types.
- Don't log or return PII beyond the user ID. No names, DOBs, or document data in logs or errors.
- Sanctions/KYC data is regulated: decisions should be auditable. Log the decision, user ID, action, reason codes, and the screening ID relied upon, using structured `slog` fields.
- Keep packages small and named for what they provide (`kyc`, not `utils`/`common`).
- Comments explain *why*; exported identifiers get doc comments.
- Don't add abstractions, config, or dependencies ahead of a concrete need.

## Working agreements for agents

- Make the smallest change that satisfies the task; don't refactor unrelated code.
- Respect the dependency rule above; if a change seems to need violating it, stop and raise it.
- Ask before adding dependencies, new top-level directories, or changing the public JSON contract.
- Don't commit secrets, real user data, or real sanctions-list data. Use synthetic fixtures.
