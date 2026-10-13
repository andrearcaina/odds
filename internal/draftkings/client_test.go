package draftkings

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andrearcaina/draftkings-odds/internal/logger"
)

func TestClientPreservesSnapshotAfterUpstreamFailure(t *testing.T) {
	firstResponse := rawResponse{
		Events: []rawEvent{{
			ID:       "game-1",
			LeagueID: "88808",
			Name:     "BUF Bills @ LA Rams",
			Status:   "NOT_STARTED",
		}},
		Markets: []rawMarket{{
			ID:            "ml-1",
			EventID:       "game-1",
			LeagueID:      "88808",
			SubcategoryID: "4518",
			Name:          "Moneyline",
		}},
		Selections: []rawSelection{{
			MarketID:    "ml-1",
			Label:       "BUF Bills",
			OutcomeType: "Away",
			DisplayOdds: rawOdds{American: "+145"},
		}},
	}

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount > 1 {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(firstResponse); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, time.Second, time.Second, logger.New())
	client.refresh(t.Context())

	initial := client.Snapshot()
	if initial.Stale {
		t.Fatal("initial successful snapshot should not be stale")
	}
	if len(initial.Games) != 1 || initial.Games[0].ID != "game-1" {
		t.Fatal("expected initial snapshot to contain game-1")
	}

	client.refresh(t.Context())

	stale := client.Snapshot()
	if len(stale.Games) != 1 || stale.Games[0].ID != "game-1" {
		t.Fatal("expected the last successful games to remain available")
	}
	if !stale.Stale {
		t.Fatal("expected the snapshot to be stale after the upstream failure")
	}
	if stale.SourceStatus != http.StatusForbidden {
		t.Fatalf("source status = %d, want %d", stale.SourceStatus, http.StatusForbidden)
	}
}
