package draftkings

func isMainMarket(name string) bool {
	return name == "Moneyline" || name == "Spread" || name == "Total"
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}

	return false
}
