package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
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

func (c *ClientHTTP) Do(ctx context.Context, req *Request) (*Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode request")
	}
	r, err := http.NewRequest(
		http.MethodPost,
		c.baseURL,
		bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	r.Header.Set("content-type", "application/json")

	res, err := c.http.Do(r.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	jres, err := DecodeFrom(res.Body)
	if err != nil {
		return nil, err
	}
	return jres.(*Response), nil
}

func (c *ClientHTTP) DoBatch(ctx context.Context, req []*Request) ([]*Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode request")
	}
	r, err := http.NewRequest(
		http.MethodPost,
		c.baseURL,
		bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	r.Header.Set("content-type", "application/json")

	res, err := c.http.Do(r.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	jres, err := DecodeFrom(res.Body)
	if err != nil {
		return nil, err
	}
	l := jres.([]interface{})
	out := make([]*Response, len(l))
	for i := range l {
		out[i] = l[i].(*Response)
	}
	return out, nil
}
