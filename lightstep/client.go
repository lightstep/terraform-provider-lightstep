package lightstep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Headers map[string]string

type APIResponseCarrier interface {
	GetHTTPResponse() *http.Response
}

// APIClientError contains the HTTP Response(for inspection of the error code) as well as the error message
type APIClientError struct {
	Response *http.Response
	Message  string
}

func (a APIClientError) Error() string {
	return a.Message
}
func (a APIClientError) GetHTTPResponse() *http.Response {
	return a.Response
}

type Body struct {
	Data   interface{}            `json:"data,omitempty"`
	Errors []string               `json:"errors,omitempty"`
	Links  map[string]interface{} `json:"links,omitempty"`
}

type Client struct {
	apiKey      string
	baseURL     string
	client      *http.Client
	contentType string
}

// NewClient gets a client for the public API
func NewClient(ctx context.Context, apiKey string, orgName string) *Client {
	baseUrl := os.Getenv("LIGHTSTEP_HOST")
	if baseUrl == "" {
		baseUrl = "https://api-staging.lightstep.com/public/v0.1/" // Hardcoding to staging for now
	}
	baseURLWithOrg := fmt.Sprintf("%v/%v/", baseUrl, orgName)

	return &Client{
		apiKey:      apiKey,
		baseURL:     baseURLWithOrg,
		client:      http.DefaultClient,
		contentType: "application/vnd.api+json",
	}
}

// CallAPI calls the given API and unmarshals the result to into result.
func (c *Client) CallAPI(httpMethod string, suffix string, data interface{}, result interface{}) error {
	return callAPI(
		context.TODO(),
		c.client,
		fmt.Sprintf("%v/%v", c.baseURL, suffix),
		httpMethod,
		Headers{
			"Authorization": fmt.Sprintf("bearer %v", c.apiKey),
			"Content-Type":  c.contentType,
			"Accept":        c.contentType,
		},
		data,
		result,
	)
}

func executeAPIRequest(client *http.Client, req *http.Request, result interface{}) error {
	resp, err := client.Do(req)
	if err != nil {
		return APIClientError{
			Response: resp,
			Message:  fmt.Sprintf("%v failed: %v: %v", req.Method, req.URL, err),
		}
	}
	defer resp.Body.Close() // nolint: errcheck

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return APIClientError{
			Response: resp,
			Message:  fmt.Sprintf("status %d (%s): %q", resp.StatusCode, resp.Status, string(body)),
		}
	}

	if len(body) == 0 {
		return APIClientError{
			Response: resp,
			Message:  fmt.Sprintf("body empty. status=%v", resp.StatusCode),
		}
	}

	contentType := resp.Header.Get("Content-Type")
	if len(contentType) > 0 && contentType != req.Header.Get("Accept") {
		return APIClientError{
			Response: resp,
			Message: fmt.Sprintf(
				"content type (%s) is not \"%s\": %q",
				contentType,
				req.Header.Get("Accept"),
				string(body),
			),
		}
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return APIClientError{
				Response: resp,
				Message:  fmt.Sprintf("status %d (%s): %q: %v", resp.StatusCode, resp.Status, string(body), err),
			}
		}
	}

	return nil
}

func createJSONRequest(
	ctx context.Context,
	httpMethod string,
	url string,
	data interface{},
	headers map[string]string,
) (*http.Request, error) {
	var body io.Reader

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		if !httpMethodSupportsRequestBody(httpMethod) {
			log.Printf("this HTTP method does not support a request body: %v", httpMethod)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(httpMethod, url, body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

// callAPI is a helper function that enables flexibly issuing an API Request
func callAPI(
	ctx context.Context,
	client *http.Client,
	url string,
	httpMethod string,
	headers Headers,
	data interface{},
	result interface{},
) error {
	req, err := createJSONRequest(
		ctx,
		httpMethod,
		url,
		data,
		headers,
	)
	if err != nil {
		log.Print(err)
		return err
	}

	// Do the request.
	return executeAPIRequest(client, req, result)
}

func httpMethodSupportsRequestBody(method string) bool {
	return method != "GET" && method != "DELETE"
}
