package jsonrpc

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseResponse(t *testing.T) {
	tt := map[string]struct {
		In  []byte
		Out interface{}
		Err bool
	}{
		"response with result": {
			In:  []byte(`{"jsonrpc":"2.0","id":1,"result":"foo"}`),
			Out: &Response{Version: "2.0", ID: 1, Result: json.RawMessage(`"foo"`)},
		},
		"response with result and null error": {
			In:  []byte(`{"jsonrpc":"2.0","id":1,"result":"foo","error":null}`),
			Out: &Response{Version: "2.0", ID: 1, Result: json.RawMessage(`"foo"`)},
		},
		"response with error": {
			In:  []byte(`{"jsonrpc":"2.0","id":1,"error":{"code":1,"message":"foo"}}`),
			Out: &Response{Version: "2.0", ID: 1, Error: &Error{Code: 1, Message: "foo"}},
		},
		"response with error and null result": {
			In:  []byte(`{"jsonrpc":"2.0","id":1,"resutl":null,"error":{"code":1,"message":"foo"}}`),
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
			out, err := DecodeFrom(bytes.NewReader(tc.In))
			if tc.Err {
				require.Error(t, err)
				require.Nil(t, out)
				t.Log(err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.Out, out)
		})
	}
}
