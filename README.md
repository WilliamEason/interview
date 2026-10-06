# KYC Service

Domain-driven Go service that answers: *has this user passed the KYC checks required to perform this action?*

First use case: **can User A BUY?** — requires a valid sanctions screening with no hits.

See [AGENTS.md](AGENTS.md) for project layout, tooling, and conventions.

## Development

```
go fmt ./... && go vet ./... && go test ./...
go run ./cmd/kycd
```

## Try it

Start the server (seeded with synthetic demo users) on `localhost:8080`:

```
go run ./cmd/kycd
```

User A (valid screening, no hits) is approved:

```
curl -X POST localhost:8080/v1/users/11111111-1111-4111-8111-111111111111/eligibility   -d '{"ACTION":"BUY"}'
# {"USER_ID":"11111111-1111-4111-8111-111111111111","ACTION":"BUY","PERMISSION":"APPROVED"}
```

User B (sanctions hit) is denied, and the server logs `sanctions hit <user_id>`:

```
curl -X POST localhost:8080/v1/users/22222222-2222-4222-8222-222222222222/eligibility   -d '{"ACTION":"BUY"}'
# {"USER_ID":"22222222-2222-4222-8222-222222222222","ACTION":"BUY","PERMISSION":"DENIED","REASONS":["SANCTIONS_HIT"]}
```
