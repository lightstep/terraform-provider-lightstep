package lightstep

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Stream struct {
	Type       string           `json:"type,omitempty"`
	ID         string           `json:"id,omitempty"`
	Attributes StreamAttributes `json:"attributes,omitempty"`
}

type StreamAttributes struct {
	Name  string `json:"name"`
	Query string `json:"query"`

	// "custom_data" on set, but "custom-data" on get
	CustomData map[string]map[string]string `json:"custom_data,omitempty"`

	// Hack until https://lightstep.atlassian.net/browse/LS-26494 is fixed.
	CustomDataGet map[string]map[string]string `json:"custom-data,omitempty"`
}

func CustomDataConvert(customData []interface{}) map[string]map[string]string {

	// This is what Lightstep expects
	//"custom_data": {
	//	"object1": {
	//		"url": "http://",
	//		"key": "value"
	//	},
	//	"object2": {
	//		"key": "value"
	//	}
	//},
	lsCustomData := make(map[string]map[string]string)

	// This is what we have (terraform doesn't support a map of maps natively.
	//	custom_data = [
	//	  {
	//      // This name field is special and becomes the key
	//	  	"name": "object1"
	//  	"url" = "https://lightstep.atlassian.net/l/c/M7b0rBsj",
	//      "key" = "value",
	//    },
	//  ]

	// The "name" key is special and must exist
	for _, value := range customData {
		v := value.(map[string]interface{})
		name := v["name"].(string)

		lsCustomData[name] = make(map[string]string)
		for key, value := range v {
			if key == "name" {
				continue
			}
			lsCustomData[name][key] = value.(string)
		}
	}

	return lsCustomData
}

func (c *Client) CreateStream(
	projectName string,
	name string,
	query string,
	customData []interface{},
) (Stream, error) {

	var (
		s    Stream
		resp Envelope
	)

	lsCustomData := CustomDataConvert(customData)

	bytes, err := json.Marshal(
		Stream{
			Type: "stream",
			Attributes: StreamAttributes{
				Name:       name,
				Query:      query,
				CustomData: lsCustomData,
			},
		})
	if err != nil {
		return s, err
	}

	err = c.CallAPI(
		"POST",
		fmt.Sprintf("projects/%v/streams", projectName),
		Envelope{Data: bytes},
		&resp)
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(resp.Data, &s)
	if err != nil {
		return s, err
	}

	return s, err
}

func (c *Client) ListStreams(projectName string) ([]Stream, error) {
	var (
		s    []Stream
		resp Envelope
	)

	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/streams", projectName), nil, &resp)
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(resp.Data, &s)
	if err != nil {
		return s, err
	}
	return s, err
}

func (c *Client) GetStream(projectName string, StreamID string) (*Stream, error) {
	var (
		s    *Stream
		resp Envelope
	)

	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/streams/%v", projectName, StreamID), nil, &resp)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp.Data, &s)
	if err != nil {
		return s, err
	}
	return s, err
}

func (c *Client) UpdateStream(projectName string,
	streamID string,
	stream Stream,
) (Stream, error) {

	var (
		s    Stream
		resp Envelope
	)

	bytes, err := json.Marshal(&stream)
	if err != nil {
		return s, err
	}

	err = c.CallAPI("PATCH", fmt.Sprintf("projects/%v/streams/%v", projectName, streamID), Envelope{Data: bytes}, &resp)
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(resp.Data, &s)
	return s, err
}

func (c *Client) DeleteStream(projectName string, StreamID string) error {
	err := c.CallAPI("DELETE", fmt.Sprintf("projects/%v/streams/%v", projectName, StreamID), nil, nil)
	if err != nil {
		apiClientError := err.(APIResponseCarrier)
		if apiClientError.GetHTTPResponse().StatusCode != http.StatusNoContent {
			return err
		}
	}
	return nil
}
