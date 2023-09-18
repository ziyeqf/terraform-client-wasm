package wasmconn

// the message have to be different or they will be encoded to same wasm message

type Request struct {
	ConnectStr string `json:"request_connect_str"`
	ConnId     string `json:"request_conn_id"`
}

type Response struct {
	ConnId string `json:"response_conn_id"`
}

type Close struct {
	ConnId string `json:"close_conn_id"`
}

type Message struct {
	ConnId string `json:"conn_id"`
	Bytes  []byte `json:"bytes"`
}
