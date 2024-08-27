package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ClientHTTP struct {
	baseURL string
	http    *http.Client
}

func NewClientHTTP(url string) *ClientHTTP {
	return &ClientHTTP{
		baseURL: url,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *ClientHTTP) Call(ctx context.Context, req *Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal jsonrpc request: %w", err)
	}

	httpResponse, err := c.do(ctx, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var res Response
	err = json.Unmarshal(httpResponse, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal jsonrpc response: %w", err)
	}

	return &res, nil
}

func (c *ClientHTTP) BatchCall(ctx context.Context, batch []Request) ([]Response, error) {
	body, err := json.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal jsonrpc requests batch: %w", err)
	}

	httpResponse, err := c.do(ctx, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var res []Response
	err = json.Unmarshal(httpResponse, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal jsonrpc response batch: %w", err)
	}

	return res, nil
}

func (c *ClientHTTP) do(ctx context.Context, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, c.baseURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to init http request: %w", err)
	}
	req.Header.Set("content-type", "application/json")

	res, err := c.http.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to do http request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http request failed with status code: %d", res.StatusCode)
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read http response body: %w", err)
	}
	return content, nil
}
