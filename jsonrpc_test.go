package jsonrpc

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

func TestParseResponse(t *testing.T) {
	tt := map[string]struct {
		In  []byte
		Out any
		Err bool
	}{
		"response with result": {
			In:  []byte(`{"jsonrpc":"2.0","id":1,"result":"foo"}`),
			Out: &Response{Version: "2.0", ID: 1, Result: ptr(json.RawMessage(`"foo"`))},
		},
		"response with result and null error": {
			In:  []byte(`{"jsonrpc":"2.0","id":1,"result":"foo","error":null}`),
			Out: &Response{Version: "2.0", ID: 1, Result: ptr(json.RawMessage(`"foo"`))},
		},
		"response with error": {
			In:  []byte(`{"jsonrpc":"2.0","id":1,"error":{"code":1,"message":"foo"}}`),
			Out: &Response{Version: "2.0", ID: 1, Error: &Error{Code: 1, Message: "foo"}},
		},
		"response with error and null result": {
			In:  []byte(`{"jsonrpc":"2.0","id":1,"result":null,"error":{"code":1,"message":"foo"}}`),
			Out: &Response{Version: "2.0", ID: 1, Error: &Error{Code: 1, Message: "foo"}},
		},
		"malformed json struct": {
			In:  []byte(`foo`),
			Err: true,
		},
		"malformed json values": {
			In:  []byte(`{"foo": "bar"}`),
			Err: true,
		},
	}
	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			var actual Response
			err := DecodeFrom(bytes.NewReader(tc.In), &actual)
			if tc.Err {
				require.Error(t, err)
				t.Log("[debug] got error:", err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.Out, &actual)
			}
		})
	}
}
