package lightstep_sdk

import "time"

type ListServicesAPIResponse struct {
	Data *ListServicesResponse `json:"data,omitempty"`
}

type ListServicesResponse struct {
	Type  string            `json:"type"`
	Links Links             `json:"links"`
	Items []ServiceResponse `json:"items"`
}

type ListServicesRequest struct {
	Offset *int `json:"offset"`
	Limit  *int `json:"limit"`
}

type ServiceResponse struct {
	ID            string               `json:"id"`
	Attributes    ServiceAttributes    `json:"attributes,omitempty"`
	Relationships ServiceRelationships `json:"relationships,omitempty"`
}

type ServiceAttributes struct {
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type ServiceRelationships struct {
	Project LinksObj `json:"project"`
}
