package jsonrpc

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/require"
)

func WebsocketEchoFooParams(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()

	for {
		_, b, err := conn.Read(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				conn.Close(websocket.StatusNormalClosure, "")
			} else {
				conn.Close(websocket.StatusInternalError, err.Error())
			}
			return
		}

		l, status := handleEchoFoo(bytes.NewReader(b))
		if status != http.StatusOK {
			conn.Close(websocket.StatusInternalError, err.Error())
			return
		}

		if len(l) == 1 {
			err = wsjson.Write(ctx, conn, l[0])
		} else {
			err = wsjson.Write(ctx, conn, l)
		}

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				conn.Close(websocket.StatusNormalClosure, "")
			} else {
				conn.Close(websocket.StatusInternalError, err.Error())
			}
			return
		}
	}
}

func TestWebsocketClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(WebsocketEchoFooParams))
	t.Cleanup(func() { srv.Close() })
	c, err := ConnectWebsocket(context.Background(), srv.URL, slog.Default())
	require.NoError(t, err)
	testClient(t, c)
}
