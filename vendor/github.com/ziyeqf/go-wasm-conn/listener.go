package wasmconn

import (
	"net"

	"github.com/magodo/go-wasmww"
)

type Listener struct {
	connectStr      string
	postMessageFunc PostFunc
	eventChan       jsMessageChan
	cancelFunc      wasmww.WebWorkerCloseFunc
	conns           map[string]*WasmConn
	connChs         map[string]chan WasmConnMessage
}

func NewListener(connectStr string, postMessageFunc PostFunc, eventChan jsMessageChan, cancelFunc wasmww.WebWorkerCloseFunc) *Listener {
	return &Listener{
		connectStr,
		postMessageFunc,
		eventChan,
		cancelFunc,
		make(map[string]*WasmConn, 0),
		make(map[string]chan WasmConnMessage, 0),
	}
}

func (w *Listener) Accept() (net.Conn, error) {
	for event := range w.eventChan {
		if data, err := event.Data(); err == nil {
			if req, err := ParseWasmMsg[wasmConnRequest](data); err == nil {
				if req.ConnectStr == w.connectStr {
					connEventCh := make(chan WasmConnMessage, 0)
					conn := NewWasmConn(req.ConnId, w.postMessageFunc, connEventCh)
					w.conns[req.ConnId] = conn
					w.connChs[req.ConnId] = connEventCh

					if err := w.postMessageFunc(EncodeWasmMsg(WasmConnResponse{
						ConnId: req.ConnId,
					})); err != nil {
						return nil, err
					}
					return conn, nil
				}
			}
			if msg, err := ParseWasmMsg[WasmConnMessage](data); err == nil {
				if connEventCh, ok := w.connChs[msg.ConnId]; ok {
					connEventCh <- *msg
				}
			}
			if c, err := ParseWasmMsg[wasmConnClose](data); err == nil {
				if conn, ok := w.conns[c.ConnId]; ok {
					conn.done = true
					delete(w.conns, c.ConnId)
					close(w.connChs[c.ConnId])
					delete(w.connChs, c.ConnId)
				}
			}

		}
	}
	return nil, nil
}

func (w *Listener) Close() error {
	return w.cancelFunc()
}

func (w *Listener) Addr() net.Addr {
	return NewWasmAddr(w.connectStr)
}
