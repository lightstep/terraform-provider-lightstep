package lightstep_sdk

type Response struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type ResourceIDObject struct {
	Response
}

// Serializes as `{ "links": { "key" : "value" } }`
type LinksObj struct {
	Links Links `json:"links,omitempty"`
}

type Links map[string]interface{}
