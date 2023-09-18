package wasmconn

import (
	"bytes"
	"encoding/gob"
	"log"
	"syscall/js"

	"github.com/hack-pad/safejs"
	"github.com/magodo/go-webworkers/types"
)

type jsMessageChan <-chan types.MessageEventMessage

type PostFunc func(data safejs.Value, transfers []safejs.Value) error

type MsgChan[T WasmMsg] chan T

type WasmMsg interface {
	wasmConnRequest | WasmConnResponse | wasmConnClose | WasmConnMessage
}

func EncodeWasmMsg[T WasmMsg](m T) (safejs.Value, []safejs.Value) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(m)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	jsArray := js.Global().Get("Uint8Array").New(buf.Len())
	js.CopyBytesToJS(jsArray, buf.Bytes())
	safeArray := safejs.Safe(jsArray)
	return safeArray, []safejs.Value{}
}

func ParseWasmMsg[T WasmMsg](jsMsg safejs.Value) (*T, error) {
	len, err := jsMsg.Length()
	if err != nil {
		return nil, err
	}
	buffer := make([]byte, len)
	if _, err := safejs.CopyBytesToGo(buffer, jsMsg); err != nil {
		return nil, err
	}
	var msg T

	dec := gob.NewDecoder(bytes.NewBuffer(buffer))
	err = dec.Decode(&msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}
