package jsonrpc

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func handleEchoFoo(r io.Reader) ([]*Response, int) {
	rd := bufio.NewReader(r)

	peek, err := bufio.NewReader(rd).Peek(1)
	if err != nil {
		return nil, http.StatusInternalServerError
	}

	var req []Request
	if peek[0] == '[' {
		err := json.NewDecoder(rd).Decode(&req)
		if err != nil {
			return nil, http.StatusBadRequest
		}
	} else {
		req = append(req, Request{})
		err := json.NewDecoder(rd).Decode(&req[0])
		if err != nil {
			return nil, http.StatusBadRequest
		}
	}

	var l []*Response
	for i := range req {
		var res *Response
		if req[i].Method != "foo" {
			res = NewError(req[i].ID, ErrMethodNotFound)
		} else {
			res = NewResult(req[i].ID, req[i].Params)
		}
		l = append(l, res)
	}
	return l, http.StatusOK
}

func EchoFooParams(w http.ResponseWriter, r *http.Request) {
	l, status := handleEchoFoo(r.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if len(l) == 1 {
		json.NewEncoder(w).Encode(l[0])
	} else {
		json.NewEncoder(w).Encode(l)
	}
}

type Client interface {
	Close() error
	Call(ctx context.Context, req *Request) (*Response, error)
	BatchCall(ctx context.Context, batch []*Request) ([]*Response, error)
}

func TestHTTPClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(EchoFooParams))
	t.Cleanup(func() { srv.Close() })
	c := NewHTTPClient(srv.URL)
	testClient(t, c)
}

func testClient(t *testing.T, c Client) {
	t.Helper()
	ctx := context.Background()
	t.Run("call with result", func(t *testing.T) {
		req := NewRequest(1, "foo", "hello")
		res, err := c.Call(ctx, req)
		require.NoError(t, err)
		require.Equal(t, req.ID, res.ID)
		require.Nil(t, res.Error)
		require.Equal(t, req.Params, res.Result)
	})
	t.Run("call with error", func(t *testing.T) {
		req := NewRequest(1, "bar", "hello")
		res, err := c.Call(ctx, req)
		require.NoError(t, err)
		require.Equal(t, req.ID, res.ID)
		require.Equal(t, &ErrMethodNotFound, res.Error)
		require.Nil(t, res.Result)
	})
	t.Run("batch call with result", func(t *testing.T) {
		batch := []*Request{
			NewRequest(1, "foo", "hello"),
			NewRequest(2, "foo", "world"),
		}
		expected := []*Response{
			NewResult(1, "hello"),
			NewResult(2, "world"),
		}
		actual, err := c.BatchCall(ctx, batch)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	})
	t.Run("batch call with error", func(t *testing.T) {
		batch := []*Request{
			NewRequest(1, "bar", "hello"),
			NewRequest(2, "bar", "world"),
		}
		expected := []*Response{
			NewError(1, ErrMethodNotFound),
			NewError(2, ErrMethodNotFound),
		}
		actual, err := c.BatchCall(ctx, batch)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	})
	t.Run("batch call mixed", func(t *testing.T) {
		batch := []*Request{
			NewRequest(1, "foo", "hello"),
			NewRequest(2, "bar", "world"),
		}
		expected := []*Response{
			NewResult(1, "hello"),
			NewError(2, ErrMethodNotFound),
		}
		actual, err := c.BatchCall(ctx, batch)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	})
}
