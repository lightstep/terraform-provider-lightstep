package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/time/rate"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	DefaultRateLimitPerSecond = 5
	DefaultRetryMax           = 3
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

// Envelope represents a generic response from the API
type Envelope struct {
	Data json.RawMessage `json:"data"`
}

type Body struct {
	Data   interface{}            `json:"data,omitempty"`
	Errors []string               `json:"errors,omitempty"`
	Links  map[string]interface{} `json:"links,omitempty"`
}

type Client struct {
	apiKey      string
	baseURL     string
	orgName     string
	client      *retryablehttp.Client
	rateLimiter *rate.Limiter
	contentType string
	context     context.Context
}

// NewClient gets a client for the public API
func NewClient(ctx context.Context, apiKey string, orgName string, env string) *Client {
	var baseURL string

	if env == "public" {
		baseURL = fmt.Sprintf("https://api.lightstep.com/public/v0.2/%v", orgName)
	} else {
		baseURL = fmt.Sprintf("https://api-%v.lightstep.com/public/v0.2/%v", env, orgName)
	}

	return &Client{
		apiKey:      apiKey,
		orgName:     orgName,
		baseURL:     baseURL,
		context:     ctx,
		rateLimiter: rate.NewLimiter(rate.Limit(DefaultRateLimitPerSecond), 1),
		client: &retryablehttp.Client{
			HTTPClient:   http.DefaultClient,
			CheckRetry:   checkHTTPRetry,
			RetryWaitMin: 3 * time.Second,
			Backoff:      retryablehttp.DefaultBackoff,
			RetryMax:     DefaultRetryMax,
		},
		contentType: "application/vnd.api+json",
	}
}

// checkHTTPRetry inspects HTTP errors from the Lightstep API for known transient errors
func checkHTTPRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if resp.StatusCode == http.StatusInternalServerError {
		return true, nil
	}
	return false, nil
}

// CallAPI calls the given API and unmarshals the result to into result.
func (c *Client) CallAPI(httpMethod string, suffix string, data interface{}, result interface{}) error {
	return callAPI(
		c.context,
		c,
		fmt.Sprintf("%v/%v", c.baseURL, suffix),
		httpMethod,
		Headers{
			"Authorization":   fmt.Sprintf("bearer %v", c.apiKey),
			"User-Agent":      "terraform-provider-lightstep",
			"X-Lightstep-Org": c.orgName,
			"Content-Type":    c.contentType,
			"Accept":          c.contentType,
		},
		data,
		result,
	)
}

func executeAPIRequest(c *Client, req *retryablehttp.Request, result interface{}) error {
	if err := c.rateLimiter.Wait(c.context); err != nil {
		return err
	}

	resp, err := c.client.Do(req)
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
) (*retryablehttp.Request, error) {
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

	req, err := retryablehttp.NewRequest(httpMethod, url, body)
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
	c *Client,
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
		return err
	}

	// Do the request.
	return executeAPIRequest(c, req, result)
}

func httpMethodSupportsRequestBody(method string) bool {
	return method != "GET" && method != "DELETE"
}

func (c *Client) GetStreamIDByLink(url string) (string, error) {
	response := Envelope{}
	str := Stream{}
	err := callAPI(c.context,
		c,
		url,
		"GET",
		Headers{
			"Authorization": fmt.Sprintf("bearer %v", c.apiKey),
			"Content-Type":  c.contentType,
			"Accept":        c.contentType,
		}, nil, &response)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(response.Data, &str)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response json: %v", err)
	}

	return str.ID, nil
}
