package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/coder/websocket"
)

type WebsocketClient struct {
	log  Logger
	conn *websocket.Conn

	notifications chan Notification

	mx      sync.Mutex
	waiting map[uint64]chan<- Response
	err     error
}

func ConnectWebsocket(ctx context.Context, wsURL string, logger ...Logger) (*WebsocketClient, error) {
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		return nil, err
	}

	var log Logger = nopLogger{}
	if len(logger) == 1 {
		log = logger[0]
	}

	ws := &WebsocketClient{
		conn:          conn,
		log:           log,
		notifications: make(chan Notification),
		waiting:       make(map[uint64]chan<- Response),
	}
	go ws.fanOutMessages(ctx)
	return ws, nil
}

func (ws *WebsocketClient) Close() error {
	return ws.conn.Close(websocket.StatusNormalClosure, "")
}

func (ws *WebsocketClient) Call(ctx context.Context, req *Request) (*Response, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal jsonrpc request: %w", err)
	}

	err = ws.conn.Write(ctx, websocket.MessageText, b)
	if err != nil {
		return nil, fmt.Errorf("failed to write request to websocket: %w", err)
	}

	l, err := ws.waitForResponse(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return l[0], nil
}

func (ws *WebsocketClient) BatchCall(ctx context.Context, batch []*Request) ([]*Response, error) {
	b, err := json.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal jsonrpc requests batch: %w", err)
	}

	err = ws.conn.Write(ctx, websocket.MessageText, b)
	if err != nil {
		return nil, fmt.Errorf("failed to write batch to websocket: %w", err)
	}

	id := make([]uint64, len(batch))
	for i := range batch {
		id[i] = batch[i].ID
	}

	return ws.waitForResponse(ctx, id...)
}

func (ws *WebsocketClient) waitForResponse(ctx context.Context, id ...uint64) ([]*Response, error) {
	done := make(chan Response, len(id))
	defer close(done)

	ws.mx.Lock()
	if ws.err != nil {
		ws.mx.Unlock()
		return nil, ws.err
	}

	for i := range id {
		reqID := id[i]
		ws.waiting[reqID] = done
	}
	ws.mx.Unlock()

	m := make(map[uint64]*Response, len(id))
	for len(m) < len(id) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case res := <-done:
			m[res.ID] = &res
		}
	}

	l := make([]*Response, len(id))
	for i := range id {
		reqID := id[i]
		l[i] = m[reqID]
	}
	return l, nil
}

func (ws *WebsocketClient) fanOutMessages(ctx context.Context) {
	for {
		typ, b, err := ws.conn.Read(ctx)
		if err != nil {
			ws.mx.Lock()
			ws.err = fmt.Errorf("failed to read message from websocket: %w", err)
			ws.mx.Unlock()
			return
		}

		if typ != websocket.MessageText {
			ws.log.Warn("received non-text message from websocket")
			continue
		}

		var l []ResponseOrNotification
		if len(b) > 0 && b[0] == '[' {
			err = json.Unmarshal(b, &l)
		} else {
			l = append(l, ResponseOrNotification{})
			err = json.Unmarshal(b, &l[0])
		}
		if err != nil {
			ws.log.Warn("failed to unmarshal jsonrpc message", "error", err)
			continue
		}

		for _, rn := range l {
			if rn.IsResponse() {
				err = ws.fanOutResponse(ctx, *rn.Response)
			} else {
				err = ws.fanOutNotification(ctx, *rn.Notification)
			}

			if err != nil {
				ws.mx.Lock()
				err = fmt.Errorf("failed to fan out jsonrpc message: %w", err)
				ws.mx.Unlock()
				return
			}
		}
	}
}

func (ws *WebsocketClient) fanOutNotification(ctx context.Context, n Notification) error {
	select {
	case ws.notifications <- n:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (ws *WebsocketClient) fanOutResponse(ctx context.Context, r Response) error {
	ws.mx.Lock()
	defer ws.mx.Unlock()

	ch, ok := ws.waiting[r.ID]
	if ok {
		select {
		case ch <- r:
			delete(ws.waiting, r.ID)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
