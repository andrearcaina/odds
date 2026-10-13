package draftkings

import (
	"sync"
	"time"
)

type snapshotCache struct {
	mu       sync.RWMutex
	snapshot Snapshot
}

func newSnapshotCache() *snapshotCache {
	return &snapshotCache{
		snapshot: Snapshot{
			Games: []Game{},
			Stale: true,
			Error: "Waiting for first upstream response",
		},
	}
}

func (c *snapshotCache) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	copy := c.snapshot
	copy.Games = cloneGames(c.snapshot.Games)

	return copy
}

func (c *snapshotCache) Store(games []Game, sourceStatus int, fetchDuration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	annotateChanges(c.snapshot.Games, games)

	c.snapshot = Snapshot{
		Games:           cloneGames(games),
		FetchedAt:       time.Now().UTC(),
		FetchDurationMS: fetchDuration.Milliseconds(),
		SourceStatus:    sourceStatus,
	}
}

func (c *snapshotCache) MarkStale(status int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.snapshot.Stale = true
	c.snapshot.Error = err.Error()
	c.snapshot.SourceStatus = status
}

func annotateChanges(previous, current []Game) {
	previousPrices := make(map[string]Price)

	for _, game := range previous {
		for market, prices := range map[string][]Price{
			"Moneyline": game.Moneyline,
			"Spread":    game.Spread,
			"Total":     game.Total,
		} {
			for _, price := range prices {
				previousPrices[priceKey(game.ID, market, price)] = price
			}
		}
	}

	for gameIndex := range current {
		for priceIndex := range current[gameIndex].Moneyline {
			annotatePrice(current[gameIndex].ID, "Moneyline", &current[gameIndex].Moneyline[priceIndex], previousPrices)
		}
		for priceIndex := range current[gameIndex].Spread {
			annotatePrice(current[gameIndex].ID, "Spread", &current[gameIndex].Spread[priceIndex], previousPrices)
		}
		for priceIndex := range current[gameIndex].Total {
			annotatePrice(current[gameIndex].ID, "Total", &current[gameIndex].Total[priceIndex], previousPrices)
		}
	}
}

func annotatePrice(gameID, market string, current *Price, previousPrices map[string]Price) {
	previous, ok := previousPrices[priceKey(gameID, market, *current)]
	if !ok {
		return
	}

	if previous.Odds == current.Odds && sameLine(previous.Line, current.Line) {
		current.PreviousOdds = previous.PreviousOdds
		current.PreviousLine = previous.PreviousLine
		return
	}

	current.PreviousOdds = previous.Odds
	current.PreviousLine = previous.Line
	current.Changed = true
}

func priceKey(gameID, market string, price Price) string {
	return gameID + "|" + market + "|" + price.Side + "|" + price.Label
}

func sameLine(left, right *float64) bool {
	if left == nil || right == nil {
		return left == right
	}

	return *left == *right
}

func cloneGames(games []Game) []Game {
	cloned := make([]Game, len(games))
	copy(cloned, games)

	for index := range cloned {
		cloned[index].Moneyline = append([]Price(nil), games[index].Moneyline...)
		cloned[index].Spread = append([]Price(nil), games[index].Spread...)
		cloned[index].Total = append([]Price(nil), games[index].Total...)
	}

	return cloned
}
