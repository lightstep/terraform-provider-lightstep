package lightstep

import (
	"encoding/json"
	"fmt"

)

type Condition struct {
	Type          string                 `json:"type,omitempty"`
	ID            string                 `json:"id,omitempty"`
	Attributes    ConditionAttributes 	`json:"attributes"`
	Relationships ConditionRelationships `json:"relationships,omitempty"`
	Links         Links                  `json:"links"`
}

type ConditionAttributes struct {
	Name               string                `json:"name"`
	EvaluationWindowMs int64                  `json:"eval-window-ms"`
	Expression         string                 `json:"expression"`
	CustomData         map[string]interface{} `json:"custom-data,omitempty"`
}

type ConditionRelationships struct {
	Project LinksObj `json:"project"`
	Stream  LinksObj `json:"stream"`
}

//
//func(c *Client) CreateCondition(
//	projectName string,
//	conditionName string,
//	relationships map[string]string,) (Condition, error){
//	bytes, err := json.Marshal(
//		Condition{
//			Type: "condition",
//			Attributes: ConditionAttributes{
//				Name:    conditionName,
//				EvaluationWindowMs: ,
//				Expression:,
//				CustomData: ,
//			},
//			Relationships: ConditionRelationships{
//				Project:
//					Stream:
//			}
//			},
//		})
//	if err != nil {
//		log.Printf("error marshalling data: %v", err)
//	}
//
//	var resp Envelope
//	var cond Condition
//
//	req := Envelope{
//		Data: bytes,
//	}
//
//	err = c.CallAPI("POST", fmt.Sprintf("projects/%v/conditions", projectName), req, &resp)
//	if err != nil {
//		log.Printf("error getting dashboard: %v", err)
//		return cond, err
//	}
//
//	err = json.Unmarshal(resp.Data, &c)
//	if err != nil {
//		log.Printf("error unmarshalling: %v", err)
//		return cond, err
//	}
//	return cond, err
//}

func (c *Client)GetCondition(projectName string, conditionID string) (Condition, error) {

	var cond Condition
	var resp Envelope

	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/conditions/%v", projectName, conditionID), nil, &resp)
	if err != nil {
		fmt.Printf("error getting condition %v", err)
		return cond, err
	}

	err = json.Unmarshal(resp.Data, &cond)
	if err != nil {
		fmt.Printf("error unmarshalling: %v", err)
		return cond, err
	}
	return cond, err
}

//type Condition struct {
//	Type          string                 `json:"type,omitempty"`
//	ID            string                 `json:"id,omitempty"`
//	Attributes    ConditionAttributes 	`json:"attributes"`
//	Relationships ConditionRelationships `json:"relationships,omitempty"`
//	Links         Links                  `json:"links"`
//}
//
//type ConditionAttributes struct {
//	Name               string                `json:"name"`
//	EvaluationWindowMs int64                  `json:"eval-window-ms"`
//	Expression         string                 `json:"expression"`
//	CustomData         map[string]interface{} `json:"custom-data,omitempty"`
//}
//
//type ConditionRelationships struct {
//	Project LinksObj `json:"project"`
//	Stream  LinksObj `json:"stream"`
//}

func(c *Client) UpdateCondition(projectName string, condition Condition) (Condition, error){
	resp := Envelope{}

	bytes, err := json.Marshal(condition)
	if err != nil {
		fmt.Printf("error marshalling %v", err)
	}

	err = c.CallAPI("PATCH", fmt.Sprintf("projects/%v/conditions/%v", projectName, condition.ID),Envelope{Data:bytes}, &resp)
	if err != nil {
		fmt.Printf("error calling api %v", err)
	}
	cond := Condition{}
	err = json.Unmarshal(resp.Data, &cond)
	if err != nil {
		fmt.Printf("error unmarshalling %v", err)
	}
	return cond, err
}

func(c *Client) DeleteCondition(projectName string, conditionID string) error {
	err := c.CallAPI("DELETE", fmt.Sprintf("projects/%v/conditions/%v", projectName, conditionID), nil, nil)
	if err != nil {
		fmt.Printf("error getting condition %v", err)
	}
	return err
}


