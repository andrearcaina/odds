package draftkings

import "time"

type Snapshot struct {
	Games        []Game    `json:"games"`
	FetchedAt    time.Time `json:"fetchedAt"`
	Stale        bool      `json:"stale"`
	Error        string    `json:"error,omitempty"`
	SourceStatus int       `json:"sourceStatus,omitempty"`
}

type Game struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	StartTime string  `json:"startTime"`
	Status    string  `json:"status"`
	Moneyline []Price `json:"moneyline,omitempty"`
	Spread    []Price `json:"spread,omitempty"`
	Total     []Price `json:"total,omitempty"`
}

type Price struct {
	Side  string   `json:"side"`
	Label string   `json:"label"`
	Line  *float64 `json:"line,omitempty"`
	Odds  string   `json:"odds"`
}

type rawResponse struct {
	Events     []rawEvent     `json:"events"`
	Markets    []rawMarket    `json:"markets"`
	Selections []rawSelection `json:"selections"`
}

type rawEvent struct {
	ID             string           `json:"id"`
	LeagueID       string           `json:"leagueId"`
	Name           string           `json:"name"`
	StartEventDate string           `json:"startEventDate"`
	Status         string           `json:"status"`
	Participants   []rawParticipant `json:"participants"`
}

type rawParticipant struct {
	Name      string `json:"name"`
	VenueRole string `json:"venueRole"`
}

type rawMarket struct {
	ID            string `json:"id"`
	EventID       string `json:"eventId"`
	Name          string `json:"name"`
	LeagueID      string `json:"leagueId"`
	SubcategoryID string `json:"subcategoryId"`
}

type rawSelection struct {
	MarketID    string   `json:"marketId"`
	Label       string   `json:"label"`
	Points      *float64 `json:"points"`
	OutcomeType string   `json:"outcomeType"`
	DisplayOdds rawOdds  `json:"displayOdds"`
	Tags        []string `json:"tags"`
}

type rawOdds struct {
	American string `json:"american"`
}
