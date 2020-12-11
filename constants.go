package main

var (
	validQueryNames = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

	validEvaluationWindowInput = []string{"2m", "5m", "10m", "15m", "30m", "1h", "2h", "4h"}
	validEvaluationWindow      = map[string]int{
		"2m":         120000,
		"5m":         300000,
		"10m":        600000,
		"15m":        900000,
		"30m":        1800000,
		"1h":         3600000,
		"2h":         7200000,
		"4h": 14400000,
	}

	validRenotifyInput = []string{"10m", "20m", "30m", "40m", "50m", "1h", "1.5h", "2h", "3h", "4h", "5h", "6h", "12h", "1d"}
	validRenotify      = map[string]int{
		"10m": 600000,
		"20m": 1200000,
		"30m": 1800000,
		"40m": 2400000,
		"50m": 3000000,
		"1h":  3600000,
		"1":   5400000,
		"2h":  7200000,
		"3h":  10800000,
		"4h":  14400000,
		"5h":  18000000,
		"6h":  21600000,
		"12h": 43200000,
		"1d":  86400000,
	}
)
