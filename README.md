# DraftKings NFL Odds

A small Go application that fetches NFL odds from DraftKings Ontario and displays them in a browser.

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

Set the DraftKings league endpoint in `.env`:

```env
PORT=8080
REFRESH_INTERVAL=5
HTTP_TIMEOUT=10
DRAFTKINGS_URL=<DraftKings NFL league endpoint>
```

The interval and timeout values are in seconds. Start the application with:

```sh
make run
```

Then open <http://localhost:8080>.

Other useful commands:

```sh
make test
make build
make check
```

`.env` is used for local development and is ignored by Git. Production environments can provide the same values as regular environment variables.

## How it works

DraftKings returns a complete snapshot containing events, markets, and selections. The Go client joins those records together using their event and market IDs, keeps upcoming NFL games, and selects only the Moneyline, Spread, and Total markets.

Spread and Total selections use DraftKings’ `MainPointLine` tag so alternate lines are not displayed. The normalized result is stored in an in-memory cache and served to the frontend through `/api/odds`.

The server refreshes DraftKings on a timer. The browser polls the Go server at the same interval, and the Refresh button can trigger an immediate update. Multiple refreshes happening at the same time are coalesced into one upstream request.

If DraftKings is unavailable, times out, rate-limits the request, or returns data the client cannot parse, the last successful snapshot remains available and the UI marks it as stale.

## Why this approach

The DraftKings page already exposes structured JSON, so using that feed is more reliable than scraping rendered HTML. It also provides the relationships needed to map each game to its markets and selections.

Fetching through one Go process means the application does not make a separate DraftKings request for every visitor. Visitors receive the same cached snapshot, while the client handles polling, failures, and request coalescing in one place.

The frontend uses plain HTML, CSS, and Alpine.js. That keeps the page small while still providing reactive updates for the table, refresh button, loading state, and stale-data indicator.

## Freshness

With `REFRESH_INTERVAL=5`, the displayed odds are normally no more than about five seconds old, plus the time required for the DraftKings request and the browser request. This is polling rather than a true streaming connection, since the endpoint does not provide a public source-change timestamp or a documented fixed rate limit.

One running instance makes approximately twelve upstream requests per minute at a five-second interval. Visitors do not each create their own DraftKings request because they share the server-side snapshot.

The public odds response does not require a login, token, or persistent cookie. Geo-restrictions, bot protection, and rate limits depend on the hosting location and DraftKings environment, so the application reports upstream failures and keeps the last valid data visible instead of hiding the problem.

## Extending the application

Additional sportsbooks could be supported by giving each sportsbook its own client and normalizer while keeping the same internal game, market, and price model. A second league could use the same structure with a different league configuration and endpoint.

AI or developer tooling could help identify upstream schema changes, generate mapping tests, and flag missing markets. It should not be responsible for parsing the known response format at runtime.
