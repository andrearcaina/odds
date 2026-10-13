package draftkings

import "testing"

func TestNormalizeKeepsOnlyMainNFLMarkets(t *testing.T) {
	spread := 3.0
	altSpread := 7.0

	payload := rawResponse{
		Events: []rawEvent{{ID: "game-1", LeagueID: "88808", Name: "BUF Bills @ LA Rams", Status: "NOT_STARTED"}},
		Markets: []rawMarket{
			{ID: "ml", EventID: "game-1", LeagueID: "88808", SubcategoryID: "4518", Name: "Moneyline"},
			{ID: "sp", EventID: "game-1", LeagueID: "88808", SubcategoryID: "4518", Name: "Spread"},
			{ID: "alt", EventID: "game-1", LeagueID: "88808", SubcategoryID: "4518", Name: "Spread"},
			{ID: "prop", EventID: "game-1", LeagueID: "88808", SubcategoryID: "12438", Name: "Anytime TD Scorer"},
		},
		Selections: []rawSelection{
			{MarketID: "ml", Label: "BUF Bills", OutcomeType: "Away", DisplayOdds: rawOdds{American: "+145"}},
			{MarketID: "sp", Label: "BUF Bills", OutcomeType: "Away", Points: &spread, Tags: []string{"MainPointLine"}, DisplayOdds: rawOdds{American: "-118"}},
			{MarketID: "alt", Label: "BUF Bills", OutcomeType: "Away", Points: &altSpread, Tags: []string{"Alternate"}, DisplayOdds: rawOdds{American: "+100"}},
			{MarketID: "prop", Label: "Player", DisplayOdds: rawOdds{American: "+500"}},
		},
	}

	games, err := normalize(payload)
	if err != nil {
		t.Fatalf("normalize returned error: %v", err)
	}

	if len(games) != 1 {
		t.Fatalf("games = %d, want 1", len(games))
	}
	if len(games[0].Moneyline) != 1 || len(games[0].Spread) != 1 || len(games[0].Total) != 0 {
		t.Fatalf("unexpected market counts: moneyline=%d spread=%d total=%d", len(games[0].Moneyline), len(games[0].Spread), len(games[0].Total))
	}
}
