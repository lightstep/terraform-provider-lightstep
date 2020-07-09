package lightstep

import "encoding/json"

type Response struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type ResourceIDObject struct {
	Response
}

type APIResponse struct {
	Data string `json:"data"`
}
type Links map[string]string

type LinksObj struct {
	Links Links `json:"links"`
}

type Envelope struct {
	Data  json.RawMessage `json:"data"`
	Links Links           `json:"links,omitempty"`
}
