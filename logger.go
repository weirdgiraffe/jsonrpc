package jsonrpc

type Logger interface {
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type nopLogger struct{}

func (nopLogger) Warn(msg string, args ...any) {
}
func (nopLogger) Error(msg string, args ...any) {
}
