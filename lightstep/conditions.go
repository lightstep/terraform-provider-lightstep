package lightstep

type ConditionAPIResponse struct {
	Data *ConditionResponse
}

type ConditionResponse struct {
	Response
	Attributes    ConditionAttributes    `json:"attributes,omitempty"`
	Relationships ConditionRelationships `json:"relationships,omitempty"`
	Links         Links                  `json:"links"`
}

type ConditionAttributes struct {
	Name               *string                `json:"name"`
	EvaluationWindowMs int64                  `json:"eval-window-ms"`
	Expression         string                 `json:"expression"`
	CustomData         map[string]interface{} `json:"custom-data,omitempty"`
}

type ConditionRelationships struct {
	Project LinksObj `json:"project"`
	Search  LinksObj `json:"search"`
}

type ListConditionsAPIResponse struct {
	Data *ListConditionsResponse `json:"data,omitempty"`
}

type ListConditionsResponse []ConditionResponse

type ConditionStatusAPIResponse struct {
	Data *ConditionStatusResponse
}

type ConditionStatusResponse struct {
	Response
	Attributes    ConditionStatusAttributes    `json:"attributes,omitempty"`
	Relationships ConditionStatusRelationships `json:"relationships,omitempty"`
}

type ConditionStatusAttributes struct {
	Expression  string `json:"expression"`
	State       string `json:"state"`
	Description string `json:"description"`
}

type ConditionStatusRelationships struct {
	Condition LinksObj `json:"condition"`
}

type ConditionRequestBody struct {
	Data *ConditionRequest `json:"data"`
}

type ConditionRequest struct {
	Response
	Attributes    ConditionRequestAttributes    `json:"attributes"`
	Relationships ConditionRequestRelationships `json:"relationships"`
}

type ConditionRequestAttributes struct {
	Name               *string                 `json:"name"`
	Expression         *string                 `json:"expression"`
	EvaluationWindowMs *int64                  `json:"eval-window-ms"`
	CustomData         *map[string]interface{} `json:"custom-data"`
}

type ConditionRequestRelationships struct {
	Search ResourceIDObject `json:"search"`
}
