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
make test        # runs all test files
make fmt         # formats all Go files
make fmt-check   # checks formatting
make build       # builds the server binary
make vet         # runs go vet on all files
make tidy        # runs go mod tidy
make check       # runs fmt-check, tidy, test, vet build
```

## Production container

The `Dockerfile` is for the production build only. Local development only requires Go and the Makefile commands above. The production image uses a multi-stage build and runs the compiled server in a minimal non-root image:

```sh
docker build -t draftkings-odds .
docker run --env-file .env -p 8080:8080 draftkings-odds
```

`.env` is used for local development and is ignored by Git. Production environments can provide the same values as regular environment variables.

## How it works

DraftKings provides a full JSON snapshot of events, markets, and selections. The Go client joins these records by ID, keeps upcoming NFL games, and selects the main Moneyline, Spread, and Total markets. The normalized result is stored in memory and served through `/api/odds`.

The data flow is:

```plaintext
DraftKings feed -> validate and normalize -> cache snapshot -> browser polls /api/odds every few seconds -> display in table
```

The server refreshes the feed on a timer, while the browser uses short polling to read the cached snapshot. All users read the same snapshot, so their requests do not create extra DraftKings requests. With one server instance, only its background loop polls DraftKings. Price changes include the previous value so the frontend can highlight movement.

Transient network errors, `429` responses, and `5xx` responses are retried. If refreshes fail, the last successful snapshot remains visible and is marked stale.

## Why this approach

The DraftKings page already exposes structured JSON, so using that feed is more reliable than scraping rendered HTML. It also provides the relationships needed to map each game to its markets and selections.

The frontend uses plain HTML, CSS, and Alpine.js. That keeps the page small while still providing reactive updates for the table, loading state, and stale-data indicator.

## How updated is the data

With `REFRESH_INTERVAL=5`, the displayed odds are normally no more than about five seconds old, plus request time. I chose five seconds because it keeps the odds reasonably fresh without constantly requesting DraftKings, which means one server instance makes about twelve upstream requests per minute. The page shows the snapshot age and fetch duration, and marks the data stale after two missed polling intervals.

## Hosting

I initially deployed the app on Render, but DraftKings blocked its Virginia datacenter IP and returned `403` responses. Render did not offer a Canadian region for this deployment, so I moved the app to Fly.io in Toronto (`yyz`), where the DraftKings feed works successfully.

## Extending the application

If I added another sportsbook, I would give it its own client and normalizer while keeping the same internal game, market, and price model. If I added another league, I would reuse the same structure with a different league configuration and endpoint.

I could use AI or other developer tools to identify upstream schema changes, generate mapping tests, and flag missing markets.
