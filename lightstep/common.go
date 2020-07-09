package lightstep

import "encoding/json"

type Envelope struct {
	Data  json.RawMessage `json:"data"`
	Links Links           `json:"links,omitempty"`
}

type LinksObj struct {
	Links Links `json:"links"`
}

type Links map[string]string
