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
	copy.Games = append([]Game(nil), c.snapshot.Games...)

	return copy
}

func (c *snapshotCache) Store(games []Game, sourceStatus int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.snapshot = Snapshot{
		Games:        games,
		FetchedAt:    time.Now().UTC(),
		SourceStatus: sourceStatus,
	}
}

func (c *snapshotCache) MarkStale(status int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.snapshot.Stale = true
	c.snapshot.Error = err.Error()
	c.snapshot.SourceStatus = status
}
