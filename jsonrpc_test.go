package jsonrpc

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalResponse(t *testing.T) {
	tt := map[string]struct {
		In  []byte
		Out any
		Err bool
	}{
		"without version": {
			In:  []byte(`{"id":1,"result":"foo"}`),
			Err: true,
		},
		"without null version": {
			In:  []byte(`{"version":null, "id":1,"result":"foo"}`),
			Err: true,
		},
		"with result": {
			In:  []byte(`{"jsonrpc":"2.0","id":1,"result":"foo"}`),
			Out: &Response{Version: "2.0", ID: 1, Result: json.RawMessage(`"foo"`)},
		},
		"with error": {
			In:  []byte(`{"jsonrpc":"2.0","id":1,"error":{"code":1,"message":"foo"}}`),
			Out: &Response{Version: "2.0", ID: 1, Error: &Error{Code: 1, Message: "foo"}},
		},
		"response with result and error": {
			In:  []byte(`{"jsonrpc":"2.0","id":1,"result":"foo","error":{"code":1,"message":"foo"}}`),
			Err: true,
		},
		"without id": {
			In:  []byte(`{"jsonrpc":"2.0","result":"foo"}`),
			Err: true,
		},
		"without null id": {
			In:  []byte(`{"jsonrpc":"2.0","id":null,"result":"foo"}`),
			Err: true,
		},
		"malformed json": {
			In:  []byte(`"foo"`),
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

func TestUnmarshalNotification(t *testing.T) {
	tt := map[string]struct {
		In  []byte
		Out any
		Err bool
	}{
		"complete": {
			In:  []byte(`{"jsonrpc":"2.0","method":"foo","params":"bar"}`),
			Out: &Notification{Version: "2.0", Method: "foo", Params: json.RawMessage(`"bar"`)},
		},
		"with null version": {
			In:  []byte(`{"jsonrpc":null,"method":"foo", "params":"bar"}`),
			Err: true,
		},
		"without version": {
			In:  []byte(`{"method":"foo", "params":"bar"}`),
			Err: true,
		},
		"with null method": {
			In:  []byte(`{"jsonrpc":"2.0","method":null,"params":"bar"}`),
			Err: true,
		},
		"without method": {
			In:  []byte(`{"jsonrpc":"2.0","params":"bar"}`),
			Err: true,
		},
		"with null params": {
			In:  []byte(`{"jsonrpc":"2.0", "method":"foo","params":null}`),
			Out: &Notification{Version: "2.0", Method: "foo"},
		},
		"without params": {
			In:  []byte(`{"jsonrpc":"2.0", "method":"foo"}`),
			Out: &Notification{Version: "2.0", Method: "foo"},
		},
		"malformed json": {
			In:  []byte(`"foo"`),
			Err: true,
		},
	}
	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			var actual Notification
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
