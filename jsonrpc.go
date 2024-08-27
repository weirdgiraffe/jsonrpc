package jsonrpc

import (
	"encoding/json"
	"fmt"
	"io"
)

type Request struct {
	Version string          `json:"jsonrpc"`
	ID      uint64          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func NewRequest(id uint64, method string, params ...interface{}) *Request {
	var p json.RawMessage
	if len(params) > 0 {
		p, _ = json.Marshal(params)
	}
	return &Request{
		Version: "2.0",
		ID:      id,
		Method:  method,
		Params:  p,
	}
}

type Response struct {
	Version string           `json:"jsonrpc"`
	ID      uint64           `json:"id"`
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *Error           `json:"error,omitempty"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e Error) Error() string {
	return fmt.Sprintf("code: %d message: %s", e.Code, e.Message)
}

type Notification struct {
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonrpcSingleMessage interface {
	Request | Response | Notification
}

type jsonrpcBatchMesssage interface {
	[]Request | []Response | []Notification
}

type jsonrpcMessage interface {
	jsonrpcSingleMessage | jsonrpcBatchMesssage
}

func EncodeTo[T jsonrpcMessage](w io.Writer, msg T) error {
	return json.NewEncoder(w).Encode(msg)
}

func DecodeFrom[T jsonrpcMessage](r io.Reader, dst *T) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
