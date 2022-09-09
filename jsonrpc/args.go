package jsonrpc

type Response struct {
	Result string
	Error  error
	id     int
}

type SendTransactionArgs struct {
	From      string
	To        string
	Value     int
	Nonce     int
	Signature []byte
}
