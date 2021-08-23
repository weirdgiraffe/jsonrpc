package jsonrpc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
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

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e Error) Error() string {
	return fmt.Sprintf("code: %d message: %s", e.Code, e.Message)
}

type Response struct {
	Version string          `json:"jsonrpc"`
	ID      uint64          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

type Notification struct {
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func EncodeTo(w io.Writer, msg interface{}) error {
	switch m := msg.(type) {
	case *Request, []*Request, Request, []Request, *Response, []*Response, *Notification, Notification:
		return json.NewEncoder(w).Encode(m)
	default:
		return errors.New("unexpected msg type")
	}
}

func DecodeFrom(r io.Reader) (interface{}, error) {
	br := bufio.NewReader(r)

	sample, err := br.Peek(10)
	if err == io.EOF {
		fmt.Println(sample)
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	for c, _ := range sample {
		if c == '[' {
			return decodeBatch(br)
		}
		return decodeObject(br)
	}

	return nil, errors.New("jsonrpc: malformed json")
}

type object struct {
	Version string          `json:"jsonrpc"`
	ID      *uint64         `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

func decodeBatch(r io.Reader) (interface{}, error) {
	var l []object
	err := json.NewDecoder(r).Decode(&l)
	if err != nil {
		return nil, err
	}
	out := make([]interface{}, len(l))
	for i := range l {
		out[i], err = formatObject(l[i])
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func decodeObject(r io.Reader) (interface{}, error) {
	var o object
	if err := json.NewDecoder(r).Decode(&o); err != nil {
		return nil, err
	}
	return formatObject(o)
}

func formatObject(o object) (interface{}, error) {
	switch {
	case o.ID == nil && o.Method != "":
		return &Notification{Version: o.Version, Method: o.Method, Params: o.Params}, nil
	case o.ID != nil && o.Method != "":
		return &Request{Version: o.Version, ID: *o.ID, Method: o.Method, Params: o.Params}, nil
	case o.ID != nil && (o.Result != nil || o.Error != nil):
		return &Response{Version: o.Version, ID: *o.ID, Result: o.Result, Error: o.Error}, nil
	default:
		return nil, errors.New("not a jsonrpc message")
	}
}
