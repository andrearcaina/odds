# AGENTS.md

## Project overview

This is a small Go monolith for displaying normalized NFL odds from DraftKings. The server embeds and serves a lightweight HTML, CSS, and Alpine.js frontend.

## Scope

The application is limited to:

- NFL (`leagueId=88808`, `sportId=3`)
- Upcoming games
- Moneyline, Spread, and Total markets

Player props, alternate lines, and other leagues are outside the current scope.

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
```

Direct Go commands are also fine:

```sh
go run ./cmd/server
go test ./...
gofmt -w cmd internal
```

Local configuration belongs in `.env`, which is ignored by Git. Keep `.env.example` free of local URLs and credentials.

## Containers

The `Dockerfile` is for the production image only. Local development should use Go and the Makefile directly.

## Implementation guidelines

- Keep the runtime deterministic. Do not add an AI or agent framework to the application.
- Do not hard-code DraftKings cookies, tokens, or session state.
- Treat `403`, `429`, timeouts, malformed JSON, empty responses, and schema changes as feed failures. Preserve the last successful snapshot and mark it stale when a refresh ultimately fails.
- Keep DraftKings polling centralized in the client.
- Keep caching and request coalescing in the DraftKings client rather than in HTTP handlers.
- Retry only transient network errors, `429` responses, and `5xx` responses with a bounded backoff. Do not retry `403` responses or invalid payloads.
- Update tests when the DraftKings response mapping changes.
- Prefer small, focused changes over introducing unnecessary dependencies or frontend infrastructure.
