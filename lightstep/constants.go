package lightstep

var (
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
		"2h":  7200000,
		"3h":  10800000,
		"4h":  14400000,
		"5h":  18000000,
		"6h":  21600000,
		"12h": 43200000,
		"1d":  86400000,
		"7d":  604800000,
		"14d": 1209600000,
	}
)

func GetValidUpdateInterval() []string {
	var res []string
	for k := range validUpdateInterval {
		res = append(res, k)
	}
	return res
}

func GetUpdateIntervalValue(in int) interface{} {
	for k, v := range validUpdateInterval {
		if v == in {
			return k
		}
	}
	if in == 0 {
		return ""
	}
	// This isn't a valid value according to the validation fun in the schema. The only
	// reason we return this here is so terraform can tell there's a difference between
	// no update interval and an update interval not supported by terraform.
	return "invalid"
}
