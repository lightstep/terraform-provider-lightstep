package lightstep

import (
	"fmt"
	"github.com/lightstep/terraform-provider-lightstep/client"
)

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
		"1":   5400000,
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

func GetUpdateIntervalValue(in int) string {
	for k, v := range validUpdateInterval {
		if v == in {
			return k
		}
	}
	return ""
}

//extractLabels transforms labels from the API call into TF resource labels
func extractLabels(incomingLabels []client.Label) []interface{} {
	var labels []interface{}
	for _, l := range incomingLabels {
		label := map[string]interface{}{}
		if l.Key != "" {
			label["key"] = l.Key
		}
		label["value"] = l.Value
		labels = append(labels, label)
	}
	return labels
}

//BuildLabels transforms labels from the TF resource into labels for the API request
func BuildLabels(labelsIn []interface{}) ([]client.Label, error) {
	var labels []client.Label

	for _, l := range labelsIn {
		label, ok := l.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("bad format, %v", l)
		}

		if len(label) == 0 {
			continue
		}

		// label keys can be omitted for labels without the key:value syntax
		k := label["key"]
		if k == nil {
			k = ""
		}

		key, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("label key must be a string, %v", k)
		}

		v, ok := label["value"].(string)
		if !ok {
			return nil, fmt.Errorf("label value is a required field, %v", v)
		}

		labels = append(labels, client.Label{
			Key:   key,
			Value: v,
		})
	}

	return labels, nil
}
