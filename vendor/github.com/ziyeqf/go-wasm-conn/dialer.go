package wasmconn

import (
	"net"

	"github.com/google/uuid"
	"github.com/magodo/go-wasmww"
	"github.com/magodo/go-webworkers/types"
)

type Dialer struct {
	connectStr string
	workerConn *wasmww.WasmWebWorkerConn
}

func NewWasmDialer(connectStr string, workerConn *wasmww.WasmWebWorkerConn) *Dialer {
	return &Dialer{
		connectStr,
		workerConn,
	}
}

func (d *Dialer) Dial() (net.Conn, error) {
	connId := uuid.New().String()

	connReceived := make(chan interface{}, 0)
	// it needs to listen before sending the request
	go func() {
		for event := range d.workerConn.EventChannel() {
			if data, err := event.Data(); err == nil {
				if resp, err := ParseWasmMsg[Response](data); err == nil {
					if resp.ConnId == connId {
						if err := d.workerConn.PostMessage(EncodeWasmMsg(Response{
							ConnId: connId,
						})); err != nil {
							panic(err)
						}
						connReceived <- struct{}{}
						return
					}
				}
			}
		}
	}()

	if err := d.workerConn.PostMessage(EncodeWasmMsg(Request{
		ConnectStr: d.connectStr,
		ConnId:     connId,
	})); err != nil {
		panic(err)
	}

	<-connReceived

	msgCh := make(chan Message, 0)
	conn := NewWasmConn(connId, d.workerConn.PostMessage, msgCh)
	startMsgChanProxy(msgCh, d.workerConn.EventChannel(), conn)
	return conn, nil
}

func startMsgChanProxy(msgCh chan Message, eventChan <-chan types.MessageEventMessage, conn *WasmConn) <-chan Message {
	go func() {
		for event := range eventChan {
			if data, err := event.Data(); err == nil {
				if msg, err := ParseWasmMsg[Message](data); err == nil {
					if msg.ConnId == conn.connId {
						msgCh <- *msg
						continue
					}
				}
				if c, err := ParseWasmMsg[Close](data); err == nil {
					if c.ConnId == conn.connId {
						conn.done = true
						close(msgCh)
						return
					}
				}
			}
		}
	}()
	return msgCh
}
