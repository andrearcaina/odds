package draftkings

import (
	"errors"
	"testing"
	"time"
)

func TestSnapshotCacheTracksPriceChanges(t *testing.T) {
	cache := newSnapshotCache()
	oldLine := 3.0
	newLine := 3.5

	cache.Store([]Game{{
		ID: "game-1",
		Moneyline: []Price{{
			Side:  "Away",
			Label: "BUF Bills",
			Odds:  "+120",
		}},
		Spread: []Price{{
			Side:  "Away",
			Label: "BUF Bills",
			Line:  &oldLine,
			Odds:  "-110",
		}},
	}}, 200, 100*time.Millisecond)

	cache.Store([]Game{{
		ID: "game-1",
		Moneyline: []Price{{
			Side:  "Away",
			Label: "BUF Bills",
			Odds:  "+125",
		}},
		Spread: []Price{{
			Side:  "Away",
			Label: "BUF Bills",
			Line:  &newLine,
			Odds:  "-115",
		}},
	}}, 200, 125*time.Millisecond)

	snapshot := cache.Snapshot()
	if snapshot.Games[0].Moneyline[0].PreviousOdds != "+120" || snapshot.Games[0].Moneyline[0].PreviousLine != nil {
		t.Fatal("moneyline change was matched with a spread price")
	}

	price := snapshot.Games[0].Spread[0]
	if !price.Changed {
		t.Fatal("expected changed price")
	}
	if price.PreviousOdds != "-110" {
		t.Fatalf("previous odds = %q, want -110", price.PreviousOdds)
	}
	if price.PreviousLine == nil || *price.PreviousLine != oldLine {
		t.Fatalf("previous line = %v, want %v", price.PreviousLine, oldLine)
	}
}

func TestSnapshotCachePreservesDataWhenStale(t *testing.T) {
	cache := newSnapshotCache()
	cache.Store([]Game{{ID: "game-1"}}, 200, time.Millisecond)

	cache.MarkStale(429, errors.New("rate limited"))
	snapshot := cache.Snapshot()

	if len(snapshot.Games) != 1 || snapshot.Games[0].ID != "game-1" {
		t.Fatal("expected last successful games to remain available")
	}
	if !snapshot.Stale {
		t.Fatal("expected snapshot to be stale")
	}
	if snapshot.Error != "rate limited" {
		t.Fatalf("error = %q, want rate limited", snapshot.Error)
	}
}
