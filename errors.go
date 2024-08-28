package jsonrpc

var (
	ErrPraseRequest   = Error{Code: -32700, Message: "Parse error"}
	ErrInvalidRequest = Error{Code: -32600, Message: "Invalid Request"}
	ErrInvalidParams  = Error{Code: -32602, Message: "Invalid params"}
	ErrInternal       = Error{Code: -32603, Message: "Internal error"}
	ErrMethodNotFound = Error{Code: -32601, Message: "Method not found"}
)
