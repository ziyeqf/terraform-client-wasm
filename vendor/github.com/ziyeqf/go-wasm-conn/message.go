package wasmconn

type wasmConnRequest struct {
	ConnectStr string
	ConnId     string
}

type WasmConnResponse struct {
	ConnId string
}

type wasmConnClose struct {
	ConnId string
}

type WasmConnMessage struct {
	ConnId string
	Bytes  []byte
}
