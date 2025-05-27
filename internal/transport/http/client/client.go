// Package client sending requests to network services
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client for client
type Client interface {
	Post(ctx context.Context, reqURL string, headers map[string]string, in, out interface{}) error
	Get(ctx context.Context, reqURL string, headers map[string]string, in, out interface{}) error
	Put(ctx context.Context, reqURL string, headers map[string]string, in, out interface{}) error
	GetRawBody(ctx context.Context, reqURL string, headers map[string]string, in interface{}) ([]byte, error)
}

func New(hostname string, httpClient *http.Client, log *zap.SugaredLogger, opts ...Option) Client {
	// Disable SSL/TLS certificate verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	clnt := *httpClient
	clnt.Transport = tr

	c := &client{}
	c.hostname = hostname
	c.log = log
	c.httpClient = &clnt

	defaultOptions := options{}
	for _, o := range opts {
		o.apply(&defaultOptions)
	}

	c.external = defaultOptions.external

	return c
}

type client struct {
	hostname   string
	httpClient *http.Client
	log        *zap.SugaredLogger
	external   bool
}

func (c *client) GetRawBody(ctx context.Context, reqURL string, headers map[string]string, in interface{}) ([]byte, error) {
	var err error

	reqURL = c.requestURL(reqURL)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/pdf")

	httpResponse, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer httpResponse.Body.Close()

	// Check response status code
	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("something went wrong requesting %s", reqURL)
	}

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func (c *client) Get(ctx context.Context, reqURL string, headers map[string]string, in, out interface{}) error {
	var err error

	reqURL = c.requestURL(reqURL)

	var requestBody string

	if value, exists := headers["Content-Type"]; exists && value == "application/x-www-form-urlencoded" {
		requestBody, err = c.prepareRequestBody(in, "urlencoded")
	} else {
		requestBody, err = c.prepareRequestBody(in, "body")
	}

	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, strings.NewReader(requestBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	httpResponse, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}

	return c.parseResponseBody(responseBody, httpResponse.StatusCode, out)
}

func (c *client) Post(ctx context.Context, reqURL string, headers map[string]string, in, out interface{}) error {
	var err error

	reqURL = c.requestURL(reqURL)

	var requestBody string

	if value, exists := headers["Content-Type"]; exists && value == "application/x-www-form-urlencoded" {
		requestBody, err = c.prepareRequestBody(in, "urlencoded")
	} else {
		requestBody, err = c.prepareRequestBody(in, "body")
	}

	if err != nil {
		c.log.Error("Error marshalling request data to JSON", err)
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(requestBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	httpResponse, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		c.log.Error("Error reading response body", err)
		return err
	}
	return c.parseResponseBody(responseBody, httpResponse.StatusCode, out)
}

func (c *client) Put(ctx context.Context, reqURL string, headers map[string]string, in, out interface{}) error {
	var err error

	reqURL = c.requestURL(reqURL)

	// Convert the map to JSON
	jsonData, err := json.Marshal(in)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	httpResponse, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		c.log.Error("Error reading response body", err)
		return err
	}

	return c.parseResponseBody(responseBody, httpResponse.StatusCode, out)
}

func (c *client) prepareRequestBody(req interface{}, reqBodyType string) (string, error) {
	if reqBodyType == "body" {
		requestByteJSON, err := json.Marshal(req)
		if err != nil {
			return "", err
		}

		bodyStr := string(requestByteJSON)
		if bodyStr == "null" {
			bodyStr = ""
		}

		return bodyStr, nil
	} else if reqBodyType == "urlencoded" {
		values := url.Values{}
		// convert req to map[string]string
		data := req.(map[string]string)
		for key, value := range data {
			values.Set(key, value)
		}
		return values.Encode(), nil
	}

	return "", nil
}

func (c *client) parseResponseBody(body []byte, statusCode int, out interface{}) error {
	if c.external {
		if statusCode < 200 || statusCode >= 300 {
			// all failures from external (3rd party) are Internal Server Error (500) for us
			return &ErrInvalidResponse{statusCode, 0, string(body)}
		}
		if out == nil {
			return nil
		}
		return json.Unmarshal(body, out)
	}

	resp := &response{}
	if err := json.Unmarshal(body, resp); err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return &ErrInvalidResponse{statusCode, resp.Error.Code, resp.Error.Message}
	}

	if out == nil {
		return nil
	}

	if resp.Data == nil {
		return json.Unmarshal(body, out)
	}

	outJSON, err := json.Marshal(resp.Data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(outJSON, out); err != nil {
		return err
	}

	return nil
}

func (c *client) requestURL(apiMethod string) string {
	hostname := strings.TrimSuffix(c.hostname, "/")
	apiMethod = strings.TrimPrefix(apiMethod, "/")
	params := []string{hostname, apiMethod}

	return strings.Join(params, "/")
}
