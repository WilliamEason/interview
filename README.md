# KYC Service

Domain-driven Go service that answers: *has this user passed the KYC checks required to perform this action?*

First use case: **can User A BUY?** — requires a valid sanctions screening with no hits.

See [AGENTS.md](AGENTS.md) for project layout, tooling, and conventions.

## Development

```
go fmt ./... && go vet ./... && go test ./...
go run ./cmd/kycd
```
