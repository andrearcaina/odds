package draftkings

import (
	"context"
	"fmt"
	"time"

	"github.com/andrearcaina/draftkings-odds/internal/logger"
	"resty.dev/v3"
)

type Client struct {
	url      string
	interval time.Duration
	client   *resty.Client
	logger   *logger.Logger

	cache *snapshotCache
}

const maxAttempts = 3

var retryDelays = []time.Duration{
	250 * time.Millisecond,
	500 * time.Millisecond,
}

func NewClient(url string, interval, timeout time.Duration, logger *logger.Logger) *Client {
	return &Client{
		url:      url,
		interval: interval,
		client:   resty.New().SetTimeout(timeout),
		logger:   logger,
		cache:    newSnapshotCache(),
	}
}

func (c *Client) Run(ctx context.Context) {
	c.refresh(ctx)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.refresh(ctx)
		}
	}
}

func (c *Client) Snapshot() Snapshot {
	return c.cache.Snapshot()
}

func (c *Client) refresh(ctx context.Context) {
	c.refreshUpstream(ctx)
}

func (c *Client) refreshUpstream(ctx context.Context) {
	startedAt := time.Now()
	var payload rawResponse
	var response *resty.Response
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		payload = rawResponse{}
		response, err = c.client.R().
			SetContext(ctx).
			SetHeaders(map[string]string{
				"Accept":     "application/json",
				"Referer":    "https://sportsbook.draftkings.com/",
				"Origin":     "https://sportsbook.draftkings.com",
				"User-Agent": "draftkings-nfl-odds/1.0",
			}).
			SetResult(&payload).
			Get(c.url)

		if !shouldRetry(response, err) || attempt == maxAttempts {
			break
		}

		c.logger.Warn(ctx, "odds refresh retrying", "attempt", attempt+1, "max_attempts", maxAttempts)
		if !waitBeforeRetry(ctx, retryDelays[attempt-1]) {
			return
		}
	}

	if err != nil {
		c.fail(0, err)
		return
	}

	if response.StatusCode() < 200 || response.StatusCode() >= 300 {
		c.fail(response.StatusCode(), fmt.Errorf("DraftKings feed unavailable (HTTP %d)", response.StatusCode()))
		return
	}

	games, err := normalize(payload)
	if err != nil {
		c.fail(response.StatusCode(), err)
		return
	}

	fetchDuration := time.Since(startedAt)
	c.cache.Store(games, response.StatusCode(), fetchDuration)
	c.logger.Info(ctx, "odds refreshed", "games", len(games), "fetch_duration_ms", fetchDuration.Milliseconds())
}

func shouldRetry(response *resty.Response, err error) bool {
	if err != nil {
		return true
	}
	if response == nil {
		return false
	}

	return retryableStatus(response.StatusCode())
}

func retryableStatus(status int) bool {
	return status == 429 || status >= 500
}

func waitBeforeRetry(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (c *Client) fail(status int, err error) {
	c.cache.MarkStale(status, err)
	c.logger.Warn(context.Background(), "odds refresh failed", "status", status, "error", err)
}

func normalize(payload rawResponse) ([]Game, error) {
	marketByID := make(map[string]rawMarket, len(payload.Markets))

	for _, market := range payload.Markets {
		if market.LeagueID == "88808" && market.SubcategoryID == "4518" && isMainMarket(market.Name) {
			marketByID[market.ID] = market
		}
	}
	if len(payload.Events) == 0 {
		return nil, fmt.Errorf("upstream response contained no events")
	}

	byEvent := make(map[string][]rawSelection)

	for _, selection := range payload.Selections {
		if _, ok := marketByID[selection.MarketID]; ok {
			byEvent[selection.MarketID] = append(byEvent[selection.MarketID], selection)
		}
	}

	games := make([]Game, 0, len(payload.Events))

	for _, event := range payload.Events {
		if event.LeagueID != "88808" || event.Status != "NOT_STARTED" {
			continue
		}
		game := Game{
			ID:        event.ID,
			Name:      event.Name,
			StartTime: event.StartEventDate,
			Status:    event.Status,
		}

		for marketID, market := range marketByID {
			if market.EventID != event.ID {
				continue
			}
			for _, selection := range byEvent[marketID] {
				isLineMarket := market.Name == "Spread" || market.Name == "Total"
				if isLineMarket && !contains(selection.Tags, "MainPointLine") {
					continue
				}

				price := Price{
					Side:  selection.OutcomeType,
					Label: selection.Label,
					Line:  selection.Points,
					Odds:  selection.DisplayOdds.American,
				}

				switch market.Name {
				case "Moneyline":
					game.Moneyline = append(game.Moneyline, price)
				case "Spread":
					game.Spread = append(game.Spread, price)
				case "Total":
					game.Total = append(game.Total, price)
				}
			}
		}
		games = append(games, game)
	}

	return games, nil
}
