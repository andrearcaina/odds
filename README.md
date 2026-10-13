# DraftKings NFL Odds

A small Go application that fetches NFL odds from DraftKings and displays them in a browser.

The app currently shows the main markets for upcoming NFL games:

- Moneyline
- Spread
- Total

Player props, alternate lines, and other leagues are left out.

## Running locally

Create a local environment file:

```sh
cp .env.example .env
```

Set `DRAFTKINGS_URL` in `.env` to the public DraftKings feed you want to use. It does not contain credentials. The `REFRESH_INTERVAL` and `HTTP_TIMEOUT` values are also configurable.

To start the application, run:

```sh
make run
```

Or run the server directly:

```sh
go run ./cmd/server
```

Then open <http://localhost:8080>.

Other useful commands:

```sh
make test
make build
make check
```

## Production container and CI

The `Dockerfile` is for the production build only. Local development only requires Go and the Makefile commands above. The production image uses a multi-stage build and runs the compiled server in a minimal non-root image:

```sh
docker build -t draftkings-odds .
docker run --env-file .env -p 8080:8080 draftkings-odds
```

GitHub Actions runs the formatting check, module tidy check, tests, vet, and build. After CI succeeds on `main`, a separate workflow builds the production image and publishes it to GHCR with both the commit SHA and `latest` tags.

`.env` is used for local development and is ignored by Git. Production environments can provide the same values as regular environment variables.

## How it works

DraftKings returns a complete snapshot containing events, markets, and selections. The Go client joins those records together using their event and market IDs, keeps upcoming NFL games, and selects only the Moneyline, Spread, and Total markets.

Spread and Total selections use DraftKings’ `MainPointLine` tag so alternate lines are not displayed. The normalized result is stored in an in-memory cache and served to the frontend through `/api/odds`.

The server refreshes DraftKings on a timer. The browser polls the Go server at the same interval, and the Refresh button can trigger an immediate update. Multiple refreshes happening at the same time are coalesced into one upstream request.

When a price changes between successful snapshots, the API includes the previous price and the frontend highlights the current value.

Network errors, `429` responses, and `5xx` responses are retried up to two additional times with a short backoff. If all attempts fail, or DraftKings returns data the client cannot parse, the last successful snapshot remains available and the UI marks it as stale.

## Why this approach

The DraftKings page already exposes structured JSON, so using that feed is more reliable than scraping rendered HTML. It also provides the relationships needed to map each game to its markets and selections.

The frontend uses plain HTML, CSS, and Alpine.js. That keeps the page small while still providing reactive updates for the table, refresh button, loading state, and stale-data indicator.

## Freshness

With `REFRESH_INTERVAL=5`, the displayed odds are normally no more than about five seconds old, plus request time. The page shows the snapshot age and upstream fetch duration, and marks the snapshot stale after two polling intervals. Five seconds is a practical middle ground: the odds stay fresh while one instance makes about twelve upstream requests per minute instead of constantly requesting DraftKings. The interval can be changed, and request coalescing prevents several users refreshing at the same time from creating identical requests.

This is polling rather than a true streaming connection, since the endpoint does not provide a public source-change timestamp or a documented fixed rate limit.

The public odds response does not require a login, token, or persistent cookie. Geo-restrictions, bot protection, and rate limits depend on the hosting location and DraftKings environment, so the application reports upstream failures and keeps the last valid data visible instead of hiding the problem.

## Extending the application

Additional sportsbooks could be supported by giving each sportsbook its own client and normalizer while keeping the same internal game, market, and price model. A second league could use the same structure with a different league configuration and endpoint.

AI or developer tooling could help identify upstream schema changes, generate mapping tests, and flag missing markets. It should not be responsible for parsing the known response format at runtime.
