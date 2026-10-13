# AGENTS.md

## Project overview

This is a small Go monolith for displaying normalized NFL odds from DraftKings Ontario. The server embeds and serves a lightweight HTML, CSS, and Alpine.js frontend.

## Scope

The application is limited to:

- NFL (`leagueId=88808`, `sportId=3`)
- Upcoming games
- Moneyline, Spread, and Total markets

Player props, alternate lines, and other leagues are outside the current scope.

The page should make it clear when the data is stale. It currently highlights prices that changed since the previous successful snapshot and displays snapshot age and upstream fetch duration.

## Project layout

- `cmd/server` starts the application and handles shutdown.
- `internal/config` loads environment-based configuration.
- `internal/draftkings` fetches, normalizes, caches, and refreshes DraftKings data.
- `internal/web` contains the HTTP handlers and embedded frontend assets.
- `internal/logger` contains the application logging wrapper.

DraftKings provides a full snapshot with `events`, `markets`, and `selections`. Events are joined to markets through `market.eventId`, and markets are joined to selections through `selection.marketId`. Spread and Total markets should use selections tagged `MainPointLine`.

## Development commands

```sh
make run
make test
make build
make check
make tidy
```

Direct Go commands are also fine:

```sh
go run ./cmd/server
go test ./...
gofmt -w cmd internal
```

Local configuration belongs in `.env`, which is ignored by Git. Keep `.env.example` free of local URLs and credentials.

## Implementation guidelines

- Keep the runtime deterministic. Do not add an AI or agent framework to the application.
- Do not hard-code DraftKings cookies, tokens, or session state.
- Treat `403`, `429`, timeouts, malformed JSON, empty responses, and schema changes as feed failures.
- Preserve the last successful snapshot and mark it stale when a refresh fails.
- Keep DraftKings polling centralized in the client.
- Keep caching and request coalescing in the DraftKings client rather than in HTTP handlers.
- Retry network errors, `429` responses, and `5xx` responses with a bounded backoff.
- Update tests when the DraftKings response mapping changes.
- Prefer small, focused changes over introducing unnecessary dependencies or frontend infrastructure.

AI and other tooling can be used during development for tasks such as detecting schema changes or generating mapping tests, but the production parser should remain explicit and predictable.
