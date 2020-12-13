package main

var (
	validEvaluationWindow = map[string]int{
		"2m":  120000,
		"5m":  300000,
		"10m": 600000,
		"15m": 900000,
		"30m": 1800000,
		"1h":  3600000,
		"2h":  7200000,
		"4h":  14400000,
		"1d":  86400000,
	}

	validUpdateInterval = map[string]int{
		"2m":  120000,
		"5m":  300000,
		"10m": 600000,
		"15m": 900000,
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

func GetValidEvaluationWindows() []string {
	var res []string
	for k := range validEvaluationWindow {
		res = append(res, k)
	}
	return res
}

func GetValidUpdateInterval() []string {
	var res []string
	for k := range validUpdateInterval {
		res = append(res, k)
	}
	return res
}

func GetEvaluationWindowValue(in int) string {
	for k, v := range validEvaluationWindow {
		if v == in {
			return k
		}
	}
	return ""
}

func GetUpdateIntervalValue(in int) string {
	for k, v := range validUpdateInterval {
		if v == in {
			return k
		}
	}
	return ""
}
