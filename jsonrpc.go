package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func NewRequest(id uint64, method string, params any) *Request {
	var p json.RawMessage
	if params != nil {
		p, _ = json.Marshal(params)
	}
	return &Request{
		Version: "2.0",
		ID:      id,
		Method:  method,
		Params:  p,
	}
}

type ResponseOrNotification struct {
	*Response
	*Notification
}

func (rn *ResponseOrNotification) IsResponse() bool {
	return rn.Response != nil
}

func (rn *ResponseOrNotification) UnmarshalJSON(b []byte) error {
	var peek struct {
		ID *uint64 `json:"id"`
	}
	err := json.Unmarshal(b, &peek)
	if err != nil {
		return err
	}
	if peek.ID != nil {
		var r Response
		err = json.Unmarshal(b, &r)
		if err != nil {
			return err
		}
		rn.Response = &r
	} else {
		var n Notification
		err = json.Unmarshal(b, &n)
		if err != nil {
			return err
		}
		rn.Notification = &n
	}
	return nil
}

type Notification struct {
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func (n *Notification) UnmarshalJSON(b []byte) error {
	var raw struct {
		Version string           `json:"jsonrpc"`
		Method  string           `json:"method"`
		Params  *json.RawMessage `json:"params,omitempty"`
	}
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	if raw.Version != "2.0" {
		return fmt.Errorf("invalid jsonrpc version: %s", n.Version)
	}
	if raw.Method == "" {
		return errors.New("method field is not set")
	}
	n.Version = raw.Version
	n.Method = raw.Method
	if raw.Params != nil {
		n.Params = *raw.Params
	}
	return nil
}

type Request struct {
	Version string          `json:"jsonrpc"`
	ID      uint64          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	ID      uint64          `json:"id"`
	Version string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

func NewError(id uint64, e Error) *Response {
	return &Response{
		Version: "2.0",
		ID:      id,
		Error:   &e,
	}
}

func NewResult(id uint64, v any) *Response {
	var result json.RawMessage
	if v != nil {
		result, _ = json.Marshal(v)
	}
	return &Response{
		Version: "2.0",
		ID:      id,
		Result:  result,
	}
}

func (r *Response) UnmarshalJSON(b []byte) error {
	var raw struct {
		ID      *uint64          `json:"id"`
		Version string           `json:"jsonrpc"`
		Result  *json.RawMessage `json:"result,omitempty"`
		Error   *Error           `json:"error,omitempty"`
	}
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}
	if raw.Version != "2.0" {
		return fmt.Errorf("invalid jsonrpc version: %s", r.Version)
	}
	if raw.ID == nil {
		return errors.New("missing id field")
	}
	if raw.Result != nil && raw.Error != nil {
		return errors.New("both error and result fields are set")
	}
	r.ID = *raw.ID
	r.Version = raw.Version
	if raw.Result != nil {
		r.Result = *raw.Result
	}
	r.Error = raw.Error
	return nil
}

type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *Error) UnmarshalJSON(b []byte) error {
	var raw struct {
		Code    int              `json:"code"`
		Message string           `json:"message"`
		Data    *json.RawMessage `json:"data,omitempty"`
	}
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}
	e.Code = raw.Code
	e.Message = raw.Message
	if raw.Data != nil {
		e.Data = *raw.Data
	}
	return nil
}

func (e Error) Error() string {
	return fmt.Sprintf("code: %d message: %s", e.Code, e.Message)
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
